package core

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

// Standard curve for all blockchain operations
var StandardCurve = elliptic.P256()

type Wallet struct {
	PrivateKey []byte `json:"private_key"`
	PublicKey  []byte `json:"public_key"`
	Address    string `json:"address"`
	WalletFile string `json:"-"`
}

type WalletData struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
	CreatedAt  string `json:"created_at"`
}

// NewWallet creates a new wallet with generated keys
func NewWallet() (*Wallet, error) {
	// Generate ECDH P-256 key pair
	curve := ecdh.P256()
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	publicKey := privateKey.PublicKey()

	// Get the raw bytes
	privateKeyBytes := privateKey.Bytes()
	publicKeyBytes := publicKey.Bytes()

	// Generate address from public key (simplified - using SHA256 hash)
	address := generateAddress(publicKeyBytes)

	wallet := &Wallet{
		PrivateKey: privateKeyBytes,
		PublicKey:  publicKeyBytes,
		Address:    address,
		WalletFile: "wallet.json",
	}

	return wallet, nil
}

// LoadWallet loads an existing wallet from disk
func LoadWallet(walletFile string) (*Wallet, error) {
	file, err := os.Open(walletFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open wallet file: %v", err)
	}
	defer file.Close()

	var walletData WalletData
	if err := json.NewDecoder(file).Decode(&walletData); err != nil {
		return nil, fmt.Errorf("failed to decode wallet data: %v", err)
	}

	privateKeyBytes, err := hex.DecodeString(walletData.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %v", err)
	}

	publicKeyBytes, err := hex.DecodeString(walletData.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %v", err)
	}

	wallet := &Wallet{
		PrivateKey: privateKeyBytes,
		PublicKey:  publicKeyBytes,
		Address:    walletData.Address,
		WalletFile: walletFile,
	}

	return wallet, nil
}

// SaveToDisk saves the wallet to disk
func (w *Wallet) SaveToDisk() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(w.WalletFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create wallet directory: %v", err)
	}

	walletData := WalletData{
		PrivateKey: hex.EncodeToString(w.PrivateKey),
		PublicKey:  hex.EncodeToString(w.PublicKey),
		Address:    w.Address,
		CreatedAt:  fmt.Sprintf("%d", getCurrentTimestamp()),
	}

	file, err := os.Create(w.WalletFile)
	if err != nil {
		return fmt.Errorf("failed to create wallet file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(walletData); err != nil {
		return fmt.Errorf("failed to encode wallet data: %v", err)
	}

	return nil
}

// GetAddressHex returns the wallet address as hex string
func (w *Wallet) GetAddressHex() string {
	return w.Address
}

// GetPrivateKeyHex returns the private key as hex string
func (w *Wallet) GetPrivateKeyHex() string {
	return hex.EncodeToString(w.PrivateKey)
}

// GetPublicKeyHex returns the public key as hex string
func (w *Wallet) GetPublicKeyHex() string {
	return hex.EncodeToString(w.PublicKey)
}

// SignData signs data with the wallet's private key
func (w *Wallet) SignData(data []byte) ([]byte, error) {
	// For now, we'll return a simple hash-based signature
	// In production, you'd use a proper ECDSA signature
	hash := sha256.Sum256(append(data, w.PrivateKey...))
	return hash[:], nil
}

// SignECDSA signs data with ECDSA and returns signature in r||s format
func (w *Wallet) SignECDSA(data []byte) ([]byte, error) {
	// Get ECDSA private key
	privateKey, err := w.GetECDSAPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get ECDSA private key: %v", err)
	}

	// Sign the data
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %v", err)
	}

	// Combine r and s into a single byte slice (r||s format)
	signature := make([]byte, 64) // 32 bytes for r + 32 bytes for s

	// Pad r and s to 32 bytes each
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Copy r to the first 32 bytes (right-aligned)
	copy(signature[32-len(rBytes):32], rBytes)
	// Copy s to the next 32 bytes (right-aligned)
	copy(signature[64-len(sBytes):64], sBytes)

	return signature, nil
}

// VerifySignature verifies a signature against the wallet's public key
func (w *Wallet) VerifySignature(data, signature []byte) bool {
	// Simple verification for demonstration
	expectedSig, _ := w.SignData(data)
	return hex.EncodeToString(expectedSig) == hex.EncodeToString(signature)
}

// generateAddress creates an address from a public key
func generateAddress(publicKey []byte) string {
	// Full SHA3-256 hash encoded as hex (64 chars)
	hash := sha3.Sum256(publicKey)
	return hex.EncodeToString(hash[:])
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return 1736434567 // Simplified for demo - in production use time.Now().Unix()
}

// WalletExists checks if a wallet file exists
func WalletExists(walletFile string) bool {
	_, err := os.Stat(walletFile)
	return !os.IsNotExist(err)
}

// DisplayWalletInfo prints wallet information
func (w *Wallet) DisplayWalletInfo() {
	fmt.Println("💰 Wallet Information:")
	fmt.Printf("   Address: %s\n", w.Address)
	fmt.Printf("   Public Key: %s\n", w.GetPublicKeyHex())
	fmt.Printf("   Private Key: %s...\n", w.GetPrivateKeyHex()[:16])
	fmt.Printf("   Wallet File: %s\n", w.WalletFile)
}

// GetECDSAPrivateKey returns the wallet's private key as an ECDSA private key
func (w *Wallet) GetECDSAPrivateKey() (*ecdsa.PrivateKey, error) {
	// For ECDSA, we need to convert from ECDH format
	// This is a simplified approach - in production you'd store the key in ECDSA format directly
	curve := ecdh.P256()
	ecdhPrivateKey, err := curve.NewPrivateKey(w.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDH private key: %v", err)
	}

	// Get the raw private key bytes and create ECDSA private key
	// This is a workaround since we can't directly convert ECDH to ECDSA
	privateKeyInt := new(big.Int).SetBytes(w.PrivateKey)

	// Extract X and Y coordinates from ECDH public key
	pubKeyBytes := ecdhPrivateKey.PublicKey().Bytes()
	x := new(big.Int).SetBytes(pubKeyBytes[1:33]) // Skip first byte (0x04)
	y := new(big.Int).SetBytes(pubKeyBytes[33:])  // Y coordinate

	ecdsaPrivateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: StandardCurve,
			X:     x,
			Y:     y,
		},
		D: privateKeyInt,
	}

	return ecdsaPrivateKey, nil
}

// GetECDSAPublicKey returns the wallet's public key as an ECDSA public key
func (w *Wallet) GetECDSAPublicKey() (*ecdsa.PublicKey, error) {
	privateKey, err := w.GetECDSAPrivateKey()
	if err != nil {
		return nil, err
	}
	return &privateKey.PublicKey, nil
}

// GetPublicKeyData returns the public key in the format expected by the server
func (w *Wallet) GetPublicKeyData() (PublicKeyData, error) {
	publicKey, err := w.GetECDSAPublicKey()
	if err != nil {
		return PublicKeyData{}, err
	}

	return PublicKeyData{
		X: hex.EncodeToString(publicKey.X.Bytes()),
		Y: hex.EncodeToString(publicKey.Y.Bytes()),
	}, nil
}

// GetLockingScript returns the wallet's locking script (SHA3-256 of its public key)
func (w *Wallet) GetLockingScript() []byte {
	return []byte(w.Address)
}

// FindSpendableUTXOs selects UTXOs from utxoSet belonging to this wallet until amount is reached
// Returns slice of keys and total amount gathered
func (w *Wallet) FindSpendableUTXOs(utxoSet map[string]TxOut, amount uint64) ([]string, uint64, error) {
	var selected []string
	var total uint64
	script := []byte(w.Address)

	for key, out := range utxoSet {
		if bytes.Equal(out.LockingScript, script) {
			selected = append(selected, key)
			total += out.Amount
			if total >= amount {
				break
			}
		}
	}

	if total < amount {
		return nil, 0, fmt.Errorf("insufficient funds: needed %d, available %d", amount, total)
	}
	return selected, total, nil
}

// BuildTransaction builds and signs a transaction sending 'amount' to destPubKey.
// utxoSet is required to choose inputs. Returns the transaction and the keys used.
func (w *Wallet) BuildTransaction(destPubKey []byte, amount uint64, utxoSet map[string]TxOut) (Tx, []string, error) {
	inputKeys, totalIn, err := w.FindSpendableUTXOs(utxoSet, amount)
	if err != nil {
		return Tx{}, nil, err
	}

	tx := Tx{Version: 1}

	// Create inputs
	for _, key := range inputKeys {
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			return Tx{}, nil, fmt.Errorf("invalid utxo key %s", key)
		}
		txidHex := parts[0]
		idxStr := parts[1]
		idxParsed, _ := strconv.Atoi(idxStr)

		txidBytes, err := hex.DecodeString(txidHex)
		if err != nil {
			return Tx{}, nil, fmt.Errorf("invalid txid hex: %v", err)
		}

		tx.TxIns = append(tx.TxIns, TxIn{
			PrevTx:    txidBytes,
			PrevIndex: uint32(idxParsed),
			Signature: []byte{},
			PubKey:    w.PublicKey,
			Net:       "mainnet",
		})
	}

	// Outputs: recipient + change (if any)
	tx.TxOuts = append(tx.TxOuts, TxOut{
		Amount:        amount,
		LockingScript: []byte(hex.EncodeToString(HashSHA3(destPubKey))),
	})

	change := totalIn - amount
	if change > 0 {
		tx.TxOuts = append(tx.TxOuts, TxOut{
			Amount:        change,
			LockingScript: w.GetLockingScript(),
		})
	}

	// Sign inputs
	hashForSign := tx.GetHashForSigning()
	sig, err := w.SignECDSA(hashForSign)
	if err != nil {
		return Tx{}, nil, err
	}

	for i := range tx.TxIns {
		tx.TxIns[i].Signature = sig
	}

	return tx, inputKeys, nil
}

// BuildTransactionToAddress creates a tx sending 'amount' to a destination address string (locking script = address bytes)
func (w *Wallet) BuildTransactionToAddress(destAddress string, amount uint64, utxoSet map[string]TxOut) (Tx, []string, error) {
	inputKeys, totalIn, err := w.FindSpendableUTXOs(utxoSet, amount)
	if err != nil {
		return Tx{}, nil, err
	}

	tx := Tx{Version: 1}

	// inputs
	for _, key := range inputKeys {
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			return Tx{}, nil, fmt.Errorf("invalid utxo key %s", key)
		}
		txidBytes, _ := hex.DecodeString(parts[0])
		idx, _ := strconv.Atoi(parts[1])

		tx.TxIns = append(tx.TxIns, TxIn{
			PrevTx:    txidBytes,
			PrevIndex: uint32(idx),
			PubKey:    w.PublicKey,
			Signature: []byte{},
			Net:       "mainnet",
		})
	}

	// outputs: destination + change
	tx.TxOuts = append(tx.TxOuts, TxOut{Amount: amount, LockingScript: []byte(destAddress)})
	change := totalIn - amount
	if change > 0 {
		tx.TxOuts = append(tx.TxOuts, TxOut{Amount: change, LockingScript: w.GetLockingScript()})
	}

	// sign
	hash := tx.GetHashForSigning()
	sig, err := w.SignECDSA(hash)
	if err != nil {
		return Tx{}, nil, err
	}
	for i := range tx.TxIns {
		tx.TxIns[i].Signature = sig
	}
	return tx, inputKeys, nil
}

// FilterUTXOs returns a subset of utxoSet that belong to this wallet
func (w *Wallet) FilterUTXOs(utxoSet map[string]TxOut) map[string]TxOut {
	res := make(map[string]TxOut)
	script := []byte(w.Address)
	for k, v := range utxoSet {
		if bytes.Equal(v.LockingScript, script) {
			res[k] = v
		}
	}
	return res
}

// SaveUTXOs saves given utxo map to filename in JSON format
func SaveUTXOs(filename string, utxoMap map[string]TxOut) error {
	data, err := json.MarshalIndent(utxoMap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
