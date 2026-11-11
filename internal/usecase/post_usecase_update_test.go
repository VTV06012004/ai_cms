package usecase

import (
	"context"
	"errors"
	"github.com/VTV06012004/ai_cms/internal/domain"
	"testing"
	"time"
)

type postUpdateUnitOfWork struct {
	repositories TransactionRepositories
}

func (u postUpdateUnitOfWork) WithinTransaction(_ context.Context, work func(TransactionRepositories) error) error {
	return work(u.repositories)
}

type postUpdateRepository struct {
	post *domain.Post
}

func (r *postUpdateRepository) GetAll(context.Context, domain.PostListCriteria, domain.PageRequest) ([]domain.Post, int64, error) {
	return nil, 0, nil
}

func (r *postUpdateRepository) GetByID(context.Context, uint) (*domain.Post, error) {
	return r.post, nil
}

func (r *postUpdateRepository) Create(_ context.Context, post *domain.Post, _ []uint) (*domain.Post, error) {
	r.post = post
	return post, nil
}

func (r *postUpdateRepository) Update(_ context.Context, post *domain.Post, _ []uint) (*domain.Post, error) {
	r.post = post
	return post, nil
}

func (r *postUpdateRepository) Delete(context.Context, uint) error {
	return nil
}

func (r *postUpdateRepository) ListDue(context.Context, time.Time, int) ([]domain.Post, error) {
	return nil, nil
}

func (r *postUpdateRepository) PublishIfDue(context.Context, uint, time.Time) (bool, error) {
	return false, nil
}

type postUpdateRevisionRepository struct{}

func (postUpdateRevisionRepository) Create(context.Context, *domain.Revision) error { return nil }
func (postUpdateRevisionRepository) GetByPostID(context.Context, uint) ([]domain.Revision, error) {
	return nil, nil
}
func (postUpdateRevisionRepository) GetByID(context.Context, uint) (*domain.Revision, error) {
	return nil, nil
}

type postUpdateAuditRepository struct{}

func (postUpdateAuditRepository) RecordLog(context.Context, uint, string, string, string, string) error {
	return nil
}
func (postUpdateAuditRepository) GetAll(context.Context, domain.PageRequest) ([]domain.AuditLog, int64, error) {
	return nil, 0, nil
}

func newUpdateUseCase(post *domain.Post) (*PostUseCase, *postUpdateRepository) {
	posts := &postUpdateRepository{post: post}
	return NewPostUseCase(nil, postUpdateUnitOfWork{repositories: TransactionRepositories{
		Posts: posts, Revisions: postUpdateRevisionRepository{}, Audits: postUpdateAuditRepository{},
	}}), posts
}

func publishedPostOf(authorID uint) *domain.Post {
	publishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return domain.ReconstitutePost(domain.PostData{
		ID: 1, Title: "Original", Slug: "original", Content: "Original content",
		Status: domain.PostStatusPublished, CategoryID: 1, AuthorID: authorID, PublishedAt: &publishedAt,
	})
}

func TestEditorUpdatingPublishedPostWithoutStatusKeepsItPublished(t *testing.T) {
	useCase, _ := newUpdateUseCase(publishedPostOf(7))

	updated, err := useCase.UpdatePost(context.Background(), 1, 9, string(domain.UserRoleEditor), CreatePostDTO{
		Title: "Updated", Slug: "updated", Content: "Updated content", CategoryID: 1,
	}, "")
	if err != nil {
		t.Fatalf("UpdatePost returned error: %v", err)
	}
	if got := updated.Data().Status; got != domain.PostStatusPublished {
		t.Fatalf("expected status to remain %q, got %q", domain.PostStatusPublished, got)
	}
	if updated.Data().Title != "Updated" {
		t.Fatalf("expected post content to update, got title %q", updated.Data().Title)
	}
}

func TestAuthorEditingOwnPublishedPostSendsItBackToReview(t *testing.T) {
	// Form sửa bài luôn gửi lại trạng thái hiện tại, kể cả khi author không được phép đặt "published".
	for _, status := range []string{"", string(domain.PostStatusPublished)} {
		useCase, _ := newUpdateUseCase(publishedPostOf(7))

		updated, err := useCase.UpdatePost(context.Background(), 1, 7, string(domain.UserRoleAuthor), CreatePostDTO{
			Title: "Updated", Slug: "updated", Content: "Updated content", CategoryID: 1, Status: status,
		}, "")
		if err != nil {
			t.Fatalf("status %q: UpdatePost returned error: %v", status, err)
		}
		data := updated.Data()
		if data.Status != domain.PostStatusReview || data.PublishedAt != nil {
			t.Fatalf("status %q: expected review without published_at, got %q / %v", status, data.Status, data.PublishedAt)
		}
	}
}

func TestAuthorStillCannotPublishADraft(t *testing.T) {
	draft := domain.ReconstitutePost(domain.PostData{
		ID: 1, Title: "Draft", Slug: "draft", Status: domain.PostStatusDraft, CategoryID: 1, AuthorID: 7,
	})
	useCase, _ := newUpdateUseCase(draft)

	_, err := useCase.UpdatePost(context.Background(), 1, 7, string(domain.UserRoleAuthor), CreatePostDTO{
		Title: "Draft", Slug: "draft", CategoryID: 1, Status: string(domain.PostStatusPublished),
	}, "")
	if !errors.Is(err, domain.ErrPublishNotAllowed) {
		t.Fatalf("expected ErrPublishNotAllowed, got %v", err)
	}
}

func TestEditingScheduledPostWithoutScheduledAtKeepsTheSchedule(t *testing.T) {
	at := time.Now().Add(time.Hour)
	scheduled := domain.ReconstitutePost(domain.PostData{
		ID: 1, Title: "Soon", Slug: "soon", Status: domain.PostStatusScheduled,
		ScheduledAt: &at, CategoryID: 1, AuthorID: 7,
	})
	useCase, _ := newUpdateUseCase(scheduled)

	updated, err := useCase.UpdatePost(context.Background(), 1, 9, string(domain.UserRoleEditor), CreatePostDTO{
		Title: "Soon, edited", Slug: "soon", CategoryID: 1,
	}, "")
	if err != nil {
		t.Fatalf("UpdatePost returned error: %v", err)
	}
	data := updated.Data()
	if data.Status != domain.PostStatusScheduled || data.ScheduledAt == nil || !data.ScheduledAt.Equal(at) {
		t.Fatalf("schedule was lost: %q / %v", data.Status, data.ScheduledAt)
	}
}

func TestEditorPublishingDirectlySetsPublishedAt(t *testing.T) {
	useCase, _ := newUpdateUseCase(nil)

	created, err := useCase.CreatePost(context.Background(), 9, string(domain.UserRoleEditor), CreatePostDTO{
		Title: "Live", Slug: "live", CategoryID: 1, Status: string(domain.PostStatusPublished),
	}, "")
	if err != nil {
		t.Fatalf("CreatePost returned error: %v", err)
	}
	if created.Data().PublishedAt == nil {
		t.Fatal("published_at must be set when a post is published directly")
	}

	draft := domain.ReconstitutePost(domain.PostData{
		ID: 2, Title: "Draft", Slug: "draft", Status: domain.PostStatusDraft, CategoryID: 1, AuthorID: 9,
	})
	useCase, _ = newUpdateUseCase(draft)
	updated, err := useCase.UpdatePost(context.Background(), 2, 9, string(domain.UserRoleEditor), CreatePostDTO{
		Title: "Draft", Slug: "draft", CategoryID: 1, Status: string(domain.PostStatusPublished),
	}, "")
	if err != nil {
		t.Fatalf("UpdatePost returned error: %v", err)
	}
	if updated.Data().PublishedAt == nil {
		t.Fatal("published_at must be set when a draft is published by an editor")
	}
}

func TestUpdatePostWithStatusChangesStatus(t *testing.T) {
	posts := &postUpdateRepository{post: domain.ReconstitutePost(domain.PostData{
		ID: 1, Title: "Original", Slug: "original", Content: "Original content",
		Status: domain.PostStatusPublished, CategoryID: 1, AuthorID: 7,
	})}
	useCase := NewPostUseCase(nil, postUpdateUnitOfWork{repositories: TransactionRepositories{
		Posts: posts, Revisions: postUpdateRevisionRepository{}, Audits: postUpdateAuditRepository{},
	}})

	updated, err := useCase.UpdatePost(context.Background(), 1, 7, string(domain.UserRoleAuthor), CreatePostDTO{
		Title: "Updated", Slug: "updated", Content: "Updated content",
		Status: string(domain.PostStatusDraft), CategoryID: 1,
	}, "")
	if err != nil {
		t.Fatalf("UpdatePost returned error: %v", err)
	}
	if got := updated.Data().Status; got != domain.PostStatusDraft {
		t.Fatalf("expected status to change to %q, got %q", domain.PostStatusDraft, got)
	}
}
