package persistence

import (
	"testing"

	"api/internal/domain"
)

func TestUserModelPreservesRootAdminIdentity(t *testing.T) {
	rootKey := uint(1)
	root := userModelToDomain(UserModel{
		ID: 42, Email: "root@example.com", Role: string(domain.UserRoleAdmin),
		RootAdminKey: &rootKey,
	})
	if !root.Data().RootAdmin {
		t.Fatal("root admin marker was lost when converting persistence model to domain")
	}

	ordinary := userModelToDomain(UserModel{
		ID: 1, Email: "author@example.com", Role: string(domain.UserRoleAuthor),
	})
	if ordinary.Data().RootAdmin {
		t.Fatal("user ID 1 was incorrectly treated as root admin")
	}
}
