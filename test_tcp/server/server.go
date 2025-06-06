package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func main() {
	// Listen on port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on port 8080...")

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
		fmt.Printf("Received from %s: %s\n", clientAddr, message)

		// Send response back to client
		response := fmt.Sprintf("Server received message #%d at %s\n",
			messageCount, time.Now().Format("15:04:05"))

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
