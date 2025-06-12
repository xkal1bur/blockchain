package tests

import (
	"testing"

	"github.com/xkal1bur/blockchain/pkg/core"
)

func TestGenerateKeyPair(t *testing.T) {
	priv, pub, err := core.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Error generating key pair: %v", err)
	}

	if priv == nil || pub == nil {
		t.Fatal("Private or public key is nil")
	}
}

func TestBTCAddress(t *testing.T) {
	_, pub, _ := core.GenerateKeyPair() // Genera priv, pub key
	addr, err := core.GetBTCAddress(pub, true, "main")
	if err != nil {
		t.Fatalf("Error generating address: %v", err)
	}

	if len(addr) < 26 || len(addr) > 35 {
		t.Errorf("Invalid address length: %d", len(addr))
	}
}

func TestAddressToPKHash(t *testing.T) {
	_, pub, _ := core.GenerateKeyPair()
	addr, _ := core.GetBTCAddress(pub, true, "main")
	hash1 := core.Hash160(core.EncodePublicKey(pub, true))

	hash2, err := core.AddressToPKHash(addr)
	if err != nil {
		t.Fatalf("Error decoding address: %v", err)
	}

	if string(hash1) != string(hash2) {
		t.Error("Hash mismatch between pubkey and decoded address")
	}
}
