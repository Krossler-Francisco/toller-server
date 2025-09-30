package teams

import "time"

type Team struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserTeam struct {
	UserID int    `json:"user_id"`
	TeamID int    `json:"team_id"`
	Role   string `json:"role"` // admin, member
}

// Response extendido con info adicional
type TeamWithRole struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UserRole    string    `json:"user_role"` // rol del usuario actual
}

type TeamMember struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
