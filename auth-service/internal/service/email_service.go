package service

import (
	"fmt"
	"log"
	"net/smtp"
	"github.com/sales-tracker/auth-service/internal/config"
)

type EmailService interface {
	SendVerificationEmail(to string, verificationURL string) error
	SendPasswordResetEmail(to string, resetURL string) error
}

type SMTPService struct {
	config *config.Config
}

func NewSMTPService(config *config.Config) *SMTPService {
	return &SMTPService{
		config: config,
	}
}

func (s *SMTPService) SendVerificationEmail(to string, verificationURL string) error {
	from := s.config.SMTP.From
	fromName := s.config.SMTP.FromName
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
Dear user,

Please verify your email address by clicking on the link below:

%s

If you didn't create an account, please ignore this email.

Best regards,
The Sales Tracker Team
`, verificationURL)

	return s.sendEmail(to, from, fromName, subject, body)
}

func (s *SMTPService) SendPasswordResetEmail(to string, resetURL string) error {
	from := s.config.SMTP.From
	fromName := s.config.SMTP.FromName
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
Dear user,

We received a request to reset your password. Click the link below to set a new password:

%s

This link will expire in 24 hours.

If you didn't request a password reset, please ignore this email or contact support if you have any concerns.

Best regards,
The Sales Tracker Team
`, resetURL)

	return s.sendEmail(to, from, fromName, subject, body)
}

func (s *SMTPService) sendEmail(to, from, fromName, subject, body string) error {
	// Log SMTP configuration for debugging
	log.Printf("Sending email to %s via %s:%s\n", to, s.config.SMTP.Host, s.config.SMTP.Port)

	msg := []byte(fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		fromName,
		from,
		to,
		subject,
		body))


	auth := smtp.PlainAuth(
		"",
		s.config.SMTP.User,
		s.config.SMTP.Pass,
		s.config.SMTP.Host,
	)

	return smtp.SendMail(
		fmt.Sprintf("%s:%s", s.config.SMTP.Host, s.config.SMTP.Port),
		auth,
		from,
		[]string{to},
		msg,
	)
}
