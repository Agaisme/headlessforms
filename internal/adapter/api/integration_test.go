// Package api_test provides end-to-end HTTP integration tests for the API
//
//nolint:bodyclose,errcheck // Test file - response bodies are handled by ParseResponse helper
package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"headless_form/internal/adapter/api"
	"headless_form/internal/adapter/storage/sqlite"
	"headless_form/internal/core/service"
)

// TestServer encapsulates the test server setup
type TestServer struct {
	Server *httptest.Server
	Store  *sqlite.Store
	Token  string // JWT token for authenticated requests
	Mux    *http.ServeMux
}

// NewTestServer creates a new test server with in-memory database
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Create in-memory SQLite store
	store, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create services
	formService := service.NewFormService(store)
	submService := service.NewSubmissionService(store)
	statsService := service.NewStatsService(store)

	// Create router and mux
	router := api.NewRouter(formService, submService, statsService)
	mux := http.NewServeMux()

	// Register public routes (no auth for basic tests)
	router.RegisterPublicRoutes(mux, func(h http.Handler) http.Handler { return h })

	// For testing, register protected routes without auth middleware
	router.RegisterProtectedRoutes(mux, func(h http.Handler) http.Handler { return h })

	server := httptest.NewServer(mux)

	return &TestServer{
		Server: server,
		Store:  store,
		Mux:    mux,
	}
}

// Close cleans up the test server
func (ts *TestServer) Close() {
	ts.Server.Close()
	_ = ts.Store.Close()
}

// Request makes an HTTP request to the test server
func (ts *TestServer) Request(t *testing.T, method, path string, body interface{}) *http.Response {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, ts.Server.URL+path, reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if ts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+ts.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return resp
}

// ParseResponse parses a JSON response
func ParseResponse(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
}

// =============================================================================
// Health Check Tests
// =============================================================================

func TestHealthCheck(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	resp := ts.Request(t, "GET", "/api/health", nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	ParseResponse(t, resp, &result)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	if data["status"] != "healthy" {
		t.Errorf("expected healthy status, got %v", data["status"])
	}
}

// =============================================================================
// Form CRUD Tests
// =============================================================================

func TestCreateForm(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	body := map[string]interface{}{
		"name":          "Test Form",
		"redirect_url":  "https://example.com/thanks",
		"notify_emails": []string{"test@example.com"},
		"access_mode":   "public",
	}

	resp := ts.Request(t, "POST", "/api/v1/forms", body)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	ParseResponse(t, resp, &result)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	if data["name"] != "Test Form" {
		t.Errorf("expected name 'Test Form', got %v", data["name"])
	}

	if data["public_id"] == nil || data["public_id"] == "" {
		t.Error("expected public_id to be set")
	}
}

func TestListForms(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Create a form first
	ts.Request(t, "POST", "/api/v1/forms", map[string]interface{}{
		"name": "List Test Form",
	})

	// List forms
	resp := ts.Request(t, "GET", "/api/v1/forms", nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	ParseResponse(t, resp, &result)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object")
	}

	forms, ok := data["forms"].([]interface{})
	if !ok {
		t.Fatal("expected forms array")
	}

	if len(forms) == 0 {
		t.Error("expected at least one form")
	}
}

func TestGetForm(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Create a form
	createResp := ts.Request(t, "POST", "/api/v1/forms", map[string]interface{}{
		"name": "Get Test Form",
	})

	var createResult map[string]interface{}
	ParseResponse(t, createResp, &createResult)

	data := createResult["data"].(map[string]interface{})
	publicID := data["public_id"].(string)

	// Get the form
	resp := ts.Request(t, "GET", "/api/v1/forms/"+publicID, nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUpdateForm(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Create a form
	createResp := ts.Request(t, "POST", "/api/v1/forms", map[string]interface{}{
		"name": "Original Name",
	})

	var createResult map[string]interface{}
	ParseResponse(t, createResp, &createResult)

	data := createResult["data"].(map[string]interface{})
	publicID := data["public_id"].(string)

	// Update the form
	updateResp := ts.Request(t, "PUT", "/api/v1/forms/"+publicID, map[string]interface{}{
		"name":   "Updated Name",
		"status": "active",
	})

	if updateResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", updateResp.StatusCode)
	}

	var updateResult map[string]interface{}
	ParseResponse(t, updateResp, &updateResult)

	updatedData := updateResult["data"].(map[string]interface{})
	if updatedData["name"] != "Updated Name" {
		t.Errorf("expected 'Updated Name', got %v", updatedData["name"])
	}
}

func TestDeleteForm(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Create a form
	createResp := ts.Request(t, "POST", "/api/v1/forms", map[string]interface{}{
		"name": "Delete Test Form",
	})

	var createResult map[string]interface{}
	ParseResponse(t, createResp, &createResult)

	data := createResult["data"].(map[string]interface{})
	publicID := data["public_id"].(string)

	// Delete the form
	deleteResp := ts.Request(t, "DELETE", "/api/v1/forms/"+publicID, nil)

	if deleteResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", deleteResp.StatusCode)
	}

	// Verify it's deleted
	getResp := ts.Request(t, "GET", "/api/v1/forms/"+publicID, nil)
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getResp.StatusCode)
	}
}

// =============================================================================
// Submission Tests
// =============================================================================

func TestSubmitForm(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Create a form first
	createResp := ts.Request(t, "POST", "/api/v1/forms", map[string]interface{}{
		"name":        "Submission Test Form",
		"access_mode": "public",
	})

	var createResult map[string]interface{}
	ParseResponse(t, createResp, &createResult)

	data := createResult["data"].(map[string]interface{})
	publicID := data["public_id"].(string)

	// Submit to the form
	submitResp := ts.Request(t, "POST", "/api/v1/submissions/"+publicID, map[string]interface{}{
		"name":    "John Doe",
		"email":   "john@example.com",
		"message": "Hello, this is a test submission",
	})

	if submitResp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", submitResp.StatusCode)
	}

	var submitResult map[string]interface{}
	ParseResponse(t, submitResp, &submitResult)

	subData := submitResult["data"].(map[string]interface{})
	if subData["id"] == nil || subData["id"] == "" {
		t.Error("expected submission id to be set")
	}
}

func TestListSubmissions(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Create a form
	createResp := ts.Request(t, "POST", "/api/v1/forms", map[string]interface{}{
		"name": "List Submissions Form",
	})

	var createResult map[string]interface{}
	ParseResponse(t, createResp, &createResult)

	data := createResult["data"].(map[string]interface{})
	publicID := data["public_id"].(string)

	// Create some submissions
	for i := 0; i < 3; i++ {
		ts.Request(t, "POST", "/api/v1/submissions/"+publicID, map[string]interface{}{
			"name": "User " + string(rune('A'+i)),
		})
	}

	// List submissions
	listResp := ts.Request(t, "GET", "/api/v1/forms/"+publicID+"/submissions", nil)

	if listResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", listResp.StatusCode)
	}

	var listResult map[string]interface{}
	ParseResponse(t, listResp, &listResult)

	listData := listResult["data"].(map[string]interface{})
	submissions := listData["submissions"].([]interface{})

	if len(submissions) != 3 {
		t.Errorf("expected 3 submissions, got %d", len(submissions))
	}
}

// =============================================================================
// Stats Tests
// =============================================================================

func TestDashboardStats(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Create some data
	createResp := ts.Request(t, "POST", "/api/v1/forms", map[string]interface{}{
		"name": "Stats Test Form",
	})

	var createResult map[string]interface{}
	ParseResponse(t, createResp, &createResult)

	data := createResult["data"].(map[string]interface{})
	publicID := data["public_id"].(string)

	ts.Request(t, "POST", "/api/v1/submissions/"+publicID, map[string]interface{}{
		"name": "Test User",
	})

	// Get stats
	resp := ts.Request(t, "GET", "/api/v1/stats", nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	ParseResponse(t, resp, &result)

	statsData := result["data"].(map[string]interface{})
	if statsData["total_forms"] == nil {
		t.Error("expected total_forms in stats")
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestFormNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	resp := ts.Request(t, "GET", "/api/v1/forms/nonexistent-id", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestInvalidFormCreate(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Empty name should fail validation
	resp := ts.Request(t, "POST", "/api/v1/forms", map[string]interface{}{
		"name": "", // Empty name
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSubmitToNonexistentForm(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	resp := ts.Request(t, "POST", "/api/v1/submissions/nonexistent-form", map[string]interface{}{
		"name": "Test",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// =============================================================================
// Pagination Tests
// =============================================================================

func TestFormsPagination(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Create 5 forms
	for i := 0; i < 5; i++ {
		ts.Request(t, "POST", "/api/v1/forms", map[string]interface{}{
			"name": "Pagination Form " + string(rune('A'+i)),
		})
	}

	// Get page 1 with limit 2
	resp := ts.Request(t, "GET", "/api/v1/forms?page=1&limit=2", nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	ParseResponse(t, resp, &result)

	data := result["data"].(map[string]interface{})
	forms := data["forms"].([]interface{})
	pagination := data["pagination"].(map[string]interface{})

	if len(forms) != 2 {
		t.Errorf("expected 2 forms, got %d", len(forms))
	}

	if pagination["total"].(float64) != 5 {
		t.Errorf("expected total 5, got %v", pagination["total"])
	}

	if pagination["total_pages"].(float64) != 3 {
		t.Errorf("expected 3 total pages, got %v", pagination["total_pages"])
	}
}
