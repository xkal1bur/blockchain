package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type StatusResponse struct {
	BlockCount   int    `json:"block_count"`
	LatestHash   string `json:"latest_hash"`
	TotalTxs     int    `json:"total_txs"`
	TotalBalance uint64 `json:"total_balance,omitempty"`
	Message      string `json:"message,omitempty"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("📥 Dirección opcional para consultar balance (enter para omitir): ")
	address, _ := reader.ReadString('\n')
	address = strings.TrimSpace(address)

	// Conectar al nodo
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("❌ Error conectando con el nodo:", err)
		return
	}
	defer conn.Close()

	// Enviar mensaje STATUS o STATUS:<address>
	message := "STATUS"
	if address != "" {
		message += ":" + address
	}
	fmt.Fprintf(conn, message+"\n")

	// Leer respuesta
	responseData, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("❌ Error leyendo respuesta del nodo:", err)
		return
	}

	var status StatusResponse
	if err := json.Unmarshal([]byte(responseData), &status); err != nil {
		fmt.Println("❌ Error al parsear respuesta JSON:", err)
		return
	}

	fmt.Println("\n📊 Estado del nodo:")
	fmt.Printf("🔢 Bloques: %d\n", status.BlockCount)
	fmt.Printf("🔗 Último hash: %s\n", status.LatestHash)
	fmt.Printf("🧾 Transacciones totales: %d\n", status.TotalTxs)

	if address != "" {
		fmt.Printf("💰 Balance para %s: %d HORUS\n", address, status.TotalBalance)
	}
}
