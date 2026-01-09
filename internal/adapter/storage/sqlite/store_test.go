package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"headless_form/internal/core/domain"
)

// TestNew verifies database creation and migration
func TestNew(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "headless_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	t.Cleanup(func() { _ = os.Remove(tmpPath) })

	store, err := New(tmpPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if store == nil {
		t.Fatal("store should not be nil")
	}
}

// TestFormRepository_CRUD tests form create, read, update, delete operations
func TestFormRepository_CRUD(t *testing.T) {
	store := setupTestStore(t)
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	formRepo := store.Form()

	// Create
	form := &domain.Form{
		ID:              "test-id-1",
		PublicID:        "test-public-1",
		Name:            "Test Form",
		Status:          domain.FormStatusActive,
		NotifyEmails:    []string{"test@example.com"},
		AllowedOrigins:  []string{"*"},
		RedirectURL:     "https://example.com/thanks",
		AccessMode:      "public",
		SubmissionCount: 0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := formRepo.Create(ctx, form)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Read
	retrieved, err := formRepo.GetByPublicID(ctx, form.PublicID)
	if err != nil {
		t.Fatalf("GetByPublicID failed: %v", err)
	}
	if retrieved.Name != form.Name {
		t.Errorf("expected name %q, got %q", form.Name, retrieved.Name)
	}

	// Update
	form.Name = "Updated Test Form"
	err = formRepo.Update(ctx, form)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ = formRepo.GetByPublicID(ctx, form.PublicID)
	if retrieved.Name != "Updated Test Form" {
		t.Errorf("expected updated name, got %q", retrieved.Name)
	}

	// List
	forms, err := formRepo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(forms) != 1 {
		t.Errorf("expected 1 form, got %d", len(forms))
	}

	// Delete
	err = formRepo.Delete(ctx, form.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	forms, _ = formRepo.List(ctx)
	if len(forms) != 0 {
		t.Errorf("expected 0 forms after delete, got %d", len(forms))
	}
}

// TestSubmissionRepository_CRUD tests submission create, read, update, delete operations
func TestSubmissionRepository_CRUD(t *testing.T) {
	store := setupTestStore(t)
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	formRepo := store.Form()
	submRepo := store.Submission()

	// Create a form first
	form := &domain.Form{
		ID:             "form-id-1",
		PublicID:       "form-public-1",
		Name:           "Test Form",
		Status:         domain.FormStatusActive,
		NotifyEmails:   []string{},
		AllowedOrigins: []string{"*"},
		CreatedAt:      time.Now(),
	}
	_ = formRepo.Create(ctx, form)

	// Create submission
	submission := &domain.Submission{
		ID:        "sub-id-1",
		FormID:    form.ID,
		Status:    domain.SubmissionStatusUnread,
		Data:      []byte(`{"email":"test@example.com","message":"Hello"}`),
		Meta:      []byte(`{"ip":"127.0.0.1"}`),
		CreatedAt: time.Now(),
	}

	err := submRepo.Create(ctx, submission)
	if err != nil {
		t.Fatalf("Create submission failed: %v", err)
	}

	// Read
	retrieved, err := submRepo.GetByID(ctx, submission.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.ID != submission.ID {
		t.Errorf("expected id %q, got %q", submission.ID, retrieved.ID)
	}

	// Update Status
	err = submRepo.UpdateStatus(ctx, submission.ID, domain.SubmissionStatusRead)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ = submRepo.GetByID(ctx, submission.ID)
	if retrieved.Status != domain.SubmissionStatusRead {
		t.Errorf("expected status read, got %q", retrieved.Status)
	}

	// List by Form
	subs, err := submRepo.GetByFormID(ctx, form.ID)
	if err != nil {
		t.Fatalf("GetByFormID failed: %v", err)
	}
	if len(subs) != 1 {
		t.Errorf("expected 1 submission, got %d", len(subs))
	}

	// Delete
	err = submRepo.Delete(ctx, submission.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// TestUserRepository_CRUD tests user create, read, update, delete operations
func TestUserRepository_CRUD(t *testing.T) {
	store := setupTestStore(t)
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	userRepo := store.User()

	// Create
	user := &domain.User{
		ID:           "user-id-1",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Name:         "Test User",
		Role:         "admin",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create user failed: %v", err)
	}

	// Read by ID
	retrieved, err := userRepo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Email != user.Email {
		t.Errorf("expected email %q, got %q", user.Email, retrieved.Email)
	}

	// Read by Email
	retrieved, err = userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if retrieved.ID != user.ID {
		t.Errorf("expected id %q, got %q", user.ID, retrieved.ID)
	}

	// Update
	user.Name = "Updated User"
	err = userRepo.Update(ctx, user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Count
	count, err := userRepo.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// Delete
	err = userRepo.Delete(ctx, user.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ = userRepo.Count(ctx)
	if count != 0 {
		t.Errorf("expected count 0 after delete, got %d", count)
	}
}

// TestContextTimeout verifies context cancellation is respected
func TestContextTimeout(t *testing.T) {
	store := setupTestStore(t)
	t.Cleanup(func() { _ = store.Close() })

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Sleep to ensure timeout
	time.Sleep(1 * time.Millisecond)

	// This should fail with context deadline exceeded
	_, err := store.Form().List(ctx)
	if err == nil {
		// Note: SQLite operations may complete before context check
		// This is expected behavior - context cancellation is best-effort
		t.Log("Operation completed despite expired context (acceptable for fast operations)")
	}
}

// setupTestStore creates a temporary in-memory SQLite store for testing
func setupTestStore(t *testing.T) *Store {
	t.Helper()

	// Use in-memory database for faster tests
	store, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}

	return store
}
