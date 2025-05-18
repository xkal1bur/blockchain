package core

import (
	"bytes"
)

type Block struct {
	Version    uint64
	PrevBlock  []byte // 32 bytes, ojalá
	MerkleRoot []byte // 32 bytes, ojalá
	Timestamp  uint64
	Nonce      uint64
	// maybe difficulty?
}

/*
Encoding: block -> bytes
-------------------------
La estructura pensada es:
- Version: 		4 bytes
- PrevBlock: 	32 bytes
- MerkleRoot: 	32 bytes
- Timestamp: 	4 bytes
- Nonce: 		4 bytes
*/

func EncodeBlock(b *Block) ([]byte, error) {
	buf := new(bytes.Buffer)
	// Encode the version
	version, err := EncodeInt(b.Version, 4)
	if err != nil {
		return nil, err
	}

	prevBlock := b.PrevBlock
	merkleRoot := b.MerkleRoot

	// Encode the timestamp
	timestamp, err := EncodeInt(b.Timestamp, 4)
	if err != nil {
		return nil, err
	}

	// Encode the nonce
	nonce, err := EncodeInt(b.Nonce, 4)
	if err != nil {
		return nil, err
	}

	// Write to buffer
	buf.Write(version)
	buf.Write(prevBlock)
	buf.Write(merkleRoot)
	buf.Write(timestamp)
	buf.Write(nonce)

	// Return the encoded bytes :3
	return buf.Bytes(), nil
}

func DecodeBlock(encodedBlock []byte) (*Block, error) {
	buf := bytes.NewReader(encodedBlock)

	// Decode the version
	version, err := DecodeInt(buf, 4)
	if err != nil {
		return nil, err
	}

	// Decode the PrevBlock
	prevBlock := make([]byte, 32)
	_, err = buf.Read(prevBlock)
	if err != nil {
		return nil, err
	}

	// Decode the MerkleRoot
	merkleRoot := make([]byte, 32)
	_, err = buf.Read(merkleRoot)
	if err != nil {
		return nil, err
	}

	// Decode the timestamp
	timestamp, err := DecodeInt(buf, 4)
	if err != nil {
		return nil, err
	}

	// Decode the nonce
	nonce, err := DecodeInt(buf, 4)
	if err != nil {
		return nil, err
	}

	block := &Block{
		Version:    version,
		PrevBlock:  prevBlock,
		MerkleRoot: merkleRoot,
		Timestamp:  timestamp,
		Nonce:      nonce,
	}

	return block, nil
}

func ValidateBlock(b *Block) bool {
	// will depend on PoW or PoS
	return true
}
