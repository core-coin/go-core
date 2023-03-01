// Copyright 2014 by the Authors
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

package keystore

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/core-coin/go-core/v2/common"
)

func tmpKeyStoreIface(t *testing.T, encrypted bool) (dir string, ks keyStore) {
	d, err := ioutil.TempDir("", "gocore-keystore-test")
	if err != nil {
		t.Fatal(err)
	}
	if encrypted {
		ks = &keyStorePassphrase{d, veryLightScryptN, veryLightScryptP, true}
	} else {
		ks = &keyStorePlain{d}
	}
	return d, ks
}

func TestKeyStorePlain(t *testing.T) {
	dir, ks := tmpKeyStoreIface(t, false)
	defer os.RemoveAll(dir)

	pass := "" // not used but required by API
	k1, account, err := storeNewKey(ks, rand.Reader, pass)
	if err != nil {
		t.Fatal(err)
	}
	k2, err := ks.GetKey(k1.Address, account.URL.Path, pass)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(k1.Address, k2.Address) {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(k1.PrivateKey.PrivateKey(), k2.PrivateKey.PrivateKey()) {
		t.Fatal(err)
	}
}

func TestKeyStorePassphrase(t *testing.T) {
	dir, ks := tmpKeyStoreIface(t, true)
	defer os.RemoveAll(dir)

	pass := "foo"
	k1, account, err := storeNewKey(ks, rand.Reader, pass)
	if err != nil {
		t.Fatal(err)
	}
	k2, err := ks.GetKey(k1.Address, account.URL.Path, pass)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(k1.Address, k2.Address) {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(k1.PrivateKey, k2.PrivateKey) {
		t.Fatal(err)
	}
}

func TestKeyStorePassphraseDecryptionFail(t *testing.T) {
	dir, ks := tmpKeyStoreIface(t, true)
	defer os.RemoveAll(dir)

	pass := "foo"
	k1, account, err := storeNewKey(ks, rand.Reader, pass)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ks.GetKey(k1.Address, account.URL.Path, "bar"); err != ErrDecrypt {
		t.Fatalf("wrong error for invalid password\ngot %q\nwant %q", err, ErrDecrypt)
	}
}

// Test and utils for the key store tests in the Core JSON tests;
// testdataKeyStoreTests/basic_tests.json
type KeyStoreTestV3 struct {
	Json     encryptedKeyJSONV3
	Password string
	Priv     string
}

func TestV3_PBKDF2_1(t *testing.T) {
	t.Parallel()
	tests := loadKeyStoreTestV3("testdata/v3_test_vector.json", t)
	testDecryptV3(tests["wikipage_test_vector_pbkdf2"], t)
}

var testsSubmodule = filepath.Join("..", "..", "tests", "testdata", "KeyStoreTests")

func skipIfSubmoduleMissing(t *testing.T) {
	if !common.FileExist(testsSubmodule) {
		t.Skipf("can't find JSON tests from submodule at %s", testsSubmodule)
	}
}

func TestV3_PBKDF2_2(t *testing.T) {
	skipIfSubmoduleMissing(t)
	t.Parallel()
	tests := loadKeyStoreTestV3(filepath.Join(testsSubmodule, "basic_tests.json"), t)
	testDecryptV3(tests["test1"], t)
}

func TestV3_PBKDF2_3(t *testing.T) {
	skipIfSubmoduleMissing(t)
	t.Parallel()
	tests := loadKeyStoreTestV3(filepath.Join(testsSubmodule, "basic_tests.json"), t)
	testDecryptV3(tests["python_generated_test_with_odd_iv"], t)
}

func TestV3_PBKDF2_4(t *testing.T) {
	skipIfSubmoduleMissing(t)
	t.Parallel()
	tests := loadKeyStoreTestV3(filepath.Join(testsSubmodule, "basic_tests.json"), t)
	testDecryptV3(tests["evilnonce"], t)
}

func TestV3_Scrypt_1(t *testing.T) {
	t.Parallel()
	tests := loadKeyStoreTestV3("testdata/v3_test_vector.json", t)
	testDecryptV3(tests["wikipage_test_vector_scrypt"], t)
}

func TestV3_Scrypt_2(t *testing.T) {
	skipIfSubmoduleMissing(t)
	t.Parallel()
	tests := loadKeyStoreTestV3(filepath.Join(testsSubmodule, "basic_tests.json"), t)
	testDecryptV3(tests["test2"], t)
}

func testDecryptV3(test KeyStoreTestV3, t *testing.T) {
	privBytes, _, err := decryptKeyV3(&test.Json, test.Password)
	if err != nil {
		t.Fatal(err)
	}
	privHex := hex.EncodeToString(privBytes)
	if test.Priv != privHex {
		t.Fatal(fmt.Errorf("Decrypted bytes not equal to test, expected %v have %v", test.Priv, privHex))
	}
}

func loadKeyStoreTestV3(file string, t *testing.T) map[string]KeyStoreTestV3 {
	tests := make(map[string]KeyStoreTestV3)
	err := common.LoadJSON(file, &tests)
	if err != nil {
		t.Fatal(err)
	}
	return tests
}

func TestV3_31_Byte_Key(t *testing.T) {
	t.Parallel()
	tests := loadKeyStoreTestV3("testdata/v3_test_vector.json", t)
	testDecryptV3(tests["31_byte_key"], t)
}

func TestV3_30_Byte_Key(t *testing.T) {
	t.Parallel()
	tests := loadKeyStoreTestV3("testdata/v3_test_vector.json", t)
	testDecryptV3(tests["30_byte_key"], t)
}
