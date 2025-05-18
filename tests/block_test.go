package tests

import (
	"bytes"
	"testing"

	"github.com/xkal1bur/blockchain/pkg/core"
	"github.com/xkal1bur/blockchain/pkg/crypto"
)

func TestEncodeAndDecodeBlock(t *testing.T) {
	// temp variables
	prevBlock := crypto.Sha256Edu([]byte("Random string"))
	merkleRoot := crypto.Sha256Edu([]byte("Not so random string"))
	// Create a sample block
	block := core.Block{
		Version:    1,
		PrevBlock:  prevBlock[:],
		MerkleRoot: merkleRoot[:],
		Timestamp:  1234567890,
		Nonce:      0,
	}

	t.Logf("Decoded Block: Version: %d, Nonce: %d, Timestamp: %d, PrevBlock: %x, MerkleRoot: %x\n\n",
		block.Version, block.Nonce, block.Timestamp, block.PrevBlock, block.MerkleRoot)

	// Encode the block
	encodedBlock, err := core.EncodeBlock(&block)
	if err != nil {
		t.Fatalf("Failed to encode block: %v", err)
	} else {
		t.Logf("Encoded block: %x\n\n", encodedBlock)
	}

	// Now decode the block
	decodedBlock, err := core.DecodeBlock(encodedBlock)
	if err != nil {
		t.Fatalf("Failed to decode block: %v", err)
	}

	// Check if the decoded block matches the original block
	if block.Version != decodedBlock.Version {
		t.Errorf("Expected version %d, got %d", block.Version, decodedBlock.Version)
	}
	if block.Nonce != decodedBlock.Nonce {
		t.Errorf("Expected nonce %d, got %d", block.Nonce, decodedBlock.Nonce)
	}
	if block.Timestamp != decodedBlock.Timestamp {
		t.Errorf("Expected timestamp %d, got %d", block.Timestamp, decodedBlock.Timestamp)
	}
	if !bytes.Equal(block.PrevBlock, decodedBlock.PrevBlock) {
		t.Errorf("Expected PrevBlock %x, got %x", block.PrevBlock, decodedBlock.PrevBlock)
	}
	if !bytes.Equal(block.MerkleRoot, decodedBlock.MerkleRoot) {
		t.Errorf("Expected MerkleRoot %x, got %x", block.MerkleRoot, decodedBlock.MerkleRoot)
	}
	// Print the values of the decoded block
	t.Logf("Decoded Block: Version: %d, Nonce: %d, Timestamp: %d, PrevBlock: %x, MerkleRoot: %x\n",
		decodedBlock.Version, decodedBlock.Nonce, decodedBlock.Timestamp, decodedBlock.PrevBlock, decodedBlock.MerkleRoot)
}
