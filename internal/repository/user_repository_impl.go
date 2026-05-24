package repository

import (
	"database/sql"

	"github.com/Refliqx/backend-eletter/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user domain.User) error {
	query := `INSERT INTO users (email, password_hash) VALUES (?, ?)`

	_, err := r.db.Exec(query, user.Email, user.PasswordHash)
	return err
}

func (r *userRepository) IsEmailExists(email string) bool {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email=?)`
	_ = r.db.QueryRow(query, email).Scan(&exists)

	return exists
}
