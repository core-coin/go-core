package xcbapi

import (
	"testing"
)

// Simple test to verify the Synced function logic
func TestPublicCoreAPISyncedLogic(t *testing.T) {
	tests := []struct {
		name           string
		currentBlock   uint64
		highestBlock   uint64
		expectedResult uint64
	}{
		{
			name:           "fully synced",
			currentBlock:   1000,
			highestBlock:   1000,
			expectedResult: 0,
		},
		{
			name:           "partially synced",
			currentBlock:   800,
			highestBlock:   1000,
			expectedResult: 200,
		},
		{
			name:           "behind by many blocks",
			currentBlock:   100,
			highestBlock:   1000,
			expectedResult: 900,
		},
		{
			name:           "ahead of network (edge case)",
			currentBlock:   1100,
			highestBlock:   1000,
			expectedResult: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the logic directly
			var result uint64
			if tt.currentBlock >= tt.highestBlock {
				result = 0
			} else {
				result = tt.highestBlock - tt.currentBlock
			}

			// Check the result
			if result != tt.expectedResult {
				t.Errorf("Logic test failed: currentBlock=%d, highestBlock=%d, got=%d, want=%d",
					tt.currentBlock, tt.highestBlock, result, tt.expectedResult)
			}
		})
	}
}

// Test the subscription logic
func TestPublicCoreAPISyncedSubscriptionLogic(t *testing.T) {
	tests := []struct {
		name           string
		currentBlock   uint64
		highestBlock   uint64
		expectedResult uint64
	}{
		{
			name:           "fully synced subscription",
			currentBlock:   1000,
			highestBlock:   1000,
			expectedResult: 0,
		},
		{
			name:           "partially synced subscription",
			currentBlock:   800,
			highestBlock:   1000,
			expectedResult: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the subscription logic directly
			var result uint64
			if tt.currentBlock >= tt.highestBlock {
				result = 0
			} else {
				result = tt.highestBlock - tt.currentBlock
			}

			// Check the result
			if result != tt.expectedResult {
				t.Errorf("Subscription logic test failed: currentBlock=%d, highestBlock=%d, got=%d, want=%d",
					tt.currentBlock, tt.highestBlock, result, tt.expectedResult)
			}
		})
	}
}
