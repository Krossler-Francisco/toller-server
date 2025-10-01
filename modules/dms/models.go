package dms

import "time"

// DMChannelInfo representa la información de un canal de DM para la API.
// Incluye el ID del canal y los datos del otro usuario en la conversación.
type DMChannelInfo struct {
	ChannelID     int    `json:"channel_id"`
	OtherUserID   int    `json:"other_user_id"`
	OtherUsername string `json:"other_username"`
}

// Message representa un mensaje en un canal de DM.

type Message struct {
	ID        int       `json:"id"`
	ChannelID int       `json:"channel_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// LastRead representa el último mensaje leído por un usuario en un canal.
type LastRead struct {
	UserID    int       `json:"user_id"`
	ChannelID int       `json:"channel_id"`
	MessageID int       `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
}
