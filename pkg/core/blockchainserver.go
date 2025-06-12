package core

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type BlockchainServer struct {
	pendingTransactions []Tx
	blockchain          []Block
	isMining            bool
	mu                  sync.Mutex
	blockchainFile      string
	peerServers         []string // List of peer server addresses
}

// TransactionMessage represents a transaction with its public key for validation
type TransactionMessage struct {
	Transaction Tx              `json:"transaction"`
	PublicKeys  []PublicKeyData `json:"public_keys"`
}

// PublicKeyData represents a serialized public key
type PublicKeyData struct {
	X string `json:"x"`
	Y string `json:"y"`
}

// BlockMessage represents a validated block from another server
type BlockMessage struct {
	Block      Block             `json:"block"`
	PublicKeys [][]PublicKeyData `json:"public_keys"` // Public keys for each transaction
}

func NewBlockchainServer() *BlockchainServer {
	server := &BlockchainServer{
		pendingTransactions: make([]Tx, 0),
		blockchain:          make([]Block, 0),
		isMining:            false,
		blockchainFile:      "blockchain.json",
		peerServers:         []string{}, // Will be configured later
	}

	// Load existing blockchain from disk
	server.loadBlockchain()

	return server
}

// AddPeer adds a peer server address for block broadcasting
func (bs *BlockchainServer) AddPeer(peerAddress string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.peerServers = append(bs.peerServers, peerAddress)
	fmt.Printf("🔗 Added peer server: %s\n", peerAddress)
}

func (bs *BlockchainServer) HandleConnection(conn net.Conn) {
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

		var response string

		// Parse different message types
		if strings.HasPrefix(message, "TRANSACTION:") {
			txJSON := strings.TrimPrefix(message, "TRANSACTION:")
			response = bs.processTransactionMessage(txJSON)
		} else if strings.HasPrefix(message, "BLOCK:") {
			blockJSON := strings.TrimPrefix(message, "BLOCK:")
			response = bs.processBlockMessage(blockJSON)
		} else {
			response = "ERROR: Unknown message format. Use TRANSACTION:<json> or BLOCK:<json>"
		}

		// Send response back to client
		conn.Write([]byte(response + "\n"))
	}
}

func (bs *BlockchainServer) processTransactionMessage(txJSON string) string {
	// Parse transaction message with public keys
	var txMsg TransactionMessage
	if err := json.Unmarshal([]byte(txJSON), &txMsg); err != nil {
		return fmt.Sprintf("ERROR: Invalid transaction message JSON: %v", err)
	}

	fmt.Printf("Processing transaction: %s\n", txMsg.Transaction.ID())

	// Convert public key data to ecdsa.PublicKey
	publicKeys, err := bs.parsePublicKeys(txMsg.PublicKeys)
	if err != nil {
		return fmt.Sprintf("ERROR: Invalid public keys: %v", err)
	}

	// Verify transaction with provided public keys
	if !txMsg.Transaction.Verify(publicKeys) {
		return "ERROR: Invalid transaction signature"
	}

	// Add to pending transactions
	bs.mu.Lock()
	bs.pendingTransactions = append(bs.pendingTransactions, txMsg.Transaction)
	bs.mu.Unlock()

	fmt.Printf("Transaction added to mempool: %s\n", txMsg.Transaction.ID())

	// Start mining if not already mining
	go bs.startMining()

	return fmt.Sprintf("SUCCESS: Transaction %s added to mempool", txMsg.Transaction.ID())
}

func (bs *BlockchainServer) processBlockMessage(blockJSON string) string {
	// Parse block message
	var blockMsg BlockMessage
	if err := json.Unmarshal([]byte(blockJSON), &blockMsg); err != nil {
		return fmt.Sprintf("ERROR: Invalid block message JSON: %v", err)
	}

	fmt.Printf("Received block with %d transactions\n", len(blockMsg.Block.Transactions))

	// Validate the received block
	if !bs.validateReceivedBlock(blockMsg.Block, blockMsg.PublicKeys) {
		return "ERROR: Block validation failed"
	}

	// Add block to blockchain
	bs.mu.Lock()
	bs.blockchain = append(bs.blockchain, blockMsg.Block)
	bs.mu.Unlock()

	// Save to disk
	bs.saveBlockchain()

	hash, _ := blockMsg.Block.Hash()
	fmt.Printf("✅ Block accepted and added to blockchain! Hash: %x\n", hash)

	return fmt.Sprintf("SUCCESS: Block accepted and added to blockchain")
}

func (bs *BlockchainServer) validateReceivedBlock(block Block, publicKeyData [][]PublicKeyData) bool {
	// Check if block connects to our chain
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if len(bs.blockchain) > 0 {
		lastBlockHash, err := bs.blockchain[len(bs.blockchain)-1].Hash()
		if err != nil {
			fmt.Printf("Error getting last block hash: %v\n", err)
			return false
		}

		if !bytes.Equal(block.PrevBlock, lastBlockHash) {
			fmt.Printf("Block doesn't connect to our chain\n")
			return false
		}
	}

	// Validate Proof of Work
	blockHash, err := block.Hash()
	if err != nil {
		fmt.Printf("Error getting block hash: %v\n", err)
		return false
	}

	if !block.isValidHash(blockHash) {
		fmt.Printf("Invalid Proof of Work\n")
		return false
	}

	// Validate all transactions in the block
	for i, tx := range block.Transactions {
		if i >= len(publicKeyData) {
			fmt.Printf("Missing public keys for transaction %d\n", i)
			return false
		}

		publicKeys, err := bs.parsePublicKeys(publicKeyData[i])
		if err != nil {
			fmt.Printf("Invalid public keys for transaction %d: %v\n", i, err)
			return false
		}

		if !tx.Verify(publicKeys) {
			fmt.Printf("Transaction %d verification failed\n", i)
			return false
		}
	}

	fmt.Printf("Block validation successful\n")
	return true
}

func (bs *BlockchainServer) parsePublicKeys(publicKeyData []PublicKeyData) ([]*ecdsa.PublicKey, error) {
	publicKeys := make([]*ecdsa.PublicKey, len(publicKeyData))

	for i, pkData := range publicKeyData {
		xBytes, err := hex.DecodeString(pkData.X)
		if err != nil {
			return nil, fmt.Errorf("invalid X coordinate: %v", err)
		}

		yBytes, err := hex.DecodeString(pkData.Y)
		if err != nil {
			return nil, fmt.Errorf("invalid Y coordinate: %v", err)
		}

		x := new(big.Int).SetBytes(xBytes)
		y := new(big.Int).SetBytes(yBytes)

		publicKeys[i] = &ecdsa.PublicKey{
			Curve: StandardCurve,
			X:     x,
			Y:     y,
		}
	}

	return publicKeys, nil
}

func (bs *BlockchainServer) startMining() {
	bs.mu.Lock()
	if bs.isMining || len(bs.pendingTransactions) == 0 {
		bs.mu.Unlock()
		return
	}
	bs.isMining = true
	transactions := make([]Tx, len(bs.pendingTransactions))
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

	block := Block{
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

		// Broadcast block to peer servers
		go bs.broadcastBlock(block)
	} else {
		fmt.Println("❌ Failed to mine block")
		bs.mu.Lock()
		bs.isMining = false
		bs.mu.Unlock()
	}
}

// broadcastBlock sends the mined block to all peer servers
func (bs *BlockchainServer) broadcastBlock(block Block) {
	bs.mu.Lock()
	peers := make([]string, len(bs.peerServers))
	copy(peers, bs.peerServers)
	bs.mu.Unlock()

	if len(peers) == 0 {
		fmt.Printf("No peer servers configured for broadcasting\n")
		return
	}

	// Create block message (for now without public keys - would need to store them)
	blockMsg := BlockMessage{
		Block:      block,
		PublicKeys: [][]PublicKeyData{}, // TODO: Include public keys for validation
	}

	blockJSON, err := json.Marshal(blockMsg)
	if err != nil {
		log.Printf("Error marshaling block: %v", err)
		return
	}

	// Send to all peer servers
	for _, peer := range peers {
		go func(peerAddr string) {
			conn, err := net.Dial("tcp", peerAddr)
			if err != nil {
				log.Printf("Failed to connect to peer %s: %v", peerAddr, err)
				return
			}
			defer conn.Close()

			message := fmt.Sprintf("BLOCK:%s\n", string(blockJSON))
			_, err = conn.Write([]byte(message))
			if err != nil {
				log.Printf("Failed to send block to peer %s: %v", peerAddr, err)
				return
			}

			fmt.Printf("📡 Block broadcasted to peer: %s\n", peerAddr)
		}(peer)
	}
}

func (bs *BlockchainServer) loadBlockchain() {
	file, err := os.Open(bs.blockchainFile)
	if err != nil {
		fmt.Printf("No existing blockchain found, starting fresh\n")
		return
	}
	defer file.Close()

	var blockchain []Block
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
