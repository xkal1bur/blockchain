package main

import (
	"encoding/json"
	"fmt"
	"net"
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

// NewTx creates a new transaction
func NewTx(version uint32, locktime uint32) *Tx {
	return &Tx{
		Version:   version,
		TxIns:     make([]TxIn, 0),
		TxOuts:    make([]TxOut, 0),
		Locktimei: locktime,
	}
}

// AddInput adds an input to the transaction
func (tx *Tx) AddInput(prevTx []byte, prevIndex uint32, sequence uint32, net string) {
	txin := TxIn{
		PrevTx:    prevTx,
		PrevIndex: prevIndex,
		Sequence:  sequence,
		Witness:   make([][]byte, 0),
		Net:       net,
	}
	tx.TxIns = append(tx.TxIns, txin)
}

// AddOutput adds an output to the transaction
func (tx *Tx) AddOutput(amount uint64, address string) {
	txout := TxOut{
		Amount:  amount,
		Address: address,
	}
	tx.TxOuts = append(tx.TxOuts, txout)
}

// ToJSON converts the transaction to JSON
func (tx *Tx) ToJSON() (string, error) {
	jsonData, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// CreateSampleTransaction creates a sample transaction for testing
func CreateSampleTransaction() *Tx {
	tx := NewTx(1, 0)

	// Add sample input (spending a previous transaction)
	prevTxHash := []byte("abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	tx.AddInput(prevTxHash, 0, 0xffffffff, "main")

	// Add sample outputs
	tx.AddOutput(100000, "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa") // 100,000 satoshis to Alice
	tx.AddOutput(50000, "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2")  // 50,000 satoshis to Bob
	tx.AddOutput(25000, "1C4bc762dd5423e332166702cb75f40df79") // 25,000 satoshis change

	return tx
}

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "192.168.235.29:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Create a sample transaction
	tx := CreateSampleTransaction()

	// Convert transaction to JSON
	txJSON, err := tx.ToJSON()
	if err != nil {
		fmt.Println("Error converting transaction to JSON:", err)
		return
	}

	// Create message with transaction
	message := fmt.Sprintf("TRANSACTION:%s", txJSON)

	fmt.Printf("Sending transaction to server...\n")
	fmt.Printf("Transaction JSON: %s\n", txJSON)

	// Send message to the server
	_, err = conn.Write([]byte(message + "\n"))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	// Read response from server
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Printf("Server response: %s", string(response[:n]))
}
