package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/xkal1bur/blockchain/pkg/core"
)

// walllets to send:
// ar: c8c3a919c9ca981291263d9ccdc2b04e9432bbf57ff0a84dcd53c7b74fc79aa7
// an: e4cf9ec444babdf51e5783162ba14efb5210447f40f5d842ab23c945c7dfc643

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

	targets := []string{"c8c3a919c9ca981291263d9ccdc2b04e9432bbf57ff0a84dcd53c7b74fc79aa7", "e4cf9ec444babdf51e5783162ba14efb5210447f40f5d842ab23c945c7dfc643"}

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

	// Calcular Proof of Work para el bloque génesis
	fmt.Println("⛏️  Minando bloque génesis...")
	if !genesis.CalculateValidHash() {
		fmt.Println("❌ Error: No se pudo minar el bloque génesis")
		return
	}

	hash, _ := genesis.Hash()
	fmt.Printf("✅ Bloque génesis minado! Nonce: %d, Hash: %x\n", genesis.Nonce, hash)

	// 3) Enviar el bloque génesis al servidor
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
