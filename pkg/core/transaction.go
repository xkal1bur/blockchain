package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

func DecodeInt(r io.Reader, nbytes int) (uint64, error) {
	buf := make([]byte, nbytes)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, err
	}

	var val uint64
	switch nbytes {
	case 1:
		val = uint64(buf[0])
	case 2:
		val = uint64(binary.LittleEndian.Uint16(buf))
	case 4:
		val = uint64(binary.LittleEndian.Uint32(buf))
	case 8:
		val = binary.LittleEndian.Uint64(buf)
	default:
		return 0, fmt.Errorf("unsupported size: %d", nbytes)
	}
	return val, nil
}

func EncodeInt(val uint64, nbytes int) ([]byte, error) {
	buf := new(bytes.Buffer)
	switch nbytes {
	case 1:
		buf.WriteByte(byte(val))
	case 2:
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(val))
		buf.Write(b)
	case 4:
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(val))
		buf.Write(b)
	case 8:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, val)
		buf.Write(b)
	default:
		return nil, fmt.Errorf("unsupported size: %d", nbytes)
	}
	return buf.Bytes(), nil
}

// decodeVarInt decodifica un VarInt estilo Bitcoin
func DecodeVarInt(r io.Reader) (uint64, error) {
	prefix := make([]byte, 1)
	_, err := io.ReadFull(r, prefix)
	if err != nil {
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

func EncodeVarInt(val uint64) ([]byte, error) {
	switch {
	case val < 0xfd:
		return []byte{byte(val)}, nil
	case val < 0x10000:
		b, err := EncodeInt(val, 2)
		if err != nil {
			return nil, err
		}
		return append([]byte{0xfd}, b...), nil
	case val < 0x100000000:
		b, err := EncodeInt(val, 4)
		if err != nil {
			return nil, err
		}
		return append([]byte{0xfe}, b...), nil
	case val < 0x10000000000000000:
		b, err := EncodeInt(val, 8)
		if err != nil {
			return nil, err
		}
		return append([]byte{0xff}, b...), nil
	default:
		return nil, fmt.Errorf("integer too large: %d", val)
	}
}
