package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// Standard curve for all blockchain operations
var StandardCurve = elliptic.P256()

// ------------------------------------------------------

type TxFetcher struct{ CacheDir string }
type Tx struct {
	Version   uint32
	TxIns     []TxIn
	TxOuts    []TxOut
	Locktimei uint32
}
type TxIn struct {
	PrevTx    []byte
	PrevIndex uint32
	Signature []byte
	Net       string
}
type TxOut struct {
	Amount        uint64
	LockingScript []byte
}

func (tx *Tx) ID() string {
	// Serialize the transaction data
	var buf bytes.Buffer

	// Write version
	binary.Write(&buf, binary.LittleEndian, tx.Version)

	// Write number of inputs
	binary.Write(&buf, binary.LittleEndian, uint32(len(tx.TxIns)))

	// Write each input
	for _, txIn := range tx.TxIns {
		buf.Write(txIn.PrevTx)
		binary.Write(&buf, binary.LittleEndian, txIn.PrevIndex)
		binary.Write(&buf, binary.LittleEndian, uint32(len(txIn.Signature)))
		buf.Write(txIn.Signature)
		buf.WriteString(txIn.Net)
	}

	// Write number of outputs
	binary.Write(&buf, binary.LittleEndian, uint32(len(tx.TxOuts)))

	// Write each output
	for _, txOut := range tx.TxOuts {
		binary.Write(&buf, binary.LittleEndian, txOut.Amount)
		binary.Write(&buf, binary.LittleEndian, uint32(len(txOut.LockingScript)))
		buf.Write(txOut.LockingScript)
	}

	// Write locktime
	binary.Write(&buf, binary.LittleEndian, tx.Locktimei)

	// Hash the serialized data with SHA3-256
	hash := sha3.Sum256(buf.Bytes())

	// Return as hex string
	return hex.EncodeToString(hash[:])
}

///////////////

// Verify validates the transaction structure and signatures
func (tx *Tx) Verify(publicKeys []*ecdsa.PublicKey) bool {
	// Verify signatures
	if err := tx.verifySignatures(publicKeys); err != nil {
		return false
	}

	// // Validate script consistency
	// if err := tx.validateScripts(publicKeys); err != nil {
	// 	return false
	// }

	return true
}

// verifySignatures validates all input signatures
func (tx *Tx) verifySignatures(publicKeys []*ecdsa.PublicKey) error {
	if len(publicKeys) != len(tx.TxIns) {
		return fmt.Errorf("number of public keys (%d) must match number of inputs (%d)",
			len(publicKeys), len(tx.TxIns))
	}

	// Get transaction hash for signature verification
	txHash := tx.getHashForSigning()

	for i, txIn := range tx.TxIns {
		if len(txIn.Signature) == 0 {
			return fmt.Errorf("input %d has no signature", i)
		}

		// Parse signature (assuming it's r||s format)
		if len(txIn.Signature) != 64 { // 32 bytes for r + 32 bytes for s
			return fmt.Errorf("input %d has invalid signature length", i)
		}

		r := new(big.Int).SetBytes(txIn.Signature[:32])
		s := new(big.Int).SetBytes(txIn.Signature[32:])

		// Verify signature
		if !ecdsa.Verify(publicKeys[i], txHash, r, s) {
			return fmt.Errorf("input %d has invalid signature", i)
		}
	}

	return nil
}

// // validateScripts ensures output scripts match the expected public key hashes
// func (tx *Tx) validateScripts(publicKeys []*ecdsa.PublicKey) error {
// 	for i, txOut := range tx.TxOuts {
// 		if len(txOut.LockingScript) != 32 { // SHA3-256 hash length
// 			return fmt.Errorf("output %d has invalid locking script length", i)
// 		}

// 		// For validation purposes, we assume the locking script should match
// 		// the SHA3 hash of one of the provided public keys' X coordinates
// 		scriptMatched := false
// 		for _, pubKey := range publicKeys {
// 			expectedScript := sha3.Sum256(pubKey.X.Bytes())
// 			if bytes.Equal(txOut.LockingScript, expectedScript[:]) {
// 				scriptMatched = true
// 				break
// 			}
// 		}

// 		if !scriptMatched {
// 			return fmt.Errorf("output %d locking script does not match any provided public key", i)
// 		}
// 	}

// 	return nil
// }

// getHashForSigning returns the transaction hash used for signing
func (tx *Tx) getHashForSigning() []byte {
	// Create a copy of the transaction without signatures for hashing
	txCopy := *tx
	for i := range txCopy.TxIns {
		txCopy.TxIns[i].Signature = []byte{} // Clear signatures
	}

	// Serialize and hash
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, txCopy.Version)
	binary.Write(&buf, binary.LittleEndian, uint32(len(txCopy.TxIns)))

	for _, txIn := range txCopy.TxIns {
		buf.Write(txIn.PrevTx)
		binary.Write(&buf, binary.LittleEndian, txIn.PrevIndex)
		buf.WriteString(txIn.Net)
	}

	binary.Write(&buf, binary.LittleEndian, uint32(len(txCopy.TxOuts)))
	for _, txOut := range txCopy.TxOuts {
		binary.Write(&buf, binary.LittleEndian, txOut.Amount)
		binary.Write(&buf, binary.LittleEndian, uint32(len(txOut.LockingScript)))
		buf.Write(txOut.LockingScript)
	}

	binary.Write(&buf, binary.LittleEndian, txCopy.Locktimei)

	hash := sha3.Sum256(buf.Bytes())
	return hash[:]
}
