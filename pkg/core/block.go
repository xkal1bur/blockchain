package core

import (
	"crypto/ecdsa"
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

// ValidateBlock validates the block hash and optionally all transactions
// If publicKeyMap is nil, only validates block hash
// If publicKeyMap is provided, validates both hash and all transaction signatures
func (b *Block) ValidateBlock(publicKeyMap map[int][]*ecdsa.PublicKey) bool {
	// First validate the block hash
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

	// If no public keys provided, only validate block hash
	if publicKeyMap == nil {
		fmt.Println("✅ Block hash validation successful")
		return true
	}

	// Check if all transactions inside the block are valid
	for i, tx := range b.Transactions {
		publicKeys, exists := publicKeyMap[i]
		if !exists {
			fmt.Printf("No public keys provided for transaction %d\n", i)
			return false
		}
		if !tx.Verify(publicKeys) {
			fmt.Printf("Invalid transaction %d in block\n", i)
			return false
		}
	}

	fmt.Println("✅ Block validation successful: hash and all transactions are valid")
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
