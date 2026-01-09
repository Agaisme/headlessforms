package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User roles
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin" // Site owner, full access
	RoleAdmin      UserRole = "admin"       // Manage users & all forms
	RoleUser       UserRole = "user"        // Create/manage own forms only
)

// User errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailRequired      = errors.New("email is required")
	ErrPasswordRequired   = errors.New("password is required")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrInvalidResetToken  = errors.New("invalid or expired reset token")
)

// emailRegex is a basic email validation pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// PasswordResetToken represents a password reset request
type PasswordResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// User represents an authenticated user
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	Name         string    `json:"name"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// Validate validates user fields
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrEmailRequired
	}
	// Normalize and validate email format
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidEmail
	}
	return nil
}

// UserPublic is a safe representation of User for API responses
type UserPublic struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToPublic converts User to UserPublic
func (u *User) ToPublic() *UserPublic {
	return &UserPublic{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
