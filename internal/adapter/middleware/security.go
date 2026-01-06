package middleware

import (
	"net/http"
	"strings"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	AllowedOrigins []string
	IsDevelopment  bool
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content Security Policy (relaxed for admin SPA)
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline'; "+
					"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
					"font-src 'self' https://fonts.gstatic.com; "+
					"img-src 'self' data: blob:; "+
					"connect-src 'self'")

			next.ServeHTTP(w, r)
		})
	}
}

// HTTPSRedirect redirects HTTP to HTTPS in production
// Enable by setting FORCE_HTTPS=true environment variable
func HTTPSRedirect(forceHTTPS bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if forceHTTPS {
				// Check X-Forwarded-Proto header (set by reverse proxies)
				proto := r.Header.Get("X-Forwarded-Proto")
				if proto == "http" {
					httpsURL := "https://" + r.Host + r.URL.RequestURI()
					http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware creates CORS middleware with configurable origins
func CORSMiddleware(config SecurityConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			if config.IsDevelopment || len(config.AllowedOrigins) == 0 {
				// In development or if no origins specified, allow all
				allowed = true
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			} else {
				for _, o := range config.AllowedOrigins {
					if o == "*" || o == origin {
						allowed = true
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Logging middleware for request logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging for static assets
		if strings.HasPrefix(r.URL.Path, "/_app/") ||
			strings.HasPrefix(r.URL.Path, "/favicon") {
			next.ServeHTTP(w, r)
			return
		}

		// Log the request (could be enhanced with structured logging)
		// log.Printf("%s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
