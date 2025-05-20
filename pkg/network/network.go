package network

import (
	"log"
	"net"
	"sync"
)

// Network representa la capa de red del nodo
type Network struct {
	Peers     map[string]*Peer
	peerMutex sync.Mutex
	handler   *Handler
}

// NewNetwork crea una nueva instancia de Network
func NewNetwork(handler *Handler) *Network {
	return &Network{
		Peers:   make(map[string]*Peer),
		handler: handler,
	}
}

// Listen inicia un servidor TCP en la dirección dada y acepta conexiones entrantes
func (n *Network) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("🟢 Listening on %s\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}
		peer := NewPeer(conn)
		n.RegisterPeer(peer)
		go n.ListenPeer(peer)
	}
}

// Connect intenta conectarse a un peer remoto en la dirección dada
func (n *Network) Connect(addr string) (*Peer, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	peer := NewPeer(conn)
	n.RegisterPeer(peer)
	go n.ListenPeer(peer)
	return peer, nil
}

// Escucha mensajes del peer y los delega al handler
func (n *Network) ListenPeer(p *Peer) {
	for {
		select {
		case msg := <-p.Inbox:
			go n.handler.Handle(msg, p)
		case <-p.Quit:
			log.Printf("❌ Peer desconectado: %s\n", p.Conn.RemoteAddr())
			n.UnregisterPeer(p)
			return
		}
	}
}

// Enviar mensaje a todos los peers conectados
func (n *Network) Broadcast(msg *Message) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	for _, peer := range n.Peers {
		peer.SendMessage(msg)
	}
}

// Registrar un nuevo peer conectado
func (n *Network) RegisterPeer(p *Peer) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	addr := p.Conn.RemoteAddr().String()
	n.Peers[addr] = p
	log.Printf("📌 Peer registrado: %s\n", addr)
}

// Eliminar un peer desconectado
func (n *Network) UnregisterPeer(p *Peer) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	addr := p.Conn.RemoteAddr().String()
	delete(n.Peers, addr)
	log.Printf("🗑️ Peer removido: %s\n", addr)
}
