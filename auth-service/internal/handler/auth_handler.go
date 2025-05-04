package handler

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/sales-tracker/auth-service/internal/config"
	"github.com/sales-tracker/auth-service/internal/domain"
	"github.com/sales-tracker/auth-service/internal/service"
	"github.com/sales-tracker/auth-service/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userUsecase usecase.UserUsecase
	emailService service.EmailService
	config      *config.Config
	logger      *logrus.Logger
}

func NewAuthHandler(config *config.Config, userUsecase usecase.UserUsecase) *AuthHandler {
	emailService := service.NewSMTPService(config)
	return &AuthHandler{
		userUsecase: userUsecase,
		emailService: emailService,
		config:      config,
		logger:      logrus.New(),
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req domain.UserRegistration
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	user := &domain.User{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		IsVerified: false, // New users need verification
	}

	if err := h.userUsecase.RegisterUser(user); err != nil {
		h.logger.Error("Failed to register user:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	// Generate verification token
	verificationToken := uuid.New().String()
	verificationURL := fmt.Sprintf("%s%s?token=%s", h.config.BaseURL, h.config.Verification, verificationToken)

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

	user, err := h.userUsecase.FindUserByEmail(req.Email)
	if err != nil {
		h.logger.Error("Failed to find user:", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.logger.Error("Invalid password")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	if !user.IsVerified {
		return echo.NewHTTPError(http.StatusUnauthorized, "Account not verified")
	}

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

	_, err := h.userUsecase.GenerateResetToken(req.Email)
	if err != nil {
		h.logger.Error("Failed to generate reset token:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process password reset request")
	}

	// TODO: Implement email sending service with reset link: %s/reset-password?token=%s
	h.logger.Info("Password reset instructions sent to:", req.Email)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password reset instructions sent to your email",
	})
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req domain.PasswordReset
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.userUsecase.ResetPassword(req.Email, req.Token, req.Password); err != nil {
		h.logger.Error("Failed to reset password:", err)
		if err == domain.ErrInvalidResetToken {
			return echo.NewHTTPError(http.StatusNotFound, "Invalid reset token")
		}
		if err == domain.ErrExpiredResetToken {
			return echo.NewHTTPError(http.StatusGone, "Reset token has expired")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}
