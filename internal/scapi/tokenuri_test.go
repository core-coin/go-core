package scapi

import (
	"context"
	"math/big"
	"testing"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core/state"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/internal/xcbapi"
	"github.com/core-coin/go-core/v2/rpc"
)

func TestTokenURIFunctionSelector(t *testing.T) {
	t.Logf("Testing CIP-721 function selector for TokenURI:")
	t.Logf("  - tokenURI(uint256): 0xa89da637")

	// Test tokenURI selector
	tokenURISelector := "0xa89da637"
	tokenURIData := hexutil.MustDecode(tokenURISelector)
	t.Logf("  tokenURI selector: %s -> %d bytes", tokenURISelector, len(tokenURIData))

	// Verify selector is 4 bytes
	if len(tokenURIData) != 4 {
		t.Errorf("Expected tokenURI selector to be 4 bytes, got %d", len(tokenURIData))
	}
}

func TestTokenURIDataConstruction(t *testing.T) {
	t.Logf("Testing TokenURI call data construction:")

	// Test token ID
	tokenId := big.NewInt(12345)
	t.Logf("  Token ID: %s", tokenId.String())

	// Expected selector
	expectedSelector := "0xa89da637"
	expectedData := hexutil.MustDecode(expectedSelector)
	t.Logf("  Expected selector: %s -> %d bytes", expectedSelector, len(expectedData))

	// Expected tokenId encoding (32 bytes, big-endian)
	expectedTokenIdBytes := make([]byte, 32)
	tokenId.FillBytes(expectedTokenIdBytes)
	expectedTokenIdHex := hexutil.Encode(expectedTokenIdBytes)
	t.Logf("  Expected tokenId encoding: %s (%d bytes)", expectedTokenIdHex, len(expectedTokenIdBytes))

	// Verify tokenId is properly padded to 32 bytes
	if len(expectedTokenIdBytes) != 32 {
		t.Errorf("Expected tokenId to be padded to 32 bytes, got %d", len(expectedTokenIdBytes))
	}

	// Verify the tokenId is properly encoded (check a few bytes)
	// For tokenId 12345, the last few bytes should contain the value
	expectedValue := big.NewInt(12345)
	if new(big.Int).SetBytes(expectedTokenIdBytes).Cmp(expectedValue) != 0 {
		t.Errorf("Expected tokenId bytes to encode %s, got %s", expectedValue.String(), new(big.Int).SetBytes(expectedTokenIdBytes).String())
	}
}

func TestTokenURIWithMockContract(t *testing.T) {
	t.Logf("Testing TokenURI with mock CoreNFT contract:")

	// Test contract address
	contractAddr, err := common.HexToAddress("cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c")
	if err != nil {
		t.Fatalf("Failed to parse address: %v", err)
	}
	t.Logf("  Contract: %s", contractAddr.Hex())

	// Test token ID
	tokenId := big.NewInt(1)
	t.Logf("  Token ID: %s", tokenId.String())

	// Expected token URI
	expectedURI := "https://example.com/metadata/1.json"
	t.Logf("  Expected URI: %s", expectedURI)

	// Create mock backend
	mockTokenURIBackend := &mockTokenURIBackend{
		callContractFunc: func(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
			// Verify the call data structure
			if len(call.DataBytes) != 36 { // 4 bytes selector + 32 bytes tokenId
				t.Errorf("Expected call data to be 36 bytes, got %d", len(call.DataBytes))
			}

			// Verify selector
			selector := hexutil.Encode(call.DataBytes[:4])
			if selector != "0xa89da637" {
				t.Errorf("Expected selector 0xa89da637, got %s", selector)
			}

			// Verify tokenId encoding
			tokenIdBytes := call.DataBytes[4:]
			decodedTokenId := new(big.Int).SetBytes(tokenIdBytes)
			if decodedTokenId.Cmp(tokenId) != 0 {
				t.Errorf("Expected tokenId %s, got %s", tokenId.String(), decodedTokenId.String())
			}

			// Return mock tokenURI response
			// This is a simplified mock - in reality it would be properly ABI encoded
			return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 35, 104, 116, 116, 112, 115, 58, 47, 47, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109, 47, 109, 101, 116, 97, 100, 97, 116, 97, 47, 49, 46, 106, 115, 111, 110}, nil
		},
	}

	// Create API instance
	api := NewPublicSmartContractAPI(mockTokenURIBackend)

	// Test TokenURI call
	t.Run("Get token URI", func(t *testing.T) {
		result, err := api.TokenURI(context.Background(), contractAddr, tokenId)
		if err != nil {
			t.Fatalf("TokenURI failed: %v", err)
		}

		t.Logf("  Result: %s", result)

		// Basic validation
		if result == "" {
			t.Error("Expected non-empty token URI")
		}
		if result != expectedURI {
			t.Errorf("Expected URI '%s', got '%s'", expectedURI, result)
		}
	})
}

// Mock backend for testing
type mockTokenURIBackend struct {
	callContractFunc func(context.Context, xcbapi.CallMsg, rpc.BlockNumber) ([]byte, error)
}

func (m *mockTokenURIBackend) CallContract(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error) {
	return m.callContractFunc(ctx, call, blockNumber)
}

func (m *mockTokenURIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	return nil, nil, nil
}
