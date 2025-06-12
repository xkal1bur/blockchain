package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	"github.com/xkal1bur/blockchain/pkg/core"
	"golang.org/x/crypto/sha3"
)

func TestAliceBobTransactionVerification(t *testing.T) {
	// Generate Alice's private key using our standard curve
	alicePrivateKey, err := ecdsa.GenerateKey(core.StandardCurve, rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate Alice's private key: %v", err)
	}
	alicePublicKey := &alicePrivateKey.PublicKey

	// Generate Bob's private key using our standard curve
	bobPrivateKey, err := ecdsa.GenerateKey(core.StandardCurve, rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate Bob's private key: %v", err)
	}
	bobPublicKey := &bobPrivateKey.PublicKey

	bobHash := sha3.Sum256(bobPublicKey.X.Bytes())
	aliceHash := sha3.Sum256(alicePublicKey.X.Bytes())

	// Create transaction input (Alice's previous output that she's spending)
	txIn1 := core.TxIn{
		PrevTx:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		PrevIndex: 0,
		Signature: []byte{}, // Will be filled after signing
		Net:       "main",
	}

	// Create transaction outputs
	// Alice sends 1 BTC to Bob
	txOutToBob := core.TxOut{
		Amount:        100000000, // 1 BTC to Bob (100,000,000 satoshis)
		LockingScript: bobHash[:],
	}

	// Alice keeps 1.5 BTC as change (assuming she had 2.5 BTC input)
	txOutToAliceChange := core.TxOut{
		Amount:        150000000, // 1.5 BTC change back to Alice (150,000,000 satoshis)
		LockingScript: aliceHash[:],
	}

	// Create the transaction: Alice sends to Bob and gets change back
	tx1 := core.Tx{
		Version: 1,
		TxIns:   []core.TxIn{txIn1},
		TxOuts:  []core.TxOut{txOutToBob, txOutToAliceChange},
	}

	// Get the transaction hash for signing (BEFORE adding the signature)
	// This must match what the Verify method uses
	txHash := tx1.GetHashForSigning()

	r, s, err := ecdsa.Sign(rand.Reader, alicePrivateKey, txHash)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	signature := append(r.Bytes(), s.Bytes()...)
	tx1.TxIns[0].Signature = signature

	// Create and test the block
	block := core.Block{
		Version:      1,
		PrevBlock:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Timestamp:    1234567890,
		Nonce:        0,
		Bits:         0,
		Transactions: []core.Tx{tx1},
	}

	publicKeys := []*ecdsa.PublicKey{alicePublicKey}

	isTransactionValid := tx1.Verify(publicKeys)

	if isTransactionValid {
		t.Logf("✅ Transaction verification: PASSED")
	} else {
		t.Errorf("❌ Transaction verification: FAILED")
	}

	publicKeyMap := map[int][]*ecdsa.PublicKey{
		0: {alicePublicKey}, // Transaction 0 (our tx1) needs Alice's key
	}

	isBlockValid := block.ValidateBlock(publicKeyMap)
	if isBlockValid {
		t.Logf("✅ Block validation: PASSED")
	} else {
		t.Errorf("❌ Block validation: FAILED")
	}
}
