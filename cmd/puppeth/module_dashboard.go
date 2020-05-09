// Copyright 2017 The go-core Authors
// This file is part of go-core.
//
// go-core is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-core is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-core. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/core-coin/go-core/log"
)

// dashboardDockerfile is the Dockerfile required to build an dashboard container
// to aggregate various private network services under one easily accessible page.
var dashboardDockerfile = `
FROM mhart/alpine-node:latest

RUN \
	npm install connect serve-static && \
	\
	echo 'var connect = require("connect");'                                > server.js && \
	echo 'var serveStatic = require("serve-static");'                      >> server.js && \
	echo 'connect().use(serveStatic("/dashboard")).listen(80, function(){' >> server.js && \
	echo '    console.log("Server running on 80...");'                     >> server.js && \
	echo '});'                                                             >> server.js

ADD {{.Network}}.json /dashboard/{{.Network}}.json
ADD {{.Network}}-cpp.json /dashboard/{{.Network}}-cpp.json
ADD {{.Network}}-harmony.json /dashboard/{{.Network}}-harmony.json
ADD {{.Network}}-parity.json /dashboard/{{.Network}}-parity.json
ADD {{.Network}}-python.json /dashboard/{{.Network}}-python.json
ADD index.html /dashboard/index.html
ADD puppeth.png /dashboard/puppeth.png

EXPOSE 80

CMD ["node", "/server.js"]
`

// dashboardComposefile is the docker-compose.yml file required to deploy and
// maintain an service aggregating dashboard.
var dashboardComposefile = `
version: '2'
services:
  dashboard:
    build: .
    image: {{.Network}}/dashboard{{if not .VHost}}
    ports:
      - "{{.Port}}:80"{{end}}
    environment:
      - XCESTATS_PAGE={{.CorestatsPage}}
      - EXPLORER_PAGE={{.ExplorerPage}}
      - WALLET_PAGE={{.WalletPage}}
      - FAUCET_PAGE={{.FaucetPage}}{{if .VHost}}
      - VIRTUAL_HOST={{.VHost}}{{end}}
    logging:
      driver: "json-file"
      options:
        max-size: "1m"
        max-file: "10"
    restart: always
`

// deployDashboard deploys a new dashboard container to a remote machine via SSH,
// docker and docker-compose. If an instance with the specified network name
// already exists there, it will be overwritten!
func deployDashboard(client *sshClient, network string, conf *config, config *dashboardInfos, nocache bool) ([]byte, error) {
	// Generate the content to upload to the server
	workdir := fmt.Sprintf("%d", rand.Int63())
	files := make(map[string][]byte)

	dockerfile := new(bytes.Buffer)
	template.Must(template.New("").Parse(dashboardDockerfile)).Execute(dockerfile, map[string]interface{}{
		"Network": network,
	})
	files[filepath.Join(workdir, "Dockerfile")] = dockerfile.Bytes()

	composefile := new(bytes.Buffer)
	template.Must(template.New("").Parse(dashboardComposefile)).Execute(composefile, map[string]interface{}{
		"Network":      network,
		"Port":         config.port,
		"VHost":        config.host,
		"CorestatsPage": config.xcestats,
		"ExplorerPage": config.explorer,
		"WalletPage":   config.wallet,
		"FaucetPage":   config.faucet,
	})
	files[filepath.Join(workdir, "docker-compose.yaml")] = composefile.Bytes()

	statsLogin := fmt.Sprintf("yournode:%s", conf.xcestats)
	if !config.trusted {
		statsLogin = ""
	}
	indexfile := new(bytes.Buffer)
	bootCpp := make([]string, len(conf.bootnodes))
	for i, boot := range conf.bootnodes {
		bootCpp[i] = "required:" + strings.TrimPrefix(boot, "enode://")
	}
	bootHarmony := make([]string, len(conf.bootnodes))
	for i, boot := range conf.bootnodes {
		bootHarmony[i] = fmt.Sprintf("-Dpeer.active.%d.url=%s", i, boot)
	}
	bootPython := make([]string, len(conf.bootnodes))
	for i, boot := range conf.bootnodes {
		bootPython[i] = "'" + boot + "'"
	}
	template.Must(template.New("").ParseFiles("template_dashboard.html")).Execute(indexfile, map[string]interface{}{
		"Network":           network,
		"NetworkID":         conf.Genesis.Config.ChainID,
		"NetworkTitle":      strings.Title(network),
		"CorestatsPage":      config.xcestats,
		"ExplorerPage":      config.explorer,
		"WalletPage":        config.wallet,
		"FaucetPage":        config.faucet,
		"GcoreGenesis":       network + ".json",
		"Bootnodes":         conf.bootnodes,
		"BootnodesFlat":     strings.Join(conf.bootnodes, ","),
		"Xcestats":          statsLogin,
		"Ethash":            conf.Genesis.Config.Ethash != nil,
		"CppGenesis":        network + "-cpp.json",
		"CppBootnodes":      strings.Join(bootCpp, " "),
		"HarmonyGenesis":    network + "-harmony.json",
		"HarmonyBootnodes":  strings.Join(bootHarmony, " "),
		"ParityGenesis":     network + "-parity.json",
		"PythonGenesis":     network + "-python.json",
		"PythonBootnodes":   strings.Join(bootPython, ","),
		"Homestead":         conf.Genesis.Config.HomesteadBlock,
		"Tangerine":         conf.Genesis.Config.EIP150Block,
		"Spurious":          conf.Genesis.Config.EIP155Block,
		"Byzantium":         conf.Genesis.Config.ByzantiumBlock,
		"Constantinople":    conf.Genesis.Config.ConstantinopleBlock,
		"ConstantinopleFix": conf.Genesis.Config.PetersburgBlock,
	})
	files[filepath.Join(workdir, "index.html")] = indexfile.Bytes()

	// Marshal the genesis spec files for go-core and all the other clients
	genesis, _ := conf.Genesis.MarshalJSON()
	files[filepath.Join(workdir, network+".json")] = genesis

	if conf.Genesis.Config.Ethash != nil {
		cppSpec, err := newAlethGenesisSpec(network, conf.Genesis)
		if err != nil {
			return nil, err
		}
		cppSpecJSON, _ := json.Marshal(cppSpec)
		files[filepath.Join(workdir, network+"-cpp.json")] = cppSpecJSON

		harmonySpecJSON, _ := conf.Genesis.MarshalJSON()
		files[filepath.Join(workdir, network+"-harmony.json")] = harmonySpecJSON

		paritySpec, err := newParityChainSpec(network, conf.Genesis, conf.bootnodes)
		if err != nil {
			return nil, err
		}
		paritySpecJSON, _ := json.Marshal(paritySpec)
		files[filepath.Join(workdir, network+"-parity.json")] = paritySpecJSON

		pyethSpec, err := newPyCoreGenesisSpec(network, conf.Genesis)
		if err != nil {
			return nil, err
		}
		pyethSpecJSON, _ := json.Marshal(pyethSpec)
		files[filepath.Join(workdir, network+"-python.json")] = pyethSpecJSON
	} else {
		for _, client := range []string{"cpp", "harmony", "parity", "python"} {
			files[filepath.Join(workdir, network+"-"+client+".json")] = []byte{}
		}
	}

	// Upload the deployment files to the remote server (and clean up afterwards)
	if out, err := client.Upload(files); err != nil {
		return out, err
	}
	defer client.Run("rm -rf " + workdir)

	// Build and deploy the dashboard service
	if nocache {
		return nil, client.Stream(fmt.Sprintf("cd %s && docker-compose -p %s build --pull --no-cache && docker-compose -p %s up -d --force-recreate --timeout 60", workdir, network, network))
	}
	return nil, client.Stream(fmt.Sprintf("cd %s && docker-compose -p %s up -d --build --force-recreate --timeout 60", workdir, network))
}

// dashboardInfos is returned from a dashboard status check to allow reporting
// various configuration parameters.
type dashboardInfos struct {
	host    string
	port    int
	trusted bool

	xcestats string
	explorer string
	wallet   string
	faucet   string
}

// Report converts the typed struct into a plain string->string map, containing
// most - but not all - fields for reporting to the user.
func (info *dashboardInfos) Report() map[string]string {
	return map[string]string{
		"Website address":       info.host,
		"Website listener port": strconv.Itoa(info.port),
		"Xcestats service":      info.xcestats,
		"Explorer service":      info.explorer,
		"Wallet service":        info.wallet,
		"Faucet service":        info.faucet,
	}
}

// checkDashboard does a health-check against a dashboard container to verify if
// it's running, and if yes, gathering a collection of useful infos about it.
func checkDashboard(client *sshClient, network string) (*dashboardInfos, error) {
	// Inspect a possible xcestats container on the host
	infos, err := inspectContainer(client, fmt.Sprintf("%s_dashboard_1", network))
	if err != nil {
		return nil, err
	}
	if !infos.running {
		return nil, ErrServiceOffline
	}
	// Resolve the port from the host, or the reverse proxy
	port := infos.portmap["80/tcp"]
	if port == 0 {
		if proxy, _ := checkNginx(client, network); proxy != nil {
			port = proxy.port
		}
	}
	if port == 0 {
		return nil, ErrNotExposed
	}
	// Resolve the host from the reverse-proxy and configure the connection string
	host := infos.envvars["VIRTUAL_HOST"]
	if host == "" {
		host = client.server
	}
	// Run a sanity check to see if the port is reachable
	if err = checkPort(host, port); err != nil {
		log.Warn("Dashboard service seems unreachable", "server", host, "port", port, "err", err)
	}
	// Container available, assemble and return the useful infos
	return &dashboardInfos{
		host:     host,
		port:     port,
		xcestats: infos.envvars["XCESTATS_PAGE"],
		explorer: infos.envvars["EXPLORER_PAGE"],
		wallet:   infos.envvars["WALLET_PAGE"],
		faucet:   infos.envvars["FAUCET_PAGE"],
	}, nil
}
