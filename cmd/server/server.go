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

	// Listen on TCP port 8081
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	fmt.Println("📡 Server listening on :8081")
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
			defer c.Close()

			// Use a custom reader to intercept and print transaction details
			reader := bufio.NewReader(c)

			for {
				message, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("Client disconnected: %v", err)
					return
				}

				message = strings.TrimSpace(message)

				// Print transaction details if it's a transaction message
				if strings.HasPrefix(message, "TRANSACTION:") {
					printTransaction(message)
				}

				// Process the message using the blockchain server
				response := processMessage(server, message)

				// Send response back to client
				_, writeErr := c.Write([]byte(response + "\n"))
				if writeErr != nil {
					log.Printf("Error writing response: %v", writeErr)
					return
				}
			}
		}(conn)
	}
}

// processMessage processes a single message and returns the response
func processMessage(server *core.BlockchainServer, message string) string {
	var response string

	// Parse different message types
	if strings.HasPrefix(message, "TRANSACTION:") {
		txJSON := strings.TrimPrefix(message, "TRANSACTION:")
		response = server.ProcessTransactionMessage(txJSON)
	} else if strings.HasPrefix(message, "BLOCK:") {
		blockJSON := strings.TrimPrefix(message, "BLOCK:")
		response = server.ProcessBlockMessage(blockJSON)
	} else {
		response = "ERROR: Unknown message format. Use TRANSACTION:<json> or BLOCK:<json>"
	}

	return response
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
