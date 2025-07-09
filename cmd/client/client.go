package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/xkal1bur/blockchain/pkg/core"
)

func main() {
	// Create or load wallet
	fmt.Println("🔑 Initializing wallet...")
	var wallet *core.Wallet
	var err error

	walletFile := "wallet.json"
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

	// Load UTXOs from file
	fmt.Println("\n📄 Loading UTXOs from utxos.json...")
	utxoData, err := os.ReadFile("utxos.json")
	if err != nil {
		fmt.Printf("❌ Error reading utxos.json: %v\n", err)
		fmt.Println("💡 Make sure to run the genesis block script first!")
		return
	}

	var utxoSet map[string]core.TxOut
	if err := json.Unmarshal(utxoData, &utxoSet); err != nil {
		fmt.Printf("❌ Error parsing utxos.json: %v\n", err)
		return
	}

	fmt.Printf("📊 Loaded %d UTXOs\n", len(utxoSet))

	// Find UTXOs that belong to our wallet
	myUTXOs := wallet.FilterUTXOs(utxoSet)
	fmt.Printf("💰 Found %d UTXOs belonging to this wallet\n", len(myUTXOs))

	if len(myUTXOs) == 0 {
		fmt.Println("❌ No UTXOs found for this wallet. Cannot create transaction.")
		fmt.Println("💡 This wallet address:", wallet.Address)
		return
	}

	// Show available UTXOs
	fmt.Println("\n💳 Available UTXOs:")
	totalBalance := uint64(0)
	for key, utxo := range myUTXOs {
		fmt.Printf("  %s: %d satoshis\n", key, utxo.Amount)
		totalBalance += utxo.Amount
	}
	fmt.Printf("💵 Total balance: %d satoshis\n", totalBalance)

	// Create a transaction sending to another address (or self)
	sendAmount := uint64(500000)                                                             // Send 0.5M satoshis
	destinationAddress := "e4cf9ec444babdf51e5783162ba14efb5210447f40f5d842ab23c945c7dfc643" // Second target from genesis

	fmt.Printf("\n🚀 Creating transaction to send %d satoshis to %s\n", sendAmount, destinationAddress)

	// Build the transaction using wallet's BuildTransactionToAddress method
	tx, usedKeys, err := wallet.BuildTransactionToAddress(destinationAddress, sendAmount, utxoSet)
	if err != nil {
		fmt.Printf("❌ Error building transaction: %v\n", err)
		return
	}

	fmt.Printf("✅ Transaction built successfully!\n")
	fmt.Printf("📝 Transaction ID: %s\n", tx.ID())
	fmt.Printf("🔑 Used %d UTXOs as inputs\n", len(usedKeys))

	// Connect to the server
	fmt.Println("\n🌐 Connecting to blockchain server...")
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Printf("❌ Error connecting to server: %v\n", err)
		fmt.Println("💡 Make sure the blockchain server is running on port 8081")
		return
	}
	defer conn.Close()

	// Create the transaction message (PublicKeys field deprecated, left empty)
	txMsg := core.TransactionMessage{
		Transaction: tx,
		PublicKeys:  []core.PublicKeyData{},
	}

	// Marshal to JSON
	txJSON, err := json.Marshal(txMsg)
	if err != nil {
		fmt.Println("❌ Error marshaling transaction:", err)
		return
	}

	// Send the transaction
	fmt.Println("\n📤 Sending transaction to server...")
	message := fmt.Sprintf("TRANSACTION:%s\n", string(txJSON))
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("❌ Error sending transaction:", err)
		return
	}

	fmt.Println("✅ Transaction sent at", time.Now().Format(time.RFC3339))

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("❌ Error reading response:", err)
		return
	}

	fmt.Println("📡 Server response:", string(buf[:n]))
}
