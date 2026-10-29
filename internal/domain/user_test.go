package domain

import (
	"errors"
	"testing"
)

func TestNewUserDefaultsToAuthor(t *testing.T) {
	user, err := NewUser("  Example User  ", "user@example.com", "hashed-password")
	if err != nil {
		t.Fatalf("NewUser returned error: %v", err)
	}
	if user.Data().Name != "Example User" {
		t.Fatalf("expected trimmed name, got %q", user.Data().Name)
	}
	if user.Data().Role != UserRoleAuthor {
		t.Fatalf("expected author role, got %q", user.Data().Role)
	}
}

func TestNewUserRejectsInvalidEmail(t *testing.T) {
	_, err := NewUser("Example User", "invalid-email", "hashed-password")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestNewRootAdminUserCreatesProtectedAdmin(t *testing.T) {
	user, err := NewRootAdminUser("Admin", "admin@example.com", "hashed-password")
	if err != nil {
		t.Fatalf("NewRootAdminUser returned error: %v", err)
	}
	data := user.Data()
	if data.Role != UserRoleAdmin || !data.RootAdmin {
		t.Fatalf("expected protected admin, got role %q and root flag %t", data.Role, data.RootAdmin)
	}
	if err := user.ChangeRole(UserRoleAuthor); !errors.Is(err, ErrCannotChangeRootAdmin) {
		t.Fatalf("expected ErrCannotChangeRootAdmin, got %v", err)
	}
}

func TestNonRootUserWithIDOneCanChangeRole(t *testing.T) {
	user := ReconstituteUser(UserData{ID: 1, Role: UserRoleAuthor})
	if err := user.ChangeRole(UserRoleAdmin); err != nil {
		t.Fatalf("ID 1 without root identity should be changeable: %v", err)
	}
}

func TestUserChangeRoleRejectsUnknownRole(t *testing.T) {
	user := ReconstituteUser(UserData{Role: UserRoleAuthor})
	if err := user.ChangeRole(UserRole("owner")); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
	if user.Data().Role != UserRoleAuthor {
		t.Fatalf("role changed after invalid update: %q", user.Data().Role)
	}
}

func TestUserRolePostPermissions(t *testing.T) {
	if UserRoleAuthor.CanSubmitPostStatus(PostStatusPublished) {
		t.Fatal("author should not publish directly")
	}
	if !UserRoleAuthor.CanSubmitPostStatus(PostStatusReview) {
		t.Fatal("author should be allowed to submit for review")
	}
	if !UserRoleEditor.CanSubmitPostStatus(PostStatusPublished) {
		t.Fatal("editor should be allowed to publish")
	}
	if UserRoleEditor.CanSubmitPostStatus(PostStatusRejected) {
		t.Fatal("rejection should be reserved for moderation")
	}
}

func TestUserDataIsDetachedFromEntity(t *testing.T) {
	user := ReconstituteUser(UserData{Email: Email("user@example.com"), Role: UserRoleAuthor})
	data := user.Data()
	data.Role = UserRoleAdmin
	if user.Data().Role != UserRoleAuthor {
		t.Fatal("modifying exported data changed the user entity")
	}
}
