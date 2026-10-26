package http

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"api/internal/domain"
)

func TestUserResponseKeepsAPIShapeAndOmitsPasswordHash(t *testing.T) {
	user := domain.ReconstituteUser(domain.UserData{
		ID: 9, Email: domain.Email("user@example.com"), Name: "User",
		PasswordHash: "must-not-leak", Role: domain.UserRoleAuthor,
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	})
	encoded, err := json.Marshal(toUserResponse(*user))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	body := string(encoded)
	for _, expected := range []string{`"id":9`, `"email":"user@example.com"`, `"role":"author"`, `"created_at":"2026-01-02T03:04:05Z"`} {
		if !strings.Contains(body, expected) {
			t.Errorf("response missing %s: %s", expected, body)
		}
	}
	if strings.Contains(body, "must-not-leak") || strings.Contains(body, "PasswordHash") {
		t.Fatalf("response exposed a password hash: %s", body)
	}
}

func TestPostResponseMapsNestedDomainObjects(t *testing.T) {
	post := domain.ReconstitutePost(domain.PostData{
		ID: 1, Title: "Title", Status: domain.PostStatusDraft,
		Author:   domain.ReconstituteUser(domain.UserData{ID: 2, Email: domain.Email("author@example.com"), Role: domain.UserRoleAuthor}),
		Category: domain.ReconstituteCategory(domain.CategoryData{ID: 3, Name: "News"}),
		Tags:     []domain.Tag{domain.ReconstituteTag(domain.TagData{ID: 4, Name: "Go"})},
	})
	encoded, err := json.Marshal(toPostResponse(*post))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	body := string(encoded)
	for _, expected := range []string{`"status":"draft"`, `"author":{"id":2`, `"category":{"id":3`, `"tags":[{"id":4`} {
		if !strings.Contains(body, expected) {
			t.Errorf("response missing %s: %s", expected, body)
		}
	}
}
