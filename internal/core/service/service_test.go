package service

import (
	"context"
	"testing"

	"headless_form/internal/core/domain"
	"headless_form/internal/core/ports"
)

// MockRepository implements ports.Repository for testing
type MockRepository struct {
	forms       map[string]*domain.Form
	submissions map[string][]*domain.Submission
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		forms:       make(map[string]*domain.Form),
		submissions: make(map[string][]*domain.Submission),
	}
}

func (m *MockRepository) Tx(ctx context.Context, fn func(ports.Repository) error) error {
	return fn(m)
}

func (m *MockRepository) Form() ports.FormRepository {
	return &MockFormRepository{forms: m.forms}
}

func (m *MockRepository) Submission() ports.SubmissionRepository {
	return &MockSubmissionRepository{submissions: m.submissions, forms: m.forms}
}

func (m *MockRepository) Stats() ports.StatsRepository {
	return &MockStatsRepository{forms: m.forms, submissions: m.submissions}
}

func (m *MockRepository) User() ports.UserRepository {
	return &MockUserRepository{}
}

// MockUserRepository for testing
type MockUserRepository struct {
	users map[string]*domain.User
}

func (r *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return nil
}

func (r *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}

func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

func (r *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return nil
}

func (r *MockUserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (r *MockUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	return nil, nil
}

func (r *MockUserRepository) Count(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *MockRepository) PasswordReset() ports.PasswordResetRepository {
	return &MockPasswordResetRepository{}
}

// MockPasswordResetRepository for testing
type MockPasswordResetRepository struct{}

func (r *MockPasswordResetRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	return nil
}
func (r *MockPasswordResetRepository) GetByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	return nil, nil
}
func (r *MockPasswordResetRepository) MarkAsUsed(ctx context.Context, tokenID string) error {
	return nil
}
func (r *MockPasswordResetRepository) DeleteExpired(ctx context.Context) error { return nil }

func (m *MockRepository) Settings() ports.SettingsRepository {
	return nil // Not used in current tests
}

// MockFormRepository
type MockFormRepository struct {
	forms map[string]*domain.Form
}

func (r *MockFormRepository) Create(ctx context.Context, f *domain.Form) error {
	r.forms[f.PublicID] = f
	return nil
}

func (r *MockFormRepository) Update(ctx context.Context, f *domain.Form) error {
	r.forms[f.PublicID] = f
	return nil
}

func (r *MockFormRepository) GetByPublicID(ctx context.Context, publicID string) (*domain.Form, error) {
	return r.forms[publicID], nil
}

func (r *MockFormRepository) GetByID(ctx context.Context, id string) (*domain.Form, error) {
	for _, f := range r.forms {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, nil
}

func (r *MockFormRepository) List(ctx context.Context) ([]*domain.Form, error) {
	var list []*domain.Form
	for _, f := range r.forms {
		list = append(list, f)
	}
	return list, nil
}

func (r *MockFormRepository) Delete(ctx context.Context, id string) error {
	for pid, f := range r.forms {
		if f.ID == id {
			delete(r.forms, pid)
			break
		}
	}
	return nil
}

func (r *MockFormRepository) IncrementSubmissionCount(ctx context.Context, formID string) error {
	for _, f := range r.forms {
		if f.ID == formID {
			f.SubmissionCount++
			break
		}
	}
	return nil
}

func (r *MockFormRepository) ListPaginated(ctx context.Context, limit, offset int) ([]*domain.Form, int, error) {
	var list []*domain.Form
	for _, f := range r.forms {
		list = append(list, f)
	}
	// Simple pagination simulation
	total := len(list)
	if offset >= len(list) {
		return []*domain.Form{}, total, nil
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end], total, nil
}

func (r *MockFormRepository) ListByOwnerPaginated(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Form, int, error) {
	var list []*domain.Form
	for _, f := range r.forms {
		if f.OwnerID == ownerID {
			list = append(list, f)
		}
	}
	total := len(list)
	if offset >= len(list) {
		return []*domain.Form{}, total, nil
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end], total, nil
}

// MockSubmissionRepository
type MockSubmissionRepository struct {
	submissions map[string][]*domain.Submission
	forms       map[string]*domain.Form
}

func (r *MockSubmissionRepository) Create(ctx context.Context, s *domain.Submission) error {
	r.submissions[s.FormID] = append(r.submissions[s.FormID], s)
	return nil
}

func (r *MockSubmissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	for _, subs := range r.submissions {
		for _, s := range subs {
			if s.ID == id {
				return s, nil
			}
		}
	}
	return nil, nil
}

func (r *MockSubmissionRepository) GetByFormID(ctx context.Context, formID string) ([]*domain.Submission, error) {
	return r.submissions[formID], nil
}

func (r *MockSubmissionRepository) GetByFormIDPaginated(ctx context.Context, formID string, limit, offset int) ([]*domain.Submission, int, error) {
	subs := r.submissions[formID]
	total := len(subs)
	if offset >= len(subs) {
		return []*domain.Submission{}, total, nil
	}
	end := offset + limit
	if end > len(subs) {
		end = len(subs)
	}
	return subs[offset:end], total, nil
}

func (r *MockSubmissionRepository) UpdateStatus(ctx context.Context, id string, status domain.SubmissionStatus) error {
	for _, subs := range r.submissions {
		for _, s := range subs {
			if s.ID == id {
				s.Status = status
				break
			}
		}
	}
	return nil
}

func (r *MockSubmissionRepository) Delete(ctx context.Context, id string) error {
	for formID, subs := range r.submissions {
		for i, s := range subs {
			if s.ID == id {
				r.submissions[formID] = append(subs[:i], subs[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

// MockStatsRepository
type MockStatsRepository struct {
	forms       map[string]*domain.Form
	submissions map[string][]*domain.Submission
}

func (r *MockStatsRepository) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	total := 0
	for _, subs := range r.submissions {
		total += len(subs)
	}
	return &domain.DashboardStats{
		TotalForms:       len(r.forms),
		ActiveForms:      len(r.forms),
		TotalSubmissions: total,
	}, nil
}

func (r *MockStatsRepository) GetFormStats(ctx context.Context, formID string) (*domain.FormStats, error) {
	return &domain.FormStats{
		FormID:           formID,
		TotalSubmissions: len(r.submissions[formID]),
	}, nil
}

// Tests
func TestFormService_CreateForm(t *testing.T) {
	repo := NewMockRepository()
	svc := NewFormService(repo)

	form, err := svc.CreateForm(context.Background(), "Contact Form", "", nil, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form.Name != "Contact Form" {
		t.Errorf("expected name 'Contact Form', got '%s'", form.Name)
	}
	if form.Status != domain.FormStatusActive {
		t.Errorf("expected status 'active', got '%s'", form.Status)
	}
}

func TestFormService_CreateForm_ValidationError(t *testing.T) {
	repo := NewMockRepository()
	svc := NewFormService(repo)

	_, err := svc.CreateForm(context.Background(), "", "", nil, "", "", "")
	if err != domain.ErrFormNameRequired {
		t.Errorf("expected ErrFormNameRequired, got %v", err)
	}
}

func TestFormService_ListForms(t *testing.T) {
	repo := NewMockRepository()
	svc := NewFormService(repo)

	_, _ = svc.CreateForm(context.Background(), "Form 1", "", nil, "", "", "")
	_, _ = svc.CreateForm(context.Background(), "Form 2", "", nil, "", "", "")

	forms, err := svc.ListForms(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forms) != 2 {
		t.Errorf("expected 2 forms, got %d", len(forms))
	}
}

func TestSubmissionService_Submit(t *testing.T) {
	repo := NewMockRepository()
	formSvc := NewFormService(repo)
	submSvc := NewSubmissionService(repo)

	form, _ := formSvc.CreateForm(context.Background(), "Test Form", "", nil, "", "", "")

	sub, err := submSvc.Submit(context.Background(), form.PublicID, map[string]interface{}{"email": "test@example.com"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.FormID != form.ID {
		t.Errorf("expected form_id '%s', got '%s'", form.ID, sub.FormID)
	}
	if sub.Status != domain.SubmissionStatusUnread {
		t.Errorf("expected status 'unread', got '%s'", sub.Status)
	}
}

func TestSubmissionService_Submit_FormNotFound(t *testing.T) {
	repo := NewMockRepository()
	submSvc := NewSubmissionService(repo)

	_, err := submSvc.Submit(context.Background(), "nonexistent", nil, nil)
	if err != domain.ErrFormNotFound {
		t.Errorf("expected ErrFormNotFound, got %v", err)
	}
}

func TestSubmissionService_ListSubmissions(t *testing.T) {
	repo := NewMockRepository()
	formSvc := NewFormService(repo)
	submSvc := NewSubmissionService(repo)

	form, _ := formSvc.CreateForm(context.Background(), "Test Form", "", nil, "", "", "")
	_, _ = submSvc.Submit(context.Background(), form.PublicID, map[string]interface{}{"email": "a@b.com"}, nil)
	_, _ = submSvc.Submit(context.Background(), form.PublicID, map[string]interface{}{"email": "c@d.com"}, nil)

	subs, err := submSvc.ListSubmissions(context.Background(), form.PublicID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs) != 2 {
		t.Errorf("expected 2 submissions, got %d", len(subs))
	}
}
