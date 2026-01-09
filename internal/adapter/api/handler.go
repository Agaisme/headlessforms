package api

import (
	"net/http"
	"strconv"

	"headless_form/internal/adapter/api/response"
	"headless_form/internal/adapter/spam"
	"headless_form/internal/core/service"
)

// =============================================================================
// Router - Core API Handler
// =============================================================================

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

// Router is the main API handler that routes requests to appropriate handlers
type Router struct {
	formService       *service.FormService
	submissionService *service.SubmissionService
	statsService      *service.StatsService
	spamDetector      *spam.Detector
}

// NewRouter creates a new Router with the given services
func NewRouter(formService *service.FormService, submService *service.SubmissionService, statsService *service.StatsService) *Router {
	return &Router{
		formService:       formService,
		submissionService: submService,
		statsService:      statsService,
		spamDetector:      spam.NewDetector(spam.DefaultConfig()),
	}
}

// =============================================================================
// Route Registration
// =============================================================================

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
	mux.Handle("GET /api/v1/submissions/{sub_id}", authMiddleware(http.HandlerFunc(h.HandleGetSubmission)))
	mux.Handle("PUT /api/v1/submissions/{sub_id}/read", authMiddleware(http.HandlerFunc(h.HandleMarkAsRead)))
	mux.Handle("PUT /api/v1/submissions/{sub_id}/unread", authMiddleware(http.HandlerFunc(h.HandleMarkAsUnread)))
	mux.Handle("DELETE /api/v1/submissions/{sub_id}", authMiddleware(http.HandlerFunc(h.HandleDeleteSubmission)))

	// Admin / Testing (protected)
	mux.Handle("POST /api/v1/admin/seed", authMiddleware(http.HandlerFunc(h.HandleSeed)))
}

// =============================================================================
// Base Handlers
// =============================================================================

// HandleHealthCheck: GET /api/health
// Returns health status with database connectivity check
func (h *Router) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	checks := make(map[string]interface{})

	// Check database connection by fetching dashboard stats
	_, err := h.statsService.GetDashboardStats(r.Context())
	if err != nil {
		checks["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		status = "degraded"
	} else {
		checks["database"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	response.Success(w, map[string]interface{}{
		"status":  status,
		"version": "1.2.0",
		"checks":  checks,
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
