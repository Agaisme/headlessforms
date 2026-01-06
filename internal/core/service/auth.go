package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"headless_form/internal/core/domain"
	"headless_form/internal/core/ports"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string
	TokenDuration time.Duration
}

// AuthService handles authentication operations
type AuthService struct {
	repo   ports.Repository
	config AuthConfig
}

// NewAuthService creates a new auth service
func NewAuthService(repo ports.Repository, config AuthConfig) *AuthService {
	if config.TokenDuration == 0 {
		config.TokenDuration = 24 * time.Hour // Default 24 hours
	}
	return &AuthService{repo: repo, config: config}
}

// Claims represents JWT claims
type Claims struct {
	UserID string          `json:"user_id"`
	Email  string          `json:"email"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	// Check if user already exists
	existing, _ := s.repo.User().GetByEmail(ctx, email)
	if existing != nil {
		return nil, domain.ErrUserExists
	}

	// Determine role - first user is super_admin
	role := domain.RoleUser
	count, _ := s.repo.User().Count(ctx)
	if count == 0 {
		role = domain.RoleSuperAdmin
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.User().Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	user, err := s.repo.User().GetByEmail(ctx, email)
	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	if !user.CheckPassword(password) {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.User().GetByID(ctx, id)
}

// generateToken creates a new JWT token for a user
func (s *AuthService) generateToken(user *domain.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// HasUsers returns true if there are any users in the system
func (s *AuthService) HasUsers(ctx context.Context) (bool, error) {
	count, err := s.repo.User().Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListUsers returns all users in the system (admin only)
func (s *AuthService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.repo.User().List(ctx)
}

// DeleteUser removes a user from the system (admin only)
func (s *AuthService) DeleteUser(ctx context.Context, userID string) error {
	// Prevent deleting the last admin
	user, err := s.repo.User().GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// If deleting an admin, make sure there's at least one other admin
	if user.Role == domain.RoleAdmin {
		users, err := s.repo.User().List(ctx)
		if err != nil {
			return err
		}
		adminCount := 0
		for _, u := range users {
			if u.Role == domain.RoleAdmin {
				adminCount++
			}
		}
		if adminCount <= 1 {
			return errors.New("cannot delete the last admin user")
		}
	}

	return s.repo.User().Delete(ctx, userID)
}

// CreateUser creates a new user with a specified role (admin only)
func (s *AuthService) CreateUser(ctx context.Context, email, password, name string, role domain.UserRole) (*domain.User, error) {
	// Check if user already exists
	existing, _ := s.repo.User().GetByEmail(ctx, email)
	if existing != nil {
		return nil, domain.ErrUserExists
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.User().Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// RequestPasswordReset creates a password reset token for the given email
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (*domain.PasswordResetToken, error) {
	user, err := s.repo.User().GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// Don't reveal if email exists - return nil silently
		return nil, nil
	}

	// Clean up expired tokens first
	_ = s.repo.PasswordReset().DeleteExpired(ctx)

	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	tokenStr := base64.URLEncoding.EncodeToString(tokenBytes)

	resetToken := &domain.PasswordResetToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiry
		CreatedAt: time.Now(),
	}

	if err := s.repo.PasswordReset().Create(ctx, resetToken); err != nil {
		return nil, err
	}

	return resetToken, nil
}

// ResetPassword resets the password using a valid reset token
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	resetToken, err := s.repo.PasswordReset().GetByToken(ctx, token)
	if err != nil {
		return err
	}
	if resetToken == nil {
		return domain.ErrInvalidResetToken
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return domain.ErrInvalidResetToken
	}

	// Check if token was already used
	if resetToken.UsedAt != nil {
		return domain.ErrInvalidResetToken
	}

	// Get user and update password
	user, err := s.repo.User().GetByID(ctx, resetToken.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if err := user.SetPassword(newPassword); err != nil {
		return err
	}
	user.UpdatedAt = time.Now()

	if err := s.repo.User().Update(ctx, user); err != nil {
		return err
	}

	// Mark token as used
	return s.repo.PasswordReset().MarkAsUsed(ctx, resetToken.ID)
}
