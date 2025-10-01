// modules/channels/repository.go
package channels

import (
	"database/sql"
	"errors"
	"time"
)

type Channel struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TeamID    int       `json:"team_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ChannelMember struct {
	UserID    int    `json:"user_id"`
	ChannelID int    `json:"channel_id"`
	Role      string `json:"role"` // admin, user
}

type ChannelWithRole struct {
	Channel
	UserRole string `json:"user_role"`
}

type ChannelRepository struct {
	DB *sql.DB
}

// CreateChannel cria um novo canal e adiciona o criador como admin
func (r *ChannelRepository) CreateChannel(name string, teamID, creatorID int) (*Channel, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Criar o canal
	var channel Channel
	query := `
		INSERT INTO channels (name, team_id, created_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		RETURNING id, name, team_id, created_at
	`
	err = tx.QueryRow(query, name, teamID).Scan(
		&channel.ID,
		&channel.Name,
		&channel.TeamID,
		&channel.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Adicionar o criador como admin
	_, err = tx.Exec(`
		INSERT INTO channel_users (user_id, channel_id, role)
		VALUES ($1, $2, 'admin')
	`, creatorID, channel.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &channel, nil
}

// GetChannelsByTeam retorna todos os canais de um time aos quais o usuário pertence
func (r *ChannelRepository) GetChannelsByTeam(teamID, userID int) ([]ChannelWithRole, error) {
	query := `
		SELECT c.id, c.name, c.team_id, c.created_at, cu.role
		FROM channels c
		INNER JOIN channel_users cu ON c.id = cu.channel_id
		WHERE c.team_id = $1 AND cu.user_id = $2
		ORDER BY c.created_at ASC
	`

	rows, err := r.DB.Query(query, teamID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []ChannelWithRole
	for rows.Next() {
		var ch ChannelWithRole
		err := rows.Scan(
			&ch.ID,
			&ch.Name,
			&ch.TeamID,
			&ch.CreatedAt,
			&ch.UserRole,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}

	return channels, nil
}

// GetChannelByID retorna um canal específico
func (r *ChannelRepository) GetChannelByID(channelID int) (*Channel, error) {
	var channel Channel
	query := `
		SELECT id, name, team_id, created_at
		FROM channels
		WHERE id = $1
	`
	err := r.DB.QueryRow(query, channelID).Scan(
		&channel.ID,
		&channel.Name,
		&channel.TeamID,
		&channel.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("canal não encontrado")
	}
	if err != nil {
		return nil, err
	}

	return &channel, nil
}

// IsUserInChannel verifica se um usuário pertence a um canal
func (r *ChannelRepository) IsUserInChannel(userID, channelID int) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM channel_users
			WHERE user_id = $1 AND channel_id = $2
		)
	`
	err := r.DB.QueryRow(query, userID, channelID).Scan(&exists)
	return exists, err
}

// IsUserChannelAdmin verifica se um usuário é admin de um canal
func (r *ChannelRepository) IsUserChannelAdmin(userID, channelID int) (bool, error) {
	var role string
	query := `
		SELECT role FROM channel_users
		WHERE user_id = $1 AND channel_id = $2
	`
	err := r.DB.QueryRow(query, userID, channelID).Scan(&role)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return role == "admin", nil
}

// AddUserToChannel adiciona um usuário a um canal
func (r *ChannelRepository) AddUserToChannel(userID, channelID int, role string) error {
	query := `
		INSERT INTO channel_users (user_id, channel_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, channel_id) DO UPDATE SET role = $3
	`
	_, err := r.DB.Exec(query, userID, channelID, role)
	return err
}

// RemoveUserFromChannel remove um usuário de um canal
func (r *ChannelRepository) RemoveUserFromChannel(userID, channelID int) error {
	query := `
		DELETE FROM channel_users
		WHERE user_id = $1 AND channel_id = $2
	`
	result, err := r.DB.Exec(query, userID, channelID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("usuário não está no canal")
	}

	return nil
}

// GetChannelMembers retorna todos os membros de um canal
func (r *ChannelRepository) GetChannelMembers(channelID int) ([]ChannelMember, error) {
	query := `
		SELECT user_id, channel_id, role
		FROM channel_users
		WHERE channel_id = $1
	`

	rows, err := r.DB.Query(query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []ChannelMember
	for rows.Next() {
		var member ChannelMember
		err := rows.Scan(&member.UserID, &member.ChannelID, &member.Role)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

// DeleteChannel remove um canal
func (r *ChannelRepository) DeleteChannel(channelID int) error {
	query := `DELETE FROM channels WHERE id = $1`
	result, err := r.DB.Exec(query, channelID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("canal não encontrado")
	}

	return nil
}

// IsUserInTeam verifica se um usuário pertence a um time
func (r *ChannelRepository) IsUserInTeam(userID, teamID int) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_teams
			WHERE user_id = $1 AND team_id = $2
		)
	`
	err := r.DB.QueryRow(query, userID, teamID).Scan(&exists)
	return exists, err
}
