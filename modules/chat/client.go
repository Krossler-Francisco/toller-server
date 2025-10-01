package chat

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type Client struct {
	conn      *websocket.Conn
	send      chan OutgoingMessage
	userID    int64
	channelID int64
	hub       *Hub
	repo      *Repository
}

type IncomingMessage struct {
	Type    string `json:"type"`    // "message", "typing", etc
	Content string `json:"content"` // text
}

type OutgoingMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	UserID    int64  `json:"user_id"`
	ChannelID int64  `json:"channel_id"`
	MessageID int64  `json:"message_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c, c.channelID)
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var im IncomingMessage
		if err := c.conn.ReadJSON(&im); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("readPump error:", err)
			}
			break
		}

		// LOG CRÍTICO 1: Confirma que a mensagem foi lida do WebSocket.
		log.Printf("CLIENT %d: Mensagem lida do WebSocket. Tipo=%s.", c.userID, im.Type)

		// process message types
		switch im.Type {
		case "message":
			// persistir
			msgID, createdAt, err := c.repo.SaveMessage(c.channelID, c.userID, im.Content)
			if err != nil {
				log.Println("SaveMessage error:", err)
				continue
			}
			out := OutgoingMessage{
				Type:      "message",
				Content:   im.Content,
				UserID:    c.userID,
				ChannelID: c.channelID,
				MessageID: msgID,
				CreatedAt: createdAt.Format("2006-01-02 15:04:05"),
			}

			// LOG CRÍTICO 2: Confirma que a mensagem foi salva e será enviada ao Hub.
			log.Printf("CLIENT %d: Mensagem salva (ID %d). Chamando Broadcast para o canal %d.", c.userID, msgID, c.channelID)

			c.hub.Broadcast(c, c.channelID, out)
		case "typing":
			// opcional: retransmitir estado "typing"
			out := OutgoingMessage{
				Type:      "typing",
				Content:   "",
				UserID:    c.userID,
				ChannelID: c.channelID,
			}
			c.hub.Broadcast(c, c.channelID, out)
		default:
			// ignorar
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				// Se houver erro de escrita, o cliente pode ter fechado a conexão.
				log.Printf("CLIENT %d: Erro ao escrever JSON: %v. Fechando writePump.", c.userID, err)
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
