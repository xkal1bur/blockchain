package main

import (
	"fmt"
	"log"
	"os"

	"github.com/xkal1bur/blockchain/pkg/core"
)

func main() {
	walletFile := "wallet.json"

	fmt.Println("🏦 Blockchain Wallet Generator")
	fmt.Println("==============================")

	// Check if wallet already exists
	if core.WalletExists(walletFile) {
		fmt.Printf("📁 Wallet file '%s' already exists!\n", walletFile)
		fmt.Println("Loading existing wallet...")

		// Load existing wallet
		wallet, err := core.LoadWallet(walletFile)
		if err != nil {
			log.Fatalf("❌ Error loading wallet: %v", err)
		}

		fmt.Println("✅ Wallet loaded successfully!")
		wallet.DisplayWalletInfo()
		return
	}

	// Create new wallet
	fmt.Println("🔐 Generating new cryptographic keys...")
	wallet, err := core.NewWallet()
	if err != nil {
		log.Fatalf("❌ Error creating wallet: %v", err)
	}

	// Set wallet file path
	wallet.WalletFile = walletFile

	// Save wallet to disk
	fmt.Printf("💾 Saving wallet to disk: %s\n", walletFile)
	if err := wallet.SaveToDisk(); err != nil {
		log.Fatalf("❌ Error saving wallet: %v", err)
	}

	fmt.Println("✅ Wallet created and saved successfully!")
	wallet.DisplayWalletInfo()

	// Show file size
	if info, err := os.Stat(walletFile); err == nil {
		fmt.Printf("📏 Wallet file size: %d bytes\n", info.Size())
	}

	fmt.Println("\n⚠️  IMPORTANT SECURITY NOTICE:")
	fmt.Println("   - Keep your private key secure and never share it")
	fmt.Println("   - Back up your wallet file in a safe location")
	fmt.Println("   - Anyone with access to your private key can control your funds")
}
