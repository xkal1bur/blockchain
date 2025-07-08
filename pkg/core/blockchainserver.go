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

	utxoSet map[string]TxOut // Unspent transaction outputs
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

const utxoFile = "utxos.json"

func NewBlockchainServer() *BlockchainServer {
	server := &BlockchainServer{
		pendingTransactions: make([]Tx, 0),
		blockchain:          make([]Block, 0),
		isMining:            false,
		blockchainFile:      "blockchain.json",
		peerServers:         []string{}, // Will be configured later

		utxoSet: make(map[string]TxOut),
	}

	// Load existing blockchain from disk
	server.loadBlockchain()

	// Try to load persisted UTXO set
	if err := server.loadUTXOSet(); err != nil {
		fmt.Println("🗄️  No UTXO set file found, rebuilding from blockchain…")
		server.rebuildUTXOSet()
		server.saveUTXOSet()
	}

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
			response = bs.ProcessTransactionMessage(txJSON)
		} else if strings.HasPrefix(message, "BLOCK:") {
			blockJSON := strings.TrimPrefix(message, "BLOCK:")
			response = bs.ProcessBlockMessage(blockJSON)
		} else {
			response = "ERROR: Unknown message format. Use TRANSACTION:<json> or BLOCK:<json>"
		}

		// Send response back to client
		conn.Write([]byte(response + "\n"))
	}
}

func (bs *BlockchainServer) ProcessTransactionMessage(txJSON string) string {
	// Parse the transaction message (we ignore PublicKeys field now)
	var txMsg TransactionMessage
	if err := json.Unmarshal([]byte(txJSON), &txMsg); err != nil {
		return fmt.Sprintf("ERROR: Invalid transaction message JSON: %v", err)
	}

	fmt.Printf("Processing transaction: %s\n", txMsg.Transaction.ID())

	prevMap := bs.buildPrevTxMap()

	if !txMsg.Transaction.Validate(prevMap) {
		return "ERROR: Transaction validation failed"
	}

	bs.mu.Lock()
	bs.pendingTransactions = append(bs.pendingTransactions, txMsg.Transaction)
	bs.mu.Unlock()

	fmt.Printf("Transaction added to mempool: %s\n", txMsg.Transaction.ID())

	go bs.startMining()

	return fmt.Sprintf("SUCCESS: Transaction %s added to mempool", txMsg.Transaction.ID())
}

func (bs *BlockchainServer) ProcessBlockMessage(blockJSON string) string {
	// Parse block message
	var blockMsg BlockMessage
	if err := json.Unmarshal([]byte(blockJSON), &blockMsg); err != nil {
		return fmt.Sprintf("ERROR: Invalid block message JSON: %v", err)
	}

	fmt.Printf("Received block with %d transactions\n", len(blockMsg.Block.Transactions))

	// Validate the received block
	if !bs.validateReceivedBlock(blockMsg.Block) {
		return "ERROR: Block validation failed"
	}

	// Update UTXO set with the new block before adding
	bs.updateUTXOSetWithBlock(blockMsg.Block)

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

// buildPrevTxMap construye mapa txID → *Tx recorriendo toda la blockchain actual.
func (bs *BlockchainServer) buildPrevTxMap() map[string]*Tx {
	prevMap := make(map[string]*Tx)
	for _, blk := range bs.blockchain {
		for idx := range blk.Transactions {
			tx := &blk.Transactions[idx]
			prevMap[tx.ID()] = tx
		}
	}
	return prevMap
}

func (bs *BlockchainServer) validateReceivedBlock(block Block) bool {
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

	// Prepare map of previous tx for validation, including chain so far
	prevMap := bs.buildPrevTxMap()

	for i := 0; i < len(block.Transactions); i++ {
		tx := &block.Transactions[i]
		if !tx.Validate(prevMap) {
			fmt.Printf("Transaction %d validation failed\n", i)
			return false
		}
		// After validation add tx to map to allow intra-block spending
		prevMap[tx.ID()] = tx
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
	}

	block := Block{
		Version:      1,
		PrevBlock:    prevBlockHash,
		Timestamp:    uint64(time.Now().Unix()),
		Nonce:        0,
		Bits:         12, // Difficulty: 4 leading zero bits
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

		// Update UTXO set with mined block
		bs.updateUTXOSetWithBlock(block)

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

// rebuildUTXOSet reconstruye todo el conjunto UTXO recorriendo la blockchain
func (bs *BlockchainServer) rebuildUTXOSet() {
	bs.utxoSet = make(map[string]TxOut)

	for _, block := range bs.blockchain {
		bs.updateUTXOSetWithBlock(block)
	}
}

// updateUTXOSetWithBlock actualiza el conjunto UTXO al aceptar un bloque
func (bs *BlockchainServer) updateUTXOSetWithBlock(block Block) {
	// Remove spent outputs
	for _, tx := range block.Transactions {
		for _, in := range tx.TxIns {
			key := fmt.Sprintf("%x:%d", in.PrevTx, in.PrevIndex)
			delete(bs.utxoSet, key)
		}

		// Add new outputs
		txID := tx.ID()
		for idx, out := range tx.TxOuts {
			key := fmt.Sprintf("%s:%d", txID, idx)
			bs.utxoSet[key] = out
		}
	}

	// Persist updated set
	bs.saveUTXOSet()
}

// loadUTXOSet loads UTXOs from disk into memory
func (bs *BlockchainServer) loadUTXOSet() error {
	data, err := os.ReadFile(utxoFile)
	if err != nil {
		return err
	}
	var m map[string]TxOut
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	bs.utxoSet = m
	fmt.Printf("🔄 UTXO set loaded (%d entries)\n", len(m))
	return nil
}

func (bs *BlockchainServer) saveUTXOSet() {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	data, err := json.MarshalIndent(bs.utxoSet, "", "  ")
	if err != nil {
		log.Printf("Error marshaling UTXO set: %v", err)
		return
	}
	if err := os.WriteFile(utxoFile, data, 0644); err != nil {
		log.Printf("Error writing UTXO set: %v", err)
		return
	}
}
