package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/xkal1bur/blockchain/pkg/core"
)

// walllets to send: bc112bc6bc381a9a05dfed7eeeb26c87da9d2bad321, bc1623e0f3f6d64de891b0633f0a6e09a4e6ad91229

func main() {
	// 1) Cargar o crear wallet local
	walletFile := "wallet.json"
	var wallet *core.Wallet
	var err error
	if core.WalletExists(walletFile) {
		wallet, err = core.LoadWallet(walletFile)
	} else {
		wallet, err = core.NewWallet()
		wallet.WalletFile = walletFile
		wallet.SaveToDisk()
	}
	if err != nil {
		panic(err)
	}

	wallet.DisplayWalletInfo()

	// 2) Generar bloque génesis con salida hacia la dirección del wallet
	fmt.Println("⚙️  Creando bloque génesis para el wallet…")
	amountGenesis := uint64(1_000_000)

	targets := []string{"bc112bc6bc381a9a05dfed7eeeb26c87da9d2bad321", "bc1623e0f3f6d64de891b0633f0a6e09a4e6ad91229"}

	// Única coinbase con 3 salidas por destino
	var outputs []core.TxOut
	for _, addr := range targets {
		for i := 0; i < 3; i++ {
			outputs = append(outputs, core.TxOut{
				Amount:        amountGenesis,
				LockingScript: []byte(addr),
			})
		}
	}

	coinbaseTx := core.Tx{Version: 1, TxOuts: outputs}

	genesis := core.Block{
		Version:      1,
		PrevBlock:    make([]byte, 32),
		Timestamp:    uint64(time.Now().Unix()),
		Nonce:        0,
		Bits:         1,
		Transactions: []core.Tx{coinbaseTx},
	}

	// Guardar blockchain.json
	chainData, _ := json.MarshalIndent([]core.Block{genesis}, "", "  ")
	os.WriteFile("blockchain.json", chainData, 0644)
	// 3) Construir y guardar utxos.json correspondiente
	utxoSet := make(map[string]core.TxOut)
	txID := coinbaseTx.ID()
	for idx, out := range coinbaseTx.TxOuts {
		key := fmt.Sprintf("%s:%d", txID, idx)
		utxoSet[key] = out
	}
	core.SaveUTXOs("utxos.json", utxoSet)

	fmt.Printf("✅ Génesis guardado: 1 transacción coinbase con %d UTXO\n", len(utxoSet))

	// 4) Enviar el bloque génesis al servidor
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("⚠️  Inicia primero el servidor TCP para enviar transacciones")
		return
	}
	defer conn.Close()

	blockMsg := core.BlockMessage{Block: genesis, PublicKeys: [][]core.PublicKeyData{}}
	data, _ := json.Marshal(blockMsg)
	conn.Write([]byte("BLOCK:" + string(data) + "\n"))
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	fmt.Printf("📤 Bloque génesis enviado. Respuesta: %s\n", string(buf[:n]))

	// Fin
}

func hashBlock(b core.Block) string {
	h, err := b.Hash()
	if err != nil {
		return "error"
	}
	return hex.EncodeToString(h)[:16]
}
