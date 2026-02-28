package service

import (
	"fmt"
	"log"

	resend "github.com/resend/resend-go/v3"
	"github.com/mobileproxy/server/internal/domain"
)

// EmailService wraps the Resend Go SDK for transactional email delivery.
type EmailService struct {
	client  *resend.Client
	from    string
	baseURL string // dashboard URL for building links in emails
}

// NewEmailService creates an EmailService. If cfg.APIKey is empty, emails are
// logged to stdout instead of sent (development fallback).
func NewEmailService(cfg domain.ResendConfig) *EmailService {
	var client *resend.Client
	if cfg.APIKey != "" {
		client = resend.NewClient(cfg.APIKey)
	}
	return &EmailService{
		client:  client,
		from:    cfg.FromEmail,
		baseURL: cfg.BaseURL,
	}
}

// SendVerification sends an email-verification message with a link containing
// the raw (unhashed) token. If the Resend API key is not configured, the email
// contents are logged to stdout for development convenience.
func (s *EmailService) SendVerification(to, rawToken string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, rawToken)
	subject := "Verify your PocketProxy account"
	html := buildEmailHTML(
		"Verify your PocketProxy account",
		"Thanks for signing up! Click the button below to confirm your email address and activate your account.",
		"Verify Email",
		link,
		"This link expires in 24 hours.",
	)

	if s.client == nil {
		log.Printf("[EmailService] DEV — would send verification email to %s\nSubject: %s\nLink: %s\n", to, subject, link)
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}
	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}

// SendPasswordReset sends a password-reset email with a link containing the
// raw (unhashed) token. Falls back to stdout logging in development.
func (s *EmailService) SendPasswordReset(to, rawToken string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, rawToken)
	subject := "Reset your PocketProxy password"
	html := buildEmailHTML(
		"Reset your PocketProxy password",
		"We received a request to reset your password. Click the button below to choose a new one.",
		"Reset Password",
		link,
		"This link expires in 24 hours. If you didn't request this, ignore this email.",
	)

	if s.client == nil {
		log.Printf("[EmailService] DEV — would send password-reset email to %s\nSubject: %s\nLink: %s\n", to, subject, link)
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}
	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("send password-reset email: %w", err)
	}
	return nil
}

// buildEmailHTML returns a simple inline-styled transactional email body.
// Brand colour: #8b5cf6 (purple-500).
func buildEmailHTML(title, bodyText, buttonLabel, buttonURL, footerNote string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>%s</title></head>
<body style="margin:0;padding:0;background-color:#f4f4f5;font-family:sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="padding:40px 0;">
    <tr>
      <td align="center">
        <table width="560" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
          <!-- Header -->
          <tr>
            <td style="background:#8b5cf6;padding:32px 40px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:22px;font-weight:700;">PocketProxy</h1>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding:40px 40px 24px;">
              <h2 style="margin:0 0 16px;font-size:20px;color:#111827;">%s</h2>
              <p style="margin:0 0 32px;font-size:15px;color:#374151;line-height:1.6;">%s</p>
              <table cellpadding="0" cellspacing="0">
                <tr>
                  <td>
                    <a href="%s"
                       style="display:inline-block;background:#8b5cf6;color:#ffffff;text-decoration:none;padding:14px 32px;border-radius:6px;font-size:15px;font-weight:600;">
                      %s
                    </a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="padding:24px 40px 40px;border-top:1px solid #e5e7eb;">
              <p style="margin:0;font-size:13px;color:#6b7280;">%s</p>
              <p style="margin:12px 0 0;font-size:13px;color:#6b7280;">
                Or paste this link into your browser:<br>
                <a href="%s" style="color:#8b5cf6;word-break:break-all;">%s</a>
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, title, title, bodyText, buttonURL, buttonLabel, footerNote, buttonURL, buttonURL)
}
