package main

import (
	"fmt"

	"github.com/xkal1bur/blockchain/pkg/core"
)

func handleWallet(args []string) {
	if len(args) < 1 {
		fmt.Println("Uso: wallet [create|info]")
		return
	}

	switch args[0] {
	case "create":
		createWallet()
	case "info":
		printWalletInfo()
	default:
		fmt.Println("Comando wallet no reconocido.")
	}
}

func createWallet() {
	wallet, err := core.NewWallet()
	if err != nil {
		fmt.Printf("❌ Error al crear wallet: %v\n", err)
		return
	}
	err = wallet.SaveToDisk()
	if err != nil {
		fmt.Printf("❌ Error al guardar wallet: %v\n", err)
		return
	}
	fmt.Println("✅ Wallet creada con éxito.")
	wallet.DisplayWalletInfo()
}

func printWalletInfo() {
	if !core.WalletExists("wallet.json") {
		fmt.Println("❌ No existe wallet. Crea una con: wallet create")
		return
	}
	wallet, err := core.LoadWallet("wallet.json")
	if err != nil {
		fmt.Printf("❌ Error al cargar wallet: %v\n", err)
		return
	}
	wallet.DisplayWalletInfo()
}
