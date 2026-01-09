package api

import (
	"net/http"
	"os"
	"path/filepath"
)

// OpenAPIHandler serves the OpenAPI specification and Swagger UI
type OpenAPIHandler struct {
	specPath string
}

// NewOpenAPIHandler creates a new OpenAPI handler
func NewOpenAPIHandler(specPath string) *OpenAPIHandler {
	return &OpenAPIHandler{specPath: specPath}
}

// ServeSpec serves the OpenAPI YAML specification
func (h *OpenAPIHandler) ServeSpec(w http.ResponseWriter, r *http.Request) {
	// Try to read from file first (development)
	data, err := os.ReadFile(h.specPath)
	if err != nil {
		// Fall back to default response
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "OpenAPI spec not found"}`))
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write(data)
}

// ServeSwaggerUI serves a Swagger UI page
func (h *OpenAPIHandler) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>HeadlessForms API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        html { box-sizing: border-box; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; padding: 0; background: #fafafa; }
        .swagger-ui .topbar { display: none; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/api/docs/openapi.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                docExpansion: "list",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// RegisterDocsRoutes registers documentation routes
func (h *OpenAPIHandler) RegisterDocsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/docs", h.ServeSwaggerUI)
	mux.HandleFunc("GET /api/docs/", h.ServeSwaggerUI)
	mux.HandleFunc("GET /api/docs/openapi.yaml", h.ServeSpec)
}

// GetSpecPath returns the path to the OpenAPI spec file
func GetSpecPath() string {
	// Try current directory first
	if _, err := os.Stat("openapi.yaml"); err == nil {
		return "openapi.yaml"
	}

	// Try executable directory
	exe, err := os.Executable()
	if err == nil {
		specPath := filepath.Join(filepath.Dir(exe), "openapi.yaml")
		if _, err := os.Stat(specPath); err == nil {
			return specPath
		}
	}

	return "openapi.yaml" // Default
}
