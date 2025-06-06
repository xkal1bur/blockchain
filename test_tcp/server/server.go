package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Tx struct {
	Version   uint32  `json:"version"`
	TxIns     []TxIn  `json:"txins"`
	TxOuts    []TxOut `json:"txouts"`
	Locktimei uint32  `json:"locktime"`
}

type TxIn struct {
	PrevTx    []byte   `json:"prevtx"`
	PrevIndex uint32   `json:"previndex"`
	Sequence  uint32   `json:"sequence"`
	Witness   [][]byte `json:"witness"`
	Net       string   `json:"net"`
}

type TxOut struct {
	Amount  uint64 `json:"amount"`
	Address string `json:"address"`
}

// GenerateTxID generates a unique transaction ID
func (tx *Tx) GenerateTxID() string {
	jsonData, _ := json.Marshal(tx)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars as ID
}

// SaveTransaction saves a transaction to disk
func SaveTransaction(tx *Tx) error {
	// Create transactions directory if it doesn't exist
	err := os.MkdirAll("transactions", 0755)
	if err != nil {
		return fmt.Errorf("failed to create transactions directory: %v", err)
	}

	// Generate transaction ID
	txID := tx.GenerateTxID()

	// Create filename with timestamp and transaction ID
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("tx_%s_%s.json", timestamp, txID)
	filepath := filepath.Join("transactions", filename)

	// Convert transaction to pretty JSON
	jsonData, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}

	// Write to file
	err = os.WriteFile(filepath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write transaction file: %v", err)
	}

	fmt.Printf("Transaction saved: %s (ID: %s)\n", filename, txID)
	return nil
}

// ParseTransaction parses a transaction from JSON string
func ParseTransaction(jsonStr string) (*Tx, error) {
	var tx Tx
	err := json.Unmarshal([]byte(jsonStr), &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction JSON: %v", err)
	}
	return &tx, nil
}

func main() {
	// Listen on port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Transaction Server listening on port 8080...")
	fmt.Println("Ready to receive and save transactions...")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle each client connection in a separate goroutine
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("New client connected: %s\n", clientAddr)

	scanner := bufio.NewScanner(conn)
	messageCount := 0

	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		messageCount++
		fmt.Printf("Received from %s: message #%d\n", clientAddr, messageCount)

		var response string

		// Check if message is a transaction
		if strings.HasPrefix(message, "TRANSACTION:") {
			// Extract JSON part
			jsonPart := strings.TrimPrefix(message, "TRANSACTION:")

			// Parse transaction
			tx, err := ParseTransaction(jsonPart)
			if err != nil {
				response = fmt.Sprintf("ERROR: Failed to parse transaction: %v\n", err)
				fmt.Printf("Error parsing transaction from %s: %v\n", clientAddr, err)
			} else {
				// Save transaction
				err = SaveTransaction(tx)
				if err != nil {
					response = fmt.Sprintf("ERROR: Failed to save transaction: %v\n", err)
					fmt.Printf("Error saving transaction from %s: %v\n", clientAddr, err)
				} else {
					txID := tx.GenerateTxID()
					response = fmt.Sprintf("SUCCESS: Transaction saved with ID: %s\n", txID)
					fmt.Printf("Transaction successfully saved from %s (ID: %s)\n", clientAddr, txID)

					// Print transaction details
					fmt.Printf("  Version: %d, Inputs: %d, Outputs: %d, Locktime: %d\n",
						tx.Version, len(tx.TxIns), len(tx.TxOuts), tx.Locktimei)

					// Print output details
					for i, out := range tx.TxOuts {
						fmt.Printf("  Output %d: %d satoshis to %s\n", i, out.Amount, out.Address)
					}
				}
			}
		} else {
			// Regular message
			response = fmt.Sprintf("Server received message #%d at %s\n",
				messageCount, time.Now().Format("15:04:05"))
		}

		// Send response back to client
		_, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Printf("Error sending response to %s: %v\n", clientAddr, err)
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from %s: %v\n", clientAddr, err)
	}

	fmt.Printf("Client %s disconnected. Total messages received: %d\n",
		clientAddr, messageCount)
}
