package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"headless_form/internal/adapter/api/response"
	"headless_form/internal/adapter/middleware"
	"headless_form/internal/core/domain"
)

// =============================================================================
// Admin Handlers (Seed, Export)
// =============================================================================

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
			"Test Form "+string(rune('A'+i%26))+"-"+strconv.Itoa(i+1),
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
				"name":    "User " + strconv.Itoa(j+1),
				"email":   "user" + strconv.Itoa(j+1) + "@example.com",
				"message": "Test message from user " + strconv.Itoa(j+1) + " on form " + strconv.Itoa(i+1),
				"phone":   "555-" + strconv.Itoa(1000+j),
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

// HandleExportCSV: GET /api/v1/forms/{form_id}/export/csv
func (h *Router) HandleExportCSV(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("form_id")

	// Get form to verify it exists and get name
	form, err := h.formService.GetForm(r.Context(), publicID)
	if err != nil {
		if response.HandleDomainError(w, err) {
			return
		}
		response.HandleError(w, err)
		return
	}

	// Get all submissions (no pagination for export)
	submissions, err := h.submissionService.ListSubmissions(r.Context(), publicID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	// Collect all unique field keys and metadata
	fieldSet := make(map[string]bool)
	var allData []map[string]interface{}
	var allMeta []map[string]interface{}
	for _, sub := range submissions {
		var data map[string]interface{}
		if err := json.Unmarshal(sub.Data, &data); err == nil {
			for key := range data {
				fieldSet[key] = true
			}
			allData = append(allData, data)
		} else {
			allData = append(allData, nil)
		}

		// Parse meta for IP, country, spam
		var meta map[string]interface{}
		if err := json.Unmarshal(sub.Meta, &meta); err == nil {
			allMeta = append(allMeta, meta)
		} else {
			allMeta = append(allMeta, nil)
		}
	}

	// Convert to sorted slice
	var fields []string
	for key := range fieldSet {
		fields = append(fields, key)
	}

	// Build CSV content
	csv := buildCSVContent(submissions, allData, allMeta, fields)

	// Set headers for file download
	filename := form.Name + "_submissions.csv"
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	if _, err := w.Write([]byte(csv)); err != nil {
		// Log but don't return error - headers already sent
		log.Printf("[ERROR] Failed to write CSV response: %v", err)
	}
}

// buildCSVContent creates CSV string from submissions data
func buildCSVContent(submissions []*domain.Submission, allData, allMeta []map[string]interface{}, fields []string) string {
	var csv string

	// Header row: id, created_at, status, metadata columns, + dynamic fields
	csv = "id,created_at,status,ip,country,spam_score,is_spam"
	for _, f := range fields {
		csv += "," + escapeCSV(f)
	}
	csv += "\n"

	// Data rows
	for i, sub := range submissions {
		csv += escapeCSV(sub.ID) + ","
		csv += escapeCSV(sub.CreatedAt.Format("2006-01-02 15:04:05")) + ","
		csv += escapeCSV(string(sub.Status)) + ","

		// Extract metadata
		ip, country, spamScore, isSpam := extractMetadata(allMeta[i])
		csv += escapeCSV(ip) + ","
		csv += escapeCSV(country) + ","
		csv += escapeCSV(spamScore) + ","
		csv += escapeCSV(isSpam)

		// Dynamic fields
		data := allData[i]
		for _, f := range fields {
			csv += "," + escapeCSV(formatFieldValue(data, f))
		}
		csv += "\n"
	}

	return csv
}

// extractMetadata gets IP, country, and spam info from meta
func extractMetadata(meta map[string]interface{}) (ip, country, spamScore, isSpam string) {
	if meta == nil {
		return "", "", "", ""
	}

	if serverMeta, ok := meta["_server"].(map[string]interface{}); ok {
		if v, ok := serverMeta["ip"].(string); ok {
			ip = v
		}
		if v, ok := serverMeta["country"].(string); ok {
			country = v
		}
	}
	if spamMeta, ok := meta["_spam"].(map[string]interface{}); ok {
		if v, ok := spamMeta["score"].(float64); ok {
			spamScore = strconv.FormatFloat(v, 'f', 0, 64)
		}
		if v, ok := spamMeta["is_spam"].(bool); ok {
			isSpam = strconv.FormatBool(v)
		}
	}
	return
}

// formatFieldValue formats a field value for CSV output
func formatFieldValue(data map[string]interface{}, field string) string {
	if data == nil {
		return ""
	}
	v, ok := data[field]
	if !ok {
		return ""
	}

	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		// JSON encode complex types
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return ""
	}
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
