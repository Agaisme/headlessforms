package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Validation errors
var (
	ErrFormNameRequired   = errors.New("form name is required")
	ErrFormNameTooLong    = errors.New("form name must be less than 100 characters")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrFormNotFound       = errors.New("form not found")
	ErrSubmissionNotFound = errors.New("submission not found")
)

// FormStatus represents the state of a form
type FormStatus string

const (
	FormStatusActive   FormStatus = "active"
	FormStatusInactive FormStatus = "inactive"
)

// AccessMode determines who can submit to a form
type AccessMode string

const (
	AccessModePublic  AccessMode = "public"   // Anyone can submit (default)
	AccessModeWithKey AccessMode = "with_key" // Requires SubmissionKey in hidden field
	AccessModePrivate AccessMode = "private"  // Only authenticated users can submit
)

// Access control errors
var (
	ErrInvalidSubmissionKey = errors.New("invalid submission key")
	ErrAuthRequired         = errors.New("authentication required for this form")
)

// Form represents a form endpoint configuration
type Form struct {
	ID              string     `json:"id"`
	OwnerID         string     `json:"owner_id"` // User who created this form
	PublicID        string     `json:"public_id"`
	Name            string     `json:"name"`
	Status          FormStatus `json:"status"`
	NotifyEmails    []string   `json:"notify_emails"`
	AllowedOrigins  []string   `json:"allowed_origins"`
	RedirectURL     string     `json:"redirect_url"`
	WebhookURL      string     `json:"webhook_url,omitempty"`
	WebhookSecret   string     `json:"webhook_secret,omitempty"`
	AccessMode      string     `json:"access_mode"` // public, with_key, private
	SubmissionKey   string     `json:"submission_key,omitempty"`
	SubmissionCount int        `json:"submission_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Validate checks if the form data is valid
func (f *Form) Validate() error {
	f.Name = strings.TrimSpace(f.Name)
	if f.Name == "" {
		return ErrFormNameRequired
	}
	if len(f.Name) > 100 {
		return ErrFormNameTooLong
	}
	return nil
}

// SubmissionStatus represents the read state of a submission
type SubmissionStatus string

const (
	SubmissionStatusUnread SubmissionStatus = "unread"
	SubmissionStatusRead   SubmissionStatus = "read"
)

// Submission represents a form submission
type Submission struct {
	ID        string           `json:"id"`
	FormID    string           `json:"form_id"`
	Status    SubmissionStatus `json:"status"`
	Data      json.RawMessage  `json:"data"`
	Meta      json.RawMessage  `json:"meta"`
	CreatedAt time.Time        `json:"created_at"`
}

// DailySubmission represents submission count for a day
type DailySubmission struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// DashboardStats contains overview statistics
type DashboardStats struct {
	TotalForms          int               `json:"total_forms"`
	ActiveForms         int               `json:"active_forms"`
	TotalSubmissions    int               `json:"total_submissions"`
	UnreadSubmissions   int               `json:"unread_submissions"`
	SubmissionsToday    int               `json:"submissions_today"`
	SubmissionsThisWeek int               `json:"submissions_this_week"`
	DailySubmissions    []DailySubmission `json:"daily_submissions,omitempty"`
}

// FormStats contains statistics for a single form
type FormStats struct {
	FormID              string `json:"form_id"`
	TotalSubmissions    int    `json:"total_submissions"`
	UnreadSubmissions   int    `json:"unread_submissions"`
	SubmissionsToday    int    `json:"submissions_today"`
	SubmissionsThisWeek int    `json:"submissions_this_week"`
}
