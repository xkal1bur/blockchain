package main

import (
	"fmt"
	"log"
	"net"
	"os"

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

		// Handle each connection in a goroutine
		go server.HandleConnection(conn)
	}
}
