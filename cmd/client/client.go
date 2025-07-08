package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/xkal1bur/blockchain/pkg/core"
)

func main() {
	// Create or load wallet
	fmt.Println("🔑 Initializing wallet...")
	var wallet *core.Wallet
	var err error

	walletFile := "client_wallet.json"
	if core.WalletExists(walletFile) {
		fmt.Println("📂 Loading existing wallet...")
		wallet, err = core.LoadWallet(walletFile)
		if err != nil {
			fmt.Printf("Error loading wallet: %v\n", err)
			return
		}
	} else {
		fmt.Println("🆕 Creating new wallet...")
		wallet, err = core.NewWallet()
		if err != nil {
			fmt.Printf("Error creating wallet: %v\n", err)
			return
		}
		wallet.WalletFile = walletFile

		// Save the new wallet
		if err := wallet.SaveToDisk(); err != nil {
			fmt.Printf("Error saving wallet: %v\n", err)
			return
		}
		fmt.Println("💾 Wallet saved to disk")
	}

	// Display wallet info
	wallet.DisplayWalletInfo()

	// Connect to the server
	fmt.Println("\n🌐 Connecting to blockchain server...")
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	// Raw public key bytes (65)
	pubBytes := wallet.PublicKey

	// Create a test transaction first (without signature)
	tx := core.Tx{
		Version: 1,
		TxIns: []core.TxIn{
			{
				PrevTx:    []byte("prev_tx_hash_123"),
				PrevIndex: 0,
				Signature: []byte{}, // Will be filled after signing
				PubKey:    pubBytes,
				Net:       "testnet",
			},
		},
		TxOuts: []core.TxOut{
			{
				Amount: 1000000,
				// Locking script es hash SHA3-256 de la clave pública del destinatario (en este demo, nos enviamos a nosotros mismos)
				LockingScript: core.HashSHA3(pubBytes),
			},
		},
	}

	// Get the transaction hash for signing
	txHash := tx.GetHashForSigning()

	// Sign the transaction hash with ECDSA
	signature, err := wallet.SignECDSA(txHash)
	if err != nil {
		fmt.Printf("Error signing transaction: %v\n", err)
		return
	}

	// Add the signature to the transaction
	tx.TxIns[0].Signature = signature

	// Create the transaction message (PublicKeys field deprecated, left empty)
	txMsg := core.TransactionMessage{
		Transaction: tx,
		PublicKeys:  []core.PublicKeyData{},
	}

	// Marshal to JSON
	txJSON, err := json.Marshal(txMsg)
	if err != nil {
		fmt.Println("Error marshaling transaction:", err)
		return
	}

	// Send the transaction
	fmt.Println("\n📤 Sending transaction to server...")
	fmt.Printf("Transaction ID: %s\n", tx.ID())
	message := fmt.Sprintf("TRANSACTION:%s\n", string(txJSON))
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending transaction:", err)
		return
	}

	fmt.Println("✅ Transaction sent at", time.Now().Format(time.RFC3339))

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("📡 Server response:", string(buf[:n]))
}
