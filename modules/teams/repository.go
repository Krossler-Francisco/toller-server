package teams

import (
	"database/sql"
	"errors"
)

type TeamRepository struct {
	DB *sql.DB
}

// Crear un team
func (r *TeamRepository) CreateTeam(team *Team) error {
	query := `
		INSERT INTO teams (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at
	`
	return r.DB.QueryRow(query, team.Name, team.Description).
		Scan(&team.ID, &team.CreatedAt)
}

// Agregar usuario al team
func (r *TeamRepository) AddUserToTeam(userID, teamID int, role string) error {
	query := `
		INSERT INTO user_teams (user_id, team_id, role)
		VALUES ($1, $2, $3)
	`
	_, err := r.DB.Exec(query, userID, teamID, role)
	return err
}

// Obtener teams de un usuario
func (r *TeamRepository) GetUserTeams(userID int) ([]TeamWithRole, error) {
	query := `
		SELECT t.id, t.name, t.description, t.created_at, ut.role
		FROM teams t
		INNER JOIN user_teams ut ON t.id = ut.team_id
		WHERE ut.user_id = $1
		ORDER BY t.created_at DESC
	`
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []TeamWithRole
	for rows.Next() {
		var t TeamWithRole
		var desc sql.NullString
		err := rows.Scan(&t.ID, &t.Name, &desc, &t.CreatedAt, &t.UserRole)
		if err != nil {
			return nil, err
		}
		if desc.Valid {
			t.Description = desc.String
		}
		teams = append(teams, t)
	}
	return teams, nil
}

// Obtener un team por ID
func (r *TeamRepository) GetTeamByID(teamID int) (*Team, error) {
	team := &Team{}
	var desc sql.NullString
	query := `
		SELECT id, name, description, created_at
		FROM teams
		WHERE id = $1
	`
	err := r.DB.QueryRow(query, teamID).Scan(
		&team.ID, &team.Name, &desc, &team.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("equipo no encontrado")
	}
	if desc.Valid {
		team.Description = desc.String
	}
	return team, err
}

// Verificar si un usuario es miembro del team
func (r *TeamRepository) IsUserInTeam(userID, teamID int) (bool, string, error) {
	var role string
	query := `
		SELECT role FROM user_teams
		WHERE user_id = $1 AND team_id = $2
	`
	err := r.DB.QueryRow(query, userID, teamID).Scan(&role)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, role, nil
}

// Obtener miembros de un team
func (r *TeamRepository) GetTeamMembers(teamID int) ([]TeamMember, error) {
	query := `
		SELECT u.id, u.username, u.email, ut.role
		FROM users u
		INNER JOIN user_teams ut ON u.id = ut.user_id
		WHERE ut.team_id = $1
		ORDER BY ut.role DESC, u.username ASC
	`
	rows, err := r.DB.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []TeamMember
	for rows.Next() {
		var m TeamMember
		err := rows.Scan(&m.UserID, &m.Username, &m.Email, &m.Role)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

// Remover usuario de un team
func (r *TeamRepository) RemoveUserFromTeam(userID, teamID int) error {
	query := `DELETE FROM user_teams WHERE user_id = $1 AND team_id = $2`
	_, err := r.DB.Exec(query, userID, teamID)
	return err
}

// Actualizar rol de usuario en team
func (r *TeamRepository) UpdateUserRole(userID, teamID int, role string) error {
	query := `
		UPDATE user_teams 
		SET role = $1 
		WHERE user_id = $2 AND team_id = $3
	`
	_, err := r.DB.Exec(query, role, userID, teamID)
	return err
}

// Actualizar team
func (r *TeamRepository) UpdateTeam(team *Team) error {
	query := `
		UPDATE teams
		SET name = $1, description = $2
		WHERE id = $3
	`
	_, err := r.DB.Exec(query, team.Name, team.Description, team.ID)
	return err
}

// Eliminar team
func (r *TeamRepository) DeleteTeam(teamID int) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.DB.Exec(query, teamID)
	return err
}
