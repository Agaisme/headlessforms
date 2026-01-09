# HeadlessForms Implementation Plan: Phase 1

> **Phase**: Complete Missing CRUD Operations  
> **Priority**: High  
> **Estimated Effort**: 4 hours  
> **Date Created**: 2026-01-08

---

## Objective

Complete all missing CRUD operations to ensure every entity has full functionality:

1. **User Update** - Profile and password management
2. **Submission Get** - Fetch single submission by ID

---

## Tasks

### 1. Backend: User Update Operations

#### 1.1 Add UpdateUser to AuthService

**File**: `internal/core/service/auth.go`

```go
// UpdateUser updates a user's profile (admin or self)
func (s *AuthService) UpdateUser(ctx context.Context, userID, name, email string) (*domain.User, error) {
    user, err := s.repo.User().GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, domain.ErrUserNotFound
    }

    // Check if email is changing and already exists
    if email != "" && email != user.Email {
        existing, _ := s.repo.User().GetByEmail(ctx, email)
        if existing != nil && existing.ID != userID {
            return nil, domain.ErrUserExists
        }
        user.Email = email
    }

    if name != "" {
        user.Name = name
    }
    user.UpdatedAt = time.Now()

    if err := s.repo.User().Update(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

// UpdatePassword changes a user's password (requires current password)
func (s *AuthService) UpdatePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
    user, err := s.repo.User().GetByID(ctx, userID)
    if err != nil {
        return err
    }
    if user == nil {
        return domain.ErrUserNotFound
    }

    // Verify current password
    if !user.CheckPassword(currentPassword) {
        return domain.ErrInvalidCredentials
    }

    // Set new password
    if err := user.SetPassword(newPassword); err != nil {
        return err
    }
    user.UpdatedAt = time.Now()

    return s.repo.User().Update(ctx, user)
}
```

#### 1.2 Add API Endpoints

**File**: `internal/adapter/api/auth_handler.go`

Add these new handlers:

- `HandleUpdateProfile` - `PUT /api/v1/auth/profile`
- `HandleUpdatePassword` - `PUT /api/v1/auth/password`
- `HandleUpdateUser` - `PUT /api/v1/users/{user_id}` (admin)

---

### 2. Backend: Submission Get Operation

#### 2.1 Add GetSubmission to SubmissionService

**File**: `internal/core/service/service.go`

```go
// GetSubmission retrieves a single submission by ID
func (s *SubmissionService) GetSubmission(ctx context.Context, submissionID string) (*domain.Submission, error) {
    submission, err := s.repo.Submission().GetByID(ctx, submissionID)
    if err != nil {
        return nil, fmt.Errorf("get submission: %w", err)
    }
    if submission == nil {
        return nil, domain.ErrSubmissionNotFound
    }
    return submission, nil
}
```

#### 2.2 Add API Endpoint

**File**: `internal/adapter/api/handler.go`

```go
// HandleGetSubmission: GET /api/v1/submissions/{sub_id}
func (h *Router) HandleGetSubmission(w http.ResponseWriter, r *http.Request) {
    subID := r.PathValue("sub_id")

    // Get submission
    sub, err := h.submissionService.GetSubmission(r.Context(), subID)
    if err != nil {
        if err == domain.ErrSubmissionNotFound {
            response.NotFound(w, "Submission not found")
        } else {
            response.HandleError(w, err)
        }
        return
    }

    // Verify ownership through form
    form, err := h.formService.GetFormByID(r.Context(), sub.FormID)
    if err != nil || form == nil {
        response.NotFound(w, "Form not found")
        return
    }

    if !middleware.CanAccessForm(r.Context(), form.OwnerID) {
        response.Error(w, http.StatusForbidden, "Access denied", "FORBIDDEN")
        return
    }

    response.Success(w, sub)
}
```

---

### 3. Backend: Enhanced Ownership Verification

Add ownership checks to mark-as-read/unread and delete submission handlers.

---

### 4. Frontend: Type Fix

**File**: `web/src/lib/stores/auth.ts`

```typescript
interface User {
  id: string;
  email: string;
  name: string;
  role: "super_admin" | "admin" | "user"; // Fixed type
}
```

---

### 5. Frontend: Profile Page

**File**: `web/src/routes/profile/+page.svelte`

Create new profile page with:

- Display current user info
- Edit name form
- Change password form

---

## Implementation Order

1. ✅ Add `UpdateUser()` and `UpdatePassword()` to AuthService
2. ✅ Add `HandleUpdateProfile`, `HandleUpdatePassword`, `HandleUpdateUser` handlers
3. ✅ Register new routes in main.go
4. ✅ Add `GetSubmission()` to SubmissionService
5. ✅ Add `HandleGetSubmission` handler
6. ✅ Add ownership verification to submission handlers
7. ✅ Fix frontend role type
8. ✅ Create profile page

---

## Testing Checklist

- [ ] User can update their profile name
- [ ] User can change their password
- [ ] Admin can update any user
- [ ] Admin can change user roles
- [ ] User cannot update other users
- [ ] Submission can be fetched by ID
- [ ] Non-owner cannot access submissions
- [ ] All existing tests pass

---

## Files Modified

| File                                   | Changes                             |
| -------------------------------------- | ----------------------------------- |
| `internal/core/service/auth.go`        | Add UpdateUser, UpdatePassword      |
| `internal/core/service/service.go`     | Add GetSubmission                   |
| `internal/adapter/api/auth_handler.go` | Add profile/password endpoints      |
| `internal/adapter/api/handler.go`      | Add GetSubmission, ownership checks |
| `cmd/server/main.go`                   | Register new routes                 |
| `web/src/lib/stores/auth.ts`           | Fix role type                       |
| `web/src/routes/profile/+page.svelte`  | New file                            |

---

_Plan created: 2026-01-08_
