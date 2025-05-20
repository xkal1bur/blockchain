// Received messages processor and what to do with them

package network

import (
	"fmt"

	"github.com/xkal1bur/blockchain/pkg/core"
)

type Handler struct {
	Mempool []*core.Tx // Managment of transactions
}

func NewHandler() *Handler {
	return &Handler{
		Mempool: []*core.Tx{},
	}
}

// Handle processes specific messages based on their type
func (h *Handler) Handle(msg *Message, p *Peer) {
	switch msg.Type {

	case MsgPing:
		fmt.Printf("[handler] Ping recibido de %s\n", p.Address)
		pong := &Message{Type: MsgPong, Payload: []byte("pong")}
		p.SendMessage(pong)

	case MsgTx:
		fmt.Printf("[handler] Transacción recibida de %s\n", p.Address)

		tx, err := core.DecodeTxBytes(msg.Payload)
		if err != nil {
			fmt.Printf("[handler] Error al decodificar transacción: %v\n", err)
			return
		}
		fmt.Printf("[handler] Transacción válida: %+v\n", tx)

		// add mempool with no duplicates
		h.Mempool = append(h.Mempool, tx)

	case MsgBlock:
		fmt.Printf("[handler] Bloque recibido de %s\n", p.Address)
		block, err := core.DecodeBlock(msg.Payload)
		if err != nil {
			fmt.Printf("[handler] Error al decodificar bloque: %v\n", err)
			return
		}
		fmt.Printf("[handler] Bloque válido: %+v\n", block)

		// Aquí podrías validar el bloque, verificar PoW, etc.

	default:
		fmt.Printf("[handler] Tipo de mensaje desconocido %d de %s\n", msg.Type, p.Address)
	}
}
