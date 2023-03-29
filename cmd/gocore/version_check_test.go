// Copyright 2020 by the Authors
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func versionUint(v string) int {
	mustInt := func(s string) int {
		a, err := strconv.Atoi(s)
		if err != nil {
			panic(v)
		}
		return a
	}
	components := strings.Split(strings.TrimPrefix(v, "v"), ".")
	a := mustInt(components[0])
	b := mustInt(components[1])
	c := mustInt(components[2])
	return a*100*100 + b*100 + c
}

// TestMatching can be used to check that the regexps are correct
func TestMatching(t *testing.T) {
	data, _ := ioutil.ReadFile("./testdata/vcheck/vulnerabilities.json")
	var vulns []vulnJson
	if err := json.Unmarshal(data, &vulns); err != nil {
		t.Fatal(err)
	}
	check := func(version string) {
		vFull := fmt.Sprintf("Gocore/%v-unstable-15339cf1-20201204/linux-amd64/go1.15.4", version)
		for _, vuln := range vulns {
			r, err := regexp.Compile(vuln.Check)
			vulnIntro := versionUint(vuln.Introduced)
			vulnFixed := versionUint(vuln.Fixed)
			current := versionUint(version)
			if err != nil {
				t.Fatal(err)
			}
			if vuln.Name == "Denial of service due to Go CVE-2020-28362" {
				// this one is not tied to gocore-versions
				continue
			}
			if vulnIntro <= current && vulnFixed > current {
				// Should be vulnerable
				if !r.MatchString(vFull) {
					t.Errorf("Should be vulnerable, version %v, intro: %v, fixed: %v %v %v",
						version, vuln.Introduced, vuln.Fixed, vuln.Name, vuln.Check)
				}
			} else {
				if r.MatchString(vFull) {
					t.Errorf("Should not be flagged vulnerable, version %v, intro: %v, fixed: %v %v %d %d %d",
						version, vuln.Introduced, vuln.Fixed, vuln.Name, vulnIntro, current, vulnFixed)
				}
			}

		}
	}
	for major := 1; major < 2; major++ {
		for minor := 0; minor < 30; minor++ {
			for patch := 0; patch < 30; patch++ {
				vShort := fmt.Sprintf("v%d.%d.%d", major, minor, patch)
				check(vShort)
			}
		}
	}
}
