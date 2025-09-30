package auth

import (
	"database/sql"
	"errors"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) CreateUser(user *User) error {
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id`
	return r.DB.QueryRow(query, user.Username, user.Email, user.Password).Scan(&user.ID)
}

func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, password FROM users WHERE email = $1`
	err := r.DB.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return nil, errors.New("usuario no encontrado")
	}
	return user, nil
}
