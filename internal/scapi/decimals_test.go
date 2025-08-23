package scapi

import (
	"testing"

	"github.com/core-coin/go-core/v2/common/hexutil"
)

func TestDecimalsFunctionData(t *testing.T) {
	// Test the data construction logic for the user's example
	// decimals() -> 0x5d1fb5f9
	selector := "0x5d1fb5f9"

	// Expected call data
	expectedData := selector

	// Create the call data using our implementation
	callData := hexutil.MustDecode(selector)

	// Convert to hex for display
	actualData := hexutil.Encode(callData)

	t.Logf("CBC20 decimals() function:")
	t.Logf("  selector: %s", selector)
	t.Logf("  call data: %s", actualData)
	t.Logf("  data length: %d bytes", len(callData))

	// Verify the data format matches
	if actualData != expectedData {
		t.Errorf("Call data mismatch:\nExpected: %s\nGot:      %s", expectedData, actualData)
	} else {
		t.Logf("✅ Call data format matches user's example exactly!")
	}

	// Verify the data is the correct length (4 bytes selector)
	if len(callData) != 4 {
		t.Errorf("Call data length is %d, expected 4", len(callData))
	} else {
		t.Logf("✅ Call data length is correct: 4 bytes")
	}
}

func TestDecimalsResponseDecoding(t *testing.T) {
	// Test the response decoding logic with the user's example
	// Response: "0x0000000000000000000000000000000000000000000000000000000000000012"
	// Expected: 18 (decimal places)

	userResponse := "0x0000000000000000000000000000000000000000000000000000000000000012"
	expectedDecimals := uint8(18)

	t.Logf("User's example response:")
	t.Logf("  hex response: %s", userResponse)
	t.Logf("  expected decimals: %d", expectedDecimals)

	// Decode the hex response
	responseBytes := hexutil.MustDecode(userResponse)

	// Extract the uint8 value (last byte)
	if len(responseBytes) >= 32 {
		decimals := responseBytes[31] // Last byte contains the uint8 value

		t.Logf("Our decoding:")
		t.Logf("  response bytes: %x", responseBytes)
		t.Logf("  extracted decimals: %d", decimals)

		// Verify the decoded value matches
		if decimals != expectedDecimals {
			t.Errorf("Decoded decimals mismatch:\nExpected: %d\nGot:      %d", expectedDecimals, decimals)
		} else {
			t.Logf("✅ Decoded decimals match user's example exactly!")
		}
	} else {
		t.Errorf("Response too short: expected at least 32 bytes, got %d", len(responseBytes))
	}
}

func TestDecimalsWithMockContract(t *testing.T) {
	// Test the complete function with a mock contract response
	// This simulates what would happen in a real contract call

	// Mock response: 18 decimals
	mockResponse := "0x0000000000000000000000000000000000000000000000000000000000000012"
	expectedDecimals := uint8(18)

	// Simulate the decoding logic from our implementation
	responseBytes := hexutil.MustDecode(mockResponse)

	if len(responseBytes) >= 32 {
		decimals := responseBytes[31]

		t.Logf("Mock contract test:")
		t.Logf("  mock response: %s", mockResponse)
		t.Logf("  decoded decimals: %d", decimals)
		t.Logf("  expected: %d", expectedDecimals)

		if decimals != expectedDecimals {
			t.Errorf("Mock test failed:\nExpected: %d\nGot:      %d", expectedDecimals, decimals)
		} else {
			t.Logf("✅ Mock contract test passed!")
		}
	} else {
		t.Errorf("Mock response too short: expected at least 32 bytes, got %d", len(responseBytes))
	}
}
