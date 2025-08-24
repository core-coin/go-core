package scapi

import (
	"context"
	"testing"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core/state"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/internal/xcbapi"
	"github.com/core-coin/go-core/v2/rpc"
)

func TestListKVFunctionSelectors(t *testing.T) {
	t.Logf("Testing CIP-150 function selectors for ListKV:")
	t.Logf("  - count(): 0x2d7d47f2")
	t.Logf("  - listKeys(): 0xe4d90ad0")

	// Test count selector
	countSelector := "0x2d7d47f2"
	countData := hexutil.MustDecode(countSelector)
	t.Logf("  count selector: %s -> %d bytes", countSelector, len(countData))

	// Test listKeys selector
	listKeysSelector := "0xe4d90ad0"
	listKeysData := hexutil.MustDecode(listKeysSelector)
	t.Logf("  listKeys selector: %s -> %d bytes", listKeysSelector, len(listKeysData))

	// Verify selectors are 4 bytes each
	if len(countData) != 4 {
		t.Errorf("Expected count selector to be 4 bytes, got %d", len(countData))
	}
	if len(listKeysData) != 4 {
		t.Errorf("Expected listKeys selector to be 4 bytes, got %d", len(listKeysData))
	}
}

func TestListKVBasicFunctionality(t *testing.T) {
	t.Logf("Testing ListKV basic functionality:")

	// Test contract address
	contractAddr, err := common.HexToAddress("cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c")
	if err != nil {
		t.Fatalf("Failed to parse address: %v", err)
	}
	t.Logf("  Contract: %s", contractAddr.Hex())

	// Create mock backend that returns simple responses
	mockListKVBackend := &mockListKVBackend{
		callContractFunc: func(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
			// Mock responses for the CIP-150 interface
			if call.DataBytes[0] == 0x2d { // count
				// Return 0 keys (empty contract)
				return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil
			}
			// For other calls, return empty response
			return []byte{}, nil
		},
	}

	// Create API instance
	api := NewPublicSmartContractAPI(mockListKVBackend)

	// Test ListKV with empty contract
	t.Run("Empty contract", func(t *testing.T) {
		result, err := api.ListKV(context.Background(), contractAddr, false)
		if err != nil {
			t.Fatalf("ListKV failed: %v", err)
		}

		t.Logf("  Result: count=%d, keys=%v", result.Count, result.Keys)

		// Basic validation
		if result.Count != 0 {
			t.Errorf("Expected count 0, got %d", result.Count)
		}
		if len(result.Keys) != 0 {
			t.Errorf("Expected empty keys array, got %d keys", len(result.Keys))
		}
		if len(result.Sealed) != 0 {
			t.Errorf("Expected empty sealed array, got %d items", len(result.Sealed))
		}
		if len(result.Values) != 0 {
			t.Errorf("Expected empty values array, got %d items", len(result.Values))
		}
	})
}

// Mock backend for testing
type mockListKVBackend struct {
	callContractFunc func(context.Context, xcbapi.CallMsg, rpc.BlockNumber) ([]byte, error)
}

func (m *mockListKVBackend) CallContract(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
	return m.callContractFunc(ctx, call, blockNumber)
}

func (m *mockListKVBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	return nil, nil, nil
}
