package mail

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type Mailer interface {
	SendRegistrationCode(ctx context.Context, to, code string, expiresIn time.Duration) error
	SendPasswordResetProof(ctx context.Context, to, proof string, expiresIn time.Duration) error
	SendPasswordResetNotice(ctx context.Context, to string) error
}

type SMTPMailer struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPMailer(host string, port int, username, password, from string) *SMTPMailer {
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (m *SMTPMailer) SendRegistrationCode(ctx context.Context, to, code string, expiresIn time.Duration) error {
	subject := "Podcast Hub 邮箱验证码"
	body := fmt.Sprintf("你的注册验证码是：%s\n验证码将在 %d 分钟后过期。", code, int(expiresIn.Minutes()))
	return m.send(ctx, to, subject, body)
}

func (m *SMTPMailer) SendPasswordResetProof(ctx context.Context, to, proof string, expiresIn time.Duration) error {
	subject := "Podcast Hub 密码重置验证码"
	body := fmt.Sprintf("你的密码重置验证码是：%s\n验证码将在 %d 分钟后过期。", proof, int(expiresIn.Minutes()))
	return m.send(ctx, to, subject, body)
}

func (m *SMTPMailer) SendPasswordResetNotice(ctx context.Context, to string) error {
	subject := "Podcast Hub 密码已重置通知"
	body := "你的账户密码已被重置。如果这不是你的操作，请立即联系管理员。"
	return m.send(ctx, to, subject, body)
}

func (m *SMTPMailer) send(_ context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	var auth smtp.Auth
	if strings.TrimSpace(m.username) != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}
	message := strings.Join([]string{
		fmt.Sprintf("From: %s", m.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"",
		body,
	}, "\r\n")
	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(message))
}
