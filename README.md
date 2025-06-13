# Blockchain Project Documentation

## 📁 Estructura del Proyecto

Este proyecto implementa una blockchain básica en Go con arquitectura modular separada en dos directorios principales:

### 📂 `/cmd` - Aplicaciones Ejecutables
Contiene las aplicaciones de línea de comandos del proyecto:

#### 🏦 `cmd/wallet/` - Generador de Carteras
- **Archivo**: `key_creation.go`
- **Propósito**: Crear y gestionar carteras criptográficas
- **Funcionalidad**:
  - Genera nuevas claves criptográficas ECDSA usando curva P-256
  - Crea archivos de cartera en formato JSON
  - Carga carteras existentes desde disco
  - Muestra información segura de la cartera (oculta claves privadas)
  - Implementa avisos de seguridad para el manejo de claves

#### 🌐 `cmd/client/` - Cliente de Blockchain
- **Archivo**: `client.go`
- **Propósito**: Interactuar con el servidor de blockchain
- **Funcionalidad**:
  - Crea/carga carteras de cliente
  - Se conecta al servidor TCP en el puerto 8081
  - Construye y firma transacciones usando ECDSA
  - Envía transacciones al servidor con claves públicas para validación
  - Maneja respuestas del servidor

#### 🖥️ `cmd/server/` - Servidor de Blockchain
- **Archivo**: `server.go`
- **Propósito**: Ejecutar el nodo servidor de blockchain
- **Funcionalidad**:
  - Escucha conexiones TCP en el puerto 8081
  - Acepta múltiples conexiones concurrentes
  - Procesa dos tipos de mensajes:
    - `TRANSACTION:<json>` - Transacciones con claves públicas
    - `BLOCK:<json>` - Bloques validados de otros nodos
  - Muestra información detallada de transacciones recibidas
  - Maneja configuración de nodos peer

### 📦 `/pkg` - Paquetes Reutilizables

#### 🔧 `pkg/core/` - Lógica Central de Blockchain
Contiene la implementación principal de la blockchain:

##### `blockchainserver.go`
- **Clase**: `BlockchainServer`
- **Responsabilidades**:
  - Gestión de transacciones pendientes (mempool)
  - Minería de bloques con Proof of Work
  - Validación de bloques recibidos
  - Persistencia de blockchain en disco
  - Comunicación peer-to-peer
  - Validación de firmas ECDSA

##### `wallet.go`
- **Clase**: `Wallet`
- **Responsabilidades**:
  - Generación de claves ECDH/ECDSA P-256
  - Persistencia de carteras en JSON
  - Firma de datos con ECDSA
  - Generación de direcciones blockchain
  - Conversión entre formatos de claves

##### `block.go`
- **Clase**: `Block`
- **Responsabilidades**:
  - Estructura de bloques blockchain
  - Cálculo de hash SHA3-256
  - Implementación de Proof of Work
  - Validación de dificultad (bits de ceros)
  - Validación de transacciones en bloques

##### `transaction.go`
- **Clase**: `Tx`, `TxIn`, `TxOut`
- **Responsabilidades**:
  - Estructura de transacciones UTXO
  - Generación de IDs únicos
  - Validación de firmas ECDSA
  - Serialización para firma
  - Verificación de integridad

#### 🔐 `pkg/crypto/` - Implementaciones Criptográficas
Implementaciones educativas de algoritmos de hash:

##### `sha2_256.go`
- **Funcionalidad**: Implementación completa de SHA-256 desde cero
- **Características**:
  - Funciones bitwise (rotación, desplazamiento)
  - Generación de constantes K y H
  - Padding de mensajes
  - Procesamiento por bloques de 512 bits

##### `sha3_256.go`
- **Funcionalidad**: Implementación completa de SHA3-256 desde cero
- **Características**:
  - Permutación Keccak-f[1600]
  - Padding pad10*1
  - Fases de absorción y exprimido
  - 24 rondas de transformación

##### Tests incluidos
- `sha2_256_test.go` y `sha3_256_test.go` con casos de prueba

#### 🛠️ `pkg/utils/` - Utilidades
- Directorio presente pero sin contenido actual
- Preparado para funciones auxiliares futuras

## 🚀 Cómo Usar el Sistema

### 1. Generar una Cartera
```bash
cd cmd/wallet
go run key_creation.go
```

### 2. Ejecutar el Servidor
```bash
cd cmd/server
go run server.go [peer1:port] [peer2:port]
```

### 3. Ejecutar el Cliente
```bash
cd cmd/client
go run client.go
```

## 🔧 Arquitectura Técnica

### Seguridad Criptográfica
- **Curva Elíptica**: P-256 (NIST) para todas las operaciones
- **Firmas**: ECDSA con formato r||s (64 bytes)
- **Hash**: SHA3-256 para bloques, SHA-256 para direcciones
- **Claves**: Generación segura con crypto/rand

### Protocolo de Red
- **Transporte**: TCP puro (puerto 8081)
- **Formato**: Mensajes JSON estructurados
- **Tipos**: `TRANSACTION:` y `BLOCK:`
- **Concurrencia**: Goroutines para múltiples conexiones

### Consenso
- **Algoritmo**: Proof of Work
- **Dificultad**: Configurable (bits de ceros)
- **Minería**: Búsqueda incremental de nonce
- **Validación**: Hash y verificación de firmas

### Persistencia
- **Formato**: JSON para carteras y blockchain
- **Archivos**: `wallet.json`, `blockchain.json`
- **Sincronización**: Mutex para acceso concurrente

## ⚠️ Consideraciones de Seguridad

- Las claves privadas se almacenan en texto plano en JSON
- Implementación educativa, no para producción
- Falta encriptación de comunicaciones
- Validación de entrada limitada
- No implementa recuperación de claves

## 🧪 Propósito Educativo

Este proyecto está diseñado para demostrar:
- Conceptos fundamentales de blockchain
- Criptografía aplicada (ECDSA, SHA)
- Arquitectura de sistemas distribuidos
- Protocolos de consenso
- Networking TCP en Go

---

*Proyecto desarrollado con fines educativos para el curso de Ética y Seguridad de los Datos*

