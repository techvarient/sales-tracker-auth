package repository

import (
	"github.com/sales-tracker/auth-service/internal/domain"
)

type UserRepository interface {
	CreateUser(user *domain.User) error
	FindUserByEmail(email string) (*domain.User, error)
	FindUserByVerificationToken(token string) (*domain.User, error)
	FindUserByResetToken(token string) (*domain.User, error)
	UpdateUserPassword(userID int64, passwordHash string) error
	UpdateUserVerificationStatus(userID int64, isVerified bool) error
	FindUserByID(userID int64) (*domain.User, error)
	UpdateUser(user *domain.User) error
}
