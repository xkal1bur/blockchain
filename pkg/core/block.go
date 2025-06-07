package core

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/xkal1bur/blockchain/pkg/crypto"
)

type Block struct {
	Version      uint64 `json:"version"`
	PrevBlock    []byte `json:"prev_block"` // 32 bytes, ojalá
	Timestamp    uint64 `json:"timestamp"`
	Nonce        uint64 `json:"nonce"`
	Bits         uint64 `json:"bits"`         //  difficulty?
	Transactions []Tx   `json:"transactions"` // Transactions in the block
}

// Get block ID
func (b *Block) Hash() ([]byte, error) {
	blockBytes, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize (marshal) block: %v", err)
	}
	return crypto.Sha3_256(blockBytes), nil
}

// Validate the block's hash against its difficulty target
func (b *Block) ValidateBlock() bool {
	hash, err := b.Hash()
	if err != nil {
		fmt.Println("Error calculating block hash:", err)
		return false
	}
	hashInt := new(big.Int).SetBytes(hash)
	target := new(big.Int).SetUint64(b.Bits)
	return hashInt.Cmp(target) == -1

}

// ToDo: Append block to a blockchain file . How to store its hash?
