package tests

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

// --- Estructuras para el pipeline (con guion bajo para evitar conflictos) ---

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

// --- Funciones auxiliares para crear transacciones y bloques ---

func CreateSampleTransaction_() *Tx_ {
	tx := &Tx_{
		Version:   1,
		TxIns:     []TxIn_{},
		TxOuts:    []TxOut_{},
		Locktimei: 0,
	}
	prevTxHash := []byte("abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	tx.TxIns = append(tx.TxIns, TxIn_{
		PrevTx:    prevTxHash,
		PrevIndex: 0,
		Sequence:  0xffffffff,
		Witness:   [][]byte{},
		Net:       "main2",
	})
	tx.TxOuts = append(tx.TxOuts, TxOut_{Amount: 100000, Address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"})
	tx.TxOuts = append(tx.TxOuts, TxOut_{Amount: 50000, Address: "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"})
	tx.TxOuts = append(tx.TxOuts, TxOut_{Amount: 25000, Address: "1C4bc762dd5423e332166702cb75f40df79"})
	return tx
}

func CreateSampleBlock_(prevBlock []byte, txs []*Tx_) *Block_ {
	return &Block_{
		Version:    1,
		PrevBlock:  prevBlock,
		MerkleRoot: []byte("dummymerklerootdummymerklerootdummymerkl"), // 32 bytes dummy
		Timestamp:  uint64(time.Now().Unix()),
		Nonce:      0,
		Txs:        txs,
	}
}

// --- Función para enviar mensaje al servidor y recibir respuesta ---

func sendMessageToServer_(messageType string, payload interface{}) (string, error) {
	conn, err := net.Dial("tcp", "192.168.235.226:8080")
	if err != nil {
		return "", fmt.Errorf("error connecting to server: %v", err)
	}
	defer conn.Close()

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error serializing payload: %v", err)
	}

	message := fmt.Sprintf("%s:%s\n", messageType, string(jsonData))
	_, err = conn.Write([]byte(message))
	if err != nil {
		return "", fmt.Errorf("error sending message: %v", err)
	}

	response := make([]byte, 2048)
	n, err := conn.Read(response)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	return string(response[:n]), nil
}

// --- Test principal del pipeline ---

func TestBlockchainPipeline(t *testing.T) {
	// 1. Crear transacción
	tx := CreateSampleTransaction_()

	// 2. Crear bloque con la transacción
	block := CreateSampleBlock_(make([]byte, 32), []*Tx_{tx})

	// 3. Enviar transacción al servidor y verificar respuesta
	respTx, err := sendMessageToServer_("TRANSACTION", tx)
	if err != nil {
		t.Fatalf("Error al enviar transacción: %v", err)
	}
	t.Logf("Respuesta del servidor (transacción): %s", respTx)

	// 4. Enviar bloque al servidor y verificar respuesta
	respBlock, err := sendMessageToServer_("BLOCK", block)
	if err != nil {
		t.Fatalf("Error al enviar bloque: %v", err)
	}
	t.Logf("Respuesta del servidor (bloque): %s", respBlock)

	// 5. (Opcional) Verificar que los archivos se hayan creado
	if _, err := os.Stat("transactions"); os.IsNotExist(err) {
		t.Errorf("No se encontró el directorio de transacciones")
	}
	if _, err := os.Stat("blocks"); os.IsNotExist(err) {
		t.Errorf("No se encontró el directorio de bloques")
	}
}
