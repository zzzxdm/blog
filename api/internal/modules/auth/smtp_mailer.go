package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"mime"
	"net"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type EmailSender interface {
	SendEmailVerification(ctx context.Context, user User, token string) error
	SendInvitation(ctx context.Context, user User, initialPassword string, resetToken string) error
	SendPasswordSetup(ctx context.Context, user User, token string) error
}

type SMTPEmailSender struct {
	host      string
	port      int
	username  string
	password  string
	from      string
	publicURL string
	siteName  string
}

func NewSMTPEmailSender(host string, port string, username string, password string, from string, publicURL ...string) (*SMTPEmailSender, error) {
	host = strings.TrimSpace(host)
	from = strings.TrimSpace(from)
	if host == "" || from == "" {
		return nil, nil
	}

	portNumber, err := strconv.Atoi(strings.TrimSpace(port))
	if err != nil || portNumber <= 0 {
		portNumber = 587
	}

	return &SMTPEmailSender{
		host:      host,
		port:      portNumber,
		username:  strings.TrimSpace(username),
		password:  password,
		from:      from,
		publicURL: normalizedPublicURL(firstString(publicURL, "http://localhost:5173")),
		siteName:  "\u4e91\u95f4\u7b14\u8bb0",
	}, nil
}

func (sender *SMTPEmailSender) SendEmailVerification(_ context.Context, user User, token string) error {
	actionURL := sender.authLink("verify", token)
	text := strings.Join([]string{
		"Please verify your email address.",
		"",
		"Verification link:",
		actionURL,
		"",
		"Verification token:",
		token,
		"",
		"If you did not create this account, you can ignore this email.",
	}, "\r\n")
	html, err := renderEmailTemplate("email-verification.html", emailTemplateData{
		SiteName:    sender.siteName,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		ActionURL:   actionURL,
		Token:       token,
		ExpiresIn:   "24 \u5c0f\u65f6",
	})
	if err != nil {
		return err
	}
	return sender.send(user.Email, "\u9a8c\u8bc1\u4f60\u7684\u90ae\u7bb1", text, html)
}

func (sender *SMTPEmailSender) SendPasswordSetup(_ context.Context, user User, token string) error {
	actionURL := sender.authLink("reset", token)
	text := strings.Join([]string{
		"Please reset your account password.",
		"",
		"Password reset link:",
		actionURL,
		"",
		"Password reset token:",
		token,
		"",
		"If you did not request this password reset, you can ignore this email.",
	}, "\r\n")
	html, err := renderEmailTemplate("password-reset.html", emailTemplateData{
		SiteName:    sender.siteName,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		ActionURL:   actionURL,
		Token:       token,
		ExpiresIn:   "30 \u5206\u949f",
	})
	if err != nil {
		return err
	}
	return sender.send(user.Email, "\u91cd\u7f6e\u4f60\u7684\u8d26\u53f7\u5bc6\u7801", text, html)
}

func (sender *SMTPEmailSender) SendInvitation(_ context.Context, user User, initialPassword string, resetToken string) error {
	actionURL := sender.authLink("reset", resetToken)
	loginURL := sender.loginURL()
	text := strings.Join([]string{
		"You have been invited to create an account.",
		"",
		"Temporary password:",
		initialPassword,
		"",
		"Login link:",
		loginURL,
		"",
		"Password reset link:",
		actionURL,
		"",
		"Password reset token:",
		resetToken,
		"",
		"If you did not expect this invitation, you can ignore this email.",
	}, "\r\n")
	html, err := renderEmailTemplate("password-setup.html", emailTemplateData{
		SiteName:        sender.siteName,
		DisplayName:     user.DisplayName,
		Email:           user.Email,
		ActionURL:       actionURL,
		LoginURL:        loginURL,
		Token:           resetToken,
		InitialPassword: initialPassword,
		ExpiresIn:       "30 \u5206\u949f",
	})
	if err != nil {
		return err
	}
	return sender.send(user.Email, "\u4f60\u7684\u8d26\u53f7\u9080\u8bf7", text, html)
}

func (sender *SMTPEmailSender) send(to string, subject string, textBody string, htmlBody string) error {
	address := net.JoinHostPort(sender.host, strconv.Itoa(sender.port))
	client, err := sender.dial(address)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	defer client.Close()

	if sender.username != "" {
		auth := smtp.PlainAuth("", sender.username, sender.password, sender.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(sender.from); err != nil {
		return fmt.Errorf("set sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("set recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("open smtp data: %w", err)
	}
	if _, err := writer.Write([]byte(sender.message(to, subject, textBody, htmlBody))); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write smtp data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close smtp data: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("quit smtp: %w", err)
	}

	return nil
}

func (sender *SMTPEmailSender) message(to string, subject string, textBody string, htmlBody string) string {
	boundary := "blog-mail-boundary"
	return strings.Join([]string{
		"From: " + sender.from,
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"MIME-Version: 1.0",
		`Content-Type: multipart/alternative; boundary="` + boundary + `"`,
		"",
		"--" + boundary,
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		textBody,
		"--" + boundary,
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		htmlBody,
		"--" + boundary + "--",
		"",
	}, "\r\n")
}

func (sender *SMTPEmailSender) dial(address string) (*smtp.Client, error) {
	tlsConfig := &tls.Config{ServerName: sender.host, MinVersion: tls.VersionTLS12}
	if sender.port == 465 {
		connection, err := tls.Dial("tcp", address, tlsConfig)
		if err != nil {
			return nil, err
		}

		return smtp.NewClient(connection, sender.host)
	}

	client, err := smtp.Dial(address)
	if err != nil {
		return nil, err
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsConfig); err != nil {
			_ = client.Close()
			return nil, err
		}
	}

	return client, nil
}

func (sender *SMTPEmailSender) authLink(mode string, token string) string {
	values := url.Values{}
	values.Set("mode", mode)
	values.Set("token", token)
	return sender.publicURL + "/login?" + values.Encode()
}

func (sender *SMTPEmailSender) loginURL() string {
	return sender.publicURL + "/login"
}

type emailTemplateData struct {
	SiteName        string
	DisplayName     string
	Email           string
	ActionURL       string
	LoginURL        string
	Token           string
	InitialPassword string
	ExpiresIn       string
}

func renderEmailTemplate(name string, data emailTemplateData) (string, error) {
	path, err := findEmailTemplate(name)
	if err != nil {
		return "", err
	}

	parsed, err := template.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("parse email template %s: %w", name, err)
	}

	var body bytes.Buffer
	if err := parsed.Execute(&body, data); err != nil {
		return "", fmt.Errorf("render email template %s: %w", name, err)
	}
	return body.String(), nil
}

func findEmailTemplate(name string) (string, error) {
	candidates := []string{
		filepath.Join("email-templates", name),
		filepath.Join("api", "email-templates", name),
		filepath.Join("..", "..", "..", "email-templates", name),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("email template %s not found", name)
}

func normalizedPublicURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "http://localhost:5173"
	}
	return strings.TrimRight(value, "/")
}

func firstString(values []string, fallback string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return fallback
}
