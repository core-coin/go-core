// Copyright 2024 by the Authors
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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
)

// ---------------------------------------------------------------------------
// Known test vector
// ---------------------------------------------------------------------------

// Known test vector: "abandon...about" is the canonical BIP-39 test mnemonic.
// We hardcode the expected private key hex to detect any regression in the
// derivation pipeline (entropy → PBKDF2 seed → first 57 bytes → Ed448 key).
const (
	knownMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	// Expected: first 57 bytes of bip39.NewSeed(knownMnemonic, "")
	knownPrivateKeyHex = "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d"
)

// parseJSON unmarshals JSON and fails the test on error,
// preventing silent zero-value comparisons.
func parseJSON(t *testing.T, data string, v interface{}) {
	t.Helper()
	if err := json.Unmarshal([]byte(data), v); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, data)
	}
}

// ---------------------------------------------------------------------------
// Unit tests for wordsToEntropyBits
// ---------------------------------------------------------------------------

func TestWordsToEntropyBits(t *testing.T) {
	tests := []struct {
		words   int
		want    int
		wantErr bool
	}{
		{12, 128, false},
		{15, 160, false},
		{18, 192, false},
		{21, 224, false},
		{24, 256, false},
		{0, 0, true},
		{1, 0, true},
		{11, 0, true},
		{13, 0, true},
		{25, 0, true},
		{-1, 0, true},
		{100, 0, true},
	}
	for _, tt := range tests {
		bits, err := wordsToEntropyBits(tt.words)
		if (err != nil) != tt.wantErr {
			t.Errorf("wordsToEntropyBits(%d): error = %v, wantErr %v", tt.words, err, tt.wantErr)
			continue
		}
		if bits != tt.want {
			t.Errorf("wordsToEntropyBits(%d) = %d, want %d", tt.words, bits, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Unit tests for deriveKeyFromMnemonic
// ---------------------------------------------------------------------------

func TestDeriveKeyFromMnemonic_KnownVector(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	key, err := deriveKeyFromMnemonic(knownMnemonic, 0)
	if err != nil {
		t.Fatalf("derivation failed: %v", err)
	}

	got := hex.EncodeToString(crypto.MarshalPrivateKey(key))
	if got != knownPrivateKeyHex {
		t.Errorf("known vector mismatch:\n  got:  %s\n  want: %s", got, knownPrivateKeyHex)
	}
}

func TestDeriveKeyFromMnemonic_Deterministic(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	key1, err := deriveKeyFromMnemonic(knownMnemonic, 0)
	if err != nil {
		t.Fatalf("first derivation failed: %v", err)
	}
	key2, err := deriveKeyFromMnemonic(knownMnemonic, 0)
	if err != nil {
		t.Fatalf("second derivation failed: %v", err)
	}

	hex1 := hex.EncodeToString(crypto.MarshalPrivateKey(key1))
	hex2 := hex.EncodeToString(crypto.MarshalPrivateKey(key2))
	if hex1 != hex2 {
		t.Errorf("same mnemonic produced different keys:\n  %s\n  %s", hex1, hex2)
	}
}

func TestDeriveKeyFromMnemonic_KeyLength(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	key, err := deriveKeyFromMnemonic(knownMnemonic, 0)
	if err != nil {
		t.Fatalf("derivation failed: %v", err)
	}

	keyBytes := crypto.MarshalPrivateKey(key)
	if len(keyBytes) != crypto.PrivkeyLength {
		t.Errorf("private key length = %d bytes, want %d (Ed448)", len(keyBytes), crypto.PrivkeyLength)
	}

	pubBytes := key.PublicKey()
	if len(pubBytes) != crypto.PubkeyLength {
		t.Errorf("public key length = %d bytes, want %d (Ed448)", len(pubBytes), crypto.PubkeyLength)
	}
}

func TestDeriveKeyFromMnemonic_DifferentMnemonics(t *testing.T) {
	bip39.SetWordList(wordlists.English)
	m1 := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	m2 := "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong"

	key1, err := deriveKeyFromMnemonic(m1, 0)
	if err != nil {
		t.Fatalf("first derivation failed: %v", err)
	}
	key2, err := deriveKeyFromMnemonic(m2, 0)
	if err != nil {
		t.Fatalf("second derivation failed: %v", err)
	}

	hex1 := hex.EncodeToString(crypto.MarshalPrivateKey(key1))
	hex2 := hex.EncodeToString(crypto.MarshalPrivateKey(key2))
	if hex1 == hex2 {
		t.Error("different mnemonics produced the same key — derivation is broken")
	}
}

// ---------------------------------------------------------------------------
// Unit tests for setWordList
// ---------------------------------------------------------------------------

func TestSetWordList_ValidLanguages(t *testing.T) {
	languages := []string{
		"english", "japanese", "korean", "spanish",
		"french", "italian", "chinese_simplified", "chinese_traditional",
	}
	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			if err := setWordList(lang); err != nil {
				t.Errorf("setWordList(%q) returned error: %v", lang, err)
			}
		})
	}
	// Reset to English for subsequent tests.
	bip39.SetWordList(wordlists.English)
}

func TestSetWordList_InvalidLanguages(t *testing.T) {
	invalid := []string{"", "german", "russian", "ENGLISH", "English", "en"}
	for _, lang := range invalid {
		t.Run(lang, func(t *testing.T) {
			if err := setWordList(lang); err == nil {
				t.Errorf("setWordList(%q) should have returned error", lang)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unit tests for networkName
// ---------------------------------------------------------------------------

func TestNetworkName(t *testing.T) {
	original := common.DefaultNetworkID
	defer func() { common.DefaultNetworkID = original }()

	common.DefaultNetworkID = common.Mainnet
	if got := networkName(); got != "mainnet" {
		t.Errorf("networkName() for Mainnet = %q, want %q", got, "mainnet")
	}

	common.DefaultNetworkID = common.Devin
	if got := networkName(); got != "devin" {
		t.Errorf("networkName() for Devin = %q, want %q", got, "devin")
	}

	common.DefaultNetworkID = 99
	got := networkName()
	if !strings.Contains(got, "private") {
		t.Errorf("networkName() for private network = %q, want containing 'private'", got)
	}
}

// ---------------------------------------------------------------------------
// Unit tests for network-dependent address derivation
// ---------------------------------------------------------------------------

func TestAddressDependsOnNetwork(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	original := common.DefaultNetworkID
	defer func() { common.DefaultNetworkID = original }()

	// Derive key with mainnet.
	common.DefaultNetworkID = common.Mainnet
	keyMainnet, err := deriveKeyFromMnemonic(knownMnemonic, 0)
	if err != nil {
		t.Fatalf("mainnet derivation failed: %v", err)
	}
	addrMainnet := keyMainnet.Address().Hex()

	// Derive key with devin.
	common.DefaultNetworkID = common.Devin
	keyDevin, err := deriveKeyFromMnemonic(knownMnemonic, 0)
	if err != nil {
		t.Fatalf("devin derivation failed: %v", err)
	}
	addrDevin := keyDevin.Address().Hex()

	// Private keys must be identical (network only affects address encoding).
	hexMainnet := hex.EncodeToString(crypto.MarshalPrivateKey(keyMainnet))
	hexDevin := hex.EncodeToString(crypto.MarshalPrivateKey(keyDevin))
	if hexMainnet != hexDevin {
		t.Errorf("same mnemonic produced different private keys on different networks:\n  mainnet: %s\n  devin:   %s", hexMainnet, hexDevin)
	}

	// Addresses must differ (different prefix).
	if addrMainnet == addrDevin {
		t.Errorf("mainnet and devin addresses should differ, both are: %s", addrMainnet)
	}

	// Check ICAN prefixes.
	if !strings.HasPrefix(addrMainnet, "cb") {
		t.Errorf("mainnet address should start with 'cb', got %q", addrMainnet)
	}
	if !strings.HasPrefix(addrDevin, "ab") {
		t.Errorf("devin address should start with 'ab', got %q", addrDevin)
	}

	// Check address length (22 bytes = 44 hex chars).
	if len(addrMainnet) != common.AddressLength*2 {
		t.Errorf("mainnet address length = %d chars, want %d", len(addrMainnet), common.AddressLength*2)
	}
}

// ---------------------------------------------------------------------------
// Unit tests for mnemonic word counts
// ---------------------------------------------------------------------------

func TestGenerateMnemonicWordCounts(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	tests := []struct {
		words       int
		entropyBits int
	}{
		{12, 128},
		{15, 160},
		{18, 192},
		{21, 224},
		{24, 256},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d_words", tt.words), func(t *testing.T) {
			entropy, err := bip39.NewEntropy(tt.entropyBits)
			if err != nil {
				t.Fatalf("NewEntropy(%d) failed: %v", tt.entropyBits, err)
			}
			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				t.Fatalf("NewMnemonic failed: %v", err)
			}

			wordCount := len(strings.Fields(mnemonic))
			if wordCount != tt.words {
				t.Errorf("expected %d words, got %d: %s", tt.words, wordCount, mnemonic)
			}

			if !bip39.IsMnemonicValid(mnemonic) {
				t.Errorf("generated mnemonic is not valid BIP-39: %s", mnemonic)
			}

			// Verify key derivation works and produces correct length.
			key, err := deriveKeyFromMnemonic(mnemonic, 0)
			if err != nil {
				t.Fatalf("deriveKeyFromMnemonic failed: %v", err)
			}
			keyBytes := crypto.MarshalPrivateKey(key)
			if len(keyBytes) != crypto.PrivkeyLength {
				t.Errorf("derived key length = %d, want %d", len(keyBytes), crypto.PrivkeyLength)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unit tests for language-specific mnemonic generation and recovery
// ---------------------------------------------------------------------------

func TestLanguageMnemonicRoundTrip(t *testing.T) {
	original := common.DefaultNetworkID
	defer func() { common.DefaultNetworkID = original }()
	common.DefaultNetworkID = common.Mainnet

	languages := []struct {
		name string
		lang string
	}{
		{"english", "english"},
		{"japanese", "japanese"},
		{"korean", "korean"},
		{"spanish", "spanish"},
		{"french", "french"},
		{"italian", "italian"},
		{"chinese_simplified", "chinese_simplified"},
		{"chinese_traditional", "chinese_traditional"},
	}

	for _, lang := range languages {
		t.Run(lang.name, func(t *testing.T) {
			if err := setWordList(lang.lang); err != nil {
				t.Fatalf("setWordList(%s) failed: %v", lang.lang, err)
			}

			// Generate mnemonic.
			entropy, err := bip39.NewEntropy(256)
			if err != nil {
				t.Fatalf("NewEntropy failed: %v", err)
			}
			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				t.Fatalf("NewMnemonic failed: %v", err)
			}

			// Validate mnemonic with the same wordlist.
			if !bip39.IsMnemonicValid(mnemonic) {
				t.Fatalf("generated %s mnemonic is not valid BIP-39", lang.name)
			}

			// Verify word count is 24 for 256-bit entropy.
			words := strings.Fields(mnemonic)
			if len(words) != 24 {
				t.Errorf("expected 24 words for 256-bit entropy, got %d", len(words))
			}

			// Derive key twice — must be deterministic.
			key1, err := deriveKeyFromMnemonic(mnemonic, 0)
			if err != nil {
				t.Fatalf("first derivation failed: %v", err)
			}
			key2, err := deriveKeyFromMnemonic(mnemonic, 0)
			if err != nil {
				t.Fatalf("second derivation failed: %v", err)
			}

			hex1 := hex.EncodeToString(crypto.MarshalPrivateKey(key1))
			hex2 := hex.EncodeToString(crypto.MarshalPrivateKey(key2))
			if hex1 != hex2 {
				t.Errorf("non-deterministic derivation for %s:\n  first:  %s\n  second: %s", lang.name, hex1, hex2)
			}

			// Verify derived key is valid length.
			if len(crypto.MarshalPrivateKey(key1)) != crypto.PrivkeyLength {
				t.Errorf("key length = %d, want %d", len(crypto.MarshalPrivateKey(key1)), crypto.PrivkeyLength)
			}
		})
	}

	// Reset to English.
	bip39.SetWordList(wordlists.English)
}

// ---------------------------------------------------------------------------
// Integration tests: generate-mnemonic command
// ---------------------------------------------------------------------------

func TestGenerateMnemonicCommand_Default(t *testing.T) {
	cmd := runXcbkey(t, "generate-mnemonic", "--json")
	var out outputMnemonic
	_, matches := cmd.ExpectRegexp(`\{[^}]+\}`)
	cmd.ExpectExit()
	parseJSON(t, matches[0], &out)

	// Mnemonic must be 24 words and valid BIP-39.
	words := strings.Fields(out.Mnemonic)
	if len(words) != 24 {
		t.Errorf("expected 24-word mnemonic, got %d words", len(words))
	}

	bip39.SetWordList(wordlists.English)
	if !bip39.IsMnemonicValid(out.Mnemonic) {
		t.Errorf("generated mnemonic is not valid BIP-39: %s", out.Mnemonic)
	}

	// Private key: 57 bytes = 114 hex chars.
	if len(out.PrivateKey) != crypto.PrivkeyLength*2 {
		t.Errorf("private key hex length = %d, want %d", len(out.PrivateKey), crypto.PrivkeyLength*2)
	}

	// Public key: 57 bytes = 114 hex chars.
	if len(out.PublicKey) != crypto.PubkeyLength*2 {
		t.Errorf("public key hex length = %d, want %d", len(out.PublicKey), crypto.PubkeyLength*2)
	}

	// Address: 22 bytes = 44 hex chars, mainnet "cb" prefix.
	if len(out.Address) != common.AddressLength*2 {
		t.Errorf("address length = %d chars, want %d", len(out.Address), common.AddressLength*2)
	}
	if !strings.HasPrefix(out.Address, "cb") {
		t.Errorf("mainnet address should start with 'cb', got %q", out.Address)
	}

	if out.Network != "mainnet" {
		t.Errorf("network = %q, want 'mainnet'", out.Network)
	}

	// Verify the key hex is actually valid hex.
	if _, err := hex.DecodeString(out.PrivateKey); err != nil {
		t.Errorf("private key is not valid hex: %v", err)
	}
	if _, err := hex.DecodeString(out.PublicKey); err != nil {
		t.Errorf("public key is not valid hex: %v", err)
	}
}

func TestGenerateMnemonicCommand_12Words(t *testing.T) {
	cmd := runXcbkey(t, "generate-mnemonic", "--json", "--words", "12")
	var out outputMnemonic
	_, matches := cmd.ExpectRegexp(`\{[^}]+\}`)
	cmd.ExpectExit()
	parseJSON(t, matches[0], &out)

	if wordCount := len(strings.Fields(out.Mnemonic)); wordCount != 12 {
		t.Errorf("expected 12-word mnemonic, got %d words", wordCount)
	}

	bip39.SetWordList(wordlists.English)
	if !bip39.IsMnemonicValid(out.Mnemonic) {
		t.Errorf("12-word mnemonic is not valid BIP-39: %s", out.Mnemonic)
	}
}

func TestGenerateMnemonicCommand_DevinNetwork(t *testing.T) {
	cmd := runXcbkey(t, "generate-mnemonic", "--json", "--networkid", "3")
	var out outputMnemonic
	_, matches := cmd.ExpectRegexp(`\{[^}]+\}`)
	cmd.ExpectExit()
	parseJSON(t, matches[0], &out)

	if !strings.HasPrefix(out.Address, "ab") {
		t.Errorf("devin address should start with 'ab', got %q", out.Address)
	}
	if out.Network != "devin" {
		t.Errorf("network = %q, want 'devin'", out.Network)
	}
}

func TestGenerateMnemonicCommand_PrivateNetwork(t *testing.T) {
	cmd := runXcbkey(t, "generate-mnemonic", "--json", "--networkid", "99")
	var out outputMnemonic
	_, matches := cmd.ExpectRegexp(`\{[^}]+\}`)
	cmd.ExpectExit()
	parseJSON(t, matches[0], &out)

	// Private networks (ID > 3) use "ce" prefix.
	if !strings.HasPrefix(out.Address, "ce") {
		t.Errorf("private network address should start with 'ce', got %q", out.Address)
	}
	if !strings.Contains(out.Network, "private") {
		t.Errorf("network = %q, want containing 'private'", out.Network)
	}
}

func TestGenerateMnemonicCommand_AllWordCounts(t *testing.T) {
	wordCounts := []string{"12", "15", "18", "21", "24"}
	expected := []int{12, 15, 18, 21, 24}

	for i, wc := range wordCounts {
		t.Run(wc+"_words", func(t *testing.T) {
			cmd := runXcbkey(t, "generate-mnemonic", "--json", "--words", wc)
			var out outputMnemonic
			_, matches := cmd.ExpectRegexp(`\{[^}]+\}`)
			cmd.ExpectExit()
			parseJSON(t, matches[0], &out)

			if wordCount := len(strings.Fields(out.Mnemonic)); wordCount != expected[i] {
				t.Errorf("expected %d-word mnemonic, got %d", expected[i], wordCount)
			}
		})
	}
}

func TestGenerateMnemonicCommand_HumanReadable(t *testing.T) {
	cmd := runXcbkey(t, "generate-mnemonic")
	cmd.ExpectRegexp(`Network:.*mainnet[\s\S]*Private key:\s*[0-9a-f]+[\s\S]*Address:\s*cb[0-9a-f]+`)
	cmd.ExpectExit()
}

func TestGenerateMnemonicCommand_Uniqueness(t *testing.T) {
	cmd1 := runXcbkey(t, "generate-mnemonic", "--json")
	var out1 outputMnemonic
	_, matches1 := cmd1.ExpectRegexp(`\{[^}]+\}`)
	cmd1.ExpectExit()
	parseJSON(t, matches1[0], &out1)

	cmd2 := runXcbkey(t, "generate-mnemonic", "--json")
	var out2 outputMnemonic
	_, matches2 := cmd2.ExpectRegexp(`\{[^}]+\}`)
	cmd2.ExpectExit()
	parseJSON(t, matches2[0], &out2)

	if out1.Mnemonic == out2.Mnemonic {
		t.Error("two generates produced the same mnemonic — entropy source may be broken")
	}
	if out1.PrivateKey == out2.PrivateKey {
		t.Error("two generates produced the same private key")
	}
	if out1.Address == out2.Address {
		t.Error("two generates produced the same address")
	}
}

// ---------------------------------------------------------------------------
// Integration tests: mnemonic-to-key command
// ---------------------------------------------------------------------------

func TestMnemonicToKey_RoundTrip(t *testing.T) {
	// Generate a wallet.
	genCmd := runXcbkey(t, "generate-mnemonic", "--json")
	var genOut outputMnemonic
	_, matches := genCmd.ExpectRegexp(`\{[^}]+\}`)
	genCmd.ExpectExit()
	parseJSON(t, matches[0], &genOut)

	// Recover from mnemonic.
	words := strings.Fields(genOut.Mnemonic)
	args := append([]string{"mnemonic-to-key", "--json"}, words...)
	recoverCmd := runXcbkey(t, args...)
	var recoverOut outputMnemonic
	_, matches = recoverCmd.ExpectRegexp(`\{[^}]+\}`)
	recoverCmd.ExpectExit()
	parseJSON(t, matches[0], &recoverOut)

	if genOut.PrivateKey != recoverOut.PrivateKey {
		t.Errorf("private key mismatch:\n  generate:       %s\n  mnemonic-to-key: %s", genOut.PrivateKey, recoverOut.PrivateKey)
	}
	if genOut.PublicKey != recoverOut.PublicKey {
		t.Errorf("public key mismatch:\n  generate:       %s\n  mnemonic-to-key: %s", genOut.PublicKey, recoverOut.PublicKey)
	}
	if genOut.Address != recoverOut.Address {
		t.Errorf("address mismatch:\n  generate:       %s\n  mnemonic-to-key: %s", genOut.Address, recoverOut.Address)
	}
}

func TestMnemonicToKey_KnownVector(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	words := strings.Fields(knownMnemonic)
	args := append([]string{"mnemonic-to-key", "--json"}, words...)
	cmd := runXcbkey(t, args...)
	var out outputMnemonic
	_, matches := cmd.ExpectRegexp(`\{[^}]+\}`)
	cmd.ExpectExit()
	parseJSON(t, matches[0], &out)

	// Compare CLI output against the hardcoded known vector.
	cliPK := strings.ToLower(strings.TrimSpace(out.PrivateKey))
	if cliPK != knownPrivateKeyHex {
		t.Errorf("CLI output does not match known vector:\n  got:  %s\n  want: %s", cliPK, knownPrivateKeyHex)
	}

	if len(out.PrivateKey) != crypto.PrivkeyLength*2 {
		t.Errorf("private key hex length = %d, want %d", len(out.PrivateKey), crypto.PrivkeyLength*2)
	}
}

func TestMnemonicToKey_DevinNetwork(t *testing.T) {
	bip39.SetWordList(wordlists.English)
	words := strings.Fields(knownMnemonic)

	// Mainnet.
	argsMainnet := append([]string{"mnemonic-to-key", "--json"}, words...)
	cmdMainnet := runXcbkey(t, argsMainnet...)
	var outMainnet outputMnemonic
	_, matches := cmdMainnet.ExpectRegexp(`\{[^}]+\}`)
	cmdMainnet.ExpectExit()
	parseJSON(t, matches[0], &outMainnet)

	// Devin.
	argsDevin := append([]string{"mnemonic-to-key", "--json", "--networkid", "3"}, words...)
	cmdDevin := runXcbkey(t, argsDevin...)
	var outDevin outputMnemonic
	_, matches = cmdDevin.ExpectRegexp(`\{[^}]+\}`)
	cmdDevin.ExpectExit()
	parseJSON(t, matches[0], &outDevin)

	// Private key must be the same.
	if outMainnet.PrivateKey != outDevin.PrivateKey {
		t.Error("same mnemonic produced different private keys on different networks")
	}

	// Addresses must differ.
	if outMainnet.Address == outDevin.Address {
		t.Error("mainnet and devin addresses should differ")
	}

	if !strings.HasPrefix(outMainnet.Address, "cb") {
		t.Errorf("mainnet address should start with 'cb', got %q", outMainnet.Address)
	}
	if !strings.HasPrefix(outDevin.Address, "ab") {
		t.Errorf("devin address should start with 'ab', got %q", outDevin.Address)
	}
}

func TestMnemonicToKey_HumanReadable(t *testing.T) {
	cmd := runXcbkey(t, "mnemonic-to-key",
		"abandon", "abandon", "abandon", "abandon", "abandon", "abandon",
		"abandon", "abandon", "abandon", "abandon", "abandon", "about")
	cmd.ExpectRegexp(`Network:.*mainnet[\s\S]*Private key:\s*[0-9a-f]+[\s\S]*Address:\s*cb[0-9a-f]+`)
	cmd.ExpectExit()
}

func TestMnemonicToKey_AllWordCounts(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	wordCounts := []int{128, 160, 192, 224, 256}
	expectedWords := []int{12, 15, 18, 21, 24}

	for i, bits := range wordCounts {
		t.Run(fmt.Sprintf("%d_words", expectedWords[i]), func(t *testing.T) {
			entropy, err := bip39.NewEntropy(bits)
			if err != nil {
				t.Fatalf("NewEntropy(%d) failed: %v", bits, err)
			}
			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				t.Fatalf("NewMnemonic failed: %v", err)
			}

			words := strings.Fields(mnemonic)
			if len(words) != expectedWords[i] {
				t.Fatalf("expected %d words from %d-bit entropy, got %d", expectedWords[i], bits, len(words))
			}

			args := append([]string{"mnemonic-to-key", "--json"}, words...)
			cmd := runXcbkey(t, args...)
			var out outputMnemonic
			_, matches := cmd.ExpectRegexp(`\{[^}]+\}`)
			cmd.ExpectExit()
			parseJSON(t, matches[0], &out)

			if len(out.PrivateKey) != crypto.PrivkeyLength*2 {
				t.Errorf("private key hex length = %d, want %d", len(out.PrivateKey), crypto.PrivkeyLength*2)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unit: same entropy produces different mnemonics in different languages
// ---------------------------------------------------------------------------

func TestSameEntropyDifferentLanguages(t *testing.T) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		t.Fatalf("NewEntropy failed: %v", err)
	}

	bip39.SetWordList(wordlists.English)
	mnemonicEn, err := bip39.NewMnemonic(entropy)
	if err != nil {
		t.Fatalf("English NewMnemonic failed: %v", err)
	}

	bip39.SetWordList(wordlists.Japanese)
	mnemonicJa, err := bip39.NewMnemonic(entropy)
	if err != nil {
		t.Fatalf("Japanese NewMnemonic failed: %v", err)
	}

	// Mnemonics from the same entropy but different wordlists must differ.
	if mnemonicEn == mnemonicJa {
		t.Fatal("English and Japanese mnemonics from same entropy should differ")
	}

	// Derived keys must differ because PBKDF2 input strings differ.
	bip39.SetWordList(wordlists.English)
	keyEn, err := deriveKeyFromMnemonic(mnemonicEn, 0)
	if err != nil {
		t.Fatalf("English derivation failed: %v", err)
	}

	bip39.SetWordList(wordlists.Japanese)
	keyJa, err := deriveKeyFromMnemonic(mnemonicJa, 0)
	if err != nil {
		t.Fatalf("Japanese derivation failed: %v", err)
	}

	hexEn := fmt.Sprintf("%x", crypto.MarshalPrivateKey(keyEn))
	hexJa := fmt.Sprintf("%x", crypto.MarshalPrivateKey(keyJa))
	if hexEn == hexJa {
		t.Error("different mnemonic strings (EN vs JA) produced the same key — PBKDF2 derivation may be ignoring input")
	}

	// Reset.
	bip39.SetWordList(wordlists.English)
}

// ---------------------------------------------------------------------------
// Integration tests: verify-mnemonic command
// ---------------------------------------------------------------------------

func TestVerifyMnemonic_MatchingKeyAndMnemonic(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	// Generate a wallet first.
	genCmd := runXcbkey(t, "generate-mnemonic", "--json")
	var genOut outputMnemonic
	_, matches := genCmd.ExpectRegexp(`\{[^}]+\}`)
	genCmd.ExpectExit()
	parseJSON(t, matches[0], &genOut)

	// Verify match.
	words := strings.Fields(genOut.Mnemonic)
	args := append([]string{"verify-mnemonic", "--json", "--privatekey", genOut.PrivateKey}, words...)
	verifyCmd := runXcbkey(t, args...)

	type verifyOutput struct {
		Match      bool   `json:"match"`
		Mnemonic   string `json:"mnemonic"`
		DerivedKey string `json:"derived_key"`
		GivenKey   string `json:"given_key"`
		Address    string `json:"address"`
		Network    string `json:"network"`
	}
	var vOut verifyOutput
	_, matches = verifyCmd.ExpectRegexp(`\{[^}]+\}`)
	verifyCmd.ExpectExit()
	parseJSON(t, matches[0], &vOut)

	if !vOut.Match {
		t.Error("expected match=true for correct mnemonic+key pair")
	}
	if vOut.DerivedKey != vOut.GivenKey {
		t.Error("derived_key and given_key should be equal")
	}
}

func TestVerifyMnemonic_MismatchingKey(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	// Generate two wallets.
	genCmd1 := runXcbkey(t, "generate-mnemonic", "--json")
	var out1 outputMnemonic
	_, matches := genCmd1.ExpectRegexp(`\{[^}]+\}`)
	genCmd1.ExpectExit()
	parseJSON(t, matches[0], &out1)

	genCmd2 := runXcbkey(t, "generate-mnemonic", "--json")
	var out2 outputMnemonic
	_, matches = genCmd2.ExpectRegexp(`\{[^}]+\}`)
	genCmd2.ExpectExit()
	parseJSON(t, matches[0], &out2)

	// Verify mismatch: mnemonic from wallet1, key from wallet2.
	words := strings.Fields(out1.Mnemonic)
	args := append([]string{"verify-mnemonic", "--json", "--privatekey", out2.PrivateKey}, words...)
	verifyCmd := runXcbkey(t, args...)

	type verifyOutput struct {
		Match      bool   `json:"match"`
		DerivedKey string `json:"derived_key"`
		GivenKey   string `json:"given_key"`
	}
	var vOut verifyOutput
	_, matches = verifyCmd.ExpectRegexp(`\{[^}]+\}`)
	verifyCmd.ExpectExit()
	parseJSON(t, matches[0], &vOut)

	if vOut.Match {
		t.Error("expected match=false for mismatched mnemonic+key pair")
	}
	if vOut.DerivedKey == vOut.GivenKey {
		t.Error("derived and given keys should differ for mismatched pair")
	}
}

func TestVerifyMnemonic_WithKeyfile(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	key, err := deriveKeyFromMnemonic(knownMnemonic, 0)
	if err != nil {
		t.Fatalf("derivation failed: %v", err)
	}

	// Write key to temp file.
	tmpdir, err := ioutil.TempDir("", "xcbkey-test")
	if err != nil {
		t.Fatal("Can't create temp dir:", err)
	}
	defer os.RemoveAll(tmpdir)

	keyfile := filepath.Join(tmpdir, "testkey")
	if err := crypto.SaveEDDSA(keyfile, key); err != nil {
		t.Fatalf("failed to save key: %v", err)
	}

	// Verify using --keyfile.
	words := strings.Fields(knownMnemonic)
	args := append([]string{"verify-mnemonic", "--json", "--keyfile", keyfile}, words...)
	verifyCmd := runXcbkey(t, args...)

	type verifyOutput struct {
		Match bool `json:"match"`
	}
	var vOut verifyOutput
	_, matches := verifyCmd.ExpectRegexp(`\{[^}]+\}`)
	verifyCmd.ExpectExit()
	parseJSON(t, matches[0], &vOut)

	if !vOut.Match {
		t.Error("expected match=true when using --keyfile with correct key")
	}
}

func TestVerifyMnemonic_HumanReadable_Match(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	key, _ := deriveKeyFromMnemonic(knownMnemonic, 0)
	pkHex := hex.EncodeToString(crypto.MarshalPrivateKey(key))

	words := strings.Fields(knownMnemonic)
	args := append([]string{"verify-mnemonic", "--privatekey", pkHex}, words...)
	cmd := runXcbkey(t, args...)
	cmd.ExpectRegexp(`MATCH:.*[\s\S]*Address:\s*cb[0-9a-f]+`)
	cmd.ExpectExit()
}

func TestVerifyMnemonic_HumanReadable_Mismatch(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	fakeKey := strings.Repeat("ab", crypto.PrivkeyLength) // 57 bytes of 0xab

	words := strings.Fields(knownMnemonic)
	args := append([]string{"verify-mnemonic", "--privatekey", fakeKey}, words...)
	cmd := runXcbkey(t, args...)
	cmd.ExpectRegexp(`MISMATCH:.*[\s\S]*Derived key:.*[\s\S]*Given key:`)
	cmd.ExpectExit()
}

// ---------------------------------------------------------------------------
// Full end-to-end: generate-mnemonic -> mnemonic-to-key -> verify-mnemonic
// ---------------------------------------------------------------------------

func TestMnemonicFullRoundTrip(t *testing.T) {
	// Step 1: Generate.
	genCmd := runXcbkey(t, "generate-mnemonic", "--json")
	var genOut outputMnemonic
	_, matches := genCmd.ExpectRegexp(`\{[^}]+\}`)
	genCmd.ExpectExit()
	parseJSON(t, matches[0], &genOut)

	// Step 2: Recover from mnemonic.
	words := strings.Fields(genOut.Mnemonic)
	recoverArgs := append([]string{"mnemonic-to-key", "--json"}, words...)
	recoverCmd := runXcbkey(t, recoverArgs...)
	var recoverOut outputMnemonic
	_, matches = recoverCmd.ExpectRegexp(`\{[^}]+\}`)
	recoverCmd.ExpectExit()
	parseJSON(t, matches[0], &recoverOut)

	// Step 3: Verify.
	verifyArgs := append([]string{"verify-mnemonic", "--json", "--privatekey", genOut.PrivateKey}, words...)
	verifyCmd := runXcbkey(t, verifyArgs...)
	type verifyOutput struct {
		Match bool `json:"match"`
	}
	var vOut verifyOutput
	_, matches = verifyCmd.ExpectRegexp(`\{[^}]+\}`)
	verifyCmd.ExpectExit()
	parseJSON(t, matches[0], &vOut)

	if genOut.PrivateKey != recoverOut.PrivateKey {
		t.Error("generate and mnemonic-to-key produced different private keys")
	}
	if genOut.PublicKey != recoverOut.PublicKey {
		t.Error("generate and mnemonic-to-key produced different public keys")
	}
	if genOut.Address != recoverOut.Address {
		t.Error("generate and mnemonic-to-key produced different addresses")
	}
	if !vOut.Match {
		t.Error("verify-mnemonic verification failed")
	}
}

func TestMnemonicFullRoundTrip_DevinNetwork(t *testing.T) {
	// Generate on devin.
	genCmd := runXcbkey(t, "generate-mnemonic", "--json", "--networkid", "3")
	var genOut outputMnemonic
	_, matches := genCmd.ExpectRegexp(`\{[^}]+\}`)
	genCmd.ExpectExit()
	parseJSON(t, matches[0], &genOut)

	if !strings.HasPrefix(genOut.Address, "ab") {
		t.Fatalf("expected devin address prefix 'ab', got %q", genOut.Address)
	}

	// Recover on devin.
	words := strings.Fields(genOut.Mnemonic)
	recoverArgs := append([]string{"mnemonic-to-key", "--json", "--networkid", "3"}, words...)
	recoverCmd := runXcbkey(t, recoverArgs...)
	var recoverOut outputMnemonic
	_, matches = recoverCmd.ExpectRegexp(`\{[^}]+\}`)
	recoverCmd.ExpectExit()
	parseJSON(t, matches[0], &recoverOut)

	if genOut.Address != recoverOut.Address {
		t.Errorf("devin round-trip address mismatch:\n  generate:       %s\n  mnemonic-to-key: %s", genOut.Address, recoverOut.Address)
	}

	// Verify on devin.
	verifyArgs := append([]string{"verify-mnemonic", "--json", "--networkid", "3", "--privatekey", genOut.PrivateKey}, words...)
	verifyCmd := runXcbkey(t, verifyArgs...)
	type verifyOutput struct {
		Match   bool   `json:"match"`
		Network string `json:"network"`
	}
	var vOut verifyOutput
	_, matches = verifyCmd.ExpectRegexp(`\{[^}]+\}`)
	verifyCmd.ExpectExit()
	parseJSON(t, matches[0], &vOut)

	if !vOut.Match {
		t.Error("devin verify-mnemonic verification failed")
	}
	if vOut.Network != "devin" {
		t.Errorf("expected network 'devin', got %q", vOut.Network)
	}
}

func TestMnemonicFullRoundTrip_12Words(t *testing.T) {
	genCmd := runXcbkey(t, "generate-mnemonic", "--json", "--words", "12")
	var genOut outputMnemonic
	_, matches := genCmd.ExpectRegexp(`\{[^}]+\}`)
	genCmd.ExpectExit()
	parseJSON(t, matches[0], &genOut)

	if wordCount := len(strings.Fields(genOut.Mnemonic)); wordCount != 12 {
		t.Fatalf("expected 12 words, got %d", wordCount)
	}

	words := strings.Fields(genOut.Mnemonic)
	recoverArgs := append([]string{"mnemonic-to-key", "--json"}, words...)
	recoverCmd := runXcbkey(t, recoverArgs...)
	var recoverOut outputMnemonic
	_, matches = recoverCmd.ExpectRegexp(`\{[^}]+\}`)
	recoverCmd.ExpectExit()
	parseJSON(t, matches[0], &recoverOut)

	if genOut.PrivateKey != recoverOut.PrivateKey {
		t.Error("12-word round-trip failed: private keys differ")
	}
}

// ---------------------------------------------------------------------------
// Unit tests for multi-account derivation
// ---------------------------------------------------------------------------

func TestDeriveKeyFromMnemonic_Account0MatchesKnownVector(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	// Account 0 should produce the same key as the known vector (backwards compatible).
	key, err := deriveKeyFromMnemonic(knownMnemonic, 0)
	if err != nil {
		t.Fatalf("derivation failed: %v", err)
	}

	got := hex.EncodeToString(crypto.MarshalPrivateKey(key))
	if got != knownPrivateKeyHex {
		t.Errorf("account 0 does not match known vector:\n  got:  %s\n  want: %s", got, knownPrivateKeyHex)
	}
}

func TestDeriveKeyFromMnemonic_DifferentAccountsDifferentKeys(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	keys := make(map[string]uint32)
	for _, account := range []uint32{0, 1, 2, 5, 100} {
		key, err := deriveKeyFromMnemonic(knownMnemonic, account)
		if err != nil {
			t.Fatalf("derivation for account %d failed: %v", account, err)
		}
		h := hex.EncodeToString(crypto.MarshalPrivateKey(key))
		if prev, exists := keys[h]; exists {
			t.Errorf("account %d produced same key as account %d", account, prev)
		}
		keys[h] = account
	}
}

func TestDeriveKeyFromMnemonic_AccountDeterministic(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	for _, account := range []uint32{0, 1, 42} {
		key1, err := deriveKeyFromMnemonic(knownMnemonic, account)
		if err != nil {
			t.Fatalf("first derivation for account %d failed: %v", account, err)
		}
		key2, err := deriveKeyFromMnemonic(knownMnemonic, account)
		if err != nil {
			t.Fatalf("second derivation for account %d failed: %v", account, err)
		}

		hex1 := hex.EncodeToString(crypto.MarshalPrivateKey(key1))
		hex2 := hex.EncodeToString(crypto.MarshalPrivateKey(key2))
		if hex1 != hex2 {
			t.Errorf("account %d not deterministic:\n  first:  %s\n  second: %s", account, hex1, hex2)
		}
	}
}

func TestDeriveKeyFromMnemonic_AccountKeyLength(t *testing.T) {
	bip39.SetWordList(wordlists.English)

	for _, account := range []uint32{0, 1, 255} {
		key, err := deriveKeyFromMnemonic(knownMnemonic, account)
		if err != nil {
			t.Fatalf("derivation for account %d failed: %v", account, err)
		}
		keyBytes := crypto.MarshalPrivateKey(key)
		if len(keyBytes) != crypto.PrivkeyLength {
			t.Errorf("account %d: key length = %d, want %d", account, len(keyBytes), crypto.PrivkeyLength)
		}
		pubBytes := key.PublicKey()
		if len(pubBytes) != crypto.PubkeyLength {
			t.Errorf("account %d: pubkey length = %d, want %d", account, len(pubBytes), crypto.PubkeyLength)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration tests: multi-account CLI
// ---------------------------------------------------------------------------

func TestGenerateMnemonicCommand_AccountFlag(t *testing.T) {
	// Default (account 0).
	cmd0 := runXcbkey(t, "generate-mnemonic", "--json")
	var out0 outputMnemonic
	_, matches := cmd0.ExpectRegexp(`\{[^}]+\}`)
	cmd0.ExpectExit()
	parseJSON(t, matches[0], &out0)

	if out0.Account != 0 {
		t.Errorf("default account = %d, want 0", out0.Account)
	}

	// Explicit account 1.
	cmd1 := runXcbkey(t, "generate-mnemonic", "--json", "--account", "1")
	var out1 outputMnemonic
	_, matches = cmd1.ExpectRegexp(`\{[^}]+\}`)
	cmd1.ExpectExit()
	parseJSON(t, matches[0], &out1)

	if out1.Account != 1 {
		t.Errorf("explicit account = %d, want 1", out1.Account)
	}
}

func TestMnemonicToKey_AccountProducesDifferentKeys(t *testing.T) {
	bip39.SetWordList(wordlists.English)
	words := strings.Fields(knownMnemonic)

	// Account 0.
	args0 := append([]string{"mnemonic-to-key", "--json"}, words...)
	cmd0 := runXcbkey(t, args0...)
	var out0 outputMnemonic
	_, matches := cmd0.ExpectRegexp(`\{[^}]+\}`)
	cmd0.ExpectExit()
	parseJSON(t, matches[0], &out0)

	// Account 1.
	args1 := append([]string{"mnemonic-to-key", "--json", "--account", "1"}, words...)
	cmd1 := runXcbkey(t, args1...)
	var out1 outputMnemonic
	_, matches = cmd1.ExpectRegexp(`\{[^}]+\}`)
	cmd1.ExpectExit()
	parseJSON(t, matches[0], &out1)

	if out0.PrivateKey == out1.PrivateKey {
		t.Error("account 0 and account 1 produced the same private key")
	}
	if out0.Address == out1.Address {
		t.Error("account 0 and account 1 produced the same address")
	}
	if out0.Account != 0 || out1.Account != 1 {
		t.Errorf("account fields: got %d and %d, want 0 and 1", out0.Account, out1.Account)
	}
}

func TestMnemonicToKey_AccountRoundTrip(t *testing.T) {
	// Generate with account 2.
	genCmd := runXcbkey(t, "generate-mnemonic", "--json", "--account", "2")
	var genOut outputMnemonic
	_, matches := genCmd.ExpectRegexp(`\{[^}]+\}`)
	genCmd.ExpectExit()
	parseJSON(t, matches[0], &genOut)

	// Recover with account 2.
	words := strings.Fields(genOut.Mnemonic)
	recoverArgs := append([]string{"mnemonic-to-key", "--json", "--account", "2"}, words...)
	recoverCmd := runXcbkey(t, recoverArgs...)
	var recoverOut outputMnemonic
	_, matches = recoverCmd.ExpectRegexp(`\{[^}]+\}`)
	recoverCmd.ExpectExit()
	parseJSON(t, matches[0], &recoverOut)

	if genOut.PrivateKey != recoverOut.PrivateKey {
		t.Error("account 2 round-trip: private keys differ")
	}
	if genOut.Address != recoverOut.Address {
		t.Error("account 2 round-trip: addresses differ")
	}
}

func TestVerifyMnemonic_WithAccount(t *testing.T) {
	// Generate with account 3.
	genCmd := runXcbkey(t, "generate-mnemonic", "--json", "--account", "3")
	var genOut outputMnemonic
	_, matches := genCmd.ExpectRegexp(`\{[^}]+\}`)
	genCmd.ExpectExit()
	parseJSON(t, matches[0], &genOut)

	// Verify with correct account.
	words := strings.Fields(genOut.Mnemonic)
	verifyArgs := append([]string{"verify-mnemonic", "--json", "--account", "3", "--privatekey", genOut.PrivateKey}, words...)
	verifyCmd := runXcbkey(t, verifyArgs...)
	type verifyOutput struct {
		Match   bool   `json:"match"`
		Account uint32 `json:"account"`
	}
	var vOut verifyOutput
	_, matches = verifyCmd.ExpectRegexp(`\{[^}]+\}`)
	verifyCmd.ExpectExit()
	parseJSON(t, matches[0], &vOut)

	if !vOut.Match {
		t.Error("verify-mnemonic with correct account should match")
	}

	// Verify with wrong account (default 0) should mismatch.
	verifyArgs2 := append([]string{"verify-mnemonic", "--json", "--privatekey", genOut.PrivateKey}, words...)
	verifyCmd2 := runXcbkey(t, verifyArgs2...)
	var vOut2 verifyOutput
	_, matches = verifyCmd2.ExpectRegexp(`\{[^}]+\}`)
	verifyCmd2.ExpectExit()
	parseJSON(t, matches[0], &vOut2)

	if vOut2.Match {
		t.Error("verify-mnemonic with wrong account (0 vs 3) should not match")
	}
}
