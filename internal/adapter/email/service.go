package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strings"
	"time"
)

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	UseTLS   bool
	Enabled  bool
}

// Service provides email sending capabilities
type Service struct {
	config Config
}

// NewService creates a new email service
func NewService(config Config) *Service {
	return &Service{config: config}
}

// SubmissionData represents data for the submission notification email
type SubmissionData struct {
	FormName     string
	FormID       string
	SubmissionID string
	SubmittedAt  time.Time
	Fields       map[string]interface{}
	DashboardURL string
}

// SendSubmissionNotification sends a notification email for a new submission
func (s *Service) SendSubmissionNotification(to []string, data SubmissionData) error {
	if !s.config.Enabled {
		// Log instead of sending in dev mode
		fmt.Printf("[EMAIL] Would send submission notification to %v for form %s\n", to, data.FormName)
		return nil
	}

	if len(to) == 0 {
		return nil
	}

	subject := fmt.Sprintf("New submission: %s", data.FormName)
	htmlBody, err := s.renderSubmissionHTML(data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	textBody := s.renderSubmissionText(data)

	return s.sendEmail(to, subject, htmlBody, textBody)
}

// sendEmail sends an email with both HTML and plain text parts
func (s *Service) sendEmail(to []string, subject, htmlBody, textBody string) error {
	boundary := "BOUNDARY_HEADLESSFORMS_EMAIL"

	headers := map[string]string{
		"From":         fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From),
		"To":           strings.Join(to, ", "),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": fmt.Sprintf("multipart/alternative; boundary=%s", boundary),
	}

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")

	// Plain text part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n")

	// HTML part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n")

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// Send via SMTP
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	if s.config.UseTLS {
		return s.sendWithTLS(addr, auth, to, msg.Bytes())
	}

	return smtp.SendMail(addr, auth, s.config.From, to, msg.Bytes())
}

func (s *Service) sendWithTLS(addr string, auth smtp.Auth, to []string, msg []byte) error {
	// #nosec G402 - TLS 1.2 is minimum for email servers
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.config.Host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Printf("[WARN] Failed to close TLS connection: %v", cerr)
		}
	}()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Printf("[WARN] Failed to close SMTP client: %v", cerr)
		}
	}()

	if err := client.Auth(auth); err != nil {
		return err
	}

	if err := client.Mail(s.config.From); err != nil {
		return err
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	return w.Close()
}

func (s *Service) renderSubmissionHTML(data SubmissionData) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>New Form Submission</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px 20px; border-radius: 12px 12px 0 0; text-align: center;">
    <h1 style="color: white; margin: 0; font-size: 24px;">📬 New Submission</h1>
    <p style="color: rgba(255,255,255,0.9); margin: 10px 0 0;">{{.FormName}}</p>
  </div>
  
  <div style="background: #f8f9fa; padding: 20px; border: 1px solid #e9ecef; border-top: none;">
    <p style="color: #666; font-size: 14px; margin: 0;">
      Received on {{.SubmittedAt.Format "January 2, 2006 at 3:04 PM"}}
    </p>
  </div>

  <div style="background: white; padding: 25px; border: 1px solid #e9ecef; border-top: none; border-radius: 0 0 12px 12px;">
    <h2 style="font-size: 16px; color: #333; margin: 0 0 20px; padding-bottom: 10px; border-bottom: 2px solid #f0f0f0;">Submission Details</h2>
    
    <table style="width: 100%; border-collapse: collapse;">
      {{range $key, $value := .Fields}}
      <tr>
        <td style="padding: 12px 0; border-bottom: 1px solid #f0f0f0; color: #666; font-size: 13px; text-transform: uppercase; letter-spacing: 0.5px; width: 35%; vertical-align: top;">{{$key}}</td>
        <td style="padding: 12px 0; border-bottom: 1px solid #f0f0f0; color: #333; font-size: 15px;">{{$value}}</td>
      </tr>
      {{end}}
    </table>

    <div style="margin-top: 25px; text-align: center;">
      <a href="{{.DashboardURL}}" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 12px 30px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 14px;">View in Dashboard</a>
    </div>
  </div>

  <div style="text-align: center; padding: 20px; color: #999; font-size: 12px;">
    <p style="margin: 0;">Sent by HeadlessForms</p>
  </div>
</body>
</html>`

	t, err := template.New("submission").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *Service) renderSubmissionText(data SubmissionData) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("New Submission: %s\n", data.FormName))
	sb.WriteString(fmt.Sprintf("Received: %s\n\n", data.SubmittedAt.Format("January 2, 2006 at 3:04 PM")))
	sb.WriteString("Submission Details:\n")
	sb.WriteString("-------------------\n\n")

	for key, value := range data.Fields {
		sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
	}

	sb.WriteString(fmt.Sprintf("\nView in Dashboard: %s\n", data.DashboardURL))

	return sb.String()
}

// IsEnabled returns whether email sending is enabled
func (s *Service) IsEnabled() bool {
	return s.config.Enabled
}

// SendPasswordReset sends a password reset email
func (s *Service) SendPasswordReset(to, resetURL string) error {
	if !s.config.Enabled {
		fmt.Printf("[EMAIL] Would send password reset to %s with URL: %s\n", to, resetURL)
		return nil
	}

	subject := "Password Reset Request"
	htmlBody := s.renderPasswordResetHTML(resetURL)
	textBody := fmt.Sprintf("Reset your password by visiting: %s\n\nThis link expires in 1 hour.", resetURL)

	return s.sendEmail([]string{to}, subject, htmlBody, textBody)
}

func (s *Service) renderPasswordResetHTML(resetURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Password Reset</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px 20px; border-radius: 12px 12px 0 0; text-align: center;">
    <h1 style="color: white; margin: 0;">🔐 Password Reset</h1>
  </div>
  <div style="background: white; padding: 25px; border: 1px solid #e9ecef; border-top: none; border-radius: 0 0 12px 12px;">
    <p style="color: #333;">You requested a password reset for your HeadlessForms account.</p>
    <p style="color: #333;">Click the button below to set a new password:</p>
    <div style="text-align: center; margin: 25px 0;">
      <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 14px 32px; border-radius: 8px; text-decoration: none; font-weight: 600;">Reset Password</a>
    </div>
    <p style="color: #666; font-size: 14px;">This link will expire in 1 hour.</p>
    <p style="color: #999; font-size: 12px;">If you didn't request this, you can safely ignore this email.</p>
  </div>
</body>
</html>`, resetURL)
}
