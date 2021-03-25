// Copyright 2016 by the Authors
// This file is part of the go-core library.
//
// The go-core library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-core library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-core library. If not, see <http://www.gnu.org/licenses/>.

package params

import (
	"os/exec"
	"log"
	"regexp"
)

// Version holds the textual version string.
var Version = func() string {
	out, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		log.Fatal(err)
	}
	reg := regexp.MustCompile(`\b?[0-9]+\.[0-9]+\.[0-9]+?\b`)
	ver := reg.FindString(string(out))
	return ver
}()

// ArchiveVersion holds the textual version string used for Gocore archives.
// e.g. "1.8.11-dea1ce05" for stable releases, or
//      "1.8.13-unstable-21c059b6" for unstable releases
func ArchiveVersion(gitCommit string) string {
	vsn := Version
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	return vsn
}

func VersionWithCommit(gitTag, gitBranch, gitCommit, gitDate string) string {
	version := ""
	if gitTag != "" {
		reg := regexp.MustCompile(`\b?[0-9]+\.[0-9]+\.[0-9]+?\b`)
		version += reg.FindString(gitTag)
	} else if gitBranch != "" {
		version += "-" + gitBranch
	}
	if len(gitCommit) >= 8 {
		version += "-" + gitCommit[:8]
	}
	if gitDate != "" {
		version += "-" + gitDate
	}
	return version
}
