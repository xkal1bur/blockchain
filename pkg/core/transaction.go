package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// ------------------------------------------------------

type Tx struct {
	Version uint32
	TxIns   []TxIn
	TxOuts  []TxOut
}
type TxIn struct {
	PrevTx    []byte // ID de la transacción previa
	PrevIndex uint32 // Índice de la salida que se gasta
	Signature []byte // Firma r||s del dueño de la salida
	PubKey    []byte // Clave pública en formato no comprimido (0x04 + 64 bytes)
	Net       string // Red
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

	// Hash the serialized data with SHA3-256
	hash := sha3.Sum256(buf.Bytes())

	// Return as hex string
	return hex.EncodeToString(hash[:])
}

// HashSHA3 devuelve SHA3-256(data)
func HashSHA3(data []byte) []byte {
	sum := sha3.Sum256(data)
	return sum[:]
}

// ParsePubKeySafe convierte bytes (0x04 + 64) a *ecdsa.PublicKey y valida que esté en la curva
func ParsePubKeySafe(pub []byte) (*ecdsa.PublicKey, error) {
	if len(pub) != 65 || pub[0] != 0x04 {
		return nil, errors.New("clave pública debe estar en formato no comprimido (0x04 + 64 bytes)")
	}

	x := new(big.Int).SetBytes(pub[1:33])
	y := new(big.Int).SetBytes(pub[33:])

	curve := elliptic.P256()
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("la clave pública no está en la curva")
	}

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

// Validate ejecuta la verificación completa usando las transacciones previas (prevTxs)
// prevTxs es un mapa txID(hex) → *Tx
func (tx *Tx) Validate(prevTxs map[string]*Tx) bool {
	msg := tx.GetHashForSigning()

	for i, txin := range tx.TxIns {
		// 1. Obtener la transacción previa
		prevTxID := hex.EncodeToString(txin.PrevTx)
		prevTx, ok := prevTxs[prevTxID]
		if !ok {
			fmt.Printf("❌ Transacción previa %s no encontrada\n", prevTxID)
			return false
		}

		// 2. Verificar índice válido
		if int(txin.PrevIndex) >= len(prevTx.TxOuts) {
			fmt.Printf("❌ Índice inválido en input #%d\n", i)
			return false
		}
		prevOut := prevTx.TxOuts[txin.PrevIndex]

		// 3. Comparar HashSHA3(pubkey) con LockingScript
		pubKeyHash := HashSHA3(txin.PubKey)
		if !bytes.Equal(pubKeyHash, prevOut.LockingScript) {
			fmt.Printf("❌ El hash del pubkey no coincide con el LockingScript\n")
			return false
		}

		// 4. Verificar la firma
		pubKey, err := ParsePubKeySafe(txin.PubKey)
		if err != nil {
			fmt.Printf("❌ Error al parsear pubkey: %v\n", err)
			return false
		}
		if len(txin.Signature) != 64 {
			fmt.Printf("❌ Firma inválida (esperado 64 bytes, got %d)\n", len(txin.Signature))
			return false
		}
		r := new(big.Int).SetBytes(txin.Signature[:32])
		s := new(big.Int).SetBytes(txin.Signature[32:])
		if !ecdsa.Verify(pubKey, msg, r, s) {
			fmt.Printf("❌ Firma inválida en input #%d\n", i)
			return false
		}
	}

	fmt.Println("✅ Transacción válida")
	return true
}

// bytesToECDSAPublicKey is now superseded by ParsePubKeySafe; left for compatibility if needed.
func bytesToECDSAPublicKey(data []byte) (*ecdsa.PublicKey, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty public key bytes")
	}

	// Public key debe estar en formato sin comprimir (65 bytes) y prefijo 0x04
	if len(data) != 65 || data[0] != 0x04 {
		return nil, fmt.Errorf("invalid uncompressed public key format; expected 65 bytes starting with 0x04")
	}

	x := new(big.Int).SetBytes(data[1:33])
	y := new(big.Int).SetBytes(data[33:])

	return &ecdsa.PublicKey{
		Curve: StandardCurve,
		X:     x,
		Y:     y,
	}, nil
}

// GetHashForSigning returns the transaction hash used for signing
func (tx *Tx) GetHashForSigning() []byte {
	// Create a deep copy of the transaction without signatures for hashing
	txCopy := *tx
	// Deep copy the TxIns slice
	txCopy.TxIns = make([]TxIn, len(tx.TxIns))
	for i, txIn := range tx.TxIns {
		txCopy.TxIns[i] = txIn
		txCopy.TxIns[i].Signature = []byte{} // Clear signatures in the copy only
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

	hash := sha3.Sum256(buf.Bytes())
	return hash[:]
}
