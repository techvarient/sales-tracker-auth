package usecase

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/sales-tracker/auth-service/internal/domain"
	"github.com/sales-tracker/auth-service/internal/repository"
)

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

type UserUsecase struct {
	userRepository repository.UserRepository
}

type UserUsecaseInterface interface {
	FindUserByEmail(email string) (*domain.User, error)
	FindUserByResetToken(token string) (*domain.User, error)
	RegisterUser(user *domain.User) error
	GenerateResetToken(email string) (string, error)
	ResetPassword(token, newPassword string) error
	VerifyEmail(token string) error
	ResendVerificationEmail(email string) (string, error)
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
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(passwordHash)
	user.Password = "" // Clear the plaintext password
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

func (u *UserUsecase) VerifyEmail(token string) error {
	// Find the user by verification token
	user, err := u.userRepository.FindUserByVerificationToken(token)
	if err != nil {
		return err
	}

	// Update the user's verification status
	err = u.userRepository.UpdateUserVerificationStatus(user.ID, true)
	if err != nil {
		return err
	}

	// Clear the verification token
	user.VerificationToken = ""
	return u.userRepository.UpdateUser(user)
}

func (u *UserUsecase) FindUserByResetToken(token string) (*domain.User, error) {
	return u.userRepository.FindUserByResetToken(token)
}

func (u *UserUsecase) ResetPassword(token, newPassword string) error {
	user, err := u.userRepository.FindUserByResetToken(token)
	if err != nil {
		return fmt.Errorf("invalid reset token: %w", err)
	}

	if user.ResetTokenExpiresAt.Before(time.Now()) {
		return domain.ErrExpiredResetToken
	}

	// Validate password length
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Log the password before hashing (for debugging only - remove in production)
	logrus.Debugf("Resetting password for user %s (email: %s)", user.ID, user.Email)
	logrus.Debugf("New password (before hashing): %s", newPassword)

	// Generate new password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logrus.Errorf("Failed to generate password hash: %v", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user fields
	user.PasswordHash = string(hashedPassword)
	user.ResetToken = ""
	user.ResetTokenExpiresAt = time.Time{}
	user.IsVerified = true // Mark email as verified
	user.VerificationToken = ""

	logrus.Debugf("Generated password hash: %s... (length: %d)",
		hashedPassword[:min(10, len(hashedPassword))],
		len(hashedPassword))

	// Verify the new password can be used for login
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword)); err != nil {
		logrus.Errorf("Password verification failed after reset: %v", err)
		return fmt.Errorf("failed to verify new password: %w", err)
	}

	// Save the updated user
	if err := u.userRepository.UpdateUser(user); err != nil {
		logrus.Errorf("Failed to update user after password reset: %v", err)
		return fmt.Errorf("failed to update user: %w", err)
	}
	println(user)

	logrus.Infof("Successfully reset password for user %s (email: %s)", user.ID, user.Email)
	return nil
}

func (u *UserUsecase) ResendVerificationEmail(email string) (string, error) {
	user, err := u.userRepository.FindUserByEmail(email)
	if err != nil {
		return "", err
	}

	if user.IsVerified {
		return "", fmt.Errorf("email already verified")
	}

	// Generate a new verification token
	token := uuid.New().String()
	user.VerificationToken = token

	if err := u.userRepository.UpdateUser(user); err != nil {
		return "", err
	}

	return token, nil
}
