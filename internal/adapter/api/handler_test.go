package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"headless_form/internal/core/domain"
	"headless_form/internal/core/ports"
	"headless_form/internal/core/service"
)

// MockRepository for API tests
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

func (m *MockRepository) Settings() ports.SettingsRepository {
	return nil // Not used in current tests
}

// MockUserRepository for testing
type MockUserRepository struct{}

func (r *MockUserRepository) Create(ctx context.Context, user *domain.User) error { return nil }
func (r *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}
func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (r *MockUserRepository) Update(ctx context.Context, user *domain.User) error { return nil }
func (r *MockUserRepository) Delete(ctx context.Context, id string) error         { return nil }
func (r *MockUserRepository) List(ctx context.Context) ([]*domain.User, error)    { return nil, nil }
func (r *MockUserRepository) Count(ctx context.Context) (int, error)              { return 0, nil }

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
	return nil
}

func (r *MockFormRepository) ListPaginated(ctx context.Context, limit, offset int) ([]*domain.Form, int, error) {
	var list []*domain.Form
	for _, f := range r.forms {
		list = append(list, f)
	}
	return list, len(list), nil
}

func (r *MockFormRepository) ListByOwnerPaginated(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Form, int, error) {
	var list []*domain.Form
	for _, f := range r.forms {
		if f.OwnerID == ownerID {
			list = append(list, f)
		}
	}
	return list, len(list), nil
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
	return nil, nil
}

func (r *MockSubmissionRepository) GetByFormID(ctx context.Context, formID string) ([]*domain.Submission, error) {
	return r.submissions[formID], nil
}

func (r *MockSubmissionRepository) GetByFormIDPaginated(ctx context.Context, formID string, limit, offset int) ([]*domain.Submission, int, error) {
	subs := r.submissions[formID]
	return subs, len(subs), nil
}

func (r *MockSubmissionRepository) UpdateStatus(ctx context.Context, id string, status domain.SubmissionStatus) error {
	return nil
}

func (r *MockSubmissionRepository) Delete(ctx context.Context, id string) error {
	return nil
}

// MockStatsRepository
type MockStatsRepository struct {
	forms       map[string]*domain.Form
	submissions map[string][]*domain.Submission
}

func (r *MockStatsRepository) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	return &domain.DashboardStats{TotalForms: len(r.forms)}, nil
}

func (r *MockStatsRepository) GetFormStats(ctx context.Context, formID string) (*domain.FormStats, error) {
	return &domain.FormStats{FormID: formID}, nil
}

// Tests
func TestHandleCreateForm(t *testing.T) {
	repo := NewMockRepository()
	formSvc := service.NewFormService(repo)
	submSvc := service.NewSubmissionService(repo)
	statsSvc := service.NewStatsService(repo)
	router := NewRouter(formSvc, submSvc, statsSvc)

	body := `{"name": "Test Form"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forms", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleCreateForm(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "success" {
		t.Errorf("expected status 'success', got '%v'", resp["status"])
	}
}

func TestHandleListForms(t *testing.T) {
	repo := NewMockRepository()
	formSvc := service.NewFormService(repo)
	submSvc := service.NewSubmissionService(repo)
	statsSvc := service.NewStatsService(repo)
	router := NewRouter(formSvc, submSvc, statsSvc)

	// Create a form first
	_, _ = formSvc.CreateForm(context.Background(), "Test Form", "", nil, "", "", "", "public", "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forms", nil)
	w := httptest.NewRecorder()

	router.HandleListForms(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCreateForm_InvalidJSON(t *testing.T) {
	repo := NewMockRepository()
	formSvc := service.NewFormService(repo)
	submSvc := service.NewSubmissionService(repo)
	statsSvc := service.NewStatsService(repo)
	router := NewRouter(formSvc, submSvc, statsSvc)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forms", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleCreateForm(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleHealthCheck(t *testing.T) {
	repo := NewMockRepository()
	formSvc := service.NewFormService(repo)
	submSvc := service.NewSubmissionService(repo)
	statsSvc := service.NewStatsService(repo)
	router := NewRouter(formSvc, submSvc, statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	router.HandleHealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["status"] != "healthy" {
		t.Errorf("expected health status 'healthy'")
	}
}
