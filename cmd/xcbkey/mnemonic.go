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
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"
	"gopkg.in/urfave/cli.v1"

	"github.com/core-coin/go-core/v2/cmd/utils"
	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
)

type outputMnemonic struct {
	Mnemonic   string `json:"mnemonic"`
	Account    uint32 `json:"account"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
	Network    string `json:"network"`
}

var accountFlag = cli.UintFlag{
	Name:  "account",
	Usage: "account index for HD derivation (default: 0)",
	Value: 0,
}

var commandGenerateMnemonic = cli.Command{
	Name:      "generate-mnemonic",
	Usage:     "generate a new wallet with a BIP-39 mnemonic",
	ArgsUsage: "",
	Description: `
Generate a new Core wallet with a BIP-39 mnemonic phrase.

Derivation: entropy -> mnemonic -> BIP-39 seed (PBKDF2-HMAC-SHA512) -> Ed448 key.

By default generates a 24-word mnemonic (256-bit entropy) in English.
Use --words 12 for a 12-word mnemonic (128-bit entropy).
Use --lang to choose a different BIP-39 wordlist language.

Use --account N to derive multiple wallets from the same mnemonic.
Account 0 is the default. Each account index produces a different key.

The --networkid flag controls the address prefix:
  --networkid 1   mainnet (cb prefix, default)
  --networkid 3   devin testnet (ab prefix)
  --networkid >3  private network (ce prefix)
The private key is the same regardless of network, but the address changes.

Store the mnemonic securely — it IS your wallet backup!
`,
	Flags: []cli.Flag{
		jsonFlag,
		langFlag,
		accountFlag,
		utils.NetworkIdFlag,
		cli.IntFlag{
			Name:  "words",
			Usage: "mnemonic length: 12, 15, 18, 21, or 24 (default: 24)",
			Value: 24,
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.IsSet(utils.NetworkIdFlag.Name) {
			common.DefaultNetworkID = common.NetworkID(ctx.Uint64(utils.NetworkIdFlag.Name))
		}

		if err := setWordList(ctx.String(langFlag.Name)); err != nil {
			utils.Fatalf("%v", err)
		}

		// Map word count to entropy bits.
		words := ctx.Int("words")
		entropyBits, err := wordsToEntropyBits(words)
		if err != nil {
			utils.Fatalf("%v", err)
		}

		// Generate entropy and mnemonic.
		entropy, err := bip39.NewEntropy(entropyBits)
		if err != nil {
			utils.Fatalf("Failed to generate entropy: %v", err)
		}

		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			utils.Fatalf("Failed to generate mnemonic: %v", err)
		}

		// Derive private key from mnemonic.
		account := uint32(ctx.Uint("account"))
		privateKey, err := deriveKeyFromMnemonic(mnemonic, account)
		if err != nil {
			utils.Fatalf("Failed to derive key: %v", err)
		}

		out := outputMnemonic{
			Mnemonic:   mnemonic,
			Account:    account,
			PrivateKey: hex.EncodeToString(crypto.MarshalPrivateKey(privateKey)),
			PublicKey:  hex.EncodeToString(privateKey.PublicKey()[:]),
			Address:    privateKey.Address().Hex(),
			Network:    networkName(),
		}

		if ctx.Bool(jsonFlag.Name) {
			mustPrintJSON(out)
		} else {
			fmt.Println("Network:    ", out.Network)
			fmt.Println("Account:    ", out.Account)
			fmt.Println("Mnemonic:   ", out.Mnemonic)
			fmt.Println("Private key:", out.PrivateKey)
			fmt.Println("Public key: ", out.PublicKey)
			fmt.Println("Address:    ", out.Address)
		}
		return nil
	},
}

var commandMnemonicToKey = cli.Command{
	Name:      "mnemonic-to-key",
	Usage:     "derive a private key from a BIP-39 mnemonic",
	ArgsUsage: "<mnemonic>",
	Description: `
Derive an Ed448 private key from a BIP-39 mnemonic phrase (12, 15, 18, 21, or 24 words).

Uses the same derivation as generate-mnemonic:
  mnemonic -> BIP-39 seed (PBKDF2-HMAC-SHA512, 64 bytes) -> first 57 bytes -> Ed448 key.

Use --account N to derive a specific account (default: 0).
Each account index produces a different key from the same mnemonic.

Use --networkid to set the correct network for address derivation:
  --networkid 1   mainnet (cb prefix, default)
  --networkid 3   devin testnet (ab prefix)
  --networkid >3  private network (ce prefix)

If the mnemonic uses a non-English wordlist, specify --lang to match.
`,
	Flags: []cli.Flag{
		jsonFlag,
		langFlag,
		accountFlag,
		utils.NetworkIdFlag,
	},
	Action: func(ctx *cli.Context) error {
		if ctx.IsSet(utils.NetworkIdFlag.Name) {
			common.DefaultNetworkID = common.NetworkID(ctx.Uint64(utils.NetworkIdFlag.Name))
		}

		if err := setWordList(ctx.String(langFlag.Name)); err != nil {
			utils.Fatalf("%v", err)
		}

		mnemonic := strings.Join(ctx.Args(), " ")
		if mnemonic == "" {
			utils.Fatalf("Mnemonic is required. Usage: xcbkey mnemonic-to-key \"word1 word2 ... word24\"")
		}
		mnemonic = strings.TrimSpace(mnemonic)

		if !bip39.IsMnemonicValid(mnemonic) {
			utils.Fatalf("Invalid BIP-39 mnemonic (check --lang if using a non-English wordlist)")
		}

		account := uint32(ctx.Uint("account"))
		privateKey, err := deriveKeyFromMnemonic(mnemonic, account)
		if err != nil {
			utils.Fatalf("Failed to derive key: %v", err)
		}

		out := outputMnemonic{
			Mnemonic:   mnemonic,
			Account:    account,
			PrivateKey: hex.EncodeToString(crypto.MarshalPrivateKey(privateKey)),
			PublicKey:  hex.EncodeToString(privateKey.PublicKey()[:]),
			Address:    privateKey.Address().Hex(),
			Network:    networkName(),
		}

		if ctx.Bool(jsonFlag.Name) {
			mustPrintJSON(out)
		} else {
			fmt.Println("Network:    ", out.Network)
			fmt.Println("Account:    ", out.Account)
			fmt.Println("Mnemonic:   ", out.Mnemonic)
			fmt.Println("Private key:", out.PrivateKey)
			fmt.Println("Public key: ", out.PublicKey)
			fmt.Println("Address:    ", out.Address)
		}
		return nil
	},
}

var commandVerifyMnemonic = cli.Command{
	Name:      "verify-mnemonic",
	Usage:     "verify that a mnemonic matches a private key",
	ArgsUsage: "<mnemonic>",
	Description: `
Verify that a BIP-39 mnemonic derives the expected private key.

Since BIP-39 uses PBKDF2 (a one-way function), you cannot reverse a private key
back to its mnemonic. Instead, this command takes a mnemonic and a private key,
derives the key from the mnemonic, and confirms whether they match.

Use --account N to verify against a specific account index (default: 0).

Use --networkid to set the correct network for address derivation:
  --networkid 1   mainnet (cb prefix, default)
  --networkid 3   devin testnet (ab prefix)
  --networkid >3  private network (ce prefix)

If the mnemonic uses a non-English wordlist, specify --lang to match.
`,
	Flags: []cli.Flag{
		jsonFlag,
		langFlag,
		accountFlag,
		utils.NetworkIdFlag,
		cli.StringFlag{
			Name:  "privatekey",
			Usage: "private key in hex to verify against",
		},
		cli.StringFlag{
			Name:  "keyfile",
			Usage: "path to file containing the private key",
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.IsSet(utils.NetworkIdFlag.Name) {
			common.DefaultNetworkID = common.NetworkID(ctx.Uint64(utils.NetworkIdFlag.Name))
		}

		if err := setWordList(ctx.String(langFlag.Name)); err != nil {
			utils.Fatalf("%v", err)
		}

		// Get mnemonic from arguments.
		mnemonic := strings.Join(ctx.Args(), " ")
		if mnemonic == "" {
			utils.Fatalf("Mnemonic is required")
		}
		mnemonic = strings.TrimSpace(mnemonic)

		if !bip39.IsMnemonicValid(mnemonic) {
			utils.Fatalf("Invalid BIP-39 mnemonic (check --lang if using a non-English wordlist)")
		}

		// Get the private key to verify against.
		var existingKey *crypto.PrivateKey
		var err error

		if keyfile := ctx.String("keyfile"); keyfile != "" {
			existingKey, err = crypto.LoadEDDSA(keyfile)
			if err != nil {
				utils.Fatalf("Failed to load private key from file: %v", err)
			}
		} else if hexKey := ctx.String("privatekey"); hexKey != "" {
			existingKey, err = crypto.UnmarshalPrivateKeyHex(hexKey)
			if err != nil {
				utils.Fatalf("Failed to parse private key: %v", err)
			}
		} else {
			utils.Fatalf("Either --privatekey or --keyfile is required")
		}

		// Derive key from mnemonic and compare.
		account := uint32(ctx.Uint("account"))
		derivedKey, err := deriveKeyFromMnemonic(mnemonic, account)
		if err != nil {
			utils.Fatalf("Failed to derive key from mnemonic: %v", err)
		}

		existingHex := hex.EncodeToString(crypto.MarshalPrivateKey(existingKey))
		derivedHex := hex.EncodeToString(crypto.MarshalPrivateKey(derivedKey))
		match := existingHex == derivedHex

		if ctx.Bool(jsonFlag.Name) {
			mustPrintJSON(struct {
				Match      bool   `json:"match"`
				Mnemonic   string `json:"mnemonic"`
				Account    uint32 `json:"account"`
				DerivedKey string `json:"derived_key"`
				GivenKey   string `json:"given_key"`
				Address    string `json:"address"`
				Network    string `json:"network"`
			}{
				Match:      match,
				Mnemonic:   mnemonic,
				Account:    account,
				DerivedKey: derivedHex,
				GivenKey:   existingHex,
				Address:    derivedKey.Address().Hex(),
				Network:    networkName(),
			})
		} else {
			if match {
				fmt.Println("MATCH: mnemonic correctly derives the given private key")
				fmt.Println("Network:", networkName())
				fmt.Println("Account:", account)
				fmt.Println("Address:", derivedKey.Address().Hex())
			} else {
				fmt.Println("MISMATCH: mnemonic does NOT derive the given private key")
				fmt.Println("Account:    ", account)
				fmt.Println("Derived key:", derivedHex)
				fmt.Println("Given key:  ", existingHex)
			}
		}
		return nil
	},
}

// wordsToEntropyBits converts a mnemonic word count to BIP-39 entropy bits.
func wordsToEntropyBits(words int) (int, error) {
	switch words {
	case 12:
		return 128, nil
	case 15:
		return 160, nil
	case 18:
		return 192, nil
	case 21:
		return 224, nil
	case 24:
		return 256, nil
	default:
		return 0, fmt.Errorf("invalid word count %d: must be 12, 15, 18, 21, or 24", words)
	}
}

// deriveKeyFromMnemonic derives an Ed448 private key from a BIP-39 mnemonic
// and account index.
//
// Account 0: uses the BIP-39 seed directly (first 57 bytes), backwards
// compatible with single-account wallets.
//
// Account N (N > 0): derives a child key using HMAC-SHA512 with the BIP-39
// seed as key and a domain-separated account index as data. This produces a
// unique, deterministic 57-byte Ed448 key per account.
func deriveKeyFromMnemonic(mnemonic string, account uint32) (*crypto.PrivateKey, error) {
	seed := bip39.NewSeed(mnemonic, "")
	if account == 0 {
		return crypto.UnmarshalPrivateKey(seed[:crypto.PrivkeyLength])
	}
	// Derive child key: HMAC-SHA512(key=seed, data="Core Ed448 account" || uint32BE(account))
	h := hmac.New(sha512.New, seed)
	h.Write([]byte("Core Ed448 account"))
	binary.Write(h, binary.BigEndian, account)
	derived := h.Sum(nil) // 64 bytes
	return crypto.UnmarshalPrivateKey(derived[:crypto.PrivkeyLength])
}

// setWordList sets the BIP-39 wordlist based on the language string.
func setWordList(lang string) error {
	switch lang {
	case "english":
		bip39.SetWordList(wordlists.English)
	case "japanese":
		bip39.SetWordList(wordlists.Japanese)
	case "korean":
		bip39.SetWordList(wordlists.Korean)
	case "spanish":
		bip39.SetWordList(wordlists.Spanish)
	case "french":
		bip39.SetWordList(wordlists.French)
	case "italian":
		bip39.SetWordList(wordlists.Italian)
	case "chinese_simplified":
		bip39.SetWordList(wordlists.ChineseSimplified)
	case "chinese_traditional":
		bip39.SetWordList(wordlists.ChineseTraditional)
	default:
		return fmt.Errorf("unsupported language %q: use english, japanese, korean, spanish, french, italian, chinese_simplified, or chinese_traditional", lang)
	}
	return nil
}

// networkName returns a human-readable name for the current network ID.
func networkName() string {
	switch common.DefaultNetworkID {
	case common.Mainnet:
		return "mainnet"
	case common.Devin:
		return "devin"
	default:
		return fmt.Sprintf("private(%d)", common.DefaultNetworkID)
	}
}
