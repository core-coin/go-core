package scapi

import (
	"testing"
)

func TestDecodeDynString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "bytes32 fallback",
			input:    "0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000343544e0000000000000000000000000000000000000000000000000000000000",
			expected: "CTN",
			hasError: false,
		},
		{
			name:     "simple bytes32",
			input:    "0x414243440000000000000000000000000000000000000000000000000000000000",
			expected: "ABCD",
			hasError: false,
		},
		{
			name:     "empty string",
			input:    "0x0000000000000000000000000000000000000000000000000000000000000000",
			expected: "",
			hasError: false,
		},
		{
			name:     "short input",
			input:    "0x1234",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeDynString(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

func TestDecodeDynStringWithUserExample(t *testing.T) {
	// This is the exact example from the user's request
	userExample := "0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000343544e0000000000000000000000000000000000000000000000000000000000"

	result, err := decodeDynString(userExample)
	if err != nil {
		t.Fatalf("Failed to decode user example: %v", err)
	}

	expected := "CTN"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	t.Logf("Successfully decoded '%s' from hex: %s", result, userExample)
}
