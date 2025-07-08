package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/xkal1bur/blockchain/pkg/core"
)

type TxOut struct {
	TxID          string
	Index         uint32
	Amount        uint64
	LockingScript []byte
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// 1. Cargar wallet
	fmt.Print("📂 Ruta de tu wallet (ej: wallet.json): ")
	walletPath, _ := reader.ReadString('\n')
	walletPath = strings.TrimSpace(walletPath)

	wallet, err := core.LoadWallet(walletPath)
	if err != nil {
		fmt.Println("❌ Error cargando wallet:", err)
		return
	}
	wallet.DisplayWalletInfo()

	// 2. Mostrar UTXOs disponibles
	utxos := core.GetUTXOsForAddress(wallet.Address)
	if len(utxos) == 0 {
		fmt.Println("⚠️ No tienes fondos disponibles.")
		return
	}

	fmt.Println("\n💰 UTXOs disponibles:")
	for i, utxo := range utxos {
		fmt.Printf(" [%d] %d HORUS - %s:%d\n", i, utxo.Amount, utxo.TxID, utxo.Index)
	}

	// 3. Seleccionar UTXO a gastar
	fmt.Print("\n🔢 Selecciona el índice del UTXO que deseas gastar: ")
	choiceStr, _ := reader.ReadString('\n')
	choice, _ := strconv.Atoi(strings.TrimSpace(choiceStr))

	if choice < 0 || choice >= len(utxos) {
		fmt.Println("❌ Índice inválido.")
		return
	}
	selected := utxos[choice]

	// 4. Ingresar dirección del receptor y monto
	fmt.Print("📨 Dirección del receptor: ")
	receiverAddr, _ := reader.ReadString('\n')
	receiverAddr = strings.TrimSpace(receiverAddr)

	fmt.Print("💸 Monto a enviar (HORUS): ")
	amountStr, _ := reader.ReadString('\n')
	amount, _ := strconv.ParseUint(strings.TrimSpace(amountStr), 10, 64)

	// Necesitamos transformar el UTXO seleccionado a formato map[string]TxOut
	utxoMap := map[string]core.TxOut{
		fmt.Sprintf("%s:%d", selected.TxID, selected.Index): {
			Amount:        selected.Amount,
			LockingScript: []byte(wallet.Address),
		},
	}

	// 5. Crear transacción
	tx, err := core.CreateTransaction(wallet, receiverAddr, amount, utxoMap)
	if err != nil {
		fmt.Println("❌ Error al crear transacción:", err)
		return
	}

	// 6. Enviar al servidor
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("❌ No se pudo conectar al servidor:", err)
		return
	}
	defer conn.Close()

	txBytes, _ := json.Marshal(tx)
	message := "TRANSACTION:" + string(txBytes)

	conn.Write([]byte(message + "\n"))
	fmt.Println("✅ Transacción enviada correctamente.")
}
