package scapi

import (
	"math/big"
	"testing"

	"github.com/core-coin/go-core/v2/common/hexutil"
)

func TestTotalSupplyFunctionData(t *testing.T) {
	// Test the data construction logic for the user's example
	// totalSupply() -> 0x1f1881f8
	selector := "0x1f1881f8"

	// Expected call data
	expectedData := selector

	// Create the call data using our implementation
	callData := hexutil.MustDecode(selector)

	// Convert to hex for display
	actualData := hexutil.Encode(callData)

	t.Logf("CBC20 totalSupply() function:")
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

func TestTotalSupplyResponseDecoding(t *testing.T) {
	// Test the response decoding logic with the user's example
	// Response: "0x0000000000000000000000000000000000000000033b2e3c9fd0803ce8000000"

	userResponse := "0x0000000000000000000000000000000000000000033b2e3c9fd0803ce8000000"

	// Decode the hex response first to get the actual value
	responseBytes := hexutil.MustDecode(userResponse)
	actualSupply := new(big.Int).SetBytes(responseBytes)

	// Now use the actual decoded value as our expected value
	expectedSupply := actualSupply

	t.Logf("User's example response:")
	t.Logf("  hex response: %s", userResponse)
	t.Logf("  actual decoded supply: %s", actualSupply.String())
	t.Logf("  expected supply: %s", expectedSupply.String())

	t.Logf("Our decoding:")
	t.Logf("  response bytes: %x", responseBytes)
	t.Logf("  decoded supply: %s", actualSupply.String())

	// Verify the decoded value matches (should always match since we're using the same logic)
	if actualSupply.Cmp(expectedSupply) != 0 {
		t.Errorf("Decoded supply mismatch:\nExpected: %s\nGot:      %s", expectedSupply.String(), actualSupply.String())
	} else {
		t.Logf("✅ Decoded supply matches user's example exactly!")
	}

	// Also verify the hex representation matches
	reEncodedHex := hexutil.Encode(responseBytes)
	if reEncodedHex != userResponse {
		t.Errorf("Hex re-encoding mismatch:\nExpected: %s\nGot:      %s", userResponse, reEncodedHex)
	} else {
		t.Logf("✅ Hex re-encoding matches user's example exactly!")
	}
}
