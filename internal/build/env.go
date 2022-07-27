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

package build

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// These flags override values in build env.
	GitCommitFlag   = flag.String("git-commit", "", `Overrides git commit hash embedded into executables`)
	GitBranchFlag   = flag.String("git-branch", "", `Overrides git branch being built`)
	GitTagFlag      = flag.String("git-tag", "", `Overrides git tag being built`)
	BuildnumFlag    = flag.String("buildnum", "", `Overrides CI build number`)
	PullRequestFlag = flag.Bool("pull-request", false, `Overrides pull request status of the build`)
	CronJobFlag     = flag.Bool("cron-job", false, `Overrides cron job status of the build`)
)

// Environment contains metadata provided by the build environment.
type Environment struct {
	Name                      string // name of the environment
	Repo                      string // name of GitHub repo
	Commit, Date, Branch, Tag string // Git info
	Type                      string
	Buildnum                  string
	IsPullRequest             bool
	IsCronJob                 bool
}

func (env Environment) String() string {
	return fmt.Sprintf("%s env (commit:%s date:%s branch:%s tag:%s buildnum:%s pr:%t)",
		env.Name, env.Commit, env.Date, env.Branch, env.Tag, env.Buildnum, env.IsPullRequest)
}

// Env returns metadata about the current CI environment, falling back to LocalEnv
// if not running on CI.
func Env() Environment {
	if os.Getenv("CI") == "true" && os.Getenv("GITHUB_ACTIONS") == "true" {
		commit := os.Getenv("GITHUB_SHA")
		return Environment{
			Name:          "github actions",
			Repo:          os.Getenv("GITHUB_REPOSITORY"),
			Commit:        os.Getenv("GITHUB_SHA"),
			Date:          getDate(commit),
			Branch:        os.Getenv("GITHUB_REF_NAME"),
			Tag:           os.Getenv("GITHUB_REF_NAME"),
			Buildnum:      os.Getenv("GITHUB_RUN_ID"),
			Type:          os.Getenv("GITHUB_REF_TYPE"),
			IsPullRequest: false,
			IsCronJob:     false,
		}
	} else {
		return LocalEnv()
	}
}

// LocalEnv returns build environment metadata gathered from git.
func LocalEnv() Environment {
	env := applyEnvFlags(Environment{Name: "local", Repo: "core-coin/go-core"})

	head := readGitFile("HEAD")
	if fields := strings.Fields(head); len(fields) == 2 {
		head = fields[1]
	} else {
		// In this case we are in "detached head" state
		// see: https://git-scm.com/docs/git-checkout#_detached_head
		// Additional check required to verify, that file contains commit hash
		commitRe, _ := regexp.Compile("^([0-9a-f]{40})$")
		if commit := commitRe.FindString(head); commit != "" && env.Commit == "" {
			env.Commit = commit
		}
		return env
	}
	if env.Commit == "" {
		env.Commit = readGitFile(head)
	}
	env.Date = getDate(env.Commit)
	if env.Branch == "" {
		if head != "HEAD" {
			env.Branch = strings.TrimPrefix(head, "refs/heads/")
		}
	}
	if info, err := os.Stat(".git/objects"); err == nil && info.IsDir() && env.Tag == "" {
		env.Tag = firstLine(RunGit("describe", "--tags", "--abbrev=0"))
	}
	return env
}

func firstLine(s string) string {
	return strings.Split(s, "\n")[0]
}

func getDate(commit string) string {
	if commit == "" {
		return ""
	}
	out := RunGit("show", "-s", "--format=%ct", commit)
	if out == "" {
		return ""
	}
	date, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse git commit date: %v", err))
	}
	return time.Unix(date, 0).Format("20060102")
}

func applyEnvFlags(env Environment) Environment {
	if !flag.Parsed() {
		panic("you need to call flag.Parse before Env or LocalEnv")
	}
	if *GitCommitFlag != "" {
		env.Commit = *GitCommitFlag
	}
	if *GitBranchFlag != "" {
		env.Branch = *GitBranchFlag
	}
	if *GitTagFlag != "" {
		env.Tag = *GitTagFlag
	}
	if *BuildnumFlag != "" {
		env.Buildnum = *BuildnumFlag
	}
	if *PullRequestFlag {
		env.IsPullRequest = true
	}
	if *CronJobFlag {
		env.IsCronJob = true
	}
	return env
}
