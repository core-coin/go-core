// Copyright 2017 The go-core Authors
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
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cespare/cp"
	"github.com/core-coin/go-core/accounts"
	"github.com/core-coin/go-core/common"
	"github.com/davecgh/go-spew/spew"
)

var (
	addr1, err1       = common.HexToAddress("cb23feefe02c246be13166ddebab5e102eca211e496f")
	addr2, err2       = common.HexToAddress("cb32df9af33454ea4b84373adbdb3102bedb7838a228")
	addr3, err3       = common.HexToAddress("cb03af52d4eae20e8199abbe74a94392c01ed5cdcfab")
	cachetestDir, _   = filepath.Abs(filepath.Join("testdata", "keystore"))
	cachetestAccounts = []accounts.Account{
		{
			Address: addr1,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(cachetestDir, "UTC--2020-07-07T10-47-53.209137411Z--cb23feefe02c246be13166ddebab5e102eca211e496f")},
		},
		{
			Address: addr2,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(cachetestDir, "aaa")},
		},
		{
			Address: addr3,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(cachetestDir, "zzz")},
		},
	}
)

func TestWatchNewFile(t *testing.T) {
	t.Parallel()
	if err1 != nil || err2 != nil || err3 != nil {
		t.Error(err1, err1, err3)
	}
	dir, ks := tmpKeyStore(t, false)
	defer os.RemoveAll(dir)

	// Ensure the watcher is started before adding any files.
	ks.Accounts()
	time.Sleep(1000 * time.Millisecond)

	// Move in the files.
	wantAccounts := make([]accounts.Account, len(cachetestAccounts))
	for i := range cachetestAccounts {
		wantAccounts[i] = accounts.Account{
			Address: cachetestAccounts[i].Address,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, filepath.Base(cachetestAccounts[i].URL.Path))},
		}
		if err := cp.CopyFile(wantAccounts[i].URL.Path, cachetestAccounts[i].URL.Path); err != nil {
			t.Fatal(err)
		}
	}

	// ks should see the accounts.
	var list []accounts.Account
	for d := 200 * time.Millisecond; d < 5*time.Second; d *= 2 {
		list = ks.Accounts()
		if reflect.DeepEqual(list, wantAccounts) {
			// ks should have also received change notifications
			select {
			case <-ks.changes:
			default:
				t.Fatalf("wasn't notified of new accounts")
			}
			return
		}
		time.Sleep(d)
	}
	t.Errorf("got %s, want %s", spew.Sdump(list), spew.Sdump(wantAccounts))
}

func TestWatchNoDir(t *testing.T) {
	t.Parallel()

	// Create ks but not the directory that it watches.
	rand.Seed(time.Now().UnixNano())
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("xcc-keystore-watch-test-%d-%d", os.Getpid(), rand.Int()))
	ks := NewKeyStore(dir, LightScryptN, LightScryptP)

	list := ks.Accounts()
	if len(list) > 0 {
		t.Error("initial account list not empty:", list)
	}
	time.Sleep(100 * time.Millisecond)

	// Create the directory and copy a key file into it.
	os.MkdirAll(dir, 0700)
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "aaa")
	if err := cp.CopyFile(file, cachetestAccounts[0].URL.Path); err != nil {
		t.Fatal(err)
	}

	// ks should see the account.
	wantAccounts := []accounts.Account{cachetestAccounts[0]}
	wantAccounts[0].URL = accounts.URL{Scheme: KeyStoreScheme, Path: file}
	for d := 200 * time.Millisecond; d < 8*time.Second; d *= 2 {
		list = ks.Accounts()
		if reflect.DeepEqual(list, wantAccounts) {
			// ks should have also received change notifications
			select {
			case <-ks.changes:
			default:
				t.Fatalf("wasn't notified of new accounts")
			}
			return
		}
		time.Sleep(d)
	}
	t.Errorf("\ngot  %v\nwant %v", list, wantAccounts)
}

func TestCacheInitialReload(t *testing.T) {
	cache, _ := newAccountCache(cachetestDir)
	accounts := cache.accounts()
	if !reflect.DeepEqual(accounts, cachetestAccounts) {
		t.Fatalf("got initial accounts: %swant %s", spew.Sdump(accounts), spew.Sdump(cachetestAccounts))
	}
}

func TestCacheAddDeleteOrder(t *testing.T) {
	cache, _ := newAccountCache("testdata/no-such-dir")
	cache.watcher.running = true // prevent unexpected reloads
	addr1, err := common.HexToAddress("cb08095e7baea6a6c7c4c2dfeb977efac326af552d87")
	if err != nil {
		t.Error(err)
	}
	addr2, err := common.HexToAddress("cb292cac1adea150210703ba75ed097ddfe24e14f213")
	if err != nil {
		t.Error(err)
	}
	addr3, err := common.HexToAddress("cb458bda78331c916a08481428e4b07c96d3e916d165")
	if err != nil {
		t.Error(err)
	}
	addr4, err := common.HexToAddress("cb52d49ff4eeb0b2686ed89c0fc0f2b6ea533ddbbd5e")
	if err != nil {
		t.Error(err)
	}
	addr5, err := common.HexToAddress("cb72e8cf4629acb360350399b6cff367a97cf36e62b9")
	if err != nil {
		t.Error(err)
	}
	addr6, err := common.HexToAddress("cb625337dc760bff449e5c187e080797283ad42e69bd")
	if err != nil {
		t.Error(err)
	}
	addr7, err := common.HexToAddress("cb35348d6db8bfe52ab1199deeacbc4c1ffa0686d149")
	if err != nil {
		t.Error(err)
	}
	accs := []accounts.Account{
		{
			Address: addr1,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "-309830980"},
		},
		{
			Address: addr2,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "ggg"},
		},
		{
			Address: addr3,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "zzzzzz-the-very-last-one.keyXXX"},
		},
		{
			Address: addr4,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "SOMETHING.key"},
		},
		{
			Address: addr5,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "UTC--2020-02-27T09-51-06.983129511Z--cb72e8cf4629acb360350399b6cff367a97cf36e62b9"},
		},
		{
			Address: addr6,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "aaa"},
		},
		{
			Address: addr7,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "zzz"},
		},
	}
	for _, a := range accs {
		cache.add(a)
	}
	// Add some of them twice to check that they don't get reinserted.
	cache.add(accs[0])
	cache.add(accs[2])

	// Check that the account list is sorted by filename.
	wantAccounts := make([]accounts.Account, len(accs))
	copy(wantAccounts, accs)
	sort.Sort(accountsByURL(wantAccounts))
	list := cache.accounts()
	if !reflect.DeepEqual(list, wantAccounts) {
		t.Fatalf("got accounts: %s\nwant %s", spew.Sdump(accs), spew.Sdump(wantAccounts))
	}
	for _, a := range accs {
		if !cache.hasAddress(a.Address) {
			t.Errorf("expected hasAccount(%x) to return true", a.Address)
		}
	}
	wrongAcc, err := common.HexToAddress("cb29fd9bd350f08ee3c0c19b85a8e16114a11a60aa4e")
	if err != nil {
		t.Error(err)
	}
	if cache.hasAddress(wrongAcc) {
		t.Errorf("expected hasAccount(%x) to return false", wrongAcc)
	}

	// Delete a few keys from the cache.
	for i := 0; i < len(accs); i += 2 {
		cache.delete(wantAccounts[i])
	}
	cache.delete(accounts.Account{Address: wrongAcc, URL: accounts.URL{Scheme: KeyStoreScheme, Path: "something"}})

	// Check content again after deletion.
	wantAccountsAfterDelete := []accounts.Account{
		wantAccounts[1],
		wantAccounts[3],
		wantAccounts[5],
	}
	list = cache.accounts()
	if !reflect.DeepEqual(list, wantAccountsAfterDelete) {
		t.Fatalf("got accounts after delete: %s\nwant %s", spew.Sdump(list), spew.Sdump(wantAccountsAfterDelete))
	}
	for _, a := range wantAccountsAfterDelete {
		if !cache.hasAddress(a.Address) {
			t.Errorf("expected hasAccount(%x) to return true", a.Address)
		}
	}
	if cache.hasAddress(wantAccounts[0].Address) {
		t.Errorf("expected hasAccount(%x) to return false", wantAccounts[0].Address)
	}
}

func TestCacheFind(t *testing.T) {
	dir := filepath.Join("testdata", "dir")
	cache, _ := newAccountCache(dir)
	cache.watcher.running = true // prevent unexpected reloads

	addr1, err := common.HexToAddress("cb08095e7baea6a6c7c4c2dfeb977efac326af552d87")
	if err != nil {
		t.Error(err)
	}
	addr2, err := common.HexToAddress("cb292cac1adea150210703ba75ed097ddfe24e14f213")
	if err != nil {
		t.Error(err)
	}
	addr3, err := common.HexToAddress("cb52d49ff4eeb0b2686ed89c0fc0f2b6ea533ddbbd5e")
	if err != nil {
		t.Error(err)
	}
	accs := []accounts.Account{
		{
			Address: addr1,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "a.key")},
		},
		{
			Address: addr2,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "b.key")},
		},
		{
			Address: addr3,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "c.key")},
		},
		{
			Address: addr3,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "c2.key")},
		},
	}
	for _, a := range accs {
		cache.add(a)
	}

	nomatchAddr, err := common.HexToAddress("cb625337dc760bff449e5c187e080797283ad42e69bd")
	if err != nil {
		t.Error(err)
	}
	nomatchAccount := accounts.Account{
		Address: nomatchAddr,
		URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "something")},
	}
	tests := []struct {
		Query      accounts.Account
		WantResult accounts.Account
		WantError  error
	}{
		// by address
		{Query: accounts.Account{Address: accs[0].Address}, WantResult: accs[0]},
		// by file
		{Query: accounts.Account{URL: accs[0].URL}, WantResult: accs[0]},
		// by basename
		{Query: accounts.Account{URL: accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Base(accs[0].URL.Path)}}, WantResult: accs[0]},
		// by file and address
		{Query: accs[0], WantResult: accs[0]},
		// ambiguous address, tie resolved by file
		{Query: accs[2], WantResult: accs[2]},
		// ambiguous address error
		{
			Query: accounts.Account{Address: accs[2].Address},
			WantError: &AmbiguousAddrError{
				Addr:    accs[2].Address,
				Matches: []accounts.Account{accs[2], accs[3]},
			},
		},
		// no match error
		{Query: nomatchAccount, WantError: ErrNoMatch},
		{Query: accounts.Account{URL: nomatchAccount.URL}, WantError: ErrNoMatch},
		{Query: accounts.Account{URL: accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Base(nomatchAccount.URL.Path)}}, WantError: ErrNoMatch},
		{Query: accounts.Account{Address: nomatchAccount.Address}, WantError: ErrNoMatch},
	}
	for i, test := range tests {
		a, err := cache.find(test.Query)
		if !reflect.DeepEqual(err, test.WantError) {
			t.Errorf("test %d: error mismatch for query %v\ngot %q\nwant %q", i, test.Query, err, test.WantError)
			continue
		}
		if a != test.WantResult {
			t.Errorf("test %d: result mismatch for query %v\ngot %v\nwant %v", i, test.Query, a, test.WantResult)
			continue
		}
	}
}

func waitForAccounts(wantAccounts []accounts.Account, ks *KeyStore) error {
	var list []accounts.Account
	for d := 200 * time.Millisecond; d < 8*time.Second; d *= 2 {
		list = ks.Accounts()
		if reflect.DeepEqual(list, wantAccounts) {
			// ks should have also received change notifications
			select {
			case <-ks.changes:
			default:
				return fmt.Errorf("wasn't notified of new accounts")
			}
			return nil
		}
		time.Sleep(d)
	}
	return fmt.Errorf("\ngot  %v\nwant %v", list, wantAccounts)
}

// TestUpdatedKeyfileContents tests that updating the contents of a keystore file
// is noticed by the watcher, and the account cache is updated accordingly
func TestUpdatedKeyfileContents(t *testing.T) {
	t.Parallel()

	// Create a temporary kesytore to test with
	rand.Seed(time.Now().UnixNano())
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("xcc-keystore-watch-test-%d-%d", os.Getpid(), rand.Int()))
	ks := NewKeyStore(dir, LightScryptN, LightScryptP)

	list := ks.Accounts()
	if len(list) > 0 {
		t.Error("initial account list not empty:", list)
	}
	time.Sleep(100 * time.Millisecond)

	// Create the directory and copy a key file into it.
	os.MkdirAll(dir, 0700)
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "aaa")

	// Place one of our testfiles in there
	if err := cp.CopyFile(file, cachetestAccounts[0].URL.Path); err != nil {
		t.Fatal(err)
	}

	// ks should see the account.
	wantAccounts := []accounts.Account{cachetestAccounts[0]}
	wantAccounts[0].URL = accounts.URL{Scheme: KeyStoreScheme, Path: file}
	if err := waitForAccounts(wantAccounts, ks); err != nil {
		t.Error(err)
		return
	}

	// needed so that modTime of `file` is different to its current value after forceCopyFile
	time.Sleep(1000 * time.Millisecond)

	// Now replace file contents
	if err := forceCopyFile(file, cachetestAccounts[1].URL.Path); err != nil {
		t.Fatal(err)
		return
	}
	wantAccounts = []accounts.Account{cachetestAccounts[1]}
	wantAccounts[0].URL = accounts.URL{Scheme: KeyStoreScheme, Path: file}
	if err := waitForAccounts(wantAccounts, ks); err != nil {
		t.Errorf("First replacement failed")
		t.Error(err)
		return
	}

	// needed so that modTime of `file` is different to its current value after forceCopyFile
	time.Sleep(1000 * time.Millisecond)

	// Now replace file contents again
	if err := forceCopyFile(file, cachetestAccounts[2].URL.Path); err != nil {
		t.Fatal(err)
		return
	}
	wantAccounts = []accounts.Account{cachetestAccounts[2]}
	wantAccounts[0].URL = accounts.URL{Scheme: KeyStoreScheme, Path: file}
	if err := waitForAccounts(wantAccounts, ks); err != nil {
		t.Errorf("Second replacement failed")
		t.Error(err)
		return
	}

	// needed so that modTime of `file` is different to its current value after ioutil.WriteFile
	time.Sleep(1000 * time.Millisecond)

	// Now replace file contents with crap
	if err := ioutil.WriteFile(file, []byte("foo"), 0644); err != nil {
		t.Fatal(err)
		return
	}
	if err := waitForAccounts([]accounts.Account{}, ks); err != nil {
		t.Errorf("Emptying account file failed")
		t.Error(err)
		return
	}
}

// forceCopyFile is like cp.CopyFile, but doesn't complain if the destination exists.
func forceCopyFile(dst, src string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}
