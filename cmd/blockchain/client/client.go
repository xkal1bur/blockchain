package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/xkal1bur/blockchain/pkg/core"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	// Create a test transaction
	tx := core.Tx{
		Version: 1,
		TxIns: []core.TxIn{
			{
				PrevTx:    []byte("prev_tx_hash_123"),
				PrevIndex: 0,
				Signature: []byte("test_signature"),
				Net:       "testnet",
			},
		},
		TxOuts: []core.TxOut{
			{
				Amount:        1000000,
				LockingScript: []byte("recipient_address"),
			},
		},
	}

	// Create the transaction message with public key (simplified)
	txMsg := core.TransactionMessage{
		Transaction: tx,
		PublicKeys: []core.PublicKeyData{
			{
				X: "public_key_x",
				Y: "public_key_y",
			},
		},
	}

	// Marshal to JSON
	txJSON, err := json.Marshal(txMsg)
	if err != nil {
		fmt.Println("Error marshaling transaction:", err)
		return
	}

	// Send the transaction
	message := fmt.Sprintf("TRANSACTION:%s\n", string(txJSON))
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending transaction:", err)
		return
	}

	fmt.Println("Transaction sent at", time.Now().Format(time.RFC3339))

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Server response:", string(buf[:n]))
}
