package service

import (
	"fmt"
	"net/smtp"
	"github.com/sales-tracker/auth-service/internal/config"
)

type EmailService interface {
	SendVerificationEmail(to string, verificationURL string) error
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
	from := s.config.SMTPFrom
	fromName := s.config.SMTPFromName
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
Dear user,

Please verify your email address by clicking on the link below:

%s

If you didn't create an account, please ignore this email.

Best regards,
The Sales Tracker Team
`, verificationURL)

	msg := []byte(fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		fromName,
		from,
		to,
		subject,
		body))

	auth := smtp.PlainAuth(
			"",
			s.config.SMTPUser,
			s.config.SMTPPass,
			s.config.SMTPHost,
		)

	err := smtp.SendMail(
			fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort),
			auth,
			from,
			[]string{to},
			msg,
		)

	return err
}
