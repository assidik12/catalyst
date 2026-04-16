package mysql

import (
	"context"
	"database/sql"

	"github.com/assidik12/go-restfull-api/internal/domain"
)

// userRepository is the MySQL implementation of domain.UserRepository.
type userRepository struct {
	db *sql.DB
}

// NewUserRepository returns a domain.UserRepository backed by MySQL.
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Save implements domain.UserRepository.
func (r *userRepository) Save(ctx context.Context, user domain.User) (domain.User, error) {
	query := "INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)"
	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.Password, user.Role)
	if err != nil {
		return user, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return user, err
	}

	user.ID = int(id)
	return user, nil
}

// FindByEmail implements domain.UserRepository.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	query := "SELECT id, username, email, password, role FROM users WHERE email = ?"
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

// FindById implements domain.UserRepository.
func (r *userRepository) FindById(ctx context.Context, id int) (domain.User, error) {
	var user domain.User
	query := "SELECT id, username, email, role FROM users WHERE id = ?"
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}
