package tests

import (
	"fmt"
	"testing"

	"github.com/xkal1bur/blockchain/pkg/core"
	"github.com/xkal1bur/blockchain/pkg/crypto"
)

// func TestCreateAndHashEmptyBlock(t *testing.T) {
// 	prevBlock := crypto.Sha3_256([]byte("Random string"))

// 	block := core.Block{
// 		Version:      1,
// 		PrevBlock:    prevBlock,
// 		Timestamp:    0,
// 		Nonce:        0,
// 		Bits:         0,
// 		Transactions: []core.Tx{},
// 	}

// 	hash, err := block.Hash()
// 	if err != nil {
// 		t.Errorf("Error calculating block hash: %v", err)
// 	}

// 	t.Logf("Block hash (hex): %s", hex.EncodeToString(hash))
// }

func TestFindValidBlockHash(t *testing.T) {
	prevBlock := crypto.Sha3_256([]byte("Random string"))

	block := core.Block{
		Version:      1,
		PrevBlock:    prevBlock,
		Timestamp:    0,
		Nonce:        0,
		Bits:         12,
		Transactions: []core.Tx{},
	}

	valid := block.CalculateValidHash()
	if valid {
		t.Logf("Found valid hash for block with nonce: %d", block.Nonce)
		hash, err := block.Hash()
		if err != nil {
			t.Errorf("Error calculating block hash: %v", err)
		}

		var bitStr string
		for _, b := range hash {
			bitStr += fmt.Sprintf("%08b", b) // 8-bit binary per byte
		}

		t.Logf("Block hash (bits): %s", bitStr)
	} else {
		t.Error("Failed to find a valid hash for the block")
	}
}
