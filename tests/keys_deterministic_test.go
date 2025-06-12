package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcutil/base58"

	"github.com/xkal1bur/blockchain/pkg/core"
	"golang.org/x/crypto/ripemd160"
)

func TestDebugKnownKeypair(t *testing.T) {
	// 📥 1. Clave privada conocida (Mastering Bitcoin)
	hexPriv := "1E99423A4ED27608A15A2616A2B0E9E52CED330AC530EDCC32C8FFC6A526AEDD"
	fmt.Println("🔐 Clave privada (hex):", hexPriv)

	// 🔄 2. Derivar clave pública desde ella
	pub, err := core.PublicKeyFromHex(hexPriv)
	if err != nil {
		t.Fatalf("❌ Error derivando clave pública: %v", err)
	}

	// 📈 3. Mostrar coordenadas de la clave pública
	x := strings.ToUpper(pub.X().Text(16))
	y := strings.ToUpper(pub.Y().Text(16))
	x = padHex(x)
	y = padHex(y)

	fmt.Println("📍 Public Key X:", x)
	fmt.Println("📍 Public Key Y:", y)

	// ✅ 4. Comprobar con valores esperados
	expectedX := "F028892BAD7ED57D2FB57BF33081D5CFCF6F9ED3D3D7F159C2E2FFF579DC341A"
	expectedY := "07CF33DA18BD734C600B96A72BBC4749D5141C90EC8AC328AE52DDFE2E505BDB"

	if x != expectedX {
		fmt.Println("❌ X no coincide. Esperado:", expectedX)
	} else {
		fmt.Println("✅ X coincide con el esperado")
	}
	if y != expectedY {
		fmt.Println("❌ Y no coincide. Esperado:", expectedY)
	} else {
		fmt.Println("✅ Y coincide con el esperado")
	}

	// 🏁 5. Obtener dirección Bitcoin
	address, err := core.GetBTCAddress(pub, false, "main") // true = clave comprimida, false = no comprimida
	if err != nil {
		t.Fatalf("❌ Error generando dirección Bitcoin: %v", err)
	}
	fmt.Println("🏦 Dirección Bitcoin:", address)

	// 🎯 Valor esperado (opcional — por compatibilidad Base58Check)
	expectedAddr := "1424C2F4bC9JidNjjTUZCbUxv6Sa1Mt62x"
	if address != expectedAddr {
		fmt.Println("❌ Dirección no coincide. Esperada:", expectedAddr)
	} else {
		fmt.Println("✅ Dirección coincide con la esperada")
	}
}

// Añade ceros a la izquierda hasta llegar a 64 caracteres (32 bytes en hex)
func padHex(h string) string {
	return strings.Repeat("0", 64-len(h)) + h
}

func TestStepByStep(t *testing.T) {
	privHex := "1E99423A4ED27608A15A2616A2B0E9E52CED330AC530EDCC32C8FFC6A526AEDD"
	bytesPriv, _ := hex.DecodeString(privHex)
	priv, _ := btcec.PrivKeyFromBytes(bytesPriv)
	pub := priv.PubKey()

	// No comprimido
	pubBytes := pub.SerializeUncompressed()
	fmt.Println("PubKey (hex):", hex.EncodeToString(pubBytes))

	// Hash160
	sha := sha256.Sum256(pubBytes)
	r := ripemd160.New()
	r.Write(sha[:])
	h160 := r.Sum(nil)
	fmt.Println("Hash160:", hex.EncodeToString(h160))

	// Base58Check
	version := byte(0x00)
	payload := append([]byte{version}, h160...)
	check := sha256.Sum256(payload)
	check = sha256.Sum256(check[:])
	full := append(payload, check[:4]...)
	address := base58.Encode(full)

	fmt.Println("Dirección generada:", address)

	expected := "1424C2F4bC9JidNjjTUZCbUxv6Sa1Mt62x"
	if address != expected {
		t.Errorf("❌ Dirección no coincide. Esperada: %s, obtenida: %s", expected, address)
	} else {
		fmt.Println("✅ Dirección coincide con la esperada")
	}
}
