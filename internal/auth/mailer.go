package auth

import (
	"context"
	"fmt"
	"strconv"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

func NewMailer(host, port, username, password, from string) *Mailer {
	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort <= 0 {
		parsedPort = 587
	}
	if from == "" {
		from = username
	}

	dialer := gomail.NewDialer(host, parsedPort, username, password)

	return &Mailer{dialer: dialer, from: from}
}

func (m *Mailer) SendPasswordReset(ctx context.Context, email, resetURL string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.from)
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Reset your Fit the Meet password")
	message.SetBody(
		"text/plain",
		fmt.Sprintf("Use this link to reset your Fit the Meet password. The link expires in 30 minutes:\n\n%s\n\nIf you did not request this, you can ignore this email.", resetURL),
	)
	message.AddAlternative(
		"text/html",
		fmt.Sprintf(`<h2>Reset your password</h2><p>Use the link below to choose a new Fit the Meet password. It expires in 30 minutes.</p><p><a href="%s">Reset password</a></p><p>If you did not request this, you can ignore this email.</p>`, resetURL),
	)

	return m.dialer.DialAndSend(message)
}
