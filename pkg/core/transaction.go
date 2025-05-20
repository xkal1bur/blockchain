package core

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/xkal1bur/blockchain/pkg/crypto"
)

func DecodeInt(r io.Reader, nbytes int) (uint64, error) {
	buf := make([]byte, nbytes)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(append(buf, make([]byte, 8-nbytes)...)), nil
}

func EncodeInt(i uint64, nbytes int) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)
	return buf[:nbytes]
}

func DecodeVarInt(r io.Reader) (uint64, error) {
	prefix := make([]byte, 1)
	if _, err := r.Read(prefix); err != nil {
		return 0, err
	}
	switch prefix[0] {
	case 0xfd:
		return DecodeInt(r, 2)
	case 0xfe:
		return DecodeInt(r, 4)
	case 0xff:
		return DecodeInt(r, 8)
	default:
		return uint64(prefix[0]), nil
	}
}

func EncodeVarInt(i uint64) []byte {
	switch {
	case i <= 0xfc:
		return []byte{byte(i)}
	case i <= 0xffff:
		val := EncodeInt(i, 2)
		return append([]byte{0xfd}, val...)
	case i <= 0xffffffff:
		val := EncodeInt(i, 4)
		return append([]byte{0xfe}, val...)
	case i <= 0xffffffffffffffff:
		val := EncodeInt(i, 8)
		return append([]byte{0xff}, val...)
	default:
		return nil
	}
}

// ------------------------------------------------------

type TxFetcher struct{ CacheDir string }
type Tx struct {
	Version   uint32
	TxIns     []TxIn
	TxOuts    []TxOut
	Locktimei uint32
}
type TxIn struct {
	PrevTx    []byte
	PrevIndex uint32
	ScriptSig *Script
	Sequence  uint32
	Witness   [][]byte
	Net       string
}
type TxOut struct {
	Amount       uint64
	ScriptPubKey *Script
}

// Script represents a Bitcoin script
type Script struct {
	Data []byte
}

// TxFetcher
func NewTxFetcher() *TxFetcher { return &TxFetcher{CacheDir: "txdb"} }

func (f *TxFetcher) Fetch(txid string, net string) (*Tx, error) {
	txid = string(bytes.ToLower([]byte(txid)))
	cachePath := filepath.Join(f.CacheDir, txid)
	var raw []byte
	var err error
	if _, err = os.Stat(cachePath); err == nil {
		raw, err = os.ReadFile(cachePath)
	} else {
		var url string
		switch net {
		case "main":
			url = fmt.Sprintf("https://blockstream.info/api/tx/%s/hex", txid)
		case "test":
			url = fmt.Sprintf("https://blockstream.info/testnet/api/tx/%s/hex", txid)
		}
		resp, _ := http.Get(url)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("txid %s not found", txid)
		}
		body, _ := io.ReadAll(resp.Body)
		raw, err = hex.DecodeString(string(bytes.TrimSpace(body)))
		os.MkdirAll(f.CacheDir, 0755)
		os.WriteFile(cachePath, raw, 0644)
	}
	tx, err := DecodeTx(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %v", err)
	}
	if tx.ID() != txid {
		return nil, fmt.Errorf("decoded transaction id mismatch")
	}
	return tx, nil
}

// Transaction
func DecodeTx(s io.Reader) (*Tx, error) {
	var tx Tx
	var err error

	version, err := DecodeInt(s, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %v", err)
	}
	tx.Version = uint32(version)

	segwit := false
	num_inputs, err := DecodeVarInt(s)
	if err != nil {
		return nil, fmt.Errorf("failed to read number of inputs: %v", err)
	}
	if num_inputs == 0 {
		segwit = true
		num_inputs, err = DecodeVarInt(s)
		if err != nil {
			return nil, fmt.Errorf("failed to read number of inputs (segwit): %v", err)
		}
	}

	tx.TxIns = make([]TxIn, num_inputs)
	for i := range tx.TxIns {
		tx.TxIns[i], err = DecodeTxIn(s)
		if err != nil {
			return nil, fmt.Errorf("failed to decode input %d: %v", i, err)
		}
	}

	num_outputs, err := DecodeVarInt(s)
	if err != nil {
		return nil, fmt.Errorf("failed to read number of outputs: %v", err)
	}
	tx.TxOuts = make([]TxOut, num_outputs)
	for i := range tx.TxOuts {
		tx.TxOuts[i], err = DecodeTxOut(s)
		if err != nil {
			return nil, fmt.Errorf("failed to decode output %d: %v", i, err)
		}
	}

	if segwit {
		for i := range tx.TxIns {
			num_witness, err := DecodeVarInt(s)
			if err != nil {
				return nil, fmt.Errorf("failed to read number of witness items: %v", err)
			}
			tx.TxIns[i].Witness = make([][]byte, num_witness)
			for j := range tx.TxIns[i].Witness {
				witness_item, err := DecodeVarInt(s)
				if err != nil {
					return nil, fmt.Errorf("failed to read witness item %d for input %d: %v", j, i, err)
				}
				if witness_item == 0 {
					tx.TxIns[i].Witness[j] = []byte{}
				} else {
					tx.TxIns[i].Witness[j] = make([]byte, witness_item)
					if _, err := io.ReadFull(s, tx.TxIns[i].Witness[j]); err != nil {
						return nil, fmt.Errorf("failed to read witness item %d for input %d: %v", j, i, err)
					}
				}
			}
		}
	}

	locktime, err := DecodeInt(s, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to read locktime: %v", err)
	}
	tx.Locktimei = uint32(locktime)
	return &tx, nil
}

func (tx *Tx) ID() string {
	raw := tx.Encode(true, -1)
	hash := crypto.Sha256Edu(raw)
	hash2 := crypto.Sha256Edu(hash[:])
	reversed := reverseBytes(hash2[:])
	return hex.EncodeToString(reversed)
}

func reverseBytes(b []byte) []byte {
	n := len(b)
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = b[n-1-i]
	}
	return out
}

func DecodeTxIn(s io.Reader) (TxIn, error) {
	var txin TxIn
	var err error

	prevTx := make([]byte, 32)
	if _, err := io.ReadFull(s, prevTx); err != nil {
		return TxIn{}, fmt.Errorf("failed to read prev tx: %v", err)
	}
	// Reverse the bytes since transaction IDs are stored in little-endian
	txin.PrevTx = reverseBytes(prevTx)

	prevIndex, err := DecodeInt(s, 4)
	if err != nil {
		return TxIn{}, fmt.Errorf("failed to read prev index: %v", err)
	}
	txin.PrevIndex = uint32(prevIndex)

	scriptLen, err := DecodeVarInt(s)
	if err != nil {
		return TxIn{}, fmt.Errorf("failed to read script length: %v", err)
	}
	script := make([]byte, scriptLen)
	if _, err := io.ReadFull(s, script); err != nil {
		return TxIn{}, fmt.Errorf("failed to read script: %v", err)
	}
	txin.ScriptSig = &Script{Data: script}

	sequence, err := DecodeInt(s, 4)
	if err != nil {
		return TxIn{}, fmt.Errorf("failed to read sequence: %v", err)
	}
	txin.Sequence = uint32(sequence)

	return txin, nil
}

func DecodeTxOut(s io.Reader) (TxOut, error) {
	var txout TxOut
	var err error

	amount, err := DecodeInt(s, 8)
	if err != nil {
		return TxOut{}, fmt.Errorf("failed to read amount: %v", err)
	}
	txout.Amount = amount

	scriptLen, err := DecodeVarInt(s)
	if err != nil {
		return TxOut{}, fmt.Errorf("failed to read script length: %v", err)
	}
	script := make([]byte, scriptLen)
	if _, err := io.ReadFull(s, script); err != nil {
		return TxOut{}, fmt.Errorf("failed to read script: %v", err)
	}
	txout.ScriptPubKey = &Script{Data: script}

	return txout, nil
}

// Idk if this is it works xd

func DecodeTxBytes(data []byte) (*Tx, error) {
	return DecodeTx(bytes.NewReader(data))
}

// vibe coded D:

func (tx *Tx) Encode(force_legacy bool, sig_index int) []byte {
	var out []byte

	// encode metadata
	version := EncodeInt(uint64(tx.Version), 4)
	out = append(out, version...)

	// Check if this is a segwit transaction
	segwit := len(tx.TxIns) > 0 && len(tx.TxIns[0].Witness) > 0
	if segwit && !force_legacy {
		out = append(out, 0x00, 0x01) // segwit marker + flag bytes
	}

	// encode inputs
	num_inputs := EncodeVarInt(uint64(len(tx.TxIns)))
	out = append(out, num_inputs...)

	if sig_index == -1 {
		// encode all inputs normally
		for _, txin := range tx.TxIns {
			out = append(out, txin.Encode(false)...)
		}
	} else {
		// encode inputs with script override for the signing input
		for i, txin := range tx.TxIns {
			out = append(out, txin.Encode(i == sig_index)...)
		}
	}

	// encode outputs
	num_outputs := EncodeVarInt(uint64(len(tx.TxOuts)))
	out = append(out, num_outputs...)
	for _, txout := range tx.TxOuts {
		out = append(out, txout.Encode()...)
	}

	// encode witnesses
	if segwit && !force_legacy {
		for _, txin := range tx.TxIns {
			num_witness := EncodeVarInt(uint64(len(txin.Witness)))
			out = append(out, num_witness...)
			for _, item := range txin.Witness {
				if len(item) == 0 {
					// Empty witness item
					out = append(out, 0x00)
				} else {
					witness_len := EncodeVarInt(uint64(len(item)))
					out = append(out, witness_len...)
					out = append(out, item...)
				}
			}
		}
	}

	// encode locktime
	locktime := EncodeInt(uint64(tx.Locktimei), 4)
	out = append(out, locktime...)

	// Add SIGHASH_ALL if this is for signing
	if sig_index != -1 {
		sighash := EncodeInt(1, 4) // 1 = SIGHASH_ALL
		out = append(out, sighash...)
	}

	return out
}

// Add Encode methods for TxIn and TxOut
func (txin *TxIn) Encode(script_override bool) []byte {
	var out []byte

	// Previous transaction ID (32 bytes)
	out = append(out, txin.PrevTx...)

	// Previous output index (4 bytes)
	prevIndex := EncodeInt(uint64(txin.PrevIndex), 4)
	out = append(out, prevIndex...)

	// Script
	if script_override {
		// Empty script for signing
		out = append(out, 0x00)
	} else {
		scriptLen := EncodeVarInt(uint64(len(txin.ScriptSig.Data)))
		out = append(out, scriptLen...)
		out = append(out, txin.ScriptSig.Data...)
	}

	// Sequence (4 bytes)
	sequence := EncodeInt(uint64(txin.Sequence), 4)
	out = append(out, sequence...)

	return out
}

func (txout *TxOut) Encode() []byte {
	var out []byte

	// Amount (8 bytes)
	amount := EncodeInt(txout.Amount, 8)
	out = append(out, amount...)

	// Script length and script
	scriptLen := EncodeVarInt(uint64(len(txout.ScriptPubKey.Data)))
	out = append(out, scriptLen...)
	out = append(out, txout.ScriptPubKey.Data...)

	return out
}
