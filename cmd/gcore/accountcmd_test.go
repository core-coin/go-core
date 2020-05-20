// Copyright 2016 The go-core Authors
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
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cespare/cp"
)

// These tests are 'smoke tests' for the account related
// subcommands and flags.
//
// For most tests, the test files from package accounts
// are copied into a temporary keystore directory.

func tmpDatadirWithKeystore(t *testing.T) string {
	datadir := tmpdir(t)
	keystore := filepath.Join(datadir, "keystore")
	source := filepath.Join("..", "..", "accounts", "keystore", "testdata", "keystore")
	if err := cp.CopyAll(keystore, source); err != nil {
		t.Fatal(err)
	}
	return datadir
}

func TestAccountListEmpty(t *testing.T) {
	gcore := runGcore(t, "account", "list")
	gcore.ExpectExit()
}

func TestAccountList(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gcore := runGcore(t, "account", "list", "--datadir", datadir)
	defer gcore.ExpectExit()
	if runtime.GOOS == "windows" {
		gcore.Expect(`
Account #0: {c91938bff440fac4e6d4ff24234f52c5671fb0d4} keystore://{{.Datadir}}\keystore\UTC--2020-05-20T08-11-06.057753867Z--c91938bff440fac4e6d4ff24234f52c5671fb0d4
Account #1: {7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db} keystore://{{.Datadir}}\keystore\aaa
Account #2: {1033edd15730a1b56c4812879c01bf4754a6dda6} keystore://{{.Datadir}}\keystore\zzz
`)
	} else {
		gcore.Expect(`
Account #0: {c91938bff440fac4e6d4ff24234f52c5671fb0d4} keystore://{{.Datadir}}/keystore/UTC--2020-05-20T08-11-06.057753867Z--c91938bff440fac4e6d4ff24234f52c5671fb0d4
Account #1: {7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db} keystore://{{.Datadir}}/keystore/aaa
Account #2: {1033edd15730a1b56c4812879c01bf4754a6dda6} keystore://{{.Datadir}}/keystore/zzz
`)
	}
}

func TestAccountNew(t *testing.T) {
	gcore := runGcore(t, "account", "new", "--lightkdf")
	defer gcore.ExpectExit()
	gcore.Expect(`
Your new account is locked with a password. Please give a password. Do not forget this password.
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Repeat password: {{.InputLine "foobar"}}

Your new key was generated
`)
	gcore.ExpectRegexp(`
Public address of the key:   0x[0-9a-fA-F]{40}
Path of the secret key file: .*UTC--.+--[0-9a-f]{40}

- You can share your public address with anyone. Others need it to interact with you.
- You must NEVER share the secret key with anyone! The key controls access to your funds!
- You must BACKUP your key file! Without the key, it's impossible to access account funds!
- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!
`)
}

func TestAccountNewBadRepeat(t *testing.T) {
	gcore := runGcore(t, "account", "new", "--lightkdf")
	defer gcore.ExpectExit()
	gcore.Expect(`
Your new account is locked with a password. Please give a password. Do not forget this password.
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "something"}}
Repeat password: {{.InputLine "something else"}}
Fatal: Passwords do not match
`)
}

func TestAccountUpdate(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gcore := runGcore(t, "account", "update",
		"--datadir", datadir, "--lightkdf",
		"7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db")
	defer gcore.ExpectExit()
	gcore.Expect(`
Unlocking account 7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Please give a new password. Do not forget this password.
Password: {{.InputLine "foobar2"}}
Repeat password: {{.InputLine "foobar2"}}
`)
}

func TestWalletImport(t *testing.T) {
	t.Skip()
	gcore := runGcore(t, "wallet", "import", "--lightkdf", "testdata/guswallet.json")
	defer gcore.ExpectExit()
	gcore.Expect(`
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foo"}}
Address: {d4584b5f6229b7be90727b0fc8c6b91bb427821f}
`)

	files, err := ioutil.ReadDir(filepath.Join(gcore.Datadir, "keystore"))
	if len(files) != 1 {
		t.Errorf("expected one key file in keystore directory, found %d files (error: %v)", len(files), err)
	}
}

func TestWalletImportBadPassword(t *testing.T) {
	gcore := runGcore(t, "wallet", "import", "--lightkdf", "testdata/guswallet.json")
	defer gcore.ExpectExit()
	gcore.Expect(`
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "wrong"}}
Fatal: could not decrypt key with given password
`)
}

func TestUnlockFlag(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gcore := runGcore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db",
		"js", "testdata/empty.js")
	gcore.Expect(`
Unlocking account 7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
`)
	gcore.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=0x7c1C6Dd70d43d9ceE327FE1F2162fE3CA91CE8Db",
	}
	for _, m := range wantMessages {
		if !strings.Contains(gcore.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagWrongPassword(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gcore := runGcore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db")
	defer gcore.ExpectExit()
	gcore.Expect(`
Unlocking account 7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "wrong1"}}
Unlocking account 7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db | Attempt 2/3
Password: {{.InputLine "wrong2"}}
Unlocking account 7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db | Attempt 3/3
Password: {{.InputLine "wrong3"}}
Fatal: Failed to unlock account 7c1c6dd70d43d9cee327fe1f2162fe3ca91ce8db (could not decrypt key with given password)
`)
}

// https://github.com/core-coin/go-core/issues/1785
func TestUnlockFlagMultiIndex(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gcore := runGcore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "0,2",
		"js", "testdata/empty.js")
	gcore.Expect(`
Unlocking account 0 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Unlocking account 2 | Attempt 1/3
Password: {{.InputLine "foobar"}}
`)
	gcore.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=0xc91938bFF440FaC4E6d4Ff24234f52C5671fB0D4",
		"=0x1033edD15730a1b56c4812879C01bf4754A6Dda6",
	}
	for _, m := range wantMessages {
		if !strings.Contains(gcore.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagPasswordFile(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gcore := runGcore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--password", "testdata/passwords.txt", "--unlock", "0,2",
		"js", "testdata/empty.js")
	gcore.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=0xc91938bFF440FaC4E6d4Ff24234f52C5671fB0D4",
		"=0x1033edD15730a1b56c4812879C01bf4754A6Dda6",
	}
	for _, m := range wantMessages {
		if !strings.Contains(gcore.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagPasswordFileWrongPassword(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gcore := runGcore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--password", "testdata/wrong-passwords.txt", "--unlock", "0,2")
	defer gcore.ExpectExit()
	gcore.Expect(`
Fatal: Failed to unlock account 0 (could not decrypt key with given password)
`)
}

func TestUnlockFlagAmbiguous(t *testing.T) {
	store := filepath.Join("..", "..", "accounts", "keystore", "testdata", "dupes")
	gcore := runGcore(t,
		"--keystore", store, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "c91938bff440fac4e6d4ff24234f52c5671fb0d4",
		"js", "testdata/empty.js")
	defer gcore.ExpectExit()

	// Helper for the expect template, returns absolute keystore path.
	gcore.SetTemplateFunc("keypath", func(file string) string {
		abs, _ := filepath.Abs(filepath.Join(store, file))
		return abs
	})
	gcore.Expect(`
Unlocking account c91938bff440fac4e6d4ff24234f52c5671fb0d4 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Multiple key files exist for address c91938bff440fac4e6d4ff24234f52c5671fb0d4:
   keystore://{{keypath "1"}}
   keystore://{{keypath "2"}}
Testing your password against all of them...
Your password unlocked keystore://{{keypath "1"}}
In order to avoid this warning, you need to remove the following duplicate key files:
   keystore://{{keypath "2"}}
`)
	gcore.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=0xc91938bFF440FaC4E6d4Ff24234f52C5671fB0D4",
	}
	for _, m := range wantMessages {
		if !strings.Contains(gcore.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagAmbiguousWrongPassword(t *testing.T) {
	store := filepath.Join("..", "..", "accounts", "keystore", "testdata", "dupes")
	gcore := runGcore(t,
		"--keystore", store, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "c91938bff440fac4e6d4ff24234f52c5671fb0d4")
	defer gcore.ExpectExit()

	// Helper for the expect template, returns absolute keystore path.
	gcore.SetTemplateFunc("keypath", func(file string) string {
		abs, _ := filepath.Abs(filepath.Join(store, file))
		return abs
	})
	gcore.Expect(`
Unlocking account c91938bff440fac4e6d4ff24234f52c5671fb0d4 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "wrong"}}
Multiple key files exist for address c91938bff440fac4e6d4ff24234f52c5671fb0d4:
   keystore://{{keypath "1"}}
   keystore://{{keypath "2"}}
Testing your password against all of them...
Fatal: None of the listed files could be unlocked.
`)
	gcore.ExpectExit()
}
