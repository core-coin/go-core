// Copyright 2020 The CORE FOUNDATION, nadacia
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
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/log"
)

// nodeDockerfile is the Dockerfile required to run an Core node.
var nodeDockerfile = `
FROM docker.pkg.github.com/core-coin/go-core/gocore:latest

ADD genesis.json /genesis.json
{{if .Unlock}}
	ADD signer.json /signer.json
	ADD signer.pass /signer.pass
{{end}}
RUN \
  echo 'gocore --cache 512 init /genesis.json' > gocore.sh && \{{if .Unlock}}
	echo 'mkdir -p /root/core/keystore/ && cp /signer.json /root/core/keystore/' >> gocore.sh && \{{end}}
	echo $'exec gocore --networkid {{.NetworkID}} --cache 512 --port {{.Port}} --nat extip:{{.IP}} --maxpeers {{.Peers}} {{.LightFlag}} --xcbstats \'{{.Xcbstats}}\' {{if .Bootnodes}}--bootnodes {{.Bootnodes}}{{end}} {{if .Corebase}}--miner.corebase {{.Corebase}} --mine --miner.threads 1{{end}} {{if .Unlock}}--unlock 0 --password /signer.pass --mine{{end}} --miner.energytarget {{.EnergyTarget}} --miner.energylimit {{.EnergyLimit}} --miner.energyprice {{.EnergyPrice}}' >> gocore.sh

ENTRYPOINT ["/bin/sh", "gocore.sh"]
`

// nodeComposefile is the docker-compose.yml file required to deploy and maintain
// an Core node (bootnode or miner for now).
var nodeComposefile = `
version: '2'
services:
  {{.Type}}:
    build: .
    image: {{.Network}}/{{.Type}}
    container_name: {{.Network}}_{{.Type}}_1
    ports:
      - "{{.Port}}:{{.Port}}"
      - "{{.Port}}:{{.Port}}/udp"
    volumes:
      - {{.Datadir}}:/root/.core{{if .Cryptoredir}}
      - {{.Cryptoredir}}:/root/.cryptore{{end}}
    environment:
      - PORT={{.Port}}/tcp
      - TOTAL_PEERS={{.TotalPeers}}
      - LIGHT_PEERS={{.LightPeers}}
      - STATS_NAME={{.Xcbstats}}
      - MINER_NAME={{.Corebase}}
      - ENERGY_TARGET={{.EnergyTarget}}
      - ENERGY_LIMIT={{.EnergyLimit}}
      - ENERGY_PRICE={{.EnergyPrice}}
    logging:
      driver: "json-file"
      options:
        max-size: "1m"
        max-file: "10"
    restart: always
`

// deployNode deploys a new Core node container to a remote machine via SSH,
// docker and docker-compose. If an instance with the specified network name
// already exists there, it will be overwritten!
func deployNode(client *sshClient, network string, bootnodes []string, config *nodeInfos, nocache bool) ([]byte, error) {
	kind := "sealnode"
	if config.keyJSON == "" && config.corebase == "" {
		kind = "bootnode"
		bootnodes = make([]string, 0)
	}
	// Generate the content to upload to the server
	workdir := fmt.Sprintf("%d", rand.Int63())
	files := make(map[string][]byte)

	lightFlag := ""
	if config.peersLight > 0 {
		lightFlag = fmt.Sprintf("--lightpeers=%d --lightserv=50", config.peersLight)
	}
	dockerfile := new(bytes.Buffer)
	template.Must(template.New("").Parse(nodeDockerfile)).Execute(dockerfile, map[string]interface{}{
		"NetworkID":    config.network,
		"Port":         config.port,
		"IP":           client.address,
		"Peers":        config.peersTotal,
		"LightFlag":    lightFlag,
		"Bootnodes":    strings.Join(bootnodes, ","),
		"Xcbstats":     config.xcbstats,
		"Corebase":     config.corebase,
		"EnergyTarget": uint64(1000000 * config.energyTarget),
		"EnergyLimit":  uint64(1000000 * config.energyLimit),
		"EnergyPrice":  uint64(1000000000 * config.energyPrice),
		"Unlock":       config.keyJSON != "",
	})
	files[filepath.Join(workdir, "Dockerfile")] = dockerfile.Bytes()

	composefile := new(bytes.Buffer)
	template.Must(template.New("").Parse(nodeComposefile)).Execute(composefile, map[string]interface{}{
		"Type":         kind,
		"Datadir":      config.datadir,
		"Cryptoredir":  config.cryptoredir,
		"Network":      network,
		"Port":         config.port,
		"TotalPeers":   config.peersTotal,
		"Light":        config.peersLight > 0,
		"LightPeers":   config.peersLight,
		"Xcbstats":     config.xcbstats[:strings.Index(config.xcbstats, ":")],
		"Corebase":     config.corebase,
		"EnergyTarget": config.energyTarget,
		"EnergyLimit":  config.energyLimit,
		"EnergyPrice":  config.energyPrice,
	})
	files[filepath.Join(workdir, "docker-compose.yaml")] = composefile.Bytes()

	files[filepath.Join(workdir, "genesis.json")] = config.genesis
	if config.keyJSON != "" {
		files[filepath.Join(workdir, "signer.json")] = []byte(config.keyJSON)
		files[filepath.Join(workdir, "signer.pass")] = []byte(config.keyPass)
	}
	// Upload the deployment files to the remote server (and clean up afterwards)
	if out, err := client.Upload(files); err != nil {
		return out, err
	}
	defer client.Run("rm -rf " + workdir)

	// Build and deploy the boot or seal node service
	if nocache {
		return nil, client.Stream(fmt.Sprintf("cd %s && docker-compose -p %s build --pull --no-cache && docker-compose -p %s up -d --force-recreate --timeout 60", workdir, network, network))
	}
	return nil, client.Stream(fmt.Sprintf("cd %s && docker-compose -p %s up -d --build --force-recreate --timeout 60", workdir, network))
}

// nodeInfos is returned from a boot or seal node status check to allow reporting
// various configuration parameters.
type nodeInfos struct {
	genesis      []byte
	network      int64
	datadir      string
	cryptoredir  string
	xcbstats     string
	port         int
	enode        string
	peersTotal   int
	peersLight   int
	corebase     string
	keyJSON      string
	keyPass      string
	energyTarget float64
	energyLimit  float64
	energyPrice  float64
}

// Report converts the typed struct into a plain string->string map, containing
// most - but not all - fields for reporting to the user.
func (info *nodeInfos) Report() map[string]string {
	report := map[string]string{
		"Data directory":           info.datadir,
		"Listener port":            strconv.Itoa(info.port),
		"Peer count (all total)":   strconv.Itoa(info.peersTotal),
		"Peer count (light nodes)": strconv.Itoa(info.peersLight),
		"Xcbstats username":        info.xcbstats,
	}
	if info.energyTarget > 0 {
		// Miner or signer node
		report["Energy price (minimum accepted)"] = fmt.Sprintf("%0.3f Nucle", info.energyPrice)
		report["Energy floor (baseline target)"] = fmt.Sprintf("%0.3f MEnergy", info.energyTarget)
		report["Energy ceil  (target maximum)"] = fmt.Sprintf("%0.3f MEnergy", info.energyLimit)

		if info.corebase != "" {
			// Cryptore proof-of-work miner
			report["Cryptore directory"] = info.cryptoredir
			report["Miner account"] = info.corebase
		}
		if info.keyJSON != "" {
			// Clique proof-of-authority signer
			var key struct {
				Address string `json:"address"`
			}
			if err := json.Unmarshal([]byte(info.keyJSON), &key); err == nil {
				addr, err := common.HexToAddress(key.Address)
				if err != nil {
					log.Error("Failed to retrieve signer address: invalid address", "err", err)
					return report
				}
				report["Signer account"] = addr.Hex()
			} else {
				log.Error("Failed to retrieve signer address", "err", err)
			}
		}
	}
	return report
}

// checkNode does a health-check against a boot or seal node server to verify
// whether it's running, and if yes, whether it's responsive.
func checkNode(client *sshClient, network string, boot bool) (*nodeInfos, error) {
	kind := "bootnode"
	if !boot {
		kind = "sealnode"
	}
	// Inspect a possible bootnode container on the host
	infos, err := inspectContainer(client, fmt.Sprintf("%s_%s_1", network, kind))
	if err != nil {
		return nil, err
	}
	if !infos.running {
		return nil, ErrServiceOffline
	}
	// Resolve a few types from the environmental variables
	totalPeers, _ := strconv.Atoi(infos.envvars["TOTAL_PEERS"])
	lightPeers, _ := strconv.Atoi(infos.envvars["LIGHT_PEERS"])
	energyTarget, _ := strconv.ParseFloat(infos.envvars["ENERGY_TARGET"], 64)
	energyLimit, _ := strconv.ParseFloat(infos.envvars["ENERGY_LIMIT"], 64)
	energyPrice, _ := strconv.ParseFloat(infos.envvars["ENERGY_PRICE"], 64)

	// Container available, retrieve its node ID and its genesis json
	var out []byte
	if out, err = client.Run(fmt.Sprintf("docker exec %s_%s_1 gocore --exec admin.nodeInfo.enode --cache=16 attach", network, kind)); err != nil {
		return nil, ErrServiceUnreachable
	}
	enode := bytes.Trim(bytes.TrimSpace(out), "\"")

	if out, err = client.Run(fmt.Sprintf("docker exec %s_%s_1 cat /genesis.json", network, kind)); err != nil {
		return nil, ErrServiceUnreachable
	}
	genesis := bytes.TrimSpace(out)

	keyJSON, keyPass := "", ""
	if out, err = client.Run(fmt.Sprintf("docker exec %s_%s_1 cat /signer.json", network, kind)); err == nil {
		keyJSON = string(bytes.TrimSpace(out))
	}
	if out, err = client.Run(fmt.Sprintf("docker exec %s_%s_1 cat /signer.pass", network, kind)); err == nil {
		keyPass = string(bytes.TrimSpace(out))
	}
	// Run a sanity check to see if the devp2p is reachable
	port := infos.portmap[infos.envvars["PORT"]]
	if err = checkPort(client.server, port); err != nil {
		log.Warn(fmt.Sprintf("%s devp2p port seems unreachable", strings.Title(kind)), "server", client.server, "port", port, "err", err)
	}
	// Assemble and return the useful infos
	stats := &nodeInfos{
		genesis:      genesis,
		datadir:      infos.volumes["/root/.core"],
		cryptoredir:  infos.volumes["/root/.cryptore"],
		port:         port,
		peersTotal:   totalPeers,
		peersLight:   lightPeers,
		xcbstats:     infos.envvars["STATS_NAME"],
		corebase:     infos.envvars["MINER_NAME"],
		keyJSON:      keyJSON,
		keyPass:      keyPass,
		energyTarget: energyTarget,
		energyLimit:  energyLimit,
		energyPrice:  energyPrice,
	}
	stats.enode = string(enode)

	return stats, nil
}
