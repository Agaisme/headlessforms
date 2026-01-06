package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"headless_form/internal/adapter/api/response"
	"headless_form/internal/adapter/middleware"
	"headless_form/internal/core/domain"
	"headless_form/internal/core/service"
)

// parseIntParam extracts an integer query parameter with a default value
func parseIntParam(r *http.Request, name string, defaultVal int) int {
	str := r.URL.Query().Get(name)
	if str == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultVal
	}
	return val
}

type Router struct {
	formService       *service.FormService
	submissionService *service.SubmissionService
	statsService      *service.StatsService
}

func NewRouter(formService *service.FormService, submService *service.SubmissionService, statsService *service.StatsService) *Router {
	return &Router{
		formService:       formService,
		submissionService: submService,
		statsService:      statsService,
	}
}

// RegisterPublicRoutes registers routes that don't require authentication
// These are endpoints that external users/forms can access
// optionalAuth is used to extract user context if present (for private forms)
func (h *Router) RegisterPublicRoutes(mux *http.ServeMux, optionalAuth func(http.Handler) http.Handler) {
	// Health check - always public
	mux.HandleFunc("GET /api/health", h.HandleHealthCheck)

	// Endpoint Form Submission URL - public by default (access control handled in handler)
	// Uses optional auth to extract user context for private forms
	mux.Handle("POST /api/v1/submissions/{form_id}", optionalAuth(http.HandlerFunc(h.HandleSubmit)))
}

// RegisterProtectedRoutes registers routes that require JWT authentication
// All dashboard management operations require auth
func (h *Router) RegisterProtectedRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	// Stats (protected)
	mux.Handle("GET /api/v1/stats", authMiddleware(http.HandlerFunc(h.HandleDashboardStats)))

	// Forms CRUD (protected)
	mux.Handle("POST /api/v1/forms", authMiddleware(http.HandlerFunc(h.HandleCreateForm)))
	mux.Handle("GET /api/v1/forms", authMiddleware(http.HandlerFunc(h.HandleListForms)))
	mux.Handle("GET /api/v1/forms/{form_id}", authMiddleware(http.HandlerFunc(h.HandleGetForm)))
	mux.Handle("PUT /api/v1/forms/{form_id}", authMiddleware(http.HandlerFunc(h.HandleUpdateForm)))
	mux.Handle("DELETE /api/v1/forms/{form_id}", authMiddleware(http.HandlerFunc(h.HandleDeleteForm)))
	mux.Handle("GET /api/v1/forms/{form_id}/stats", authMiddleware(http.HandlerFunc(h.HandleFormStats)))

	// Submission management (protected) - viewing/managing submissions requires auth
	mux.Handle("GET /api/v1/forms/{form_id}/submissions", authMiddleware(http.HandlerFunc(h.HandleListSubmissions)))
	mux.Handle("GET /api/v1/forms/{form_id}/export/csv", authMiddleware(http.HandlerFunc(h.HandleExportCSV)))
	mux.Handle("PUT /api/v1/submissions/{sub_id}/read", authMiddleware(http.HandlerFunc(h.HandleMarkAsRead)))
	mux.Handle("PUT /api/v1/submissions/{sub_id}/unread", authMiddleware(http.HandlerFunc(h.HandleMarkAsUnread)))
	mux.Handle("DELETE /api/v1/submissions/{sub_id}", authMiddleware(http.HandlerFunc(h.HandleDeleteSubmission)))

	// Admin / Testing (protected)
	mux.Handle("POST /api/v1/admin/seed", authMiddleware(http.HandlerFunc(h.HandleSeed)))
}

// HandleHealthCheck: GET /api/health
func (h *Router) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

// HandleDashboardStats: GET /api/v1/stats
func (h *Router) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsService.GetDashboardStats(r.Context())
	if response.HandleError(w, err) {
		return
	}
	response.Success(w, stats)
}

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
		if err == domain.ErrFormNotFound {
			response.NotFound(w, "Form not found")
		} else {
			response.HandleError(w, err)
		}
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
		if err == domain.ErrFormNotFound {
			response.NotFound(w, "Form not found")
		} else {
			response.HandleError(w, err)
		}
		return
	}
	response.Success(w, stats)
}

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
		if err == domain.ErrFormNotFound {
			response.NotFound(w, "Form not found")
		} else {
			response.HandleError(w, err)
		}
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
		if err == domain.ErrFormNameRequired || err == domain.ErrFormNameTooLong {
			response.BadRequest(w, err.Error(), "VALIDATION_ERROR")
		} else {
			response.HandleError(w, err)
		}
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
		if err == domain.ErrFormNotFound {
			response.NotFound(w, "Form not found")
		} else {
			response.HandleError(w, err)
		}
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
		if err == domain.ErrFormNotFound {
			response.NotFound(w, "Form not found")
		} else if err == domain.ErrFormNameRequired || err == domain.ErrFormNameTooLong {
			response.BadRequest(w, err.Error(), "VALIDATION_ERROR")
		} else {
			response.HandleError(w, err)
		}
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
		if err == domain.ErrFormNotFound {
			response.NotFound(w, "Form not found")
		} else {
			response.HandleError(w, err)
		}
		return
	}
	if !middleware.CanAccessForm(r.Context(), form.OwnerID) {
		response.Error(w, http.StatusForbidden, "You can only delete your own forms", "FORBIDDEN")
		return
	}

	if err := h.formService.DeleteForm(r.Context(), publicID); err != nil {
		if err == domain.ErrFormNotFound {
			response.NotFound(w, "Form not found")
		} else {
			response.HandleError(w, err)
		}
		return
	}

	response.Success(w, map[string]string{"message": "Form deleted successfully"})
}

// HandleSubmit: POST /api/v1/submissions/{form_id}
// This is the Endpoint Form Submission URL - public access with form-level access control
func (h *Router) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("form_id")

	var req struct {
		Data map[string]interface{} `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body", "INVALID_BODY")
		return
	}

	// Initialize meta if nil
	if req.Meta == nil {
		req.Meta = make(map[string]interface{})
	}

	// For private forms: inject auth user ID from context (if present)
	// This allows OptionalAuthMiddleware to pass user context to submission service
	if userID := r.Context().Value("user_id"); userID != nil {
		req.Meta["_auth_user_id"] = userID.(string)
	}

	subm, err := h.submissionService.Submit(r.Context(), publicID, req.Data, req.Meta)
	if err != nil {
		switch err {
		case domain.ErrFormNotFound:
			response.NotFound(w, "Form not found")
		case domain.ErrInvalidSubmissionKey:
			response.Error(w, http.StatusForbidden, "Invalid or missing submission key", "INVALID_KEY")
		case domain.ErrAuthRequired:
			response.Error(w, http.StatusUnauthorized, "Authentication required for this form", "AUTH_REQUIRED")
		default:
			response.Error(w, http.StatusBadRequest, err.Error(), "SUBMISSION_FAILED")
		}
		return
	}

	response.Created(w, subm)
}

// HandleMarkAsRead: PUT /api/v1/submissions/{sub_id}/read
func (h *Router) HandleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	subID := r.PathValue("sub_id")

	if err := h.submissionService.MarkAsRead(r.Context(), subID); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Marked as read"})
}

// HandleMarkAsUnread: PUT /api/v1/submissions/{sub_id}/unread
func (h *Router) HandleMarkAsUnread(w http.ResponseWriter, r *http.Request) {
	subID := r.PathValue("sub_id")

	if err := h.submissionService.MarkAsUnread(r.Context(), subID); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Marked as unread"})
}

// HandleDeleteSubmission: DELETE /api/v1/submissions/{sub_id}
func (h *Router) HandleDeleteSubmission(w http.ResponseWriter, r *http.Request) {
	subID := r.PathValue("sub_id")

	if err := h.submissionService.DeleteSubmission(r.Context(), subID); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Submission deleted successfully"})
}

// HandleSeed: POST /api/v1/admin/seed
// Creates test data for performance testing
func (h *Router) HandleSeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Forms              int `json:"forms"`
		SubmissionsPerForm int `json:"submissions_per_form"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Forms = 1000
		req.SubmissionsPerForm = 100
	}

	if req.Forms <= 0 {
		req.Forms = 1000
	}
	if req.SubmissionsPerForm <= 0 {
		req.SubmissionsPerForm = 100
	}

	// Cap to prevent excessive load
	if req.Forms > 10000 {
		req.Forms = 10000
	}
	if req.SubmissionsPerForm > 1000 {
		req.SubmissionsPerForm = 1000
	}

	ctx := r.Context()
	formsCreated := 0
	submissionsCreated := 0

	for i := 0; i < req.Forms; i++ {
		// Create form (seeded forms have no specific owner)
		ownerID := middleware.GetUserID(ctx) // Use authenticated user or empty
		form, err := h.formService.CreateForm(ctx,
			"Test Form "+string(rune('A'+i%26))+"-"+itoa(i+1),
			"",
			nil,
			"", // webhook_url
			"", // webhook_secret
			ownerID,
			"public", // accessMode
			"",       // submissionKey
		)
		if err != nil {
			continue
		}
		formsCreated++

		// Create submissions for this form
		for j := 0; j < req.SubmissionsPerForm; j++ {
			data := map[string]interface{}{
				"name":    "User " + itoa(j+1),
				"email":   "user" + itoa(j+1) + "@example.com",
				"message": "Test message from user " + itoa(j+1) + " on form " + itoa(i+1),
				"phone":   "555-" + itoa(1000+j),
			}
			meta := map[string]interface{}{
				"source":    "seed",
				"userAgent": "SeedBot/1.0",
			}

			_, err := h.submissionService.Submit(ctx, form.PublicID, data, meta)
			if err == nil {
				submissionsCreated++
			}
		}
	}

	response.Success(w, map[string]interface{}{
		"message":             "Seeding complete",
		"forms_created":       formsCreated,
		"submissions_created": submissionsCreated,
	})
}

// itoa is a simple int to string helper
func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	var s string
	negative := i < 0
	if negative {
		i = -i
	}

	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}

	if negative {
		s = "-" + s
	}
	return s
}

// HandleExportCSV: GET /api/v1/forms/{form_id}/export/csv
func (h *Router) HandleExportCSV(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("form_id")

	// Get form to verify it exists and get name
	form, err := h.formService.GetForm(r.Context(), publicID)
	if err != nil {
		if err == domain.ErrFormNotFound {
			response.NotFound(w, "Form not found")
		} else {
			response.HandleError(w, err)
		}
		return
	}

	// Get all submissions (no pagination for export)
	submissions, err := h.submissionService.ListSubmissions(r.Context(), publicID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	// Collect all unique field keys
	fieldSet := make(map[string]bool)
	var allData []map[string]interface{}
	for _, sub := range submissions {
		var data map[string]interface{}
		if err := json.Unmarshal(sub.Data, &data); err == nil {
			for key := range data {
				fieldSet[key] = true
			}
			allData = append(allData, data)
		}
	}

	// Convert to sorted slice
	var fields []string
	for key := range fieldSet {
		fields = append(fields, key)
	}

	// Build CSV content
	var csv string

	// Header row: id, created_at, status, + dynamic fields
	csv = "id,created_at,status"
	for _, f := range fields {
		csv += "," + escapeCSV(f)
	}
	csv += "\n"

	// Data rows
	for i, sub := range submissions {
		csv += escapeCSV(sub.ID) + ","
		csv += escapeCSV(sub.CreatedAt.Format("2006-01-02 15:04:05")) + ","
		csv += escapeCSV(string(sub.Status))

		data := allData[i]
		for _, f := range fields {
			val := ""
			if v, ok := data[f]; ok {
				switch t := v.(type) {
				case string:
					val = t
				case float64:
					val = strconv.FormatFloat(t, 'f', -1, 64)
				case bool:
					val = strconv.FormatBool(t)
				default:
					// JSON encode complex types
					if b, err := json.Marshal(v); err == nil {
						val = string(b)
					}
				}
			}
			csv += "," + escapeCSV(val)
		}
		csv += "\n"
	}

	// Set headers for file download
	filename := form.Name + "_submissions.csv"
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Write([]byte(csv))
}

// escapeCSV escapes a value for CSV format
func escapeCSV(s string) string {
	needsQuote := false
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			needsQuote = true
			break
		}
	}
	if !needsQuote {
		return s
	}
	// Escape quotes by doubling them
	escaped := ""
	for _, c := range s {
		if c == '"' {
			escaped += "\"\""
		} else {
			escaped += string(c)
		}
	}
	return "\"" + escaped + "\""
}
