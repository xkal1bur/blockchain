package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xkal1bur/blockchain/pkg/core"
)

type BlockchainServer struct {
	pendingTransactions []core.Tx
	blockchain          []core.Block
	isMining            bool
	mu                  sync.Mutex
	blockchainFile      string
}

func NewBlockchainServer() *BlockchainServer {
	server := &BlockchainServer{
		pendingTransactions: make([]core.Tx, 0),
		blockchain:          make([]core.Block, 0),
		isMining:            false,
		blockchainFile:      "blockchain.json",
	}

	// Load existing blockchain from disk
	server.loadBlockchain()

	return server
}

func (bs *BlockchainServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		// Read message from client
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Client disconnected: %v\n", err)
			break
		}

		message = strings.TrimSpace(message)
		fmt.Printf("Received: %s\n", message)

		// Parse transaction message
		if strings.HasPrefix(message, "TRANSACTION:") {
			txJSON := strings.TrimPrefix(message, "TRANSACTION:")
			response := bs.processTransaction(txJSON)

			// Send response back to client
			conn.Write([]byte(response + "\n"))
		} else {
			conn.Write([]byte("ERROR: Unknown message format\n"))
		}
	}
}

func (bs *BlockchainServer) processTransaction(txJSON string) string {
	// Parse transaction JSON
	var tx core.Tx
	if err := json.Unmarshal([]byte(txJSON), &tx); err != nil {
		return fmt.Sprintf("ERROR: Invalid transaction JSON: %v", err)
	}

	fmt.Printf("Processing transaction: %s\n", tx.ID())

	// For simplicity, skip signature verification for now
	// In production, you'd need public keys from the client
	// if !tx.Verify(publicKeys) {
	// 	return "ERROR: Invalid transaction signature"
	// }

	// Add to pending transactions
	bs.mu.Lock()
	bs.pendingTransactions = append(bs.pendingTransactions, tx)
	bs.mu.Unlock()

	fmt.Printf("Transaction added to mempool: %s\n", tx.ID())

	// Start mining if not already mining
	go bs.startMining()

	return fmt.Sprintf("SUCCESS: Transaction %s added to mempool", tx.ID())
}

func (bs *BlockchainServer) startMining() {
	bs.mu.Lock()
	if bs.isMining || len(bs.pendingTransactions) == 0 {
		bs.mu.Unlock()
		return
	}
	bs.isMining = true
	transactions := make([]core.Tx, len(bs.pendingTransactions))
	copy(transactions, bs.pendingTransactions)
	bs.pendingTransactions = bs.pendingTransactions[:0] // Clear pending transactions
	bs.mu.Unlock()

	fmt.Printf("Starting mining with %d transactions...\n", len(transactions))

	// Create new block
	var prevBlockHash []byte
	if len(bs.blockchain) > 0 {
		hash, err := bs.blockchain[len(bs.blockchain)-1].Hash()
		if err != nil {
			log.Printf("Error getting previous block hash: %v", err)
			bs.mu.Lock()
			bs.isMining = false
			bs.mu.Unlock()
			return
		}
		prevBlockHash = hash
	} else {
		prevBlockHash = make([]byte, 32) // Genesis block
	}

	block := core.Block{
		Version:      1,
		PrevBlock:    prevBlockHash,
		Timestamp:    uint64(time.Now().Unix()),
		Nonce:        0,
		Bits:         4, // Difficulty: 4 leading zero bits
		Transactions: transactions,
	}

	// Perform Proof of Work
	start := time.Now()
	if block.CalculateValidHash() {
		duration := time.Since(start)
		hash, _ := block.Hash()

		fmt.Printf("✅ Block mined! Nonce: %d, Time: %v\n", block.Nonce, duration)
		fmt.Printf("Block hash: %x\n", hash)

		// Add block to blockchain
		bs.mu.Lock()
		bs.blockchain = append(bs.blockchain, block)
		bs.isMining = false
		bs.mu.Unlock()

		// Save to disk
		bs.saveBlockchain()
	} else {
		fmt.Println("❌ Failed to mine block")
		bs.mu.Lock()
		bs.isMining = false
		bs.mu.Unlock()
	}
}

func (bs *BlockchainServer) loadBlockchain() {
	file, err := os.Open(bs.blockchainFile)
	if err != nil {
		fmt.Printf("No existing blockchain found, starting fresh\n")
		return
	}
	defer file.Close()

	var blockchain []core.Block
	if err := json.NewDecoder(file).Decode(&blockchain); err != nil {
		log.Printf("Error loading blockchain: %v", err)
		return
	}

	bs.blockchain = blockchain
	fmt.Printf("Loaded blockchain with %d blocks\n", len(blockchain))
}

func (bs *BlockchainServer) saveBlockchain() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	file, err := os.Create(bs.blockchainFile)
	if err != nil {
		log.Printf("Error creating blockchain file: %v", err)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(bs.blockchain); err != nil {
		log.Printf("Error saving blockchain: %v", err)
		return
	}

	fmt.Printf("💾 Blockchain saved to disk (%d blocks)\n", len(bs.blockchain))
}

func main() {
	fmt.Println("🚀 Starting Blockchain TCP Server...")

	server := NewBlockchainServer()

	// Listen on TCP port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	fmt.Println("📡 Server listening on :8080")
	fmt.Println("Protocol: TRANSACTION:<json>")

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		fmt.Printf("🔗 New client connected: %s\n", conn.RemoteAddr())

		// Handle each connection in a goroutine
		go server.handleConnection(conn)
	}
}
