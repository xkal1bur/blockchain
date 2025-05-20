// encoding/decoding of the network message, format of exchange information

package network

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// MessageType represents the type of msg on the network
type MessageType byte

const ( // usually identified by the first byte of the message or a short string
	MsgPing    MessageType = iota + 1 // verifies if the peer is alive
	MsgPong                           // verifies if the peer is alive
	MsgVersion                        // version of the peer
	MsgInv                            // inventory of the peer, to announce blocks or transactions
	MsgGetData                        // get data from the peer
	MsgTx                             // to send a transaction
	MsgBlock                          // to send a block
)

// Message  represents a generic message on the network
type Message struct {
	Type    MessageType
	Payload []byte
}

// Encode serializes the message to send over the network
func (m *Message) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write byte (1 byte)
	err := buf.WriteByte(byte(m.Type))
	if err != nil {
		return nil, err
	}

	// Writes payload´s size (uint32, 4 bytes)
	payloadLen := uint32(len(m.Payload))
	err = binary.Write(buf, binary.BigEndian, payloadLen)
	if err != nil {
		return nil, err
	}

	// Write payload (variable size)
	_, err = buf.Write(m.Payload)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DecodeMessage deserializes a byte slice into a Message
func DecodeMessage(data []byte) (*Message, error) {
	buf := bytes.NewReader(data)

	// Read type (1 byte)
	msgTypeByte, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	msgType := MessageType(msgTypeByte)

	// Read payload length (uint32, 4 bytes)
	var payloadLen uint32
	err = binary.Read(buf, binary.BigEndian, &payloadLen)
	if err != nil {
		return nil, err
	}

	// Check if the payload length is larger than the available data
	if payloadLen > uint32(buf.Len()) {
		return nil, errors.New("payload length larger than available data")
	}

	// Read payload (variable size)
	payload := make([]byte, payloadLen)
	_, err = buf.Read(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:    msgType,
		Payload: payload,
	}, nil
}

// To create specific messages
func NewPingMessage() *Message {
	return &Message{Type: MsgPing, Payload: []byte{}}
}

func NewTxMessage(txBytes []byte) *Message {
	return &Message{Type: MsgTx, Payload: txBytes}
}
