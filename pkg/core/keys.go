package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/ripemd160"

	"github.com/btcsuite/btcd/btcec/v2"

	"github.com/btcsuite/btcutil/base58"
)

var (
	versionMainnet = byte(0x00)
	versionTestnet = byte(0x6f)
)

// GenerateKeyPair generates a new private/public keypair
func GenerateKeyPair() (*btcec.PrivateKey, *btcec.PublicKey, error) {
	priv, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	return priv, priv.PubKey(), nil
}

// PublicKeyFromHex returns a public key from a hex private key string
func PublicKeyFromHex(hexPriv string) (*btcec.PublicKey, error) {
	b, err := hex.DecodeString(hexPriv)
	if err != nil {
		return nil, err
	}
	priv, _ := btcec.PrivKeyFromBytes(b)
	return priv.PubKey(), nil
}

// EncodePublicKey encodes the public key in SEC format
func EncodePublicKey(pub *btcec.PublicKey, compressed bool) []byte {
	if compressed {
		return pub.SerializeCompressed()
	}
	return pub.SerializeUncompressed()
}

// DecodePublicKey decodes a SEC-encoded public key
func DecodePublicKey(sec []byte) (*btcec.PublicKey, error) {
	return btcec.ParsePubKey(sec)
}

// Hash160 = RIPEMD160(SHA256(b))
func Hash160(b []byte) []byte {
	sha := sha256.Sum256(b)
	r := ripemd160.New()
	r.Write(sha[:])
	return r.Sum(nil)
}

// Checksum = first 4 bytes of SHA256(SHA256(payload))
func Checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:4]
}

// GetBTCAddress returns a Base58Check-encoded Bitcoin-style address
func GetBTCAddress(pub *btcec.PublicKey, compressed bool, network string) (string, error) {
	var version byte
	switch network {
	case "main":
		version = versionMainnet
	case "test":
		version = versionTestnet
	default:
		return "", errors.New("invalid network")
	}

	pubKeyBytes := EncodePublicKey(pub, compressed)
	pubKeyHash := Hash160(pubKeyBytes)

	verPayload := append([]byte{version}, pubKeyHash...)
	checksum := Checksum(verPayload)
	final := append(verPayload, checksum...)

	return base58.Encode(final), nil
}

// AddressToPKHash decodes a Base58Check address and returns the pubkey hash (without version or checksum)
func AddressToPKHash(addr string) ([]byte, error) {
	decoded := base58.Decode(addr)
	if len(decoded) != 25 {
		return nil, errors.New("invalid decoded address length")
	}
	payload := decoded[:21]
	check := decoded[21:]
	if !bytes.Equal(Checksum(payload), check) {
		return nil, errors.New("invalid checksum")
	}
	return payload[1:], nil // exclude version byte
}
