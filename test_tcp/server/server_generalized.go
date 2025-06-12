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

// --- Estructuras con guion bajo para evitar conflictos ---

type Tx_ struct {
	Version   uint32   `json:"version"`
	TxIns     []TxIn_  `json:"txins"`
	TxOuts    []TxOut_ `json:"txouts"`
	Locktimei uint32   `json:"locktime"`
}

type TxIn_ struct {
	PrevTx    []byte   `json:"prevtx"`
	PrevIndex uint32   `json:"previndex"`
	Sequence  uint32   `json:"sequence"`
	Witness   [][]byte `json:"witness"`
	Net       string   `json:"net"`
}

type TxOut_ struct {
	Amount  uint64 `json:"amount"`
	Address string `json:"address"`
}

type Block_ struct {
	Version    uint32 `json:"version"`
	PrevBlock  []byte `json:"prevblock"`
	MerkleRoot []byte `json:"merkleroot"`
	Timestamp  uint64 `json:"timestamp"`
	Nonce      uint64 `json:"nonce"`
	Txs        []*Tx_ `json:"txs"`
}

// --- Utilidades para Tx_ y Block_ ---

func (tx *Tx_) GenerateTxID_() string {
	jsonData, _ := json.Marshal(tx)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])[:16]
}

func (b *Block_) GenerateBlockID_() string {
	jsonData, _ := json.Marshal(b)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])[:16]
}

// --- Guardar transacción y bloque ---

func SaveTransaction_(tx *Tx_) error {
	err := os.MkdirAll("transactions", 0755)
	if err != nil {
		return fmt.Errorf("failed to create transactions directory: %v", err)
	}
	txID := tx.GenerateTxID_()
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("tx_%s_%s.json", timestamp, txID)
	filepath := filepath.Join("transactions", filename)
	jsonData, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}
	err = os.WriteFile(filepath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write transaction file: %v", err)
	}
	fmt.Printf("Transaction saved: %s (ID: %s)\n", filename, txID)
	return nil
}

func SaveBlock_(b *Block_) error {
	err := os.MkdirAll("blocks", 0755)
	if err != nil {
		return fmt.Errorf("failed to create blocks directory: %v", err)
	}
	blockID := b.GenerateBlockID_()
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("block_%s_%s.json", timestamp, blockID)
	filepath := filepath.Join("blocks", filename)
	jsonData, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal block: %v", err)
	}
	err = os.WriteFile(filepath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write block file: %v", err)
	}
	fmt.Printf("Block saved: %s (ID: %s)\n", filename, blockID)
	return nil
}

// --- Parseo de JSON ---

func ParseTransaction_(jsonStr string) (*Tx_, error) {
	var tx Tx_
	err := json.Unmarshal([]byte(jsonStr), &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction JSON: %v", err)
	}
	return &tx, nil
}

func ParseBlock_(jsonStr string) (*Block_, error) {
	var b Block_
	err := json.Unmarshal([]byte(jsonStr), &b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block JSON: %v", err)
	}
	return &b, nil
}

// --- Main server loop ---

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Generalized Server listening on port 8080...")
	fmt.Println("Ready to receive transactions and blocks...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleClient_(conn)
	}
}

func handleClient_(conn net.Conn) {
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

		if strings.HasPrefix(message, "TRANSACTION:") {
			jsonPart := strings.TrimPrefix(message, "TRANSACTION:")
			tx, err := ParseTransaction_(jsonPart)
			if err != nil {
				response = fmt.Sprintf("ERROR: Failed to parse transaction: %v\n", err)
				fmt.Printf("Error parsing transaction from %s: %v\n", clientAddr, err)
			} else {
				err = SaveTransaction_(tx)
				if err != nil {
					response = fmt.Sprintf("ERROR: Failed to save transaction: %v\n", err)
					fmt.Printf("Error saving transaction from %s: %v\n", clientAddr, err)
				} else {
					txID := tx.GenerateTxID_()
					response = fmt.Sprintf("SUCCESS: Transaction saved with ID: %s\n", txID)
					fmt.Printf("Transaction successfully saved from %s (ID: %s)\n", clientAddr, txID)
					fmt.Printf("  Version: %d, Inputs: %d, Outputs: %d, Locktime: %d\n",
						tx.Version, len(tx.TxIns), len(tx.TxOuts), tx.Locktimei)
					for i, out := range tx.TxOuts {
						fmt.Printf("  Output %d: %d satoshis to %s\n", i, out.Amount, out.Address)
					}
				}
			}
		} else if strings.HasPrefix(message, "BLOCK:") {
			jsonPart := strings.TrimPrefix(message, "BLOCK:")
			block, err := ParseBlock_(jsonPart)
			if err != nil {
				response = fmt.Sprintf("ERROR: Failed to parse block: %v\n", err)
				fmt.Printf("Error parsing block from %s: %v\n", clientAddr, err)
			} else {
				err = SaveBlock_(block)
				if err != nil {
					response = fmt.Sprintf("ERROR: Failed to save block: %v\n", err)
					fmt.Printf("Error saving block from %s: %v\n", clientAddr, err)
				} else {
					blockID := block.GenerateBlockID_()
					response = fmt.Sprintf("SUCCESS: Block saved with ID: %s\n", blockID)
					fmt.Printf("Block successfully saved from %s (ID: %s)\n", clientAddr, blockID)
					fmt.Printf("  Version: %d, Tx count: %d, Timestamp: %d, Nonce: %d\n",
						block.Version, len(block.Txs), block.Timestamp, block.Nonce)
				}
			}
		} else {
			response = fmt.Sprintf("Server received message #%d at %s\n",
				messageCount, time.Now().Format("15:04:05"))
		}

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
