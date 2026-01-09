package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"headless_form/internal/adapter/api/request"
	"headless_form/internal/adapter/api/response"
	"headless_form/internal/adapter/middleware"
	"headless_form/internal/core/domain"
)

// =============================================================================
// Submission Handlers
// =============================================================================

// HandleListSubmissions: GET /api/v1/forms/{form_id}/submissions?page=1&limit=50
func (h *Router) HandleListSubmissions(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("form_id")
	page := parseIntParam(r, "page", 1)
	limit := parseIntParam(r, "limit", 50)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	subms, total, err := h.submissionService.ListSubmissionsPaginated(r.Context(), publicID, page, limit)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"submissions": subms,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// HandleSubmit: POST /api/v1/submissions/{form_id}
// This is the Endpoint Form Submission URL - public access with form-level access control
func (h *Router) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("form_id")
	contentType := r.Header.Get("Content-Type")

	var data map[string]interface{}
	var clientMeta map[string]interface{}
	clientMeta = make(map[string]interface{})

	// 1. Parse Payload based on Content-Type
	if strings.Contains(contentType, "application/x-www-form-urlencoded") || strings.Contains(contentType, "multipart/form-data") {
		// Standard HTML Form
		if err := r.ParseForm(); err != nil {
			response.BadRequest(w, "Invalid form data", "INVALID_FORM")
			return
		}
		data = make(map[string]interface{})
		for k, v := range r.PostForm {
			if len(v) > 0 {
				data[k] = v[0] // Simplify to single value for standard fields
			}
		}
	} else {
		// Default to JSON (API/Fetch)
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.BadRequest(w, "Invalid JSON body", "INVALID_BODY")
			return
		}

		// Hybrid JSON Support (Flat vs Nested)
		if nestedData, ok := payload["data"].(map[string]interface{}); ok {
			data = nestedData
			if nestedMeta, ok := payload["meta"].(map[string]interface{}); ok {
				clientMeta = nestedMeta
			}
		} else {
			data = payload
		}
	}

	// 2. Collect server-side metadata (TRUSTED - auto-detected from request)
	serverMeta := request.GetServerMeta(r)

	// 3. Spam detection (using singleton detector for rate limiting state)
	spamScore := h.spamDetector.Analyze(serverMeta.IP, serverMeta.UserAgent, data, 0)
	h.spamDetector.RecordSubmission(serverMeta.IP) // Track for rate limiting

	// 4. Build combined meta with separated _server, _client, and _spam
	meta := map[string]interface{}{
		"_server": serverMeta, // Trusted server-collected data
		"_client": clientMeta, // Client-provided data (may be spoofed)
		"_spam":   spamScore,  // Spam detection result
	}

	// For private forms: inject auth user ID from context
	if userID := r.Context().Value("user_id"); userID != nil {
		meta["_auth_user_id"] = userID.(string)
	}

	// 5. Submit
	subm, err := h.submissionService.Submit(r.Context(), publicID, data, meta)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusBadRequest, err.Error(), "SUBMISSION_FAILED")
		return
	}

	// 6. Handle Response (Redirect vs JSON)
	form, _ := h.formService.GetForm(r.Context(), publicID)

	redirectURL := ""
	if q := r.URL.Query().Get("redirect_to"); q != "" {
		redirectURL = q
	} else if form != nil && form.RedirectURL != "" {
		redirectURL = form.RedirectURL
	}

	// Only redirect if likely initiated by browser form (HTML content type)
	isHTMLForm := strings.Contains(contentType, "application/x-www-form-urlencoded") || strings.Contains(contentType, "multipart/form-data")

	if redirectURL != "" && isHTMLForm {
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	response.Created(w, subm)
}

// HandleGetSubmission: GET /api/v1/submissions/{sub_id}
func (h *Router) HandleGetSubmission(w http.ResponseWriter, r *http.Request) {
	subID := r.PathValue("sub_id")

	// Get submission
	sub, err := h.submissionService.GetSubmission(r.Context(), subID)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}

	// Verify ownership through form
	form, err := h.formService.GetFormByID(r.Context(), sub.FormID)
	if err != nil || form == nil {
		response.NotFound(w, "Associated form not found")
		return
	}

	if !middleware.CanAccessForm(r.Context(), form.OwnerID) {
		response.Error(w, http.StatusForbidden, "Access denied", "FORBIDDEN")
		return
	}

	response.Success(w, sub)
}

// verifySubmissionOwnership checks if the current user can access a submission
func (h *Router) verifySubmissionOwnership(r *http.Request, subID string) (*domain.Submission, error) {
	sub, err := h.submissionService.GetSubmission(r.Context(), subID)
	if err != nil {
		return nil, err
	}

	form, err := h.formService.GetFormByID(r.Context(), sub.FormID)
	if err != nil || form == nil {
		return nil, domain.ErrFormNotFound
	}

	if !middleware.CanAccessForm(r.Context(), form.OwnerID) {
		return nil, fmt.Errorf("access denied")
	}

	return sub, nil
}

// HandleMarkAsRead: PUT /api/v1/submissions/{sub_id}/read
func (h *Router) HandleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	subID := r.PathValue("sub_id")

	// Verify ownership before marking as read
	if _, err := h.verifySubmissionOwnership(r, subID); err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusForbidden, "Access denied", "FORBIDDEN")
		return
	}

	if err := h.submissionService.MarkAsRead(r.Context(), subID); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Marked as read"})
}

// HandleMarkAsUnread: PUT /api/v1/submissions/{sub_id}/unread
func (h *Router) HandleMarkAsUnread(w http.ResponseWriter, r *http.Request) {
	subID := r.PathValue("sub_id")

	// Verify ownership before marking as unread
	if _, err := h.verifySubmissionOwnership(r, subID); err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusForbidden, "Access denied", "FORBIDDEN")
		return
	}

	if err := h.submissionService.MarkAsUnread(r.Context(), subID); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Marked as unread"})
}

// HandleDeleteSubmission: DELETE /api/v1/submissions/{sub_id}
func (h *Router) HandleDeleteSubmission(w http.ResponseWriter, r *http.Request) {
	subID := r.PathValue("sub_id")

	// Verify ownership before deleting
	if _, err := h.verifySubmissionOwnership(r, subID); err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusForbidden, "Access denied", "FORBIDDEN")
		return
	}

	if err := h.submissionService.DeleteSubmission(r.Context(), subID); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Submission deleted successfully"})
}
