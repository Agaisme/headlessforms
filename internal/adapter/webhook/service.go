package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"headless_form/internal/core/domain"
)

// Payload represents the data sent to webhooks
type Payload struct {
	Event        string                 `json:"event"`
	FormID       string                 `json:"form_id"`
	FormName     string                 `json:"form_name"`
	SubmissionID string                 `json:"submission_id"`
	Timestamp    time.Time              `json:"timestamp"`
	Data         map[string]interface{} `json:"data"`
}

// Service handles webhook delivery
type Service struct {
	client  *http.Client
	retries int
}

// NewService creates a new webhook service
func NewService() *Service {
	return &Service{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		retries: 3,
	}
}

// TriggerSubmission sends a webhook for a new submission
func (s *Service) TriggerSubmission(form *domain.Form, submission *domain.Submission, data map[string]interface{}) {
	if form.WebhookURL == "" {
		return
	}

	payload := Payload{
		Event:        "submission.created",
		FormID:       form.PublicID,
		FormName:     form.Name,
		SubmissionID: submission.ID,
		Timestamp:    submission.CreatedAt,
		Data:         data,
	}

	go s.deliver(form.WebhookURL, form.WebhookSecret, payload)
}

func (s *Service) deliver(url, secret string, payload Payload) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WEBHOOK] Failed to marshal payload: %v", err)
		return
	}

	for attempt := 1; attempt <= s.retries; attempt++ {
		err := s.sendRequest(url, secret, body)
		if err == nil {
			log.Printf("[WEBHOOK] Delivered to %s (attempt %d)", url, attempt)
			return
		}

		log.Printf("[WEBHOOK] Attempt %d failed for %s: %v", attempt, url, err)

		if attempt < s.retries {
			// Exponential backoff: 1s, 2s, 4s
			time.Sleep(time.Duration(1<<(attempt-1)) * time.Second)
		}
	}

	log.Printf("[WEBHOOK] Failed after %d attempts for %s", s.retries, url)
}

func (s *Service) sendRequest(url, secret string, body []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HeadlessForms-Webhook/1.0")
	req.Header.Set("X-Webhook-Event", "submission.created")
	req.Header.Set("X-Webhook-Timestamp", time.Now().UTC().Format(time.RFC3339))

	// Sign payload with HMAC-SHA256 if secret is provided
	if secret != "" {
		signature := s.signPayload(body, secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Read and discard body to allow connection reuse
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("unexpected status %d", resp.StatusCode)
}

func (s *Service) signPayload(body []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// TestWebhook sends a test payload to verify webhook configuration
func (s *Service) TestWebhook(url, secret string) error {
	payload := Payload{
		Event:        "test",
		FormID:       "test-form-id",
		FormName:     "Test Form",
		SubmissionID: "test-submission-id",
		Timestamp:    time.Now(),
		Data: map[string]interface{}{
			"message": "This is a test webhook from HeadlessForms",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.sendRequest(url, secret, body)
}
