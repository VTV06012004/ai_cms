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

func TestNewPostRejectsRemovedApprovedStatus(t *testing.T) {
	_, err := NewPost("Title", "slug", "content", PostStatus("approved"), 1, 1)
	if !errors.Is(err, ErrInvalidPostStatus) {
		t.Fatalf("expected ErrInvalidPostStatus, got %v", err)
	}
}

func TestApplyScheduleRequiresFutureTime(t *testing.T) {
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	post := ReconstitutePost(PostData{Status: PostStatusScheduled})

	if err := post.ApplySchedule(nil, now); !errors.Is(err, ErrInvalidScheduleTime) {
		t.Fatalf("expected ErrInvalidScheduleTime for nil, got %v", err)
	}
	past := now.Add(-time.Minute)
	if err := post.ApplySchedule(&past, now); !errors.Is(err, ErrInvalidScheduleTime) {
		t.Fatalf("expected ErrInvalidScheduleTime for past, got %v", err)
	}
	future := now.Add(time.Hour)
	if err := post.ApplySchedule(&future, now); err != nil {
		t.Fatalf("ApplySchedule returned error: %v", err)
	}
	if got := post.Data().ScheduledAt; got == nil || !got.Equal(future) {
		t.Fatalf("scheduled time = %v, want %v", got, future)
	}
}

func TestApplyScheduleClearsTimeForOtherStatuses(t *testing.T) {
	at := time.Now().Add(time.Hour)
	post := ReconstitutePost(PostData{Status: PostStatusDraft, ScheduledAt: &at})

	if err := post.ApplySchedule(nil, time.Now()); err != nil {
		t.Fatalf("ApplySchedule returned error: %v", err)
	}
	if post.Data().ScheduledAt != nil {
		t.Fatal("scheduled time must be cleared for non-scheduled posts")
	}
}

func TestPublishScheduledOnlyWhenDue(t *testing.T) {
	due := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	post := ReconstitutePost(PostData{Status: PostStatusScheduled, ScheduledAt: &due})

	if err := post.PublishScheduled(due.Add(-time.Second)); !errors.Is(err, ErrPostNotDue) {
		t.Fatalf("expected ErrPostNotDue, got %v", err)
	}
	if err := post.PublishScheduled(due.Add(time.Second)); err != nil {
		t.Fatalf("PublishScheduled returned error: %v", err)
	}
	data := post.Data()
	if data.Status != PostStatusPublished || data.ScheduledAt != nil {
		t.Fatalf("unexpected state after publishing: %+v", data)
	}
	if data.PublishedAt == nil || !data.PublishedAt.Equal(due) {
		t.Fatalf("published_at = %v, want %v", data.PublishedAt, due)
	}
}

func TestAuthorCannotSetScheduledStatus(t *testing.T) {
	if err := UserRoleAuthor.EnsureCanSetPostStatus(PostStatusScheduled); !errors.Is(err, ErrPublishNotAllowed) {
		t.Fatalf("expected ErrPublishNotAllowed, got %v", err)
	}
	if err := UserRoleEditor.EnsureCanSetPostStatus(PostStatusScheduled); err != nil {
		t.Fatalf("editor should be allowed, got %v", err)
	}
}

func TestApplyScheduleKeepsExistingTimeWhenOmitted(t *testing.T) {
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	existing := now.Add(-time.Minute) // đã quá hạn, đang chờ tác vụ nền đăng
	post := ReconstitutePost(PostData{Status: PostStatusScheduled, ScheduledAt: &existing})

	if err := post.ApplySchedule(nil, now); err != nil {
		t.Fatalf("ApplySchedule returned error: %v", err)
	}
	if got := post.Data().ScheduledAt; got == nil || !got.Equal(existing) {
		t.Fatalf("scheduled time = %v, want %v", got, existing)
	}
}

func TestApplyScheduleAcceptsUnchangedTimeWithinTheSameMinute(t *testing.T) {
	now := time.Date(2026, 1, 1, 10, 0, 30, 0, time.UTC)
	existing := time.Date(2026, 1, 1, 9, 59, 45, 0, time.UTC)
	resent := time.Date(2026, 1, 1, 9, 59, 0, 0, time.UTC) // form chỉ gửi đến phút
	post := ReconstitutePost(PostData{Status: PostStatusScheduled, ScheduledAt: &existing})

	if err := post.ApplySchedule(&resent, now); err != nil {
		t.Fatalf("ApplySchedule returned error: %v", err)
	}
	if got := post.Data().ScheduledAt; got == nil || !got.Equal(existing) {
		t.Fatalf("scheduled time = %v, want the original %v", got, existing)
	}
	changedToPast := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	if err := post.ApplySchedule(&changedToPast, now); !errors.Is(err, ErrInvalidScheduleTime) {
		t.Fatalf("expected ErrInvalidScheduleTime for a new past time, got %v", err)
	}
}

func TestSettleStatusSendsAuthorEditsBackToReview(t *testing.T) {
	publishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	scheduledAt := publishedAt.Add(48 * time.Hour)
	tests := []struct {
		name string
		data PostData
		role UserRole
		want PostStatus
	}{
		{"author edits published", PostData{Status: PostStatusPublished, PublishedAt: &publishedAt}, UserRoleAuthor, PostStatusReview},
		{"author edits scheduled", PostData{Status: PostStatusScheduled, ScheduledAt: &scheduledAt}, UserRoleAuthor, PostStatusReview},
		{"editor edits published", PostData{Status: PostStatusPublished, PublishedAt: &publishedAt}, UserRoleEditor, PostStatusPublished},
		{"author edits draft", PostData{Status: PostStatusDraft}, UserRoleAuthor, PostStatusDraft},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			post := ReconstitutePost(test.data)
			post.SettleStatus(test.role, publishedAt.Add(time.Hour))
			if got := post.Data().Status; got != test.want {
				t.Fatalf("status = %q, want %q", got, test.want)
			}
			if test.want == PostStatusReview && (post.Data().PublishedAt != nil || post.Data().ScheduledAt != nil) {
				t.Fatal("a post sent back to review must lose its publication and schedule times")
			}
		})
	}
}

func TestSettleStatusRecordsPublicationTimeOnlyOnce(t *testing.T) {
	first := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	post := ReconstitutePost(PostData{Status: PostStatusPublished})

	post.SettleStatus(UserRoleEditor, first)
	post.SettleStatus(UserRoleEditor, first.Add(time.Hour))

	if got := post.Data().PublishedAt; got == nil || !got.Equal(first) {
		t.Fatalf("published_at = %v, want %v", got, first)
	}
}
