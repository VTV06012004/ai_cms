package usecase

import (
	"context"
	"errors"
	"testing"

	"api/internal/domain"
)

// newUsersWithTwoAccounts tạo hai tài khoản: ID 1 (admin@) và ID 2 (author@),
// mật khẩu của cả hai là "secret1" (stub băm bằng cách thêm tiền tố "hashed:").
func newUsersWithTwoAccounts(t *testing.T) *userRepositoryStub {
	t.Helper()
	users := newUserRepositoryStub()
	for _, email := range []string{"admin@example.com", "author@example.com"} {
		user, err := domain.NewUser("User", email, "hashed:secret1")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := users.Create(context.Background(), user); err != nil {
			t.Fatal(err)
		}
	}
	return users
}

func TestUpdateUserRoleRejectsChangingOwnRole(t *testing.T) {
	users := newUsersWithTwoAccounts(t)
	audits := &auditRepositoryStub{}
	useCase := newTestUserUseCase(users, audits)

	err := useCase.UpdateUserRole(context.Background(), 1, 1, string(domain.UserRoleEditor), "")

	if !errors.Is(err, domain.ErrCannotChangeOwnRole) {
		t.Fatalf("error = %v, want ErrCannotChangeOwnRole", err)
	}
	if len(audits.actions) != 0 {
		t.Fatalf("audit recorded for a rejected change: %v", audits.actions)
	}
}

func TestUpdateUserRoleRejectsUnknownRole(t *testing.T) {
	users := newUsersWithTwoAccounts(t)
	useCase := newTestUserUseCase(users, &auditRepositoryStub{})

	for _, role := range []string{"owner", "", "ADMIN"} {
		if err := useCase.UpdateUserRole(context.Background(), 1, 2, role, ""); !errors.Is(err, domain.ErrInvalidRole) {
			t.Errorf("role %q: error = %v, want ErrInvalidRole", role, err)
		}
	}
	if got := users.users[2].Data().Role; got != domain.UserRoleAuthor {
		t.Fatalf("role changed to %q despite invalid input", got)
	}
}

func TestUpdateUserRoleReportsMissingUser(t *testing.T) {
	users := newUsersWithTwoAccounts(t)
	useCase := newTestUserUseCase(users, &auditRepositoryStub{})

	err := useCase.UpdateUserRole(context.Background(), 1, 99, string(domain.UserRoleEditor), "")

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("error = %v, want ErrUserNotFound", err)
	}
}

func TestUpdateUserRoleChangesRoleAndRecordsAudit(t *testing.T) {
	users := newUsersWithTwoAccounts(t)
	audits := &auditRepositoryStub{}
	useCase := newTestUserUseCase(users, audits)

	if err := useCase.UpdateUserRole(context.Background(), 1, 2, string(domain.UserRoleEditor), ""); err != nil {
		t.Fatalf("UpdateUserRole returned error: %v", err)
	}

	if got := users.users[2].Data().Role; got != domain.UserRoleEditor {
		t.Fatalf("role = %q, want editor", got)
	}
	if len(audits.actions) != 1 || audits.actions[0] != "UPDATE_USER_ROLE" {
		t.Fatalf("audit actions = %v, want [UPDATE_USER_ROLE]", audits.actions)
	}
}

func TestLoginWithCorrectPasswordReturnsToken(t *testing.T) {
	useCase := newTestUserUseCase(newUsersWithTwoAccounts(t), &auditRepositoryStub{})

	result, err := useCase.Login(context.Background(), LoginDTO{Email: "admin@example.com", Password: "secret1"}, "")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if result.Token == "" {
		t.Fatal("login returned an empty token")
	}
}

func TestLoginWithWrongPasswordIsRejected(t *testing.T) {
	useCase := newTestUserUseCase(newUsersWithTwoAccounts(t), &auditRepositoryStub{})

	result, err := useCase.Login(context.Background(), LoginDTO{Email: "admin@example.com", Password: "wrong-password"}, "")

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("error = %v, want ErrInvalidCredentials", err)
	}
	if result != nil {
		t.Fatal("login returned a result for a wrong password")
	}
}
