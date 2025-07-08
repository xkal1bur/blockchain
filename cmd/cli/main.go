package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("❌ Uso: go run cmd/cli/main.go <comando>")
		fmt.Println("📚 Comandos disponibles:")
		fmt.Println("  tx       - Crear y enviar una transacción")
		fmt.Println("  status   - Consultar el estado del nodo")
		fmt.Println("  wallet   - Crear o ver una wallet")
		fmt.Println("  server   - Iniciar el servidor blockchain")
		return
	}

	switch os.Args[1] {
	case "tx":
		runCLI("tx_cli.go")
	case "status":
		runCLI("status_cli.go")
	case "wallet":
		runCLI("wallet_cli.go")
	case "server":
		runServer()
	default:
		fmt.Printf("❌ Comando no reconocido: %s\n", os.Args[1])
	}
}

func runCLI(file string) {
	cmd := exec.Command("go", "run", filepath.Join("cmd", "cli", file))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Error ejecutando %s: %v\n", file, err)
	}
}

func runServer() {
	cmd := exec.Command("go", "run", filepath.Join("cmd", "server", "server.go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Error al iniciar el servidor: %v\n", err)
	}
}
