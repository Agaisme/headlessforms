package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"

	"headless_form/internal/adapter/api/response"
	"headless_form/internal/adapter/middleware"
	"headless_form/internal/core/domain"
	"headless_form/internal/core/ports"
)

// SettingsHandler handles site settings API endpoints
type SettingsHandler struct {
	repo ports.Repository
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(repo ports.Repository) *SettingsHandler {
	return &SettingsHandler{repo: repo}
}

// HandleGetSettings returns site settings (super_admin only)
// GET /api/v1/settings
func (h *SettingsHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	// Verify super_admin role
	if !middleware.IsSuperAdmin(r.Context()) {
		response.Error(w, http.StatusForbidden, "Super admin access required", "FORBIDDEN")
		return
	}

	settings, err := h.repo.Settings().Get(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	// Return with masked password
	response.Success(w, settings.ToPublic())
}

// HandleUpdateSettings updates site settings (super_admin only)
// PUT /api/v1/settings
func (h *SettingsHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	// Verify super_admin role
	if !middleware.IsSuperAdmin(r.Context()) {
		response.Error(w, http.StatusForbidden, "Super admin access required", "FORBIDDEN")
		return
	}

	var req struct {
		SiteName     string `json:"site_name"`
		SiteURL      string `json:"site_url"`
		SMTPHost     string `json:"smtp_host"`
		SMTPPort     int    `json:"smtp_port"`
		SMTPUser     string `json:"smtp_user"`
		SMTPPassword string `json:"smtp_password"`
		SMTPFrom     string `json:"smtp_from"`
		SMTPFromName string `json:"smtp_from_name"`
		SMTPSecure   bool   `json:"smtp_secure"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body", "INVALID_BODY")
		return
	}

	// Build settings from request
	settings := &domain.SiteSettings{
		ID:           "default",
		SiteName:     req.SiteName,
		SiteURL:      req.SiteURL,
		SMTPHost:     req.SMTPHost,
		SMTPPort:     req.SMTPPort,
		SMTPUser:     req.SMTPUser,
		SMTPPassword: req.SMTPPassword,
		SMTPFrom:     req.SMTPFrom,
		SMTPFromName: req.SMTPFromName,
		SMTPSecure:   req.SMTPSecure,
		UpdatedBy:    middleware.GetUserID(r.Context()),
	}

	// If password is masked, don't update it
	if settings.SMTPPassword == "********" {
		settings.SMTPPassword = ""
	}

	if err := h.repo.Settings().Save(r.Context(), settings); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, settings.ToPublic())
}

// HandleTestSMTP tests SMTP connection (super_admin only)
// POST /api/v1/settings/test-smtp
func (h *SettingsHandler) HandleTestSMTP(w http.ResponseWriter, r *http.Request) {
	// Verify super_admin role
	if !middleware.IsSuperAdmin(r.Context()) {
		response.Error(w, http.StatusForbidden, "Super admin access required", "FORBIDDEN")
		return
	}

	var req struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		From     string `json:"from"`
		TestTo   string `json:"test_to"` // Email to send test to
		Secure   bool   `json:"secure"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body", "INVALID_BODY")
		return
	}

	if req.Host == "" || req.Port == 0 {
		response.BadRequest(w, "SMTP host and port are required", "MISSING_SMTP_CONFIG")
		return
	}

	if req.TestTo == "" {
		response.BadRequest(w, "Test email recipient is required", "MISSING_TEST_TO")
		return
	}

	// If password is masked, get from saved settings
	if req.Password == "********" || req.Password == "" {
		settings, _ := h.repo.Settings().Get(r.Context())
		if settings != nil {
			req.Password = settings.SMTPPassword
		}
	}

	// Build SMTP address
	addr := fmt.Sprintf("%s:%d", req.Host, req.Port)

	// Create test message
	from := req.From
	if from == "" {
		from = req.User
	}

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: Headless Forms - SMTP Test\r\n\r\nThis is a test email from Headless Forms.\n\nIf you received this, your SMTP configuration is working correctly!",
		from, req.TestTo,
	))

	// Attempt to send
	var auth smtp.Auth
	if req.User != "" && req.Password != "" {
		auth = smtp.PlainAuth("", req.User, req.Password, req.Host)
	}

	err := smtp.SendMail(addr, auth, from, []string{req.TestTo}, msg)
	if err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("SMTP test failed: %v", err), "SMTP_TEST_FAILED")
		return
	}

	response.Success(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Test email sent successfully to %s", req.TestTo),
	})
}
