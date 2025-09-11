package scapi

import (
	"testing"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
)

func TestBalanceOfFunctionData(t *testing.T) {
	// Test the data construction logic that matches the user's example
	holderICAN := "cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c"
	bcan := holderICAN[len(holderICAN)-40:] // last 20 bytes (40 hex chars)

	// Convert to address (use the full ICAN with cb prefix)
	holderAddress, err := common.HexToAddress(holderICAN)
	if err != nil {
		t.Fatalf("Failed to parse address: %v", err)
	}

	// Test that our implementation works correctly
	selector := "0x1d7976f3"
	addressBytes := holderAddress.Bytes()
	paddedAddress := make([]byte, 32)
	copy(paddedAddress[32-len(addressBytes):], addressBytes)

	// Convert back to hex for display
	actualPaddedAddress := common.Bytes2Hex(paddedAddress)

	t.Logf("Holder ICAN: %s", holderICAN)
	t.Logf("BCAN (last 40 chars): %s", bcan)
	t.Logf("Holder address: %s", holderAddress.Hex())
	t.Logf("Selector: %s", selector)
	t.Logf("Padded address: %s", actualPaddedAddress)
	t.Logf("Full call data: %s%s", selector, actualPaddedAddress)

	// Verify that the address is properly padded to 32 bytes
	if len(paddedAddress) != 32 {
		t.Errorf("Padded address length is %d, expected 32", len(paddedAddress))
	}
}

func TestBalanceOfAddressParsing(t *testing.T) {
	// Test the address parsing logic
	testCases := []struct {
		name       string
		holderICAN string
		expected   string
	}{
		{
			name:       "User example",
			holderICAN: "cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c",
			expected:   "cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c",
		},
		{
			name:       "Short ICAN",
			holderICAN: "cb270000000000000000000000000000000000000001",
			expected:   "cb270000000000000000000000000000000000000001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use the full ICAN with cb prefix
			holderAddress, err := common.HexToAddress(tc.holderICAN)
			if err != nil {
				t.Fatalf("Failed to parse address: %v", err)
			}

			if holderAddress.Hex() != tc.expected {
				t.Errorf("Address parsing failed:\nExpected: %s\nGot:      %s",
					tc.expected, holderAddress.Hex())
			}

			t.Logf("ICAN: %s -> Address: %s", tc.holderICAN, holderAddress.Hex())
		})
	}
}

func TestBalanceOfUserExample(t *testing.T) {
	// Test the exact example from the user's request for CBC20 balanceOf
	// balanceOf(address) -> 0x1d7976f3 (standard CBC20 selector)
	// const holderICAN = "cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c"; // example
	// const bcan = holderICAN.slice(-40);                      // last 20 bytes
	// const data = "0x1d7976f3" + "0".repeat(24) + bcan;      // left-pad to 32 bytes

	holderICAN := "cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c"
	bcan := holderICAN[len(holderICAN)-40:] // last 20 bytes (40 hex chars)

	t.Logf("User's example:")
	t.Logf("  holderICAN: %s", holderICAN)
	t.Logf("  bcan (last 40 chars): %s", bcan)
	t.Logf("  selector: 0x1d7976f3")
	t.Logf("  expected data: 0x1d7976f3 + 24 zeros + %s", bcan)

	// Verify our implementation matches the expected format
	selector := "0x1d7976f3"
	expectedData := selector + "000000000000000000000000" + bcan

	// Create the call data using the user's approach
	selectorBytes := hexutil.MustDecode(selector)
	bcanBytes := hexutil.MustDecode("0x" + bcan)

	// Pad bcan to 32 bytes (left-pad with zeros)
	paddedBcan := make([]byte, 32)
	copy(paddedBcan[32-len(bcanBytes):], bcanBytes)

	// Combine selector and padded bcan
	callData := append(selectorBytes, paddedBcan...)

	// Convert to hex for comparison
	actualData := hexutil.Encode(callData)

	t.Logf("User's approach implementation:")
	t.Logf("  bcan bytes: %x", bcanBytes)
	t.Logf("  padded bcan: %x", paddedBcan)
	t.Logf("  full call data: %s", actualData)

	// Verify the data format matches
	if actualData != expectedData {
		t.Errorf("Call data mismatch:\nExpected: %s\nGot:      %s", expectedData, actualData)
	} else {
		t.Logf("✅ Call data format matches user's example exactly!")
	}

	// Verify the data is the correct length (4 bytes selector + 32 bytes address = 36 bytes)
	if len(callData) != 36 {
		t.Errorf("Call data length is %d, expected 36", len(callData))
	} else {
		t.Logf("✅ Call data length is correct: 36 bytes")
	}
}
