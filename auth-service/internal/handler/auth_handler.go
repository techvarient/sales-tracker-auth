package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/sales-tracker/auth-service/internal/config"
	"github.com/sales-tracker/auth-service/internal/domain"
	"github.com/sales-tracker/auth-service/internal/service"
	"github.com/sales-tracker/auth-service/internal/usecase"
)

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

type AuthHandler struct {
	userUsecase  usecase.UserUsecase
	emailService service.EmailService
	config       *config.Config
	logger       *logrus.Logger
}

func NewAuthHandler(config *config.Config, userUsecase usecase.UserUsecase, emailService service.EmailService) *AuthHandler {
	return &AuthHandler{
		userUsecase:  userUsecase,
		emailService: emailService,
		config:       config,
		logger:       logrus.New(),
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req domain.UserRegistration
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Generate verification token
	token := uuid.New().String()

	user := &domain.User{
		Email:             req.Email,
		Password:          req.Password,
		Role:              req.Role,
		IsVerified:        false, // New users need verification
		VerificationToken: token,
	}

	if err := h.userUsecase.RegisterUser(user); err != nil {
		h.logger.Error("Failed to register user:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	// Generate verification URL
	verificationURL := fmt.Sprintf("%s%s?token=%s", h.config.BaseURL, h.config.Verification, token)

	// Send verification email
	if err := h.emailService.SendVerificationEmail(user.Email, verificationURL); err != nil {
		h.logger.Error("Failed to send verification email:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to send verification email")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully. Please check your email for verification.",
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req domain.UserLogin
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	h.logger.Infof("Login attempt for email: %s", req.Email)

	user, err := h.userUsecase.FindUserByEmail(req.Email)
	if err != nil {
		h.logger.Errorf("Failed to find user with email %s: %v", req.Email, err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	h.logger.Infof("User found - ID: %d, Email: %s, Verified: %v",
		user.ID, user.Email, user.IsVerified)

	// Log the first few characters of the stored hash for debugging
	h.logger.Debugf("Stored password hash: %s... (length: %d)",
		user.PasswordHash[:min(10, len(user.PasswordHash))],
		len(user.PasswordHash))

	h.logger.Debugf("Comparing with password: %s (length: %d)",
		strings.Repeat("*", len(req.Password)),
		len(req.Password))

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		h.logger.Errorf("Password comparison failed: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	if !user.IsVerified {
		h.logger.Warnf("Login attempt for unverified email: %s", req.Email)
		return echo.NewHTTPError(http.StatusUnauthorized, "Account not verified")
	}

	h.logger.Infof("User %s successfully authenticated", user.Email)

	claims := &domain.JWTClaims{
		UserID: user.ID,
		Role:   user.Role,
		Email:  user.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		h.logger.Error("Failed to sign token:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": signedToken,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req domain.UserLogin
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	resetToken, err := h.userUsecase.GenerateResetToken(req.Email)
	if err != nil {
		h.logger.Error("Failed to generate reset token:", err)
		// Return a generic message to avoid user enumeration
		return c.JSON(http.StatusOK, map[string]string{
			"message": "If an account with that email exists, a password reset link has been sent",
		})
	}

	// Use the frontend URL for the reset link
	//frontendURL := "http://localhost:3000" // Using port 3000 for the frontend
	frontendURL := "https://sales-tracker-reset-password.onrender.com"
	resetURL := fmt.Sprintf("%s/reset-password.html?token=%s", frontendURL, resetToken)

	if err := h.emailService.SendPasswordResetEmail(req.Email, resetURL); err != nil {
		h.logger.Error("Failed to send password reset email:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to send password reset email")
	}

	h.logger.Info("Password reset instructions sent to:", req.Email)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "If an account with that email exists, a password reset link has been sent",
	})
}

type passwordResetRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req passwordResetRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind reset password request:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	h.logger.Infof("Processing password reset for token: %s", req.Token)

	// Get user by token before reset to log details
	user, err := h.userUsecase.FindUserByResetToken(req.Token)
	if err != nil {
		h.logger.Warnf("User not found for reset token: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired reset token")
	}

	h.logger.Infof("Resetting password for user: %s (ID: %d, Verified: %v)",
		user.Email, user.ID, user.IsVerified)

	if len(req.Password) < 8 {
		h.logger.Warn("Password is too short")
		return echo.NewHTTPError(http.StatusBadRequest, "Password must be at least 8 characters long")
	}

	if err := h.userUsecase.ResetPassword(req.Token, req.Password); err != nil {
		h.logger.Errorf("Failed to reset password: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to reset password. Please try again.")
	}

	h.logger.Infof("Successfully reset password for user: %s", user.Email)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password has been reset successfully. You can now log in with your new password.",
	})
}

// ResendVerificationEmail handles resending verification emails
func (h *AuthHandler) ResendVerificationEmail(c echo.Context) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	token, err := h.userUsecase.ResendVerificationEmail(req.Email)
	if err != nil {
		h.logger.Error("Failed to resend verification email:", err)
		if err.Error() == "email already verified" {
			return echo.NewHTTPError(http.StatusBadRequest, "Email is already verified")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to resend verification email")
	}

	// Generate verification URL
	verificationURL := fmt.Sprintf("%s%s?token=%s", h.config.BaseURL, h.config.Verification, token)

	// Send verification email
	if err := h.emailService.SendVerificationEmail(req.Email, verificationURL); err != nil {
		h.logger.Error("Failed to send verification email:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to send verification email")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Verification email resent successfully",
	})
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Verification token is required",
		})
	}

	// Verify the token and mark email as verified
	err := h.userUsecase.VerifyEmail(token)
	if err != nil {
		h.logger.Error("Failed to verify email:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Invalid or expired verification token",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Email verified successfully. You can now log in.",
	})
}
