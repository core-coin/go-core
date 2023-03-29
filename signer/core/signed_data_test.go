// Copyright 2019 by the Authors
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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"testing"

	"github.com/core-coin/go-core/v2/accounts/keystore"
	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/common/math"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/signer/core"
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
			Name: "networkId",
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
            "name": "networkId",
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
        "networkId": "1",
        "verifyingContract": "cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba"
      },
      "message": {
        "from": {
          "name": "Cow",
		  "test": 3,
          "wallet": "cb76a631db606f1452ddc2432931d611f1d5b126f848"
        },
        "to": {
          "name": "Bob",
          "wallet": "cb27de521e43741cf785cbad450d5649187b9612018f"
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
	"cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba",
	"",
}

var messageStandard = map[string]interface{}{
	"from": map[string]interface{}{
		"name":   "Cow",
		"wallet": "cb76a631db606f1452ddc2432931d611f1d5b126f848",
	},
	"to": map[string]interface{}{
		"name":   "Bob",
		"wallet": "cb27de521e43741cf785cbad450d5649187b9612018f",
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
	if signature == nil || len(signature) != crypto.ExtendedSignatureLength {
		t.Errorf("Expected crypto.ExtendedSignatureLength byte signature (got %d bytes)", len(signature))
	}
	// data/typed
	control.approveCh <- "Y"
	control.inputCh <- "a_long_password"
	signature, err = api.SignTypedData(context.Background(), a, typedData)
	if err != nil {
		t.Fatal(err)
	}
	if signature == nil || len(signature) != crypto.ExtendedSignatureLength {
		t.Errorf("Expected crypto.ExtendedSignatureLength byte signature (got %d bytes)", len(signature))
	}
}

func TestDomainNetworkId(t *testing.T) {
	withoutNetworkID := core.TypedData{
		Types: core.Types{
			"CIP712Domain": []core.Type{
				{Name: "name", Type: "string"},
			},
		},
		Domain: core.TypedDataDomain{
			Name: "test",
		},
	}

	if _, ok := withoutNetworkID.Domain.Map()["networkId"]; ok {
		t.Errorf("Expected the networkId key to not be present in the domain map")
	}
	// should encode successfully
	if _, err := withoutNetworkID.HashStruct("CIP712Domain", withoutNetworkID.Domain.Map()); err != nil {
		t.Errorf("Expected the typedData to encode the domain successfully, got %v", err)
	}
	withNetworkID := core.TypedData{
		Types: core.Types{
			"CIP712Domain": []core.Type{
				{Name: "name", Type: "string"},
				{Name: "networkId", Type: "uint256"},
			},
		},
		Domain: core.TypedDataDomain{
			Name:      "test",
			NetworkId: math.NewHexOrDecimal256(1),
		},
	}

	if _, ok := withNetworkID.Domain.Map()["networkId"]; !ok {
		t.Errorf("Expected the networkId key be present in the domain map")
	}
	// should encode successfully
	if _, err := withNetworkID.HashStruct("CIP712Domain", withNetworkID.Domain.Map()); err != nil {
		t.Errorf("Expected the typedData to encode the domain successfully, got %v", err)
	}
}

func TestHashStruct(t *testing.T) {
	hash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		t.Fatal(err)
	}
	mainHash := fmt.Sprintf("0x%s", common.Bytes2Hex(hash))
	if mainHash != "0xef6fddde45efef4974d865911ed95b5713bad9416a70aefe6f1c9cf7cf2effdb" {
		t.Errorf("Expected different hashStruct result (got %s)", mainHash)
	}

	hash, err = typedData.HashStruct("CIP712Domain", typedData.Domain.Map())
	if err != nil {
		t.Error(err)
	}
	domainHash := fmt.Sprintf("0x%s", common.Bytes2Hex(hash))
	if domainHash != "0x453039a903a6cb0751f4b709f858d104351ba885a17a7daf98d7e6442d9f324d" {
		t.Errorf("Expected different domain hashStruct result (got %s)", domainHash)
	}
}

func TestEncodeType(t *testing.T) {
	domainTypeEncoding := string(typedData.EncodeType("CIP712Domain"))
	if domainTypeEncoding != "CIP712Domain(string name,string version,uint256 networkId,address verifyingContract)" {
		t.Errorf("Expected different encodeType result (got %s)", domainTypeEncoding)
	}

	mailTypeEncoding := string(typedData.EncodeType(typedData.PrimaryType))
	if mailTypeEncoding != "Mail(Person from,Person to,string contents)Person(string name,address wallet)" {
		t.Errorf("Expected different encodeType result (got %s)", mailTypeEncoding)
	}
}

func TestTypeHash(t *testing.T) {
	mailTypeHash := fmt.Sprintf("0x%s", common.Bytes2Hex(typedData.TypeHash(typedData.PrimaryType)))
	if mailTypeHash != "0xda8b122f9405015467a4c2d2b5d72f976d0dcd07f39d640df998cb582f24622b" {
		t.Errorf("Expected different typeHash result (got %s)", mailTypeHash)
	}
}

func TestEncodeData(t *testing.T) {
	hash, err := typedData.EncodeData(typedData.PrimaryType, typedData.Message, 0)
	if err != nil {
		t.Fatal(err)
	}
	dataEncoding := fmt.Sprintf("0x%s", common.Bytes2Hex(hash))
	if dataEncoding != "0xda8b122f9405015467a4c2d2b5d72f976d0dcd07f39d640df998cb582f24622bb5a2006b708c2c1a6ecb8e7e28b588a0c0f1cc91e1ad7f834b1b85a8b04fbc31b7287cdbb157715c966c50826f17d51a9311dec62c8957819d6b4de0bbc55b9cb58543c145f315ad2c9210b45c29c13e6c9fc5396a140d3b07f766925fda360e" {
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
	sighash := crypto.SHA3(rawData)
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

var gnosisTypedData = `
{
	"types": {
		"CIP712Domain": [
			{ "type": "address", "name": "verifyingContract" }
		],
		"SafeTx": [
			{ "type": "address", "name": "to" },
			{ "type": "uint256", "name": "value" },
			{ "type": "bytes", "name": "data" },
			{ "type": "uint8", "name": "operation" },
			{ "type": "uint256", "name": "safeTxEnergy" },
			{ "type": "uint256", "name": "baseEnergy" },
			{ "type": "uint256", "name": "energyPrice" },
			{ "type": "address", "name": "energyToken" },
			{ "type": "address", "name": "refundReceiver" },
			{ "type": "uint256", "name": "nonce" }
		]
	},
	"domain": {
		"verifyingContract": "cb45f23d9ab6aefb2c22dfff511e29703435a3b50f50"
	},
	"primaryType": "SafeTx",
	"message": {
		"to": "cb65e49851f010cd7d81b5b4969f3b0e8325c415359d",
		"value": "20000000000000000",
		"data": "0x",
		"operation": 0,
		"safeTxEnergy": 27845,
		"baseEnergy": 0,
		"energyPrice": "0",
		"energyToken": "cb540000000000000000000000000000000000000000",
		"refundReceiver": "cb540000000000000000000000000000000000000000",
		"nonce": 3
	}
}`

var gnosisTx = `
{
      "safe": "cb45f23d9ab6aefb2c22dfff511e29703435a3b50f50",
      "to": "cb65e49851f010cd7d81b5b4969f3b0e8325c415359d",
      "value": "20000000000000000",
      "data": null,
      "operation": 0,
      "energyToken": "cb540000000000000000000000000000000000000000",
      "safeTxEnergy": 27845,
      "baseEnergy": 0,
      "energyPrice": "0",
      "refundReceiver": "cb540000000000000000000000000000000000000000",
      "nonce": 3,
      "executionDate": null,
      "submissionDate": "2020-09-15T21:59:23.815748Z",
      "modified": "2020-09-15T21:59:23.815748Z",
      "blockNumber": null,
      "transactionHash": null,
      "safeTxHash": "0x28bae2bd58d894a1d9b69e5e9fde3570c4b98a6fc5499aefb54fb830137e831f",
      "executor": null,
      "isExecuted": false,
      "isSuccessful": null,
      "xcbEnergyPrice": null,
      "energyUsed": null,
      "fee": null,
      "origin": null,
      "dataDecoded": null,
      "confirmationsRequired": null,
      "confirmations": [
        {
          "owner": "0xAd2e180019FCa9e55CADe76E4487F126Fd08DA34",
          "submissionDate": "2020-09-15T21:59:28.281243Z",
          "transactionHash": null,
          "confirmationType": "CONFIRMATION",
          "signature": "0x5e562065a0cb15d766dac0cd49eb6d196a41183af302c4ecad45f1a81958d7797753f04424a9b0aa1cb0448e4ec8e189540fbcdda7530ef9b9d95dfc2d36cb521b",
          "signatureType": "EOA"
        }
      ],
      "signatures": null
    }
`

// TestGnosisTypedData tests the scenario where a user submits a full CIP-712
// struct without using the gnosis-specific endpoint
func TestGnosisTypedData(t *testing.T) {
	var td core.TypedData
	err := json.Unmarshal([]byte(gnosisTypedData), &td)
	if err != nil {
		t.Fatalf("unmarshalling failed '%v'", err)
	}
	_, sighash, err := sign(td)
	if err != nil {
		t.Fatal(err)
	}
	expSigHash := common.FromHex("0x402d167803f194fe5e448757cbac9a9c17fb4674a7da5c061233f49118c379b7")
	if !bytes.Equal(expSigHash, sighash) {
		t.Fatalf("Error, got %x, wanted %x", sighash, expSigHash)
	}
}

// TestGnosisCustomData tests the scenario where a user submits only the gnosis-safe
// specific data, and we fill the TypedData struct on our side
func TestGnosisCustomData(t *testing.T) {
	var tx core.GnosisSafeTx
	err := json.Unmarshal([]byte(gnosisTx), &tx)
	if err != nil {
		t.Fatal(err)
	}
	var td = tx.ToTypedData()
	_, sighash, err := sign(td)
	if err != nil {
		t.Fatal(err)
	}
	expSigHash := common.FromHex("0x402d167803f194fe5e448757cbac9a9c17fb4674a7da5c061233f49118c379b7")
	if !bytes.Equal(expSigHash, sighash) {
		t.Fatalf("Error, got %x, wanted %x", sighash, expSigHash)
	}
}
