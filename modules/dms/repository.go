package dms

import "database/sql"

type DMRepository struct {
	DB *sql.DB
}

func NewDMRepository(db *sql.DB) *DMRepository {
	return &DMRepository{DB: db}
}

func (r *DMRepository) CreateDMChannel(user1ID, user2ID int) (int, error) {
	var channelID int
	err := r.DB.QueryRow(`
		SELECT channel_id FROM (
			SELECT cu1.channel_id
			FROM channel_users cu1
			JOIN channel_users cu2 ON cu1.channel_id = cu2.channel_id
			JOIN channels c ON cu1.channel_id = c.id
			WHERE c.is_dm = TRUE AND cu1.user_id = $1 AND cu2.user_id = $2
		) AS existing_channel
	`, user1ID, user2ID).Scan(&channelID)

	if err == nil {
		return channelID, nil // DM channel already exists
	}

	if err != sql.ErrNoRows {
		return 0, err // Real error
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	// Create a new channel
	err = tx.QueryRow("INSERT INTO channels (name, is_dm) VALUES ('', TRUE) RETURNING id").Scan(&channelID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Add both users to the channel
	_, err = tx.Exec("INSERT INTO channel_users (channel_id, user_id) VALUES ($1, $2), ($1, $3)", channelID, user1ID, user2ID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return channelID, tx.Commit()
}

// ListDMChannels devuelve una lista de todos los canales de DM de un usuario.
func (r *DMRepository) ListDMChannels(userID int) ([]DMChannelInfo, error) {
	query := `
		SELECT c.id, u.id, u.username
		FROM channels c
		JOIN channel_users cu_self ON c.id = cu_self.channel_id
		JOIN channel_users cu_other ON c.id = cu_other.channel_id
		JOIN users u ON cu_other.user_id = u.id
		WHERE c.is_dm = TRUE
		  AND cu_self.user_id = $1
		  AND cu_other.user_id != $1
	`
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dms []DMChannelInfo
	for rows.Next() {
		var dm DMChannelInfo
		if err := rows.Scan(&dm.ChannelID, &dm.OtherUserID, &dm.OtherUsername); err != nil {
			return nil, err
		}
		dms = append(dms, dm)
	}

	return dms, nil
}

func (r *DMRepository) GetMessagesByChannelID(channelID int) ([]Message, error) {
	rows, err := r.DB.Query("SELECT id, channel_id, user_id, content, created_at FROM messages WHERE channel_id = $1 ORDER BY created_at ASC", channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *DMRepository) IsUserInDMChannel(userID, channelID int) (bool, error) {
	var exists bool
	err := r.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM channel_users WHERE user_id = $1 AND channel_id = $2)", userID, channelID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *DMRepository) MarkChannelAsRead(userID, channelID int) error {
	var lastMessageID int
	err := r.DB.QueryRow("SELECT id FROM messages WHERE channel_id = $1 ORDER BY created_at DESC LIMIT 1", channelID).Scan(&lastMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No messages in channel, nothing to mark as read
			return nil
		}
		return err
	}

	query := `
		INSERT INTO last_read (user_id, channel_id, message_id, timestamp)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, channel_id)
		DO UPDATE SET message_id = $3, timestamp = NOW();
	`
	_, err = r.DB.Exec(query, userID, channelID, lastMessageID)
	return err
}