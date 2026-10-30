package domain

import (
	"errors"
	"testing"
	"time"
)

func TestPostModerateRequiresReviewStatus(t *testing.T) {
	post := ReconstitutePost(PostData{Status: PostStatusDraft})

	err := post.Moderate(PostStatusPublished, "", 7, time.Now())
	if !errors.Is(err, ErrPostNotInReview) {
		t.Fatalf("expected ErrPostNotInReview, got %v", err)
	}
	if post.Data().Status != PostStatusDraft {
		t.Fatalf("status changed after rejected moderation: %q", post.Data().Status)
	}
}

func TestPostModerateRequiresReasonForRejection(t *testing.T) {
	post := ReconstitutePost(PostData{Status: PostStatusReview})

	err := post.Moderate(PostStatusRejected, "", 7, time.Now())
	if !errors.Is(err, ErrRejectionReasonRequired) {
		t.Fatalf("expected ErrRejectionReasonRequired, got %v", err)
	}
	if post.Data().Status != PostStatusReview {
		t.Fatalf("status changed after rejected moderation: %q", post.Data().Status)
	}
}

func TestPostModeratePublishesReviewPost(t *testing.T) {
	now := time.Now()
	post := ReconstitutePost(PostData{Status: PostStatusReview, RejectionReason: "old reason"})

	if err := post.Moderate(PostStatusPublished, "", 7, now); err != nil {
		t.Fatalf("Moderate returned error: %v", err)
	}
	data := post.Data()
	if data.Status != PostStatusPublished {
		t.Fatalf("expected published status, got %q", data.Status)
	}
	if data.RejectionReason != "" {
		t.Fatalf("expected rejection reason to be cleared, got %q", data.RejectionReason)
	}
	if data.ModeratorID == nil || *data.ModeratorID != 7 {
		t.Fatalf("expected moderator ID 7, got %v", data.ModeratorID)
	}
	if data.PublishedAt == nil || !data.PublishedAt.Equal(now) {
		t.Fatalf("expected publication time %v, got %v", now, data.PublishedAt)
	}
}

func TestNewPostValidatesRequiredFieldsAndStatus(t *testing.T) {
	if _, err := NewPost("", "slug", "body", PostStatusDraft, 1, 7); !errors.Is(err, ErrPostTitleRequired) {
		t.Fatalf("expected ErrPostTitleRequired, got %v", err)
	}
	if _, err := NewPost("title", "slug", "body", PostStatus("unknown"), 1, 7); !errors.Is(err, ErrInvalidPostStatus) {
		t.Fatalf("expected ErrInvalidPostStatus, got %v", err)
	}
}

func TestPostUpdateRejectsInvalidDataWithoutMutating(t *testing.T) {
	post := ReconstitutePost(PostData{
		Title: "old", Slug: "old-slug", Content: "old body",
		Status: PostStatusDraft, CategoryID: 1,
	})

	err := post.Update("new", "new-slug", "new body", PostStatus("unknown"), 2)
	if !errors.Is(err, ErrInvalidPostStatus) {
		t.Fatalf("expected ErrInvalidPostStatus, got %v", err)
	}
	data := post.Data()
	if data.Title != "old" || data.Slug != "old-slug" || data.Content != "old body" || data.CategoryID != 1 {
		t.Fatalf("post changed after rejected update: %+v", post)
	}
}

func TestPostUpdatePreservesStatusWhenOmitted(t *testing.T) {
	post := ReconstitutePost(PostData{
		Title: "old", Slug: "old-slug", Content: "old body",
		Status: PostStatusPublished, CategoryID: 1,
	})

	if err := post.Update("new", "new-slug", "new body", "", 2); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if got := post.Data().Status; got != PostStatusPublished {
		t.Fatalf("expected status to remain %q, got %q", PostStatusPublished, got)
	}
}

func TestPostCanBeManagedByRoleAndOwner(t *testing.T) {
	post := ReconstitutePost(PostData{AuthorID: 12})
	tests := []struct {
		role UserRole
		id   uint
		want bool
	}{
		{role: UserRoleAuthor, id: 12, want: true},
		{role: UserRoleAuthor, id: 13, want: false},
		{role: UserRoleEditor, id: 13, want: true},
		{role: UserRoleAdmin, id: 13, want: true},
		{role: UserRole("unknown"), id: 12, want: false},
	}
	for _, test := range tests {
		if got := post.CanBeManagedBy(test.id, test.role); got != test.want {
			t.Errorf("CanBeManagedBy(%d, %q) = %t, want %t", test.id, test.role, got, test.want)
		}
	}
}

func TestPostDataIsDetachedFromEntity(t *testing.T) {
	moderatorID := uint(4)
	publishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	parentID := uint(8)
	post := ReconstitutePost(PostData{
		Title: "original", ModeratorID: &moderatorID, PublishedAt: &publishedAt,
		Category: ReconstituteCategory(CategoryData{ParentID: &parentID}),
		Tags:     []Tag{ReconstituteTag(TagData{ID: 3})},
	})

	data := post.Data()
	*data.ModeratorID = 10
	*data.PublishedAt = publishedAt.Add(time.Hour)
	*data.Category.Data().ParentID = 11
	data.Tags[0] = ReconstituteTag(TagData{ID: 99})

	got := post.Data()
	if *got.ModeratorID != moderatorID || !got.PublishedAt.Equal(publishedAt) {
		t.Fatal("modifying exported pointer data changed the post entity")
	}
	if *got.Category.Data().ParentID != parentID || got.Tags[0].Data().ID != 3 {
		t.Fatal("modifying exported related data changed the post entity")
	}
}
