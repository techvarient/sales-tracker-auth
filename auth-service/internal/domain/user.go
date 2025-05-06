package domain

import (
	"errors"
	"time"
	"github.com/dgrijalva/jwt-go"
)

var (
	ErrInvalidResetToken = errors.New("invalid reset token")
	ErrExpiredResetToken = errors.New("reset token has expired")
)
type User struct {
	ID                  int64     `json:"id"`
	Email               string    `json:"email"`
	PasswordHash        string    `json:"-"`
	Password            string    `json:"-"`
	Role                string    `json:"role"`
	IsVerified          bool      `json:"is_verified"`
	Name                string    `json:"name"`
	VerificationToken   string    `json:"-"`
	ResetToken          string    `json:"-"`
	ResetTokenExpiresAt time.Time `json:"-"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type UserRegistration struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=client sales_rep admin"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type PasswordReset struct {
	Email    string `json:"email" validate:"required,email"`
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

// Valid implements the jwt.Claims interface
func (c *JWTClaims) Valid() error {
	return c.StandardClaims.Valid()
}
