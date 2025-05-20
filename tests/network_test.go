package tests

import (
	"net"
	"testing"
	"time"

	"github.com/xkal1bur/blockchain/pkg/network"
)

// Definir el handler de prueba fuera de la función de test
type testHandler struct {
	*network.Handler // embebemos Handler para heredar métodos y campos
	msgReceived      chan *network.Message
	peer2            *network.Peer
}

func (th *testHandler) Handle(msg *network.Message, p *network.Peer) {
	if p == th.peer2 {
		th.msgReceived <- msg
	}
	// Opcionalmente, también llamar al método original si quieres
	// th.Handler.Handle(msg, p)
}

func TestNetworkBasicMessaging(t *testing.T) {
	// Crear conexiones simuladas con net.Pipe()
	conn1, conn2 := net.Pipe()

	// Crear peers con las conexiones
	peer1 := network.NewPeer(conn1)
	peer2 := network.NewPeer(conn2)

	// Variable para capturar mensajes recibidos en peer2
	msgReceived := make(chan *network.Message, 1)

	// Crear el handler de prueba y pasarle canal y peer2
	th := &testHandler{
		Handler:     network.NewHandler(),
		msgReceived: msgReceived,
		peer2:       peer2,
	}

	// Crear la red con el handler de prueba
	netw := network.NewNetwork(th.Handler)

	// Registrar peers
	netw.RegisterPeer(peer1)
	netw.RegisterPeer(peer2)

	// Lanzar goroutines para escuchar mensajes (simulando listenPeer)
	go netw.ListenPeer(peer1)
	go netw.ListenPeer(peer2)

	// Enviar un mensaje desde peer1 a peer2
	testPayload := []byte("test message")
	msg := &network.Message{
		Type:    network.MsgPing,
		Payload: testPayload,
	}

	peer1.SendMessage(msg) // no retorna error

	// Esperar a que peer2 reciba el mensaje o timeout
	select {
	case received := <-msgReceived:
		if string(received.Payload) != string(testPayload) {
			t.Errorf("Payload no coincide, esperado %s, recibido %s", testPayload, received.Payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout esperando mensaje en peer2")
	}

	// Cerrar conexiones
	peer1.Close()
	peer2.Close()
}
