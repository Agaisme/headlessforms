package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"headless_form/internal/core/domain"
	"headless_form/internal/core/ports"

	"github.com/google/uuid"
)

// FormService handles form-related business logic
type FormService struct {
	repo ports.Repository
}

func NewFormService(repo ports.Repository) *FormService {
	return &FormService{repo: repo}
}

func (s *FormService) CreateForm(ctx context.Context, name, redirectURL string, notifyEmails []string, webhookURL, webhookSecret, ownerID, accessMode, submissionKey string) (*domain.Form, error) {
	id := uuid.New().String()
	publicID := uuid.New().String()
	now := time.Now()

	// Set default access mode if not provided
	if accessMode == "" {
		accessMode = "public"
	}

	form := &domain.Form{
		ID:              id,
		OwnerID:         ownerID,
		PublicID:        publicID,
		Name:            name,
		Status:          domain.FormStatusActive,
		NotifyEmails:    notifyEmails,
		AllowedOrigins:  []string{"*"},
		RedirectURL:     redirectURL,
		WebhookURL:      webhookURL,
		WebhookSecret:   webhookSecret,
		AccessMode:      accessMode,
		SubmissionKey:   submissionKey,
		SubmissionCount: 0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Validate form
	if err := form.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Form().Create(ctx, form); err != nil {
		return nil, fmt.Errorf("create form: %w", err)
	}

	return form, nil
}

func (s *FormService) GetForm(ctx context.Context, publicID string) (*domain.Form, error) {
	form, err := s.repo.Form().GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("get form: %w", err)
	}
	if form == nil {
		return nil, domain.ErrFormNotFound
	}
	return form, nil
}

func (s *FormService) ListForms(ctx context.Context) ([]*domain.Form, error) {
	return s.repo.Form().List(ctx)
}

func (s *FormService) ListFormsPaginated(ctx context.Context, page, limit int) ([]*domain.Form, int, error) {
	offset := (page - 1) * limit
	return s.repo.Form().ListPaginated(ctx, limit, offset)
}

func (s *FormService) ListFormsByOwnerPaginated(ctx context.Context, ownerID string, page, limit int) ([]*domain.Form, int, error) {
	offset := (page - 1) * limit
	return s.repo.Form().ListByOwnerPaginated(ctx, ownerID, limit, offset)
}

// GetFormByID retrieves a form by its internal ID (not public_id)
func (s *FormService) GetFormByID(ctx context.Context, id string) (*domain.Form, error) {
	form, err := s.repo.Form().GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get form by id: %w", err)
	}
	if form == nil {
		return nil, domain.ErrFormNotFound
	}
	return form, nil
}

func (s *FormService) UpdateForm(ctx context.Context, publicID string, name, redirectURL string, notifyEmails []string, status domain.FormStatus, webhookURL, webhookSecret, accessMode, submissionKey string) (*domain.Form, error) {
	form, err := s.repo.Form().GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("get form: %w", err)
	}
	if form == nil {
		return nil, domain.ErrFormNotFound
	}

	form.Name = name
	form.RedirectURL = redirectURL
	form.NotifyEmails = notifyEmails
	form.Status = status
	form.WebhookURL = webhookURL
	form.WebhookSecret = webhookSecret
	form.AccessMode = accessMode
	form.SubmissionKey = submissionKey
	form.UpdatedAt = time.Now()

	if err := form.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Form().Update(ctx, form); err != nil {
		return nil, fmt.Errorf("update form: %w", err)
	}

	return form, nil
}

func (s *FormService) DeleteForm(ctx context.Context, publicID string) error {
	form, err := s.repo.Form().GetByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("get form: %w", err)
	}
	if form == nil {
		return domain.ErrFormNotFound
	}

	if err := s.repo.Form().Delete(ctx, form.ID); err != nil {
		return fmt.Errorf("delete form: %w", err)
	}

	return nil
}

// SubmissionService handles submission-related business logic
type SubmissionService struct {
	repo            ports.Repository
	onNewSubmission func(form *domain.Form, submission *domain.Submission, data map[string]interface{})
}

func NewSubmissionService(repo ports.Repository) *SubmissionService {
	return &SubmissionService{repo: repo}
}

// SetNotificationCallback sets a callback for new submissions (for email notifications)
func (s *SubmissionService) SetNotificationCallback(fn func(form *domain.Form, submission *domain.Submission, data map[string]interface{})) {
	s.onNewSubmission = fn
}

func (s *SubmissionService) Submit(ctx context.Context, publicID string, data map[string]interface{}, meta map[string]interface{}) (*domain.Submission, error) {
	form, err := s.repo.Form().GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("invalid form: %w", err)
	}
	if form == nil {
		return nil, domain.ErrFormNotFound
	}

	// Check if form is active
	if form.Status != domain.FormStatusActive {
		return nil, fmt.Errorf("form is not accepting submissions")
	}

	// Access control validation based on form's access mode
	switch form.AccessMode {
	case string(domain.AccessModeWithKey):
		// Validate submission key from hidden field
		submittedKey, _ := data["_submission_key"].(string)
		if submittedKey == "" || submittedKey != form.SubmissionKey {
			return nil, domain.ErrInvalidSubmissionKey
		}
		// Remove the key from data so it's not stored
		delete(data, "_submission_key")
	case string(domain.AccessModePrivate):
		// For private forms, we need to check if request has auth context
		// This is passed via meta from the handler
		if authUserID, ok := meta["_auth_user_id"].(string); !ok || authUserID == "" {
			return nil, domain.ErrAuthRequired
		}
		// Remove internal auth field from meta
		delete(meta, "_auth_user_id")
		// case "public" or empty - no validation needed
	}

	dataBytes, _ := json.Marshal(data)
	metaBytes, _ := json.Marshal(meta)

	submission := &domain.Submission{
		ID:        uuid.New().String(),
		FormID:    form.ID,
		Status:    domain.SubmissionStatusUnread,
		Data:      json.RawMessage(dataBytes),
		Meta:      json.RawMessage(metaBytes),
		CreatedAt: time.Now(),
	}

	if err := s.repo.Submission().Create(ctx, submission); err != nil {
		return nil, fmt.Errorf("save submission: %w", err)
	}

	// Increment submission count
	_ = s.repo.Form().IncrementSubmissionCount(ctx, form.ID)

	// Trigger email notification (async, don't block submission)
	if s.onNewSubmission != nil {
		go s.onNewSubmission(form, submission, data)
	}

	return submission, nil
}

func (s *SubmissionService) ListSubmissions(ctx context.Context, publicID string) ([]*domain.Submission, error) {
	form, err := s.repo.Form().GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("lookup form: %w", err)
	}
	if form == nil {
		return nil, domain.ErrFormNotFound
	}

	return s.repo.Submission().GetByFormID(ctx, form.ID)
}

func (s *SubmissionService) ListSubmissionsPaginated(ctx context.Context, publicID string, page, limit int) ([]*domain.Submission, int, error) {
	form, err := s.repo.Form().GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, 0, fmt.Errorf("lookup form: %w", err)
	}
	if form == nil {
		return nil, 0, domain.ErrFormNotFound
	}

	offset := (page - 1) * limit
	return s.repo.Submission().GetByFormIDPaginated(ctx, form.ID, limit, offset)
}

func (s *SubmissionService) MarkAsRead(ctx context.Context, submissionID string) error {
	return s.repo.Submission().UpdateStatus(ctx, submissionID, domain.SubmissionStatusRead)
}

func (s *SubmissionService) MarkAsUnread(ctx context.Context, submissionID string) error {
	return s.repo.Submission().UpdateStatus(ctx, submissionID, domain.SubmissionStatusUnread)
}

func (s *SubmissionService) DeleteSubmission(ctx context.Context, submissionID string) error {
	return s.repo.Submission().Delete(ctx, submissionID)
}

// GetSubmission retrieves a single submission by ID
func (s *SubmissionService) GetSubmission(ctx context.Context, submissionID string) (*domain.Submission, error) {
	submission, err := s.repo.Submission().GetByID(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("get submission: %w", err)
	}
	if submission == nil {
		return nil, domain.ErrSubmissionNotFound
	}
	return submission, nil
}

// StatsService handles statistics business logic
type StatsService struct {
	repo ports.Repository
}

func NewStatsService(repo ports.Repository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	return s.repo.Stats().GetDashboardStats(ctx)
}

func (s *StatsService) GetFormStats(ctx context.Context, publicID string) (*domain.FormStats, error) {
	form, err := s.repo.Form().GetByPublicID(ctx, publicID)
	if err != nil || form == nil {
		return nil, domain.ErrFormNotFound
	}
	return s.repo.Stats().GetFormStats(ctx, form.ID)
}
