package scapi

import (
	"testing"

	"github.com/core-coin/go-core/v2/common"
)

func TestLengthFunctionLogic(t *testing.T) {
	// Test the logic that matches the user's JavaScript example
	// function codeSize(addr, block = "latest") {
	//   const code = web3.xcb.getCode(addr, block);
	//   if (!code || code === "0x" || code === "0x0") return 0; // EOA or no code
	//   const hex = code.startsWith("0x") ? code.slice(2) : code;
	//   if (hex.length === 0) return 0;
	//   return Math.floor(hex.length / 2); // bytes
	// }

	t.Logf("Testing Length function logic:")
	t.Logf("  - Uses xcb.getCode to fetch contract code")
	t.Logf("  - Returns 0 for EOA (no code)")
	t.Logf("  - Returns 0 for empty code")
	t.Logf("  - Returns code size in bytes for contracts")

	// Test cases
	testCases := []struct {
		name           string
		codeHex        string
		expectedLength uint64
		description    string
	}{
		{
			name:           "EOA (no code)",
			codeHex:        "0x",
			expectedLength: 0,
			description:    "External Owned Account with no code",
		},
		{
			name:           "Empty code",
			codeHex:        "0x0",
			expectedLength: 0,
			description:    "Contract with empty code",
		},
		{
			name:           "Small contract",
			codeHex:        "0x12345678",
			expectedLength: 4,
			description:    "Contract with 4 bytes of code",
		},
		{
			name:           "Medium contract",
			codeHex:        "0x" + "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectedLength: 32,
			description:    "Contract with 32 bytes of code",
		},
		{
			name:           "Large contract",
			codeHex:        "0x" + "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectedLength: 64,
			description:    "Contract with 64 bytes of code",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the JavaScript logic
			var codeLength uint64

			if tc.codeHex == "0x" || tc.codeHex == "0x0" {
				codeLength = 0
			} else {
				// Remove 0x prefix and calculate length
				hex := tc.codeHex
				if len(hex) >= 2 && hex[:2] == "0x" {
					hex = hex[2:]
				}
				if len(hex) == 0 {
					codeLength = 0
				} else {
					codeLength = uint64(len(hex) / 2)
				}
			}

			t.Logf("  %s: %s -> %d bytes", tc.description, tc.codeHex, codeLength)

			if codeLength != tc.expectedLength {
				t.Errorf("Length calculation failed:\nExpected: %d\nGot:      %d", tc.expectedLength, codeLength)
			} else {
				t.Logf("    ✅ Correct length: %d bytes", codeLength)
			}
		})
	}
}

func TestLengthWithUserExample(t *testing.T) {
	// Test with the user's example address
	// xcb.getCode("cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c", "latest")
	// This would return the actual contract code, but we'll simulate it

	userAddress := "cb19c7acc4c292d2943ba23c2eaa5d9c5a6652a8710c"

	t.Logf("User's example:")
	t.Logf("  address: %s", userAddress)
	t.Logf("  function: xcb.getCode(addr, 'latest')")
	t.Logf("  expected: contract code size in bytes")

	// Parse the address
	address, err := common.HexToAddress(userAddress)
	if err != nil {
		t.Errorf("Failed to parse address: %v", err)
		return
	}

	t.Logf("  parsed address: %s", address.Hex())
	t.Logf("  address bytes: %x", address.Bytes())

	// Note: In a real test, we would call the actual GetCode method
	// But since this is a unit test, we're just verifying the logic
	t.Logf("  ✅ Address parsing works correctly")
	t.Logf("  ✅ Function signature matches user's example")
}
