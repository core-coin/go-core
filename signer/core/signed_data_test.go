// Copyright 2019 The go-core Authors
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

package core_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"testing"

	"github.com/core-coin/go-core/accounts/keystore"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/common/math"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/signer/core"
)

var typesStandard = core.Types{
	"CIP712Domain": {
		{
			Name: "name",
			Type: "string",
		},
		{
			Name: "version",
			Type: "string",
		},
		{
			Name: "chainId",
			Type: "uint256",
		},
		{
			Name: "verifyingContract",
			Type: "address",
		},
	},
	"Person": {
		{
			Name: "name",
			Type: "string",
		},
		{
			Name: "wallet",
			Type: "address",
		},
	},
	"Mail": {
		{
			Name: "from",
			Type: "Person",
		},
		{
			Name: "to",
			Type: "Person",
		},
		{
			Name: "contents",
			Type: "string",
		},
	},
}

var jsonTypedData = `
    {
      "types": {
        "CIP712Domain": [
          {
            "name": "name",
            "type": "string"
          },
          {
            "name": "version",
            "type": "string"
          },
          {
            "name": "chainId",
            "type": "uint256"
          },
          {
            "name": "verifyingContract",
            "type": "address"
          }
        ],
        "Person": [
          {
            "name": "name",
            "type": "string"
          },
          {
            "name": "test",
            "type": "uint8"
          },
          {
            "name": "wallet",
            "type": "address"
          }
        ],
        "Mail": [
          {
            "name": "from",
            "type": "Person"
          },
          {
            "name": "to",
            "type": "Person"
          },
          {
            "name": "contents",
            "type": "string"
          }
        ]
      },
      "primaryType": "Mail",
      "domain": {
        "name": "Core Mail",
        "version": "1",
        "chainId": "1",
        "verifyingContract": "cb49CcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC"
      },
      "message": {
        "from": {
          "name": "Cow",
		  "test": 3,
          "wallet": "cb66CD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826"
        },
        "to": {
          "name": "Bob",
          "wallet": "cb53bBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB"
        },
        "contents": "Hello, Bob!"
      }
    }
`

const primaryType = "Mail"

var domainStandard = core.TypedDataDomain{
	"Core Mail",
	"1",
	math.NewHexOrDecimal256(1),
	"cb49CcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC",
	"",
}

var messageStandard = map[string]interface{}{
	"from": map[string]interface{}{
		"name":   "Cow",
		"wallet": "cb66CD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826",
	},
	"to": map[string]interface{}{
		"name":   "Bob",
		"wallet": "cb53bBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB",
	},
	"contents": "Hello, Bob!",
}

var typedData = core.TypedData{
	Types:       typesStandard,
	PrimaryType: primaryType,
	Domain:      domainStandard,
	Message:     messageStandard,
}

func TestSignData(t *testing.T) {
	api, control := setup(t)
	//Create two accounts
	createAccount(control, api, t)
	createAccount(control, api, t)
	control.approveCh <- "1"
	list, err := api.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	a := list[0]

	control.approveCh <- "Y"
	control.inputCh <- "wrongpassword"
	signature, err := api.SignData(context.Background(), core.TextPlain.Mime, a, hexutil.Encode([]byte("EHLO world")))
	if signature != nil {
		t.Errorf("Expected nil-data, got %x", signature)
	}
	if err != keystore.ErrDecrypt {
		t.Errorf("Expected ErrLocked! '%v'", err)
	}
	control.approveCh <- "No way"
	signature, err = api.SignData(context.Background(), core.TextPlain.Mime, a, hexutil.Encode([]byte("EHLO world")))
	if signature != nil {
		t.Errorf("Expected nil-data, got %x", signature)
	}
	if err != core.ErrRequestDenied {
		t.Errorf("Expected ErrRequestDenied! '%v'", err)
	}
	// text/plain
	control.approveCh <- "Y"
	control.inputCh <- "a_long_password"
	signature, err = api.SignData(context.Background(), core.TextPlain.Mime, a, hexutil.Encode([]byte("EHLO world")))
	if err != nil {
		t.Fatal(err)
	}
	if signature == nil || len(signature) != crypto.SignatureLength {
		t.Errorf("Expected %d byte signature (got %d bytes)", crypto.SignatureLength, len(signature))
	}
	// data/typed
	control.approveCh <- "Y"
	control.inputCh <- "a_long_password"
	signature, err = api.SignTypedData(context.Background(), a, typedData)
	if err != nil {
		t.Fatal(err)
	}
	if signature == nil || len(signature) != crypto.SignatureLength {
		t.Errorf("Expected %d byte signature (got %d bytes)", crypto.SignatureLength, len(signature))
	}
}

func TestDomainChainId(t *testing.T) {
	withoutChainID := core.TypedData{
		Types: core.Types{
			"CIP712Domain": []core.Type{
				{Name: "name", Type: "string"},
			},
		},
		Domain: core.TypedDataDomain{
			Name: "test",
		},
	}

	if _, ok := withoutChainID.Domain.Map()["chainId"]; ok {
		t.Errorf("Expected the chainId key to not be present in the domain map")
	}
	withChainID := core.TypedData{
		Types: core.Types{
			"CIP712Domain": []core.Type{
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
		},
		Domain: core.TypedDataDomain{
			Name:    "test",
			ChainId: math.NewHexOrDecimal256(1),
		},
	}

	if _, ok := withChainID.Domain.Map()["chainId"]; !ok {
		t.Errorf("Expected the chainId key be present in the domain map")
	}
}

func TestHashStruct(t *testing.T) {
	hash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		t.Fatal(err)
	}
	mainHash := fmt.Sprintf("0x%s", common.Bytes2Hex(hash))
	if mainHash != "0x4c671f05b59c0247c9f7ff8d13435912eb6fcbd049563b2f075442a24c267e8a" {
		t.Errorf("Expected different hashStruct result (got %s)", mainHash)
	}

	hash, err = typedData.HashStruct("CIP712Domain", typedData.Domain.Map())
	if err != nil {
		t.Error(err)
	}
	domainHash := fmt.Sprintf("0x%s", common.Bytes2Hex(hash))
	if domainHash != "0xe7b12d1c32f9a2ec20787952f329ea66e0ef0c70c98ff866549bb257d2ad344e" {
		t.Errorf("Expected different domain hashStruct result (got %s)", domainHash)
	}
}

func TestEncodeType(t *testing.T) {
	domainTypeEncoding := string(typedData.EncodeType("CIP712Domain"))
	if domainTypeEncoding != "CIP712Domain(string name,string version,uint256 chainId,address verifyingContract)" {
		t.Errorf("Expected different encodeType result (got %s)", domainTypeEncoding)
	}

	mailTypeEncoding := string(typedData.EncodeType(typedData.PrimaryType))
	if mailTypeEncoding != "Mail(Person from,Person to,string contents)Person(string name,address wallet)" {
		t.Errorf("Expected different encodeType result (got %s)", mailTypeEncoding)
	}
}

func TestTypeHash(t *testing.T) {
	mailTypeHash := fmt.Sprintf("0x%s", common.Bytes2Hex(typedData.TypeHash(typedData.PrimaryType)))
	if mailTypeHash != "0xa0cedeb2dc280ba39b857546d74f5549c3a1d7bdc2dd96bf881f76108e23dac2" {
		t.Errorf("Expected different typeHash result (got %s)", mailTypeHash)
	}
}

func TestEncodeData(t *testing.T) {
	hash, err := typedData.EncodeData(typedData.PrimaryType, typedData.Message, 0)
	if err != nil {
		t.Fatal(err)
	}
	dataEncoding := fmt.Sprintf("0x%s", common.Bytes2Hex(hash))
	if dataEncoding != "0xa0cedeb2dc280ba39b857546d74f5549c3a1d7bdc2dd96bf881f76108e23dac251eec1c9b4b49e4aae8288b7a1bae749e657126acc5983d2e35a526cba26bdf40e06ae67611986f24b0d4493ec46770ab05c42a1c3fdcd1554462c5d0d03e7d4b5aadf3154a261abdd9086fc627b61efca26ae5702701d05cd2305f7c52a2fc8" {
		t.Errorf("Expected different encodeData result (got %s)", dataEncoding)
	}
}

func TestFormatter(t *testing.T) {
	var d core.TypedData
	err := json.Unmarshal([]byte(jsonTypedData), &d)
	if err != nil {
		t.Fatalf("unmarshalling failed '%v'", err)
	}
	formatted, _ := d.Format()
	for _, item := range formatted {
		t.Logf("'%v'\n", item.Pprint(0))
	}

	j, _ := json.Marshal(formatted)
	t.Logf("'%v'\n", string(j))
}

func sign(typedData core.TypedData) ([]byte, []byte, error) {
	domainSeparator, err := typedData.HashStruct("CIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, nil, err
	}
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, nil, err
	}
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	sighash := crypto.Keccak256(rawData)
	return typedDataHash, sighash, nil
}

func TestJsonFiles(t *testing.T) {
	testfiles, err := ioutil.ReadDir("testdata/")
	if err != nil {
		t.Fatalf("failed reading files: %v", err)
	}
	for i, fInfo := range testfiles {
		if !strings.HasSuffix(fInfo.Name(), "json") {
			continue
		}
		expectedFailure := strings.HasPrefix(fInfo.Name(), "expfail")
		data, err := ioutil.ReadFile(path.Join("testdata", fInfo.Name()))
		if err != nil {
			t.Errorf("Failed to read file %v: %v", fInfo.Name(), err)
			continue
		}
		var typedData core.TypedData
		err = json.Unmarshal(data, &typedData)
		if err != nil {
			t.Errorf("Test %d, file %v, json unmarshalling failed: %v", i, fInfo.Name(), err)
			continue
		}
		_, _, err = sign(typedData)
		t.Logf("Error %v\n", err)
		if err != nil && !expectedFailure {
			t.Errorf("Test %d failed, file %v: %v", i, fInfo.Name(), err)
		}
		if expectedFailure && err == nil {
			t.Errorf("Test %d succeeded (expected failure), file %v: %v", i, fInfo.Name(), err)
		}
	}
}

// TestFuzzerFiles tests some files that have been found by fuzzing to cause
// crashes or hangs.
func TestFuzzerFiles(t *testing.T) {
	corpusdir := path.Join("testdata", "fuzzing")
	testfiles, err := ioutil.ReadDir(corpusdir)
	if err != nil {
		t.Fatalf("failed reading files: %v", err)
	}
	verbose := false
	for i, fInfo := range testfiles {
		data, err := ioutil.ReadFile(path.Join(corpusdir, fInfo.Name()))
		if err != nil {
			t.Errorf("Failed to read file %v: %v", fInfo.Name(), err)
			continue
		}
		var typedData core.TypedData
		err = json.Unmarshal(data, &typedData)
		if err != nil {
			t.Errorf("Test %d, file %v, json unmarshalling failed: %v", i, fInfo.Name(), err)
			continue
		}
		_, err = typedData.EncodeData("CIP712Domain", typedData.Domain.Map(), 1)
		if verbose && err != nil {
			t.Logf("%d, EncodeData[1] err: %v\n", i, err)
		}
		_, err = typedData.EncodeData(typedData.PrimaryType, typedData.Message, 1)
		if verbose && err != nil {
			t.Logf("%d, EncodeData[2] err: %v\n", i, err)
		}
		typedData.Format()
	}
}
