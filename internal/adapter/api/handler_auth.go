package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"headless_form/internal/adapter/api/response"
	"headless_form/internal/adapter/email"
	"headless_form/internal/adapter/middleware"
	"headless_form/internal/core/domain"
	"headless_form/internal/core/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService  *service.AuthService
	emailService *email.Service
	baseURL      string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, emailService *email.Service, baseURL string) *AuthHandler {
	return &AuthHandler{authService: authService, emailService: emailService, baseURL: baseURL}
}

// RegisterPublicRoutes registers public auth routes (no auth required)
func (h *AuthHandler) RegisterPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.HandleRegister)
	mux.HandleFunc("POST /api/v1/auth/login", h.HandleLogin)
	mux.HandleFunc("GET /api/v1/auth/setup", h.HandleSetupRequired)
	mux.HandleFunc("POST /api/v1/auth/forgot-password", h.HandleForgotPassword)
	mux.HandleFunc("POST /api/v1/auth/reset-password", h.HandleResetPassword)
}

// RegisterProtectedRoutes registers protected auth routes (auth required)
func (h *AuthHandler) RegisterProtectedRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	// Current user
	mux.Handle("GET /api/v1/auth/me", authMiddleware(http.HandlerFunc(h.HandleMe)))

	// Self-service profile management
	mux.Handle("PUT /api/v1/auth/profile", authMiddleware(http.HandlerFunc(h.HandleUpdateProfile)))
	mux.Handle("PUT /api/v1/auth/password", authMiddleware(http.HandlerFunc(h.HandleUpdatePassword)))

	// Admin user management
	mux.Handle("GET /api/v1/users", authMiddleware(http.HandlerFunc(h.HandleListUsers)))
	mux.Handle("POST /api/v1/users", authMiddleware(http.HandlerFunc(h.HandleCreateUser)))
	mux.Handle("PUT /api/v1/users/{user_id}", authMiddleware(http.HandlerFunc(h.HandleUpdateUser)))
	mux.Handle("DELETE /api/v1/users/{user_id}", authMiddleware(http.HandlerFunc(h.HandleDeleteUser)))
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string             `json:"token"`
	User  *domain.UserPublic `json:"user"`
}

// HandleRegister handles user registration
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "Email and password are required", "MISSING_FIELDS")
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		switch err {
		case domain.ErrUserExists:
			response.Error(w, http.StatusConflict, "User already exists", "USER_EXISTS")
		case domain.ErrPasswordTooShort:
			response.BadRequest(w, "Password must be at least 8 characters", "PASSWORD_TOO_SHORT")
		case domain.ErrEmailRequired:
			response.BadRequest(w, "Email is required", "EMAIL_REQUIRED")
		default:
			response.Error(w, http.StatusInternalServerError, "Failed to register", "REGISTER_FAILED")
		}
		return
	}

	// Generate token for immediate login
	token, _, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Registration successful but failed to generate token", "TOKEN_FAILED")
		return
	}

	response.Created(w, AuthResponse{
		Token: token,
		User:  user.ToPublic(),
	})
}

// HandleLogin handles user login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "Email and password are required", "MISSING_FIELDS")
		return
	}

	token, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid credentials", "INVALID_CREDENTIALS")
		return
	}

	response.Success(w, AuthResponse{
		Token: token,
		User:  user.ToPublic(),
	})
}

// HandleMe returns the currently authenticated user
func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.Success(w, user.ToPublic())
}

// HandleSetupRequired checks if initial setup is needed
func (h *AuthHandler) HandleSetupRequired(w http.ResponseWriter, r *http.Request) {
	hasUsers, err := h.authService.HasUsers(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to check setup status", "CHECK_FAILED")
		return
	}

	response.Success(w, map[string]bool{
		"setup_required": !hasUsers,
	})
}

// HandleListUsers returns all users (admin only)
func (h *AuthHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	// Check if current user is admin or super_admin
	if !middleware.IsAdmin(r.Context()) {
		response.Error(w, http.StatusForbidden, "Admin access required", "FORBIDDEN")
		return
	}

	users, err := h.authService.ListUsers(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	// Convert to public representation
	var publicUsers []*domain.UserPublic
	for _, u := range users {
		publicUsers = append(publicUsers, u.ToPublic())
	}

	response.Success(w, map[string]interface{}{
		"users": publicUsers,
		"total": len(publicUsers),
	})
}

// HandleCreateUser creates a new user (admin only)
func (h *AuthHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Check if current user is admin or super_admin
	if !middleware.IsAdmin(r.Context()) {
		response.Error(w, http.StatusForbidden, "Admin access required", "FORBIDDEN")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "Email and password are required", "MISSING_FIELDS")
		return
	}

	// Default to viewer role if not specified
	userRole := domain.RoleUser
	if req.Role == "admin" {
		userRole = domain.RoleAdmin
	}

	user, err := h.authService.CreateUser(r.Context(), req.Email, req.Password, req.Name, userRole)
	if err != nil {
		switch err {
		case domain.ErrUserExists:
			response.Error(w, http.StatusConflict, "User already exists", "USER_EXISTS")
		case domain.ErrPasswordTooShort:
			response.BadRequest(w, "Password must be at least 8 characters", "PASSWORD_TOO_SHORT")
		default:
			response.HandleError(w, err)
		}
		return
	}

	response.Created(w, user.ToPublic())
}

// HandleDeleteUser deletes a user (admin only)
func (h *AuthHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// Check if current user is admin or super_admin
	if !middleware.IsAdmin(r.Context()) {
		response.Error(w, http.StatusForbidden, "Admin access required", "FORBIDDEN")
		return
	}

	userID := r.PathValue("user_id")
	if userID == "" {
		response.BadRequest(w, "User ID required", "MISSING_USER_ID")
		return
	}

	// Prevent self-deletion
	currentUserID := middleware.GetUserID(r.Context())
	if userID == currentUserID {
		response.BadRequest(w, "Cannot delete your own account", "SELF_DELETE")
		return
	}

	if err := h.authService.DeleteUser(r.Context(), userID); err != nil {
		if err == domain.ErrUserNotFound {
			response.NotFound(w, "User not found")
		} else {
			response.Error(w, http.StatusBadRequest, err.Error(), "DELETE_FAILED")
		}
		return
	}

	response.Success(w, map[string]string{"message": "User deleted successfully"})
}

// HandleUpdateProfile updates the current user's profile (self-service)
func (h *AuthHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", "INVALID_BODY")
		return
	}

	user, err := h.authService.UpdateUser(r.Context(), userID, req.Name, req.Email, nil)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			response.NotFound(w, "User not found")
		case domain.ErrUserExists:
			response.Error(w, http.StatusConflict, "Email already in use", "EMAIL_EXISTS")
		default:
			response.HandleError(w, err)
		}
		return
	}

	response.Success(w, user.ToPublic())
}

// HandleUpdatePassword changes the current user's password
func (h *AuthHandler) HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		response.BadRequest(w, "Current and new password are required", "MISSING_FIELDS")
		return
	}

	if err := h.authService.UpdatePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			response.Error(w, http.StatusUnauthorized, "Current password is incorrect", "INVALID_PASSWORD")
		case domain.ErrPasswordTooShort:
			response.BadRequest(w, "Password must be at least 8 characters", "PASSWORD_TOO_SHORT")
		default:
			response.HandleError(w, err)
		}
		return
	}

	response.Success(w, map[string]string{"message": "Password updated successfully"})
}

// HandleUpdateUser updates a user's profile (admin only, can update role)
func (h *AuthHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Check if current user is admin or super_admin
	if !middleware.IsAdmin(r.Context()) {
		response.Error(w, http.StatusForbidden, "Admin access required", "FORBIDDEN")
		return
	}

	userID := r.PathValue("user_id")
	if userID == "" {
		response.BadRequest(w, "User ID required", "MISSING_USER_ID")
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", "INVALID_BODY")
		return
	}

	// Parse role if provided
	var role *domain.UserRole
	if req.Role != "" {
		r := domain.UserRole(req.Role)
		// Validate role
		if r != domain.RoleSuperAdmin && r != domain.RoleAdmin && r != domain.RoleUser {
			response.BadRequest(w, "Invalid role. Must be 'super_admin', 'admin', or 'user'", "INVALID_ROLE")
			return
		}
		role = &r
	}

	user, err := h.authService.UpdateUser(r.Context(), userID, req.Name, req.Email, role)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			response.NotFound(w, "User not found")
		case domain.ErrUserExists:
			response.Error(w, http.StatusConflict, "Email already in use", "EMAIL_EXISTS")
		default:
			response.HandleError(w, err)
		}
		return
	}

	response.Success(w, user.ToPublic())
}

// HandleForgotPassword initiates password reset flow
func (h *AuthHandler) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Email == "" {
		response.BadRequest(w, "Email is required", "EMAIL_REQUIRED")
		return
	}

	// Request password reset (returns nil token if email not found - don't reveal)
	resetToken, _ := h.authService.RequestPasswordReset(r.Context(), req.Email)

	// Send email if token was created (user exists)
	if resetToken != nil && h.emailService != nil {
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.baseURL, resetToken.Token)
		_ = h.emailService.SendPasswordReset(req.Email, resetURL)
	}

	// Always return success to prevent email enumeration
	response.Success(w, map[string]string{
		"message": "If an account with that email exists, a password reset link has been sent.",
	})
}

// HandleResetPassword completes password reset with token
func (h *AuthHandler) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		response.BadRequest(w, "Token and new password are required", "MISSING_FIELDS")
		return
	}

	if err := h.authService.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		switch err {
		case domain.ErrInvalidResetToken:
			response.BadRequest(w, "Invalid or expired reset token", "INVALID_TOKEN")
		case domain.ErrPasswordTooShort:
			response.BadRequest(w, "Password must be at least 8 characters", "PASSWORD_TOO_SHORT")
		default:
			response.HandleError(w, err)
		}
		return
	}

	response.Success(w, map[string]string{"message": "Password reset successfully"})
}
