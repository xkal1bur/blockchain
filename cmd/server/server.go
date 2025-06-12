package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/xkal1bur/blockchain/pkg/core"
)

func main() {
	fmt.Println("🚀 Starting Blockchain TCP Server...")

	server := core.NewBlockchainServer()

	// Configure peer servers from command line arguments or environment
	if len(os.Args) > 1 {
		for _, peer := range os.Args[1:] {
			server.AddPeer(peer)
		}
	}

	// Listen on TCP port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	fmt.Println("📡 Server listening on :8080")
	fmt.Println("Protocol Messages:")
	fmt.Println("  TRANSACTION:<json> - Submit transaction with public keys")
	fmt.Println("  BLOCK:<json>       - Receive validated block from peer")

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		fmt.Printf("🔗 New client connected: %s\n", conn.RemoteAddr())

		// Handle each connection in a goroutine with transaction printing
		go func(c net.Conn) {
			// Create a tee reader to both process and print the transaction
			reader := bufio.NewReader(c)

			for {
				message, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("Client disconnected: %v", err)
					c.Close()
					return
				}

				message = strings.TrimSpace(message)

				// Print transaction details if it's a transaction message
				if strings.HasPrefix(message, "TRANSACTION:") {
					printTransaction(message)
				}
				fmt.Println("Now handling: ")

				server.HandleConnection(c)
			}
		}(conn)
	}
}

// printTransaction prints the transaction details in a human-readable format
func printTransaction(message string) {
	txJSON := strings.TrimPrefix(message, "TRANSACTION:")

	var txMsg core.TransactionMessage
	if err := json.Unmarshal([]byte(txJSON), &txMsg); err != nil {
		fmt.Printf("⚠️ Failed to parse transaction: %v\n", err)
		return
	}

	tx := txMsg.Transaction

	fmt.Println("\n📜 Transaction Received (Preview)")
	fmt.Println("=============================")
	fmt.Printf("Transaction ID: %s\n", tx.ID())
	fmt.Printf("Version: %d\n", tx.Version)
	fmt.Printf("Inputs (%d):\n", len(tx.TxIns))
	for i, input := range tx.TxIns {
		fmt.Printf("  Input %d:\n", i+1)
		fmt.Printf("    Previous Tx: %x\n", input.PrevTx)
		fmt.Printf("    Index: %d\n", input.PrevIndex)
		fmt.Printf("    Network: %s\n", input.Net)
		fmt.Printf("    Signature: %x\n", input.Signature)
	}

	fmt.Printf("\nOutputs (%d):\n", len(tx.TxOuts))
	for i, output := range tx.TxOuts {
		fmt.Printf("  Output %d:\n", i+1)
		fmt.Printf("    Amount: %d\n", output.Amount)
		fmt.Printf("    Locking Script: %x\n", output.LockingScript)
	}
	fmt.Println("=============================\n")
}
