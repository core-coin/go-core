// Copyright 2023 by the Authors
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
	"fmt"
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
	fmt.Println(keystore)
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

func TestAccountImport(t *testing.T) {
	tests := []struct{ name, key, output string }{
		{
			name:   "correct account",
			key:    "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e",
			output: "Address: {fcad0b19bb29d4674531d6f115237e16afce377c}\n",
		},
		{
			name:   "invalid character",
			key:    "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e1",
			output: "Fatal: Failed to load the private key: invalid character '1' at end of key file\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			importAccountWithExpect(t, test.key, test.output)
		})
	}
}

func importAccountWithExpect(t *testing.T, key string, expected string) {
	dir := tmpdir(t)
	keyfile := filepath.Join(dir, "key.prv")
	if err := ioutil.WriteFile(keyfile, []byte(key), 0600); err != nil {
		t.Error(err)
	}
	passwordFile := filepath.Join(dir, "password.txt")
	if err := ioutil.WriteFile(passwordFile, []byte("foobar"), 0600); err != nil {
		t.Error(err)
	}
	gocore := runGocore(t, "account", "import", keyfile, "-password", passwordFile)
	defer gocore.ExpectExit()
	gocore.Expect(expected)
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
		"cb74db416ff2f9c53dabaf34f81142db30350ea7b144")
	defer gocore.ExpectExit()
	gocore.Expect(`
Unlocking account cb74db416ff2f9c53dabaf34f81142db30350ea7b144 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Please give a new password. Do not forget this password.
Password: {{.InputLine "foobar"}}
Repeat password: {{.InputLine "foobar"}}
`)
}

func TestUnlockFlag(t *testing.T) {
	gocore := runMinimalGocore(t, "--port", "0", "--ipcdisable", "--datadir", tmpDatadirWithKeystore(t),
		"--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144", "js", "testdata/empty.js")
	gocore.Expect(`
Unlocking account cb74db416ff2f9c53dabaf34f81142db30350ea7b144 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
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

func TestUnlockFlagWrongPassword(t *testing.T) {
	gocore := runMinimalGocore(t, "--port", "0", "--ipcdisable", "--datadir", tmpDatadirWithKeystore(t),
		"--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144", "js", "testdata/empty.js")

	defer gocore.ExpectExit()
	gocore.Expect(`
Unlocking account cb74db416ff2f9c53dabaf34f81142db30350ea7b144 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "wrong1"}}
Unlocking account cb74db416ff2f9c53dabaf34f81142db30350ea7b144 | Attempt 2/3
Password: {{.InputLine "wrong2"}}
Unlocking account cb74db416ff2f9c53dabaf34f81142db30350ea7b144 | Attempt 3/3
Password: {{.InputLine "wrong3"}}
Fatal: Failed to unlock account cb74db416ff2f9c53dabaf34f81142db30350ea7b144 (could not decrypt key with given password)
`)
}

// https://github.com/core-coin/go-core/v2/issues/1785
func TestUnlockFlagMultiIndex(t *testing.T) {
	gocore := runMinimalGocore(t, "--port", "0", "--ipcdisable", "--datadir", tmpDatadirWithKeystore(t),
		"--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144", "--unlock", "0,2", "js", "testdata/empty.js")

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
	gocore := runMinimalGocore(t, "--port", "0", "--ipcdisable", "--datadir", tmpDatadirWithKeystore(t),
		"--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144", "--password", "testdata/passwords.txt", "--unlock", "0,2", "js", "testdata/empty.js")

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
	gocore := runMinimalGocore(t, "--port", "0", "--ipcdisable", "--datadir", tmpDatadirWithKeystore(t),
		"--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144", "--password",
		"testdata/wrong-passwords.txt", "--unlock", "0,2")
	defer gocore.ExpectExit()
	gocore.Expect(`
Fatal: Failed to unlock account 0 (could not decrypt key with given password)
`)
}

func TestUnlockFlagAmbiguous(t *testing.T) {
	store := filepath.Join("..", "..", "accounts", "keystore", "testdata", "dupes")
	gocore := runMinimalGocore(t, "--port", "0", "--ipcdisable", "--datadir", tmpDatadirWithKeystore(t),
		"--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144", "--keystore",
		store, "--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144",
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
	gocore := runMinimalGocore(t, "--port", "0", "--ipcdisable", "--datadir", tmpDatadirWithKeystore(t),
		"--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144", "--keystore",
		store, "--unlock", "cb74db416ff2f9c53dabaf34f81142db30350ea7b144")

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
