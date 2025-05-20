// Represents the connection between two peers in the network

package network

import (
	"bufio"
	"fmt"
	"net"
	//"time"
)

// Peer represents a node connected to the network
type Peer struct {
	Conn    net.Conn
	Address string
	Inbox   chan *Message
	Outbox  chan *Message
	Quit    chan struct{}
}

// NewPeer create a new Peer instance with a stableship connection
func NewPeer(conn net.Conn) *Peer {
	p := &Peer{
		Conn:    conn,
		Address: conn.RemoteAddr().String(),
		Inbox:   make(chan *Message),
		Outbox:  make(chan *Message),
		Quit:    make(chan struct{}),
	}
	go p.readLoop()
	go p.writeLoop()
	return p
}

// readLoop reads messages from the connection and sends them to the Inbox channel
func (p *Peer) readLoop() {
	reader := bufio.NewReader(p.Conn)
	for {
		select {
		case <-p.Quit:
			return
		default:
			// Read header (ex. first 5 bytes: 1 byte type + 4 bytes length)
			header := make([]byte, 5)
			_, err := reader.Read(header)
			if err != nil {
				fmt.Printf("Error reading header since %s: %v\n", p.Address, err)
				close(p.Quit)
				return
			}

			msgType := header[0]
			payloadLen := int(header[1])<<24 | int(header[2])<<16 | int(header[3])<<8 | int(header[4])

			payload := make([]byte, payloadLen)
			_, err = reader.Read(payload)
			if err != nil {
				fmt.Printf("Error reading payload since %s: %v\n", p.Address, err)
				close(p.Quit)
				return
			}

			msg := &Message{
				Type:    MessageType(msgType),
				Payload: payload,
			}

			p.Inbox <- msg
		}
	}
}

// writeLoop writes messages from the Outbox channel to the connection
func (p *Peer) writeLoop() {
	for {
		select {
		case <-p.Quit:
			return
		case msg := <-p.Outbox:
			data, err := msg.Encode()
			if err != nil {
				fmt.Printf("Error encoding message for %s: %v\n", p.Address, err)
				continue
			}
			_, err = p.Conn.Write(data)
			if err != nil {
				fmt.Printf("Error sending message to %s: %v\n", p.Address, err)
				close(p.Quit)
				return
			}
		}
	}
}

// SendMessage add a message to the Outbox channel to be sent
func (p *Peer) SendMessage(msg *Message) {
	p.Outbox <- msg
}

// Close closes the connection and stops the read/write loops
func (p *Peer) Close() {
	p.Conn.Close()
	close(p.Quit)
}
