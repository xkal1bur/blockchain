package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Create a message with timestamp
	message := fmt.Sprintf("Hello from client at %s", time.Now().Format("15:04:05"))

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

	fmt.Printf("Sent: %s\n", message)
	fmt.Printf("Server response: %s", string(response[:n]))
}
