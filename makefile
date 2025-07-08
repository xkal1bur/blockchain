# 🛠 Ruta base del CLI
CLI_PATH=cmd/cli/main.go

# 🧾 Comandos disponibles
.PHONY: all help server wallet tx status

all: help

help:
	@echo "📚 Comandos disponibles:"
	@echo "  make wallet    - Crear o ver una wallet"
	@echo "  make tx        - Enviar transacción"
	@echo "  make status    - Consultar estado del nodo"
	@echo "  make server    - Iniciar el servidor blockchain"

wallet:
	go run $(CLI_PATH) wallet

tx:
	go run $(CLI_PATH) tx

status:
	go run $(CLI_PATH) status

server:
	go run $(CLI_PATH) server
