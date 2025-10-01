package chat

import (
	"database/sql"
	"time"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// SaveMessage salva e retorna id e created_at
func (r *Repository) SaveMessage(channelID int64, userID int64, content string) (int64, time.Time, error) {
	var id int64
	var createdAt time.Time
	query := `INSERT INTO messages (channel_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := r.DB.QueryRow(query, channelID, userID, content).Scan(&id, &createdAt)
	if err != nil {
		return 0, time.Time{}, err
	}
	return id, createdAt, nil
}

func (r *Repository) LoadLastMessages(channelID int64, limit int) ([]OutgoingMessage, error) {
	query := `SELECT id, user_id, content, created_at FROM messages WHERE channel_id=$1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.DB.Query(query, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OutgoingMessage{}
	for rows.Next() {
		var id int64
		var userID int64
		var content string
		var createdAt time.Time
		if err := rows.Scan(&id, &userID, &content, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, OutgoingMessage{
			Type:      "message",
			Content:   content,
			UserID:    userID,
			ChannelID: channelID,
			MessageID: id,
			CreatedAt: createdAt.Format("2006-01-02 15:04:05"),
		})
	}
	// return in chronological order
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
