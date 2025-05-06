package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/sales-tracker/auth-service/internal/domain"
)

type postgresUserRepository struct {
	db *sql.DB
}

func (r *postgresUserRepository) UpdateUser(user *domain.User) error {
	query := `UPDATE users SET
		name = $1,
		password_hash = $2,
		reset_token = $3,
		reset_token_expires_at = $4,
		updated_at = $5
	WHERE id = $6`

	_, err := r.db.Exec(query,
		user.Name,
		user.PasswordHash,
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
	query := `INSERT INTO users (email, password_hash, role, is_verified, verification_token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	return r.db.QueryRow(query,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.IsVerified,
		user.VerificationToken,
		time.Now(),
		time.Now(),
	).Scan(&user.ID)
}

func (r *postgresUserRepository) FindUserByVerificationToken(token string) (*domain.User, error) {
	var user domain.User
	var name sql.NullString
	var resetToken sql.NullString
	var resetTokenExpiresAt sql.NullTime

	query := `SELECT id, email, password_hash, role, is_verified, name, reset_token, reset_token_expires_at, created_at, updated_at 
		FROM users WHERE verification_token = $1`

	err := r.db.QueryRow(query, token).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&name,
		&resetToken,
		&resetTokenExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == nil {
		user.Name = name.String             // Convert sql.NullString to string (empty string if NULL)
		user.ResetToken = resetToken.String // Convert sql.NullString to string (empty string if NULL)
		if resetTokenExpiresAt.Valid {
			user.ResetTokenExpiresAt = resetTokenExpiresAt.Time
		}
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found with the given verification token")
		}
		return nil, err
	}

	return &user, nil
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

func (r *postgresUserRepository) FindUserByResetToken(token string) (*domain.User, error) {
	var user domain.User
	var name sql.NullString
	var verificationToken sql.NullString

	query := `SELECT id, email, password_hash, role, is_verified, name, verification_token, reset_token_expires_at, created_at, updated_at 
		FROM users WHERE reset_token = $1`

	err := r.db.QueryRow(query, token).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&name,
		&verificationToken,
		&user.ResetTokenExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == nil {
		user.Name = name.String
		user.VerificationToken = verificationToken.String
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found with the given reset token")
		}
		return nil, err
	}

	return &user, nil
}
