package xcbapi

import (
	"testing"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
)

// TestComposeTransactionArgsStructure verifies the structure of ComposeTransactionArgs
func TestComposeTransactionArgsStructure(t *testing.T) {
	// Test that we can create the args struct
	args := ComposeTransactionArgs{
		From:   common.Address{0x01},
		To:     &common.Address{0x02},
		Amount: (*hexutil.Big)(hexutil.MustDecodeBig("0xde0b6b3a7640000")),
		Sign:   false,
	}

	// Verify the struct was created correctly
	if args.From != (common.Address{0x01}) {
		t.Error("From address not set correctly")
	}
	if args.To == nil || *args.To != (common.Address{0x02}) {
		t.Error("To address not set correctly")
	}
	if args.Amount == nil {
		t.Error("Amount not set correctly")
	}
	if args.Sign != false {
		t.Error("Sign flag not set correctly")
	}
}

// TestComposeTransactionResultStructure verifies the structure of ComposeTransactionResult
func TestComposeTransactionResultStructure(t *testing.T) {
	// Test that we can create the result struct
	result := ComposeTransactionResult{
		Hash:        &common.Hash{0x01},
		Transaction: nil,
	}

	// Verify the struct was created correctly
	if result.Hash == nil {
		t.Error("Hash not set correctly")
	}
	if result.Transaction != nil {
		t.Error("Transaction should be nil")
	}

	// Test with transaction instead of hash
	result2 := ComposeTransactionResult{
		Hash:        nil,
		Transaction: &ComposeTransactionTx{},
	}

	if result2.Hash != nil {
		t.Error("Hash should be nil")
	}
	if result2.Transaction == nil {
		t.Error("Transaction not set correctly")
	}
}
