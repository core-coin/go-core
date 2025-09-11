package scapi

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core/state"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/internal/xcbapi"
	"github.com/core-coin/go-core/v2/rpc"
)

// MockBackend for testing
type mockBackend struct {
	callContractFunc func(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error)
}

func (m *mockBackend) CallContract(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
	return m.callContractFunc(ctx, call, blockNumber)
}

func (m *mockBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	return nil, nil, nil
}

func TestGetKVFunctionSelectors(t *testing.T) {
	t.Logf("Testing CIP-150 function selectors:")
	t.Logf("  - hasKey(string): 0xf37e8f05")
	t.Logf("  - isSealed(string): 0xf272a162")
	t.Logf("  - getValue(string): 0xe2f3625a")

	// Test hasKey selector
	hasKeySelector := "0xf37e8f05"
	hasKeyData := hexutil.MustDecode(hasKeySelector)
	t.Logf("  hasKey selector: %s -> %d bytes", hasKeySelector, len(hasKeyData))

	// Test isSealed selector
	isSealedSelector := "0xf272a162"
	isSealedData := hexutil.MustDecode(isSealedSelector)
	t.Logf("  isSealed selector: %s -> %d bytes", isSealedSelector, len(isSealedData))

	// Test getValue selector
	getValueSelector := "0xe2f3625a"
	getValueData := hexutil.MustDecode(getValueSelector)
	t.Logf("  getValue selector: %s -> %d bytes", getValueSelector, len(getValueData))

	// Verify all selectors are 4 bytes
	if len(hasKeyData) != 4 {
		t.Errorf("hasKey selector should be 4 bytes, got %d", len(hasKeyData))
	}
	if len(isSealedData) != 4 {
		t.Errorf("isSealed selector should be 4 bytes, got %d", len(isSealedData))
	}
	if len(getValueData) != 4 {
		t.Errorf("getValue selector should be 4 bytes, got %d", len(getValueData))
	}
}

func TestGetKVStringEncoding(t *testing.T) {
	t.Logf("Testing string encoding for CIP-150 calls:")

	key := "PROSPECTUS_V1"
	keyBytes := []byte(key)
	keyLength := len(keyBytes)

	t.Logf("  Key: %s", key)
	t.Logf("  Key bytes: %x", keyBytes)
	t.Logf("  Key length: %d bytes", keyLength)

	// Simulate the encoding process
	// Offset: 32 bytes (0x20)
	offset := big.NewInt(32)
	offsetBytes := make([]byte, 32)
	offset.FillBytes(offsetBytes)

	// Length: key length
	lengthBytes := make([]byte, 32)
	big.NewInt(int64(keyLength)).FillBytes(lengthBytes)

	// Key data (padded to 32 bytes)
	keyPadded := make([]byte, 32)
	copy(keyPadded, keyBytes)

	t.Logf("  Offset (32 bytes): %x", offsetBytes)
	t.Logf("  Length (%d): %x", keyLength, lengthBytes)
	t.Logf("  Key data (padded): %x", keyPadded)

	// Total call data structure
	totalLength := 4 + 32 + 32 + 32 // selector + offset + length + data
	t.Logf("  Total call data length: %d bytes", totalLength)

	// Verify the encoding logic
	if len(offsetBytes) != 32 {
		t.Errorf("Offset should be 32 bytes, got %d", len(offsetBytes))
	}
	if len(lengthBytes) != 32 {
		t.Errorf("Length should be 32 bytes, got %d", len(lengthBytes))
	}
	if len(keyPadded) != 32 {
		t.Errorf("Key data should be padded to 32 bytes, got %d", len(keyPadded))
	}
}

func TestGetKVWithMockContract(t *testing.T) {
	t.Logf("Testing GetKV with mock CIP-150 contract:")

	// Mock contract address
	contractAddr, err := common.HexToAddress("cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c")
	if err != nil {
		t.Fatalf("Failed to parse contract address: %v", err)
	}
	key := "PROSPECTUS_V1"

	t.Logf("  Contract: %s", contractAddr.Hex())
	t.Logf("  Key: %s", key)

	// Test case 1: Key exists and is not sealed
	t.Run("Key exists, not sealed", func(t *testing.T) {
		backend := &mockBackend{
			callContractFunc: func(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
				// Mock responses for hasKey, isSealed, getValue
				if call.DataBytes[0] == 0xf3 { // hasKey
					// Return true (key exists)
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, nil
				} else if call.DataBytes[0] == 0xf2 { // isSealed
					// Return false (not sealed)
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil
				} else if call.DataBytes[0] == 0xe2 { // getValue
					// Return dynamic string: "Investment Prospectus 2024"
					// This is a simplified mock - in reality it would be properly encoded
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 22, 73, 110, 118, 101, 115, 116, 109, 101, 110, 116, 32, 80, 114, 111, 115, 112, 101, 99, 116, 117, 115, 32, 50, 48, 50, 52}, nil
				}
				return nil, fmt.Errorf("unknown function selector")
			},
		}

		api := NewPublicSmartContractAPI(backend)
		result, err := api.GetKV(context.Background(), key, contractAddr, false)

		if err != nil {
			t.Errorf("GetKV failed: %v", err)
			return
		}

		t.Logf("  Result: value='%s', sealed=%t, exists=%t", result.Value, result.Sealed, result.Exists)

		if !result.Exists {
			t.Errorf("Expected key to exist")
		}
		if result.Sealed {
			t.Errorf("Expected key to not be sealed")
		}
		if result.Value == "" {
			t.Errorf("Expected non-empty value")
		}
	})

	// Test case 2: Key exists and is sealed
	t.Run("Key exists, sealed", func(t *testing.T) {
		backend := &mockBackend{
			callContractFunc: func(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
				// Mock responses for hasKey, isSealed, getValue
				if call.DataBytes[0] == 0xf3 { // hasKey
					// Return true (key exists)
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, nil
				} else if call.DataBytes[0] == 0xf2 { // isSealed
					// Return true (sealed)
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, nil
				} else if call.DataBytes[0] == 0xe2 { // getValue
					// Return dynamic string: "Final Legal Document"
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 70, 105, 110, 97, 108, 32, 76, 101, 103, 97, 108, 32, 68, 111, 99, 117, 109, 101, 110, 116}, nil
				}
				return nil, fmt.Errorf("unknown function selector")
			},
		}

		api := NewPublicSmartContractAPI(backend)
		result, err := api.GetKV(context.Background(), key, contractAddr, false)

		if err != nil {
			t.Errorf("GetKV failed: %v", err)
			return
		}

		t.Logf("  Result: value='%s', sealed=%t, exists=%t", result.Value, result.Sealed, result.Exists)

		if !result.Exists {
			t.Errorf("Expected key to exist")
		}
		if !result.Sealed {
			t.Errorf("Expected key to be sealed")
		}
		if result.Value == "" {
			t.Errorf("Expected non-empty value")
		}
	})

	// Test case 3: Key doesn't exist
	t.Run("Key doesn't exist", func(t *testing.T) {
		backend := &mockBackend{
			callContractFunc: func(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
				// Mock response for hasKey only
				if call.DataBytes[0] == 0xf3 { // hasKey
					// Return false (key doesn't exist)
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil
				}
				return nil, fmt.Errorf("unexpected call")
			},
		}

		api := NewPublicSmartContractAPI(backend)
		result, err := api.GetKV(context.Background(), key, contractAddr, false)

		if err != nil {
			t.Errorf("GetKV failed: %v", err)
			return
		}

		t.Logf("  Result: value='%s', sealed=%t, exists=%t", result.Value, result.Sealed, result.Exists)

		if result.Exists {
			t.Errorf("Expected key to not exist")
		}
		if result.Value != "" {
			t.Errorf("Expected empty value for non-existent key")
		}
	})
}

func TestGetKVSealedOnly(t *testing.T) {
	t.Logf("Testing GetKV with sealed=true (only sealed items):")

	contractAddr, err := common.HexToAddress("cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c")
	if err != nil {
		t.Fatalf("Failed to parse contract address: %v", err)
	}
	key := "LEGAL_DOCUMENT"

	t.Logf("  Contract: %s", contractAddr.Hex())
	t.Logf("  Key: %s", key)
	t.Logf("  Request: sealed=true (only sealed items)")

	// Test case: Request sealed items but key is not sealed
	t.Run("Request sealed but key not sealed", func(t *testing.T) {
		backend := &mockBackend{
			callContractFunc: func(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
				// Mock responses for hasKey and isSealed
				if call.DataBytes[0] == 0xf3 { // hasKey
					// Return true (key exists)
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, nil
				} else if call.DataBytes[0] == 0xf2 { // isSealed
					// Return false (not sealed)
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil
				}
				return nil, fmt.Errorf("unexpected call")
			},
		}

		api := NewPublicSmartContractAPI(backend)
		result, err := api.GetKV(context.Background(), key, contractAddr, true)

		if err != nil {
			t.Errorf("GetKV failed: %v", err)
			return
		}

		t.Logf("  Result: value='%s', sealed=%t, exists=%t", result.Value, result.Sealed, result.Exists)

		if !result.Exists {
			t.Errorf("Expected key to exist")
		}
		if result.Sealed {
			t.Errorf("Expected key to not be sealed")
		}
		if result.Value != "" {
			t.Errorf("Expected empty value when requesting sealed items but key is not sealed")
		}
	})

	// Test case: Request sealed items and key is sealed
	t.Run("Request sealed and key is sealed", func(t *testing.T) {
		backend := &mockBackend{
			callContractFunc: func(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
				// Mock responses for hasKey, isSealed, and getValue
				if call.DataBytes[0] == 0xf3 { // hasKey
					// Return true (key exists)
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, nil
				} else if call.DataBytes[0] == 0xf2 { // isSealed
					// Return true (sealed)
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, nil
				} else if call.DataBytes[0] == 0xe2 { // getValue
					// Return dynamic string: "Sealed Legal Document"
					// Offset: 32 (0x20), Length: 22 (0x16), Data: "Sealed Legal Document"
					// Need to pad the data to 32 bytes for proper alignment
					return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 22, 83, 101, 97, 108, 101, 100, 32, 76, 101, 103, 97, 108, 32, 68, 111, 99, 117, 109, 101, 110, 116, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil
				}
				return nil, fmt.Errorf("unknown function selector")
			},
		}

		api := NewPublicSmartContractAPI(backend)
		result, err := api.GetKV(context.Background(), key, contractAddr, true)

		if err != nil {
			t.Errorf("GetKV failed: %v", err)
			return
		}

		t.Logf("  Result: value='%s', sealed=%t, exists=%t", result.Value, result.Sealed, result.Exists)

		if !result.Exists {
			t.Errorf("Expected key to exist")
		}
		if !result.Sealed {
			t.Errorf("Expected key to be sealed")
		}
		if result.Value == "" {
			t.Errorf("Expected non-empty value for sealed key")
		}
	})
}

func TestGetKVWithUserExample(t *testing.T) {
	t.Logf("Testing GetKV with user's example scenario:")

	// Example from CIP-150: RWA token with prospectus metadata
	contractAddr, err := common.HexToAddress("cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c")
	if err != nil {
		t.Fatalf("Failed to parse contract address: %v", err)
	}
	key := "PROSPECTUS"

	t.Logf("  Contract: %s", contractAddr.Hex())
	t.Logf("  Key: %s", key)
	t.Logf("  Expected: Investment prospectus document reference")

	// Mock backend that simulates a CIP-150 compliant contract
	backend := &mockBackend{
		callContractFunc: func(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
			// Mock responses for the CIP-150 interface
			if call.DataBytes[0] == 0xf3 { // hasKey
				// Return true (key exists)
				return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, nil
			} else if call.DataBytes[0] == 0xf2 { // isSealed
				// Return false (not sealed - can be updated)
				return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil
			} else if call.DataBytes[0] == 0xe2 { // getValue
				// Return dynamic string: "Investment Prospectus 2024 - Core Blockchain RWA Token"
				// Offset: 32 (0x20), Length: 47 (0x2F), Data: full string
				// The complete string is 47 characters: "Investment Prospectus 2024 - Core Blockchain RWA Token"
				return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 54, 73, 110, 118, 101, 115, 116, 109, 101, 110, 116, 32, 80, 114, 111, 115, 112, 101, 99, 116, 117, 115, 32, 50, 48, 50, 52, 32, 45, 32, 67, 111, 114, 101, 32, 66, 108, 111, 99, 107, 99, 104, 97, 105, 110, 32, 82, 87, 65, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil
			}
			return nil, fmt.Errorf("unknown function selector")
		},
	}

	api := NewPublicSmartContractAPI(backend)

	// Test 1: Get value with sealed status (default behavior)
	t.Run("Get value with sealed status", func(t *testing.T) {
		result, err := api.GetKV(context.Background(), key, contractAddr, false)

		if err != nil {
			t.Errorf("GetKV failed: %v", err)
			return
		}

		t.Logf("  Result (sealed=false):")
		t.Logf("    Value: %s", result.Value)
		t.Logf("    Sealed: %t", result.Sealed)
		t.Logf("    Exists: %t", result.Exists)

		if !result.Exists {
			t.Errorf("Expected key to exist")
		}
		if result.Sealed {
			t.Errorf("Expected key to not be sealed")
		}
		if result.Value == "" {
			t.Errorf("Expected non-empty value")
		}

		// Verify the value matches our expectation
		expectedValue := "Investment Prospectus 2024 - Core Blockchain RWA Token"
		if result.Value != expectedValue {
			t.Errorf("Expected value '%s', got '%s'", expectedValue, result.Value)
		}
	})

	// Test 2: Get only sealed items (should return nothing since this key is not sealed)
	t.Run("Get only sealed items", func(t *testing.T) {
		result, err := api.GetKV(context.Background(), key, contractAddr, true)

		if err != nil {
			t.Errorf("GetKV failed: %v", err)
			return
		}

		t.Logf("  Result (sealed=true):")
		t.Logf("    Value: %s", result.Value)
		t.Logf("    Sealed: %t", result.Sealed)
		t.Logf("    Exists: %t", result.Exists)

		if !result.Exists {
			t.Errorf("Expected key to exist")
		}
		if result.Sealed {
			t.Errorf("Expected key to not be sealed")
		}
		if result.Value != "" {
			t.Errorf("Expected empty value when requesting sealed items but key is not sealed")
		}
	})
}
