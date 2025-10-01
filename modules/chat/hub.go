package chat

import (
	"log"
	"sync"
)

type Hub struct {
	// map channelID -> set of clients
	rooms map[int64]map[*Client]bool
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[int64]map[*Client]bool),
	}
}

func (h *Hub) Register(c *Client, channelID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[channelID]; !ok {
		h.rooms[channelID] = make(map[*Client]bool)
	}
	h.rooms[channelID][c] = true
	log.Printf("HUB: Cliente %d registrado no Canal %d. Total: %d", c.userID, channelID, len(h.rooms[channelID]))
}

func (h *Hub) Unregister(c *Client, channelID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.rooms[channelID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.rooms, channelID)
		}
	}
	log.Printf("HUB: Cliente %d desregistrado do Canal %d", c.userID, channelID)
}

func (h *Hub) Broadcast(sender *Client, channelID int64, msg OutgoingMessage) {
	// LOG CRÍTICO 3: Confirma que a função Broadcast foi chamada.
	log.Printf("HUB: Broadcast chamado pelo remetente %d para o Canal %d.", sender.userID, channelID)

	h.mu.RLock()
	clients, ok := h.rooms[channelID]
	h.mu.RUnlock()

	if !ok {
		log.Printf("HUB: Broadcast falhou. Canal %d não encontrado.", channelID)
		return
	}

	// Total de clientes - 1 (o remetente)
	log.Printf("HUB: Distribuindo mensagem de %d para %d clientes no Canal %d", sender.userID, len(clients)-1, channelID)

	for c := range clients {
		// ESSENCIAL: Ignora o cliente que enviou a mensagem (sender)
		if c == sender {
			continue
		}

		select {
		case c.send <- msg:
			// LOG CRÍTICO 4: Confirma que a mensagem foi colocada no canal de envio.
			log.Printf("HUB: Mensagem colocada com sucesso no canal de envio do Cliente %d.", c.userID)
		default:
			// canal cheio -> desconecta
			log.Printf("HUB: Buffer de envio do Cliente %d cheio. Desregistrando.", c.userID)
			h.Unregister(c, channelID)
			// É seguro fechar o canal aqui, o writePump vai parar.
			close(c.send)
		}
	}
}
