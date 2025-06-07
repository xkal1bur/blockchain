package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"
	"time"

	"github.com/xkal1bur/blockchain/pkg/core"
	"golang.org/x/crypto/sha3"
)

func TestAliceBobTransactionVerification(t *testing.T) {
	// Generate Alice's private key using our standard curve
	alicePrivateKey, err := ecdsa.GenerateKey(StandardCurve, rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate Alice's private key: %v", err)
	}
	alicePublicKey := &alicePrivateKey.PublicKey

	// Generate Bob's private key using our standard curve
	bobPrivateKey, err := ecdsa.GenerateKey(StandardCurve, rand.Reader)
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
		Version:   1,
		TxIns:     []core.TxIn{txIn1},
		TxOuts:    []core.TxOut{txOutToBob, txOutToAliceChange},
		Locktimei: uint32(time.Now().Unix()),
	}

	// Use SHA3-256 instead of SHA2-256 for transaction hashing
	hasher := sha3.New256()
	hasher.Write([]byte(tx1.ID()))
	txHash := hasher.Sum(nil)

	// ECDSA Digital Signature Process:
	// 1. Generate random nonce k
	// 2. Calculate point (x,y) = k * G (where G is generator point)
	// 3. r = x mod n (where n is curve order)
	// 4. s = k^(-1) * (hash + r * private_key) mod n
	// 5. Signature is (r, s)
	// Use Alice's private key to sign the transaction
	r, s, err := ecdsa.Sign(rand.Reader, alicePrivateKey, txHash)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Combine r and s into a single signature (DER format would be better in production)
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

	hash, err := block.Hash()
	if err != nil {
		t.Fatalf("failed to calculate block hash: %v", err)
	}

	t.Logf("Block hash: %x", hash)
	t.Logf("Transaction signature (r,s): %x", signature)
	t.Logf("Alice's address: %x", aliceHash)
	t.Logf("Bob's address: %x", bobHash)
	t.Logf("TxOut to Bob script: %x", bobHash)
	t.Logf("TxOut to Alice (change) script: %x", aliceHash)

	// Verify ECDSA signature using Alice's public key
	isValid := ecdsa.Verify(alicePublicKey, txHash, r, s)
	t.Logf("Signature verification: %v", isValid)

	// Log the key pairs for reference
	t.Logf("Alice's public key: %x", alicePublicKey.X.Bytes())
	t.Logf("Bob's public key: %x", bobPrivateKey.PublicKey.X.Bytes())
}
