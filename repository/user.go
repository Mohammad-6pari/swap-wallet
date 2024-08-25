package repository

import (
	"database/sql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetUsername(userId int) (string, error) {
	query := `SELECT username FROM users WHERE id = $1`

	var username string
	err := r.db.QueryRow(query, userId).Scan(
		&username,
	)

	if err != nil {
		return "", err
	}

	return username, nil
}
