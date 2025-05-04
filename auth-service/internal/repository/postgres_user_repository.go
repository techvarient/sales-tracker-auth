package repository

import (
	"database/sql"
	"errors"
	"github.com/sales-tracker/auth-service/internal/domain"
	"time"
)

type postgresUserRepository struct {
	db *sql.DB
}

func (r *postgresUserRepository) UpdateUser(user *domain.User) error {
	query := `UPDATE users SET
		name = $1,
		reset_token = $2,
		reset_token_expires_at = $3,
		updated_at = $4
	WHERE id = $5`
	
	_, err := r.db.Exec(query,
		user.Name,
		user.ResetToken,
		user.ResetTokenExpiresAt,
		time.Now(),
		user.ID,
	)
	return err
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(user *domain.User) error {
	query := `INSERT INTO users (email, password_hash, role, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	return r.db.QueryRow(query,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.IsVerified,
		time.Now(),
		time.Now(),
	).Scan(&user.ID)
}

func (r *postgresUserRepository) FindUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password_hash, role, is_verified, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *postgresUserRepository) UpdateUserPassword(userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, passwordHash, time.Now(), userID)
	return err
}

func (r *postgresUserRepository) UpdateUserVerificationStatus(userID int64, isVerified bool) error {
	query := `UPDATE users SET is_verified = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, isVerified, time.Now(), userID)
	return err
}

func (r *postgresUserRepository) FindUserByID(userID int64) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password_hash, role, is_verified, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
