package tests

import (
	"testing"
	"github.com/xkal1bur/blockchain/pkg/core"
	"fmt"
	"bytes"
	"encoding/hex"
)
/*
func TestVarIntEncodingDecoding(t *testing.T) {
	valores := []uint64{
		0xfc,           // 1 byte
		0xfd,           // exacto para 0xfd prefix
		0x1234,         // 2 bytes
		0x10000,        // 4 bytes
		0x12345678,     // 4 bytes
		0x100000000,    // 8 bytes
		0x123456789abc, // 8 bytes
		65535,
	}

	for _, v := range valores {
		// Encode
		encoded, err := core.EncodeVarInt(v)
		if err != nil {
			t.Errorf("error al codificar %d: %v", v, err)
			continue
		}

		buf := bytes.NewBuffer(encoded)
		decoded, err := core.DecodeVarjnt(buf)
		if err != nil {
			t.Errorf("error al decodificar %d: %v", v, err)
			continue
		}

		// Print resultado completo
		fmt.Printf("Original: %d\n", v)
		fmt.Printf("Encoded (%d bytes): ", len(encoded))
		for _, b := range encoded {
			fmt.Printf("%02x ", b)
		}
		fmt.Printf("\nDecoded: %d\n", decoded)
		fmt.Println("-------------")

		if decoded != v {
			t.Errorf("valor distinto: original=%d decodificado=%d", v, decoded)
		}
	}
}
	*/

func TestLegacyTransactionDecode(t *testing.T) {
	// Example taken from Programming Bitcoin, Chapter 5
	raw, err := hex.DecodeString("0100000001813f79011acb80925dfe69b3def355fe914bd1d96a3f5f71bf8303c6a989c7d1000000006b483045022100ed81ff192e75a3fd2304004dcadb746fa5e24c5031ccfcf21320b0277457c98f02207a986d955c6e0cb35d446a89d3f56100f4d7f67801c31967743a9c8e10615bed01210349fc4e631e3624a545de3f89f5d8684c7b8138bd94bdd531d2e213bf016b278afeffffff02a135ef01000000001976a914bc3b654dca7e56b04dca18f2566cdaf02e8d9ada88ac99c39800000000001976a9141c4bc762dd5423e332166702cb75f40df79fea1288ac19430600")
	if err != nil {
		t.Fatalf("failed to decode hex: %v", err)
	}

	tx, err := core.DecodeTx(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("failed to decode transaction: %v", err)
	}

	// Test metadata parsing
	if tx.Version != 1 {
		t.Errorf("expected version 1, got %d", tx.Version)
	}

	// Test input parsing
	if len(tx.TxIns) != 1 {
		t.Errorf("expected 1 input, got %d", len(tx.TxIns))
	}

	expectedPrevTx, _ := hex.DecodeString("d1c789a9c60383bf715f3f6ad9d14b91fe55f3deb369fe5d9280cb1a01793f81")
	if !bytes.Equal(tx.TxIns[0].PrevTx, expectedPrevTx) {
		t.Errorf("expected prev_tx %x, got %x", expectedPrevTx, tx.TxIns[0].PrevTx)
	}

	if tx.TxIns[0].PrevIndex != 0 {
		t.Errorf("expected prev_index 0, got %d", tx.TxIns[0].PrevIndex)
	}

	if tx.TxIns[0].Sequence != 0xfffffffe {
		t.Errorf("expected sequence 0xfffffffe, got %x", tx.TxIns[0].Sequence)
	}

	// Test output parsing
	if len(tx.TxOuts) != 2 {
		t.Errorf("expected 2 outputs, got %d", len(tx.TxOuts))
	}

	if tx.TxOuts[0].Amount != 32454049 {
		t.Errorf("expected amount 32454049, got %d", tx.TxOuts[0].Amount)
	}

	if tx.TxOuts[1].Amount != 10011545 {
		t.Errorf("expected amount 10011545, got %d", tx.TxOuts[1].Amount)
	}

	// Test locktime parsing
	if tx.Locktimei != 410393 {
		t.Errorf("expected locktime 410393, got %d", tx.Locktimei)
	}

	// Print transaction details for debugging
	fmt.Printf("Transaction decoded successfully:\n")
	fmt.Printf("Version: %d\n", tx.Version)
	fmt.Printf("Inputs: %d\n", len(tx.TxIns))
	fmt.Printf("Outputs: %d\n", len(tx.TxOuts))
	fmt.Printf("Locktime: %d\n", tx.Locktimei)
}


