// Copyright 2016 by the Authors
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
	gocore := runGocore(t, "account", "list")
	gocore.ExpectExit()
}

func TestAccountList(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gocore := runGocore(t, "account", "list", "--datadir", datadir)
	defer gocore.ExpectExit()
	if runtime.GOOS == "windows" {
		gocore.Expect(`
Account #0: {cb27de521e43741cf785cbad450d5649187b9612018f} keystore://{{.Datadir}}\keystore\UTC--2020-07-20T17-37-08.515483762Z--cb27de521e43741cf785cbad450d5649187b9612018f
Account #1: {cb74db416ff2f9c53dabaf34f81142db30350ea7b144} keystore://{{.Datadir}}\keystore\aaa
Account #2: {cb65e49851f010cd7d81b5b4969f3b0e8325c415359d} keystore://{{.Datadir}}\keystore\zzz
`)
	} else {
		gocore.Expect(`
Account #0: {cb27de521e43741cf785cbad450d5649187b9612018f} keystore://{{.Datadir}}/keystore/UTC--2020-07-20T17-37-08.515483762Z--cb27de521e43741cf785cbad450d5649187b9612018f
Account #1: {cb74db416ff2f9c53dabaf34f81142db30350ea7b144} keystore://{{.Datadir}}/keystore/aaa
Account #2: {cb65e49851f010cd7d81b5b4969f3b0e8325c415359d} keystore://{{.Datadir}}/keystore/zzz
`)
	}
}

func TestAccountNew(t *testing.T) {
	gocore := runGocore(t, "account", "new", "--lightkdf")
	defer gocore.ExpectExit()
	gocore.Expect(`
Your new account is locked with a password. Please give a password. Do not forget this password.
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Repeat password: {{.InputLine "foobar"}}

Your new key was generated
`)
	gocore.ExpectRegexp(`
Public address of the key:   [0-9a-fA-F]{44}
Path of the secret key file: .*UTC--.+--[0-9a-f]{44}

- You can share your public address with anyone. Others need it to interact with you.
- You must NEVER share the secret key with anyone! The key controls access to your funds!
- You must BACKUP your key file! Without the key, it's impossible to access account funds!
- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!
`)
}

func TestAccountNewBadRepeat(t *testing.T) {
	gocore := runGocore(t, "account", "new", "--lightkdf")
	defer gocore.ExpectExit()
	gocore.Expect(`
Your new account is locked with a password. Please give a password. Do not forget this password.
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "something"}}
Repeat password: {{.InputLine "something else"}}
Fatal: Passwords do not match
`)
}

func TestAccountUpdate(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gocore := runGocore(t, "account", "update",
		"--datadir", datadir, "--lightkdf",
		"cb65e49851f010cd7d81b5b4969f3b0e8325c415359d")
	defer gocore.ExpectExit()
	gocore.Expect(`
Unlocking account cb65e49851f010cd7d81b5b4969f3b0e8325c415359d | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Please give a new password. Do not forget this password.
Password: {{.InputLine "foobar"}}
Repeat password: {{.InputLine "foobar"}}
`)
}

func TestWalletImport(t *testing.T) {
	t.Skip()
	gocore := runGocore(t, "wallet", "import", "--lightkdf", "testdata/guswallet.json")
	defer gocore.ExpectExit()
	gocore.Expect(`
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foo"}}
Address: {d4584b5f6229b7be90727b0fc8c6b91bb427821f}
`)

	files, err := ioutil.ReadDir(filepath.Join(gocore.Datadir, "keystore"))
	if len(files) != 1 {
		t.Errorf("expected one key file in keystore directory, found %d files (error: %v)", len(files), err)
	}
}

func TestWalletImportBadPassword(t *testing.T) {
	gocore := runGocore(t, "wallet", "import", "--lightkdf", "testdata/guswallet.json")
	defer gocore.ExpectExit()
	gocore.Expect(`
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "wrong"}}
Fatal: could not decrypt key with given password
`)
}

func TestUnlockFlag(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gocore := runGocore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "cb65e49851f010cd7d81b5b4969f3b0e8325c415359d",
		"js", "testdata/empty.js")
	gocore.Expect(`
Unlocking account cb65e49851f010cd7d81b5b4969f3b0e8325c415359d | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
`)
	gocore.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=cb65e49851f010cd7d81b5b4969f3b0e8325c415359d",
	}
	for _, m := range wantMessages {
		if !strings.Contains(gocore.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagWrongPassword(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gocore := runGocore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "cb65e49851f010cd7d81b5b4969f3b0e8325c415359d")
	defer gocore.ExpectExit()
	gocore.Expect(`
Unlocking account cb65e49851f010cd7d81b5b4969f3b0e8325c415359d | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar1"}}
Unlocking account cb65e49851f010cd7d81b5b4969f3b0e8325c415359d | Attempt 2/3
Password: {{.InputLine "foobar1"}}
Unlocking account cb65e49851f010cd7d81b5b4969f3b0e8325c415359d | Attempt 3/3
Password: {{.InputLine "foobar1"}}
Fatal: Failed to unlock account cb65e49851f010cd7d81b5b4969f3b0e8325c415359d (could not decrypt key with given password)
`)
}

// https://github.com/core-coin/go-core/issues/1785
func TestUnlockFlagMultiIndex(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gocore := runGocore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "0,2",
		"js", "testdata/empty.js")
	gocore.Expect(`
Unlocking account 0 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Unlocking account 2 | Attempt 1/3
Password: {{.InputLine "foobar"}}
`)
	gocore.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=cb27de521e43741cf785cbad450d5649187b9612018f",
		"=cb65e49851f010cd7d81b5b4969f3b0e8325c415359d",
	}
	for _, m := range wantMessages {
		if !strings.Contains(gocore.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagPasswordFile(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gocore := runGocore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--password", "testdata/passwords.txt", "--unlock", "0,2",
		"js", "testdata/empty.js")
	gocore.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=cb27de521e43741cf785cbad450d5649187b9612018f",
		"=cb65e49851f010cd7d81b5b4969f3b0e8325c415359d",
	}
	for _, m := range wantMessages {
		if !strings.Contains(gocore.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagPasswordFileWrongPassword(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	gocore := runGocore(t,
		"--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--password", "testdata/wrong-passwords.txt", "--unlock", "0,2")
	defer gocore.ExpectExit()
	gocore.Expect(`
Fatal: Failed to unlock account 0 (could not decrypt key with given password)
`)
}

func TestUnlockFlagAmbiguous(t *testing.T) {
	store := filepath.Join("..", "..", "accounts", "keystore", "testdata", "dupes")
	gocore := runGocore(t,
		"--keystore", store, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144",
		"js", "testdata/empty.js")
	defer gocore.ExpectExit()

	// Helper for the expect template, returns absolute keystore path.
	gocore.SetTemplateFunc("keypath", func(file string) string {
		abs, _ := filepath.Abs(filepath.Join(store, file))
		return abs
	})
	gocore.Expect(`
Unlocking account cb74db416ff2f9c53dabaf34f81142db30350ea7b144 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Multiple key files exist for address cb74db416ff2f9c53dabaf34f81142db30350ea7b144:
   keystore://{{keypath "1"}}
   keystore://{{keypath "2"}}
Testing your password against all of them...
Your password unlocked keystore://{{keypath "1"}}
In order to avoid this warning, you need to remove the following duplicate key files:
   keystore://{{keypath "2"}}
`)
	gocore.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=cb74db416ff2f9c53dabaf34f81142db30350ea7b144",
	}
	for _, m := range wantMessages {
		if !strings.Contains(gocore.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagAmbiguousWrongPassword(t *testing.T) {
	store := filepath.Join("..", "..", "accounts", "keystore", "testdata", "dupes")
	gocore := runGocore(t,
		"--keystore", store, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144")
	defer gocore.ExpectExit()

	// Helper for the expect template, returns absolute keystore path.
	gocore.SetTemplateFunc("keypath", func(file string) string {
		abs, _ := filepath.Abs(filepath.Join(store, file))
		return abs
	})
	gocore.Expect(`
Unlocking account cb74db416ff2f9c53dabaf34f81142db30350ea7b144 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "wrong"}}
Multiple key files exist for address cb74db416ff2f9c53dabaf34f81142db30350ea7b144:
   keystore://{{keypath "1"}}
   keystore://{{keypath "2"}}
Testing your password against all of them...
Fatal: None of the listed files could be unlocked.
`)
	gocore.ExpectExit()
}
