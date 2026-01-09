package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"headless_form/internal/core/service"
)

// ContextKey type for context keys
type ContextKey string

const (
	UserIDKey ContextKey = "user_id"
	EmailKey  ContextKey = "email"
	RoleKey   ContextKey = "role"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, `{"status":"fail","message":"Authorization header required"}`, http.StatusUnauthorized)
				return
			}

			// Extract Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeJSONError(w, `{"status":"fail","message":"Invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				writeJSONError(w, `{"status":"fail","message":"Invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware extracts user info if present but doesn't require it
func OptionalAuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					claims, err := authService.ValidateToken(parts[1])
					if err == nil {
						ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
						ctx = context.WithValue(ctx, EmailKey, claims.Email)
						ctx = context.WithValue(ctx, RoleKey, claims.Role)
						r = r.WithContext(ctx)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole creates middleware that requires a specific role
func RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(string)
			if !ok || role != requiredRole {
				writeJSONError(w, `{"status":"fail","message":"Insufficient permissions"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserEmail extracts user email from context
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(EmailKey).(string); ok {
		return email
	}
	return ""
}

// GetUserRole extracts user role from context
func GetUserRole(ctx context.Context) string {
	// Role is stored as domain.UserRole, not string
	if role, ok := ctx.Value(RoleKey).(string); ok {
		return role
	}
	// Also try interface type and convert
	if role := ctx.Value(RoleKey); role != nil {
		return fmt.Sprintf("%v", role)
	}
	return ""
}

// RequireAnyRole creates middleware that requires any of the specified roles
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			for _, role := range roles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeJSONError(w, `{"status":"fail","message":"Insufficient permissions"}`, http.StatusForbidden)
		})
	}
}

// IsSuperAdmin checks if user has super_admin role
func IsSuperAdmin(ctx context.Context) bool {
	return GetUserRole(ctx) == "super_admin"
}

// IsAdmin checks if user has admin or super_admin role
func IsAdmin(ctx context.Context) bool {
	role := GetUserRole(ctx)
	return role == "super_admin" || role == "admin"
}

// CanManageAllForms returns true if user can manage all forms (super_admin or admin)
func CanManageAllForms(ctx context.Context) bool {
	return IsAdmin(ctx)
}

// CanAccessForm checks if user can access a specific form (owner or admin)
func CanAccessForm(ctx context.Context, formOwnerID string) bool {
	// Admins and super_admins can access all forms
	if IsAdmin(ctx) {
		return true
	}
	// Users can only access their own forms
	return GetUserID(ctx) == formOwnerID
}

// writeJSONError writes a JSON error response with proper Content-Type
func writeJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write([]byte(message)); err != nil {
		log.Printf("[ERROR] Failed to write error response: %v", err)
	}
}
