// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package lib

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"time"
)

func SendEmail(to, subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = os.Getenv("MAIL_FROM")
	}

	if host == "" || from == "" {
		fmt.Printf("[Email] Skipped sending to %s (no config)\nSubject: %s\nBody: %s\n", to, subject, body)
		return nil
	}

	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "boop.cat"
	}

	msg := []byte(fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s\r\n", fromName, from, to, subject, body))

	addr := fmt.Sprintf("%s:%s", host, port)

	if port == "465" || os.Getenv("SMTP_SECURE") == "true" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to dial tls: %w", err)
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("failed to create smtp client: %w", err)
		}
		defer c.Quit()

		if user != "" && pass != "" {
			auth := smtp.PlainAuth("", user, pass, host)
			if err = c.Auth(auth); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
		}

		if err = c.Mail(from); err != nil {
			return err
		}
		if err = c.Rcpt(to); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(msg)
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
		return nil
	}

	auth := smtp.PlainAuth("", user, pass, host)
	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		fmt.Printf("[Email] Failed to send: %v\n", err)
		return err
	}
	return nil
}

func SendVerificationEmail(to, token, username string) error {
	displayName := username
	if displayName == "" {
		displayName = "there"
	}
	url := fmt.Sprintf("%s/auth/verify-email?token=%s", os.Getenv("PUBLIC_URL"), token)
	subject := "Verify your email - boop.cat"
	body := buildEmailTemplate("Verify your email",
		fmt.Sprintf("Hey %s! Click the button below to verify your email address and activate your account.", displayName),
		"Verify Email", url,
		"This link expires in 24 hours.")
	return SendEmail(to, subject, body)
}

func SendPasswordResetEmail(to, token, username string) error {
	displayName := username
	if displayName == "" {
		displayName = "there"
	}
	url := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("PUBLIC_URL"), token)
	subject := "Reset your password - boop.cat"
	body := buildEmailTemplate("Reset your password",
		fmt.Sprintf("Hey %s! Someone requested a password reset for your account. Click the button below to set a new password.", displayName),
		"Reset Password", url,
		"This link expires in 1 hour. If you didn't request this, ignore this email.")
	return SendEmail(to, subject, body)
}

func buildEmailTemplate(heading, message, buttonText, buttonURL, footer string) string {
	brandName := "boop.cat"
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="color-scheme" content="light dark">
  <meta name="supported-color-schemes" content="light dark">
  <title>%s</title>
  <!--[if mso]>
  <style type="text/css">
    body, table, td, p, a, h1 {font-family: Arial, sans-serif !important;}
  </style>
  <![endif]-->
  <style>
    :root { color-scheme: light dark; }
    body { margin: 0; padding: 0; }
    @media (prefers-color-scheme: dark) {
      .email-bg { background-color: #1a1a2e !important; }
      .email-card { background-color: #16213e !important; }
      .email-title { color: #f1f5f9 !important; }
      .email-text { color: #cbd5e1 !important; }
      .email-muted { color: #94a3b8 !important; }
      .email-link { color: #60a5fa !important; }
      .email-divider { border-color: #334155 !important; }
    }
  </style>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; -webkit-font-smoothing: antialiased;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" class="email-bg" style="background-color: #f1f5f9;">
    <tr>
      <td align="center" style="padding: 40px 16px;">
        <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="max-width: 480px;">
          <!-- Logo -->
          <tr>
            <td align="center" style="padding-bottom: 24px;">
              <span class="email-title" style="font-size: 20px; font-weight: 700; color: #0f172a;">%s</span>
            </td>
          </tr>
          <!-- Card -->
          <tr>
            <td>
              <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" class="email-card" style="background-color: #ffffff; border-radius: 12px;">
                <tr>
                  <td style="padding: 32px 28px;">
                    <h1 class="email-title" style="margin: 0 0 8px; font-size: 22px; font-weight: 700; color: #0f172a; text-align: center;">%s</h1>
                    <p class="email-text" style="margin: 0 0 24px; font-size: 15px; color: #475569; text-align: center; line-height: 1.6;">%s</p>
                    <!-- Button -->
                    <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0">
                      <tr>
                        <td align="center" style="padding: 4px 0;">
                          <a href="%s" target="_blank" style="display: inline-block; padding: 12px 28px; background-color: #2563eb; color: #ffffff; font-size: 15px; font-weight: 600; text-decoration: none; border-radius: 8px;">%s</a>
                        </td>
                      </tr>
                    </table>
                    <p class="email-text" style="margin: 24px 0 0; font-size: 13px; color: #64748b; text-align: center; line-height: 1.6;">
                      Or copy this link into your browser:<br>
                      <a href="%s" class="email-link" style="color: #2563eb; word-break: break-all; text-decoration: underline;">%s</a>
                    </p>
                    <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="margin-top: 24px;">
                      <tr>
                        <td class="email-divider" style="border-top: 1px solid #e2e8f0; padding-top: 16px;">
                          <p class="email-muted" style="margin: 0; font-size: 12px; color: #94a3b8; text-align: center;">%s</p>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td align="center" style="padding-top: 24px;">
              <p class="email-muted" style="margin: 0; font-size: 13px; color: #64748b; line-height: 1.6;">
                &copy; %d %s
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, heading, brandName, heading, message, buttonURL, buttonText, buttonURL, buttonURL, footer, time.Now().Year(), brandName)
}
