package usecase

import (
	"time"

	"github.com/sales-tracker/auth-service/internal/domain"
	"github.com/sales-tracker/auth-service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepository repository.UserRepository
}

func (u *UserUsecase) FindUserByEmail(email string) (*domain.User, error) {
	return u.userRepository.FindUserByEmail(email)
}

func NewUserUsecase(userRepository repository.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepository: userRepository,
	}
}

func (u *UserUsecase) RegisterUser(user *domain.User) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(passwordHash)
	return u.userRepository.CreateUser(user)
}

func (u *UserUsecase) GenerateResetToken(email string) (string, error) {
	user, err := u.userRepository.FindUserByEmail(email)
	if err != nil {
		return "", err
	}

	resetToken := uuid.New().String()
	resetTokenExpiration := time.Now().Add(24 * time.Hour)

	user.ResetToken = resetToken
	user.ResetTokenExpiresAt = resetTokenExpiration
	if err := u.userRepository.UpdateUser(user); err != nil {
		return "", err
	}

	return resetToken, nil
}

func (u *UserUsecase) ResetPassword(email, token, newPassword string) error {
	user, err := u.userRepository.FindUserByEmail(email)
	if err != nil {
		return err
	}

	if user.ResetToken != token {
		return domain.ErrInvalidResetToken
	}

	if user.ResetTokenExpiresAt.Before(time.Now()) {
		return domain.ErrExpiredResetToken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(passwordHash)
	user.ResetToken = ""
	user.ResetTokenExpiresAt = time.Time{}
	return u.userRepository.UpdateUser(user)
}
