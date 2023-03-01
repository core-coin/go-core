// Copyright 2016 The go-core Authors
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

/*
The ci command is called from Continuous Integration scripts.

Usage: go run build/ci.go <command> <command flags/arguments>

Available commands are:

	install    [ -arch architecture ] [ -cc compiler ] [ packages... ]                          -- builds packages and executables
	test       [ -coverage ] [ packages... ]                                                    -- runs the tests
	lint                                                                                        -- runs certain pre-selected linters
	archive    [ -arch architecture ] [ -type zip|tar ] [ -signer key-envvar ] [ -signify key-envvar ] [ -upload dest ] -- archives build artifacts
	importkeys                                                                                  -- imports signing keys from env
	debsrc     [ -signer key-id ] [ -upload dest ]                                              -- creates a debian source package
	nsis                                                                                        -- creates a Windows NSIS installer
	aar        [ -local ] [ -sign key-id ] [-deploy repo] [ -upload dest ]                      -- creates an Android archive
	xcode      [ -local ] [ -sign key-id ] [-deploy repo] [ -upload dest ]                      -- creates an iOS XCode framework
	xgo        [ -alltools ] [ options ]                                                        -- cross builds according to options
	purge      [ -store blobstore ] [ -days threshold ]                                         -- purges old archives from the blobstore

For all commands, -n prevents execution of external programs (dry run mode).
*/
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/core-coin/go-core/v2/internal/build"
)

var (
	// Files that end up in the gocore*.zip archive.
	gocoreArchiveFiles = []string{
		"COPYING",
		executablePath("gocore"),
	}

	// Files that end up in the gocore-alltools*.zip archive.
	allToolsArchiveFiles = []string{
		"COPYING",
		executablePath("abigen"),
		executablePath("bootnode"),
		executablePath("cvm"),
		executablePath("gocore"),
		executablePath("rlpdump"),
		executablePath("clef"),
	}
	// This is the version of go that will be downloaded by
	//
	//     go run ci.go install -dlgo
	dlgoVersion = "1.15.6"
)

var GOBIN, _ = filepath.Abs(filepath.Join("build", "bin"))

func executablePath(name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(GOBIN, name)
}

func main() {
	log.SetFlags(log.Lshortfile)

	if _, err := os.Stat(filepath.Join("build", "ci.go")); os.IsNotExist(err) {
		log.Fatal("this script must be run from the root of the repository")
	}
	if len(os.Args) < 2 {
		log.Fatal("need subcommand as first argument")
	}
	switch os.Args[1] {
	case "install":
		doInstall(os.Args[2:])
	case "test":
		doTest(os.Args[2:])
	case "lint":
		doLint(os.Args[2:])
	case "xgo":
		doXgo(os.Args[2:])
	case "purge":
		doPurge(os.Args[2:])
	default:
		log.Fatal("unknown command ", os.Args[1])
	}
}

// Compiling

func doInstall(cmdline []string) {
	var (
		dlgo = flag.Bool("dlgo", false, "Download Go and build with it")
		arch = flag.String("arch", "", "Architecture to cross build for")
		cc   = flag.String("cc", "", "C compiler to cross build with")
	)
	flag.CommandLine.Parse(cmdline)
	env := build.Env()

	// Check local Go version. People regularly open issues about compilation
	// failure with outdated Go. This should save them the trouble.
	if !strings.Contains(runtime.Version(), "devel") {
		// Figure out the minor version number since we can't textually compare (1.10 < 1.9)
		var minor int
		fmt.Sscanf(strings.TrimPrefix(runtime.Version(), "go1."), "%d", &minor)
		if minor < 13 {
			log.Println("You have Go version", runtime.Version())
			log.Println("go-core requires at least Go version 1.13 and cannot")
			log.Println("be compiled with an earlier version. Please upgrade your Go installation.")
			os.Exit(1)
		}
	}

	// Choose which go command we're going to use.
	var gobuild *exec.Cmd
	if !*dlgo {
		// Default behavior: use the go version which runs ci.go right now.
		gobuild = goTool("build")
	} else {
		// Download of Go requested. This is for build environments where the
		// installed version is too old and cannot be upgraded easily.
		cachedir := filepath.Join("build", "cache")
		goroot := downloadGo(runtime.GOARCH, runtime.GOOS, cachedir)
		gobuild = localGoTool(goroot, "build")
	}

	// Configure environment for cross build.
	if *arch != "" && *arch != runtime.GOARCH {
		gobuild.Env = append(gobuild.Env, "CGO_ENABLED=1")
		gobuild.Env = append(gobuild.Env, "GOARCH="+*arch)
	}

	// Configure C compiler.
	if *cc != "" {
		gobuild.Env = append(gobuild.Env, "CC="+*cc)
	} else if os.Getenv("CC") != "" {
		gobuild.Env = append(gobuild.Env, "CC="+os.Getenv("CC"))
	}

	// arm64 CI builders are memory-constrained and can't handle concurrent builds,
	// better disable it. This check isn't the best, it should probably
	// check for something in env instead.
	if runtime.GOARCH == "arm64" {
		gobuild.Args = append(gobuild.Args, "-p", "1")
	}

	// Put the default settings in.
	gobuild.Args = append(gobuild.Args, buildFlags(env)...)

	// We use -trimpath to avoid leaking local paths into the built executables.
	gobuild.Args = append(gobuild.Args, "-trimpath")

	// Show packages during build.
	gobuild.Args = append(gobuild.Args, "-v")

	if runtime.GOOS == "windows" {
		gobuild.Args = append(gobuild.Args, "-buildmode=exe")
	}
	// Now we choose what we're even building.
	// Default: collect all 'main' packages in cmd/ and build those.
	packages := flag.Args()
	if len(packages) == 0 {
		packages = build.FindMainPackages("./cmd")
	}

	// Do the build!
	for _, pkg := range packages {
		args := make([]string, len(gobuild.Args))
		copy(args, gobuild.Args)
		args = append(args, "-o", executablePath(path.Base(pkg)))
		args = append(args, pkg)
		build.MustRun(&exec.Cmd{Path: gobuild.Path, Args: args, Env: gobuild.Env})
	}
}

// buildFlags returns the go tool flags for building.
func buildFlags(env build.Environment) (flags []string) {
	var ld []string
	if env.Commit != "" {
		ld = append(ld, "-X", "main.gitTag="+env.Tag)
		ld = append(ld, "-X", "main.gitCommit="+env.Commit)
		ld = append(ld, "-X", "main.gitDate="+env.Date)
	}
	// Strip DWARF on darwin. This used to be required for certain things,
	// and there is no downside to this, so we just keep doing it.
	if runtime.GOOS == "darwin" {
		ld = append(ld, "-s")
	}
	if len(ld) > 0 {
		flags = append(flags, "-ldflags", strings.Join(ld, " "))
	}
	return flags
}

// goTool returns the go tool. This uses the Go version which runs ci.go.
func goTool(subcmd string, args ...string) *exec.Cmd {
	cmd := build.GoTool(subcmd, args...)
	goToolSetEnv(cmd)
	return cmd
}

// localGoTool returns the go tool from the given GOROOT.
func localGoTool(goroot string, subcmd string, args ...string) *exec.Cmd {
	gotool := filepath.Join(goroot, "bin", "go")
	cmd := exec.Command(gotool, subcmd)
	goToolSetEnv(cmd)
	cmd.Env = append(cmd.Env, "GOROOT="+goroot)
	cmd.Args = append(cmd.Args, args...)
	return cmd
}

// goToolSetEnv forwards the build environment to the go tool.
func goToolSetEnv(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, "GOBIN="+GOBIN)
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOBIN=") || strings.HasPrefix(e, "CC=") {
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}
}

// Running The Tests
//
// "tests" also includes static analysis tools such as vet.

func doTest(cmdline []string) {
	coverage := flag.Bool("coverage", false, "Whether to record code coverage")
	verbose := flag.Bool("v", false, "Whether to log verbosely")
	flag.CommandLine.Parse(cmdline)
	env := build.Env()

	packages := []string{"./..."}
	if len(flag.CommandLine.Args()) > 0 {
		packages = flag.CommandLine.Args()
	}

	// Run the actual tests.
	// Test a single package at a time. CI builders are slow
	// and some tests run into timeouts under load.
	gotest := goTool("test", buildFlags(env)...)
	gotest.Args = append(gotest.Args, "-p", "1", "-timeout", "20m")
	if *coverage {
		gotest.Args = append(gotest.Args, "-covermode=atomic", "-cover")
	}
	if *verbose {
		gotest.Args = append(gotest.Args, "-v")
	}

	gotest.Args = append(gotest.Args, packages...)

	// windows sometimes fails to remove build temp dir
	if runtime.GOOS == "windows" {
		fmt.Println(">>>", strings.Join(gotest.Args, " "))
		var buf bytes.Buffer
		gotest.Stderr = &buf
		gotest.Stdout = &buf

		err := gotest.Run()
		res := buf.String()
		fmt.Println(res)
		if err != nil {
			if strings.Contains(res, "Access is denied") && !strings.Contains(res, "FAIL") {
				return
			} else {
				log.Fatal(err)
			}
		} else {
			return
		}
	}

	build.MustRun(gotest)
}

// doLint runs golangci-lint on requested packages.
func doLint(cmdline []string) {
	var (
		cachedir = flag.String("cachedir", "./build/cache", "directory for caching golangci-lint binary.")
	)
	flag.CommandLine.Parse(cmdline)
	packages := []string{"./..."}
	if len(flag.CommandLine.Args()) > 0 {
		packages = flag.CommandLine.Args()
	}

	linter := downloadLinter(*cachedir)
	lflags := []string{"run", "--config", ".golangci.yml"}
	build.MustRunCommand(linter, append(lflags, packages...)...)
	fmt.Println("You have achieved perfection.")
}

// downloadLinter downloads and unpacks golangci-lint.
func downloadLinter(cachedir string) string {
	const version = "1.27.0"

	csdb := build.MustLoadChecksums("build/checksums.txt")
	base := fmt.Sprintf("golangci-lint-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://github.com/golangci/golangci-lint/releases/download/v%s/%s.tar.gz", version, base)
	archivePath := filepath.Join(cachedir, base+".tar.gz")
	if err := csdb.DownloadFile(url, archivePath); err != nil {
		log.Fatal(err)
	}
	if err := build.ExtractArchive(archivePath, cachedir); err != nil {
		log.Fatal(err)
	}
	return filepath.Join(cachedir, base, "golangci-lint")
}

// downloadGoSources downloads the Go source tarball.
func downloadGoSources(cachedir string) string {
	csdb := build.MustLoadChecksums("build/checksums.txt")
	file := fmt.Sprintf("go%s.src.tar.gz", dlgoVersion)
	url := "https://dl.google.com/go/" + file
	dst := filepath.Join(cachedir, file)
	if err := csdb.DownloadFile(url, dst); err != nil {
		log.Fatal(err)
	}
	return dst
}

// downloadGo downloads the Go binary distribution and unpacks it into a temporary
// directory. It returns the GOROOT of the unpacked toolchain.
func downloadGo(goarch, goos, cachedir string) string {
	if goarch == "arm" {
		goarch = "armv6l"
	}

	csdb := build.MustLoadChecksums("build/checksums.txt")
	file := fmt.Sprintf("go%s.%s-%s", dlgoVersion, goos, goarch)
	if goos == "windows" {
		file += ".zip"
	} else {
		file += ".tar.gz"
	}
	url := "https://golang.org/dl/" + file
	dst := filepath.Join(cachedir, file)
	if err := csdb.DownloadFile(url, dst); err != nil {
		log.Fatal(err)
	}

	ucache, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}
	godir := filepath.Join(ucache, fmt.Sprintf("gocore-go-%s-%s-%s", dlgoVersion, goos, goarch))
	if err := build.ExtractArchive(dst, godir); err != nil {
		log.Fatal(err)
	}
	goroot, err := filepath.Abs(filepath.Join(godir, "go"))
	if err != nil {
		log.Fatal(err)
	}
	return goroot
}

// Cross compilation

func doXgo(cmdline []string) {
	var (
		alltools = flag.Bool("alltools", false, `Flag whether we're building all known tools, or only on in particular`)
	)
	flag.CommandLine.Parse(cmdline)
	env := build.Env()

	// Make sure xgo is available for cross compilation
	gogetxgo := goTool("get", "github.com/karalabe/xgo")
	build.MustRun(gogetxgo)

	// If all tools building is requested, build everything the builder wants
	args := append(buildFlags(env), flag.Args()...)

	if *alltools {
		args = append(args, []string{"--dest", GOBIN}...)
		for _, res := range allToolsArchiveFiles {
			if strings.HasPrefix(res, GOBIN) {
				// Binary tool found, cross build it explicitly
				args = append(args, "./"+filepath.Join("cmd", filepath.Base(res)))
				xgo := xgoTool(args)
				build.MustRun(xgo)
				args = args[:len(args)-1]
			}
		}
		return
	}
	// Otherwise xxecute the explicit cross compilation
	path := args[len(args)-1]
	args = append(args[:len(args)-1], []string{"--dest", GOBIN, path}...)

	xgo := xgoTool(args)
	build.MustRun(xgo)
}

func xgoTool(args []string) *exec.Cmd {
	cmd := exec.Command(filepath.Join(GOBIN, "xgo"), args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, []string{
		"GOBIN=" + GOBIN,
	}...)
	return cmd
}

// Binary distribution cleanups

func doPurge(cmdline []string) {
	var (
		store = flag.String("store", "", `Destination from where to purge archives (usually "gocorestore/builds")`)
		limit = flag.Int("days", 30, `Age threshold above which to delete unstable archives`)
	)
	flag.CommandLine.Parse(cmdline)

	if env := build.Env(); !env.IsCronJob {
		log.Printf("skipping because not a cron job")
		os.Exit(0)
	}
	// Create the azure authentication and list the current archives
	auth := build.AzureBlobstoreConfig{
		Account:   strings.Split(*store, "/")[0],
		Token:     os.Getenv("AZURE_BLOBSTORE_TOKEN"),
		Container: strings.SplitN(*store, "/", 2)[1],
	}
	blobs, err := build.AzureBlobstoreList(auth)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d blobs\n", len(blobs))

	// Iterate over the blobs, collect and sort all unstable builds
	for i := 0; i < len(blobs); i++ {
		if !strings.Contains(blobs[i].Name, "unstable") {
			blobs = append(blobs[:i], blobs[i+1:]...)
			i--
		}
	}
	for i := 0; i < len(blobs); i++ {
		for j := i + 1; j < len(blobs); j++ {
			if blobs[i].Properties.LastModified.After(blobs[j].Properties.LastModified) {
				blobs[i], blobs[j] = blobs[j], blobs[i]
			}
		}
	}
	// Filter out all archives more recent that the given threshold
	for i, blob := range blobs {
		if time.Since(blob.Properties.LastModified) < time.Duration(*limit)*24*time.Hour {
			blobs = blobs[:i]
			break
		}
	}
	fmt.Printf("Deleting %d blobs\n", len(blobs))
	// Delete all marked as such and return
	if err := build.AzureBlobstoreDelete(auth, blobs); err != nil {
		log.Fatal(err)
	}
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
