package api

import (
	"encoding/json"
	"net/http"

	"headless_form/internal/adapter/api/response"
	"headless_form/internal/adapter/middleware"
	"headless_form/internal/core/domain"
)

// =============================================================================
// Form CRUD Handlers
// =============================================================================

// HandleListForms: GET /api/v1/forms?page=1&limit=20
func (h *Router) HandleListForms(w http.ResponseWriter, r *http.Request) {
	page := parseIntParam(r, "page", 1)
	limit := parseIntParam(r, "limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var forms []*domain.Form
	var total int
	var err error

	// Check user role - admin/super_admin see all forms, users see only their own
	if middleware.IsAdmin(r.Context()) {
		forms, total, err = h.formService.ListFormsPaginated(r.Context(), page, limit)
	} else {
		ownerID := middleware.GetUserID(r.Context())
		forms, total, err = h.formService.ListFormsByOwnerPaginated(r.Context(), ownerID, page, limit)
	}

	if response.HandleError(w, err) {
		return
	}

	response.Success(w, map[string]interface{}{
		"forms": forms,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// HandleGetForm: GET /api/v1/forms/{form_id}
func (h *Router) HandleGetForm(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("form_id")
	form, err := h.formService.GetForm(r.Context(), publicID)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}

	// Check if user can access this form
	if !middleware.CanAccessForm(r.Context(), form.OwnerID) {
		response.Error(w, http.StatusForbidden, "Access denied", "FORBIDDEN")
		return
	}

	response.Success(w, form)
}

// HandleFormStats: GET /api/v1/forms/{form_id}/stats
func (h *Router) HandleFormStats(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("form_id")
	stats, err := h.statsService.GetFormStats(r.Context(), publicID)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}
	response.Success(w, stats)
}

// HandleCreateForm: POST /api/v1/forms
func (h *Router) HandleCreateForm(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string   `json:"name"`
		RedirectURL   string   `json:"redirect_url"`
		NotifyEmails  []string `json:"notify_emails"`
		WebhookURL    string   `json:"webhook_url"`
		WebhookSecret string   `json:"webhook_secret"`
		AccessMode    string   `json:"access_mode"`
		SubmissionKey string   `json:"submission_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body", "INVALID_BODY")
		return
	}

	// Get authenticated user ID for form ownership
	ownerID := middleware.GetUserID(r.Context())

	form, err := h.formService.CreateForm(r.Context(), req.Name, req.RedirectURL, req.NotifyEmails, req.WebhookURL, req.WebhookSecret, ownerID, req.AccessMode, req.SubmissionKey)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}

	response.Created(w, form)
}

// HandleUpdateForm: PUT /api/v1/forms/{form_id}
func (h *Router) HandleUpdateForm(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("form_id")

	// Check ownership - users can only edit their own forms
	form, err := h.formService.GetForm(r.Context(), publicID)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}
	if !middleware.CanAccessForm(r.Context(), form.OwnerID) {
		response.Error(w, http.StatusForbidden, "You can only edit your own forms", "FORBIDDEN")
		return
	}

	var req struct {
		Name          string   `json:"name"`
		RedirectURL   string   `json:"redirect_url"`
		NotifyEmails  []string `json:"notify_emails"`
		Status        string   `json:"status"`
		WebhookURL    string   `json:"webhook_url"`
		WebhookSecret string   `json:"webhook_secret"`
		AccessMode    string   `json:"access_mode"`
		SubmissionKey string   `json:"submission_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body", "INVALID_BODY")
		return
	}

	status := domain.FormStatusActive
	if req.Status == "inactive" {
		status = domain.FormStatusInactive
	}

	updatedForm, err := h.formService.UpdateForm(r.Context(), publicID, req.Name, req.RedirectURL, req.NotifyEmails, status, req.WebhookURL, req.WebhookSecret, req.AccessMode, req.SubmissionKey)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}

	response.Success(w, updatedForm)
}

// HandleDeleteForm: DELETE /api/v1/forms/{form_id}
func (h *Router) HandleDeleteForm(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("form_id")

	// Check ownership - users can only delete their own forms
	form, err := h.formService.GetForm(r.Context(), publicID)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}
	if !middleware.CanAccessForm(r.Context(), form.OwnerID) {
		response.Error(w, http.StatusForbidden, "You can only delete your own forms", "FORBIDDEN")
		return
	}

	if err := h.formService.DeleteForm(r.Context(), publicID); err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Form deleted successfully"})
}
