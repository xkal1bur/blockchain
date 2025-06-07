package core

import (
	"encoding/json"
	"fmt"
	"math/bits"

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

// Genesis block
func CreateGenesisBlockWithCoins() *Block {
	// Create a genesis transaction that creates the first coins
	genesisTx := Tx{
		// Your transaction structure here
		// This would typically create coins for initial addresses
		// Does not have inputs, as it's the first transaction
	}

	return &Block{
		Version:      1,
		PrevBlock:    make([]byte, 32), // 000...0
		Timestamp:    1717699200,
		Nonce:        0,
		Bits:         1,
		Transactions: []Tx{genesisTx}, // Include the coin creation transaction
	}
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
func (b *Block) CalculateValidHash() bool {
	for nonce := uint64(0); nonce < ^uint64(0); nonce++ {
		b.Nonce = nonce
		hash, err := b.Hash()
		if err != nil {
			fmt.Println("Error hashing block:", err)
			return false
		}
		if countLeadingZeroBits(hash) >= int(b.Bits) {
			return true
		}
	}
	return false
}

func (b *Block) ValidateBlock() bool {
	hash, err := b.Hash()
	if err != nil {
		fmt.Println("Error hashing block:", err)
		return false
	}
	// Check if the hash meets the difficulty target
	isHashValid := countLeadingZeroBits(hash) >= int(b.Bits)
	if !isHashValid {
		fmt.Println("Block hash does not meet difficulty target.")
		return false
	}
	// Check if the transactions inside the block are valid
	for _, tx := range b.Transactions {
		if !tx.Verify() {
			fmt.Println("Invalid transaction in block")
			return false
		}
	}
	// Check if the previous block hash is valid (if applicable)
	// This part of the code should read the previous block's hash from the blockchain
	return true
}

func countLeadingZeroBits(hash []byte) int {
	count := 0
	for _, b := range hash {
		if b == 0 {
			count += 8
		} else {
			count += bits.LeadingZeros8(b)
			break
		}
	}
	return count
}

// ToDo: Append block to a blockchain file . How to store its hash?
