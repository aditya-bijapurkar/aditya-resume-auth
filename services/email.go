package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	fromEmail    string
	baseURL      string
}

func NewEmailService() *EmailService {
	es := &EmailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUser:     os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("FROM_EMAIL"),
		baseURL:      os.Getenv("BASE_URL"),
	}

	return es
}

type VerificationEmailData struct {
	Email           string
	VerificationURL string
}

func (es *EmailService) SendVerificationEmail(toEmail, verificationToken string) error {
	verificationURL := fmt.Sprintf("%s/auth/verify?token=%s", es.baseURL, verificationToken)

	data := VerificationEmailData{
		Email:           toEmail,
		VerificationURL: verificationURL,
	}

	htmlTemplate := `
<!DOCTYPE html>
<html>
<body>
	<div class="container">
		<div class="header">
			<h3>Verify Your Email Address for Aditya's Portfolio</h3>
		</div>
		<div class="content">
			<p>Thank you for signing up! Please verify your email address by clicking the button below:</p>
			<a href="{{.VerificationURL}}" class="button">Verify Email</a>
			<p>If you didn't create an account, please ignore this email.</p>
		</div>
		<div class="footer">
			<p>This verification link will expire in 24 hours.</p>
		</div>
	</div>
</body>
</html>
`

	tmpl, err := template.New("verification").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	subject := "Verify Your Email Address"
	message := fmt.Sprintf("From: %s\r\n", es.fromEmail) +
		fmt.Sprintf("To: %s\r\n", toEmail) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body.String()

	auth := smtp.PlainAuth("", es.smtpUser, es.smtpPassword, es.smtpHost)
	addr := fmt.Sprintf("%s:%s", es.smtpHost, es.smtpPort)

	err = smtp.SendMail(addr, auth, es.fromEmail, []string{toEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Printf("Verification email sent successfully to %s", toEmail)

	return nil
}

func (es *EmailService) SendVerificationEmailAsync(toEmail, verificationToken string) {
	go func() {
		if err := es.SendVerificationEmail(toEmail, verificationToken); err != nil {
			log.Printf("Failed to send verification email to %s: %v", toEmail, err)
		}
	}()
}
