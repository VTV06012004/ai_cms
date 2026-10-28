package usecase

import (
	"testing"

	"api/internal/domain"
)

type postUpdateUnitOfWork struct {
	repositories TransactionRepositories
}

func (u postUpdateUnitOfWork) WithinTransaction(work func(TransactionRepositories) error) error {
	return work(u.repositories)
}

type postUpdateRepository struct {
	post *domain.Post
}

func (r *postUpdateRepository) GetAll(domain.PostListCriteria) ([]domain.Post, error) {
	return nil, nil
}

func (r *postUpdateRepository) GetByID(uint) (*domain.Post, error) {
	return r.post, nil
}

func (r *postUpdateRepository) Create(post *domain.Post, _ []uint) (*domain.Post, error) {
	r.post = post
	return post, nil
}

func (r *postUpdateRepository) Update(post *domain.Post, _ []uint) (*domain.Post, error) {
	r.post = post
	return post, nil
}

func (r *postUpdateRepository) Delete(uint) error {
	return nil
}

type postUpdateRevisionRepository struct{}

func (postUpdateRevisionRepository) Create(*domain.Revision) error { return nil }
func (postUpdateRevisionRepository) GetByPostID(uint) ([]domain.Revision, error) {
	return nil, nil
}
func (postUpdateRevisionRepository) GetByID(uint) (*domain.Revision, error) {
	return nil, nil
}

type postUpdateAuditRepository struct{}

func (postUpdateAuditRepository) RecordLog(uint, string, string, string, string) error {
	return nil
}
func (postUpdateAuditRepository) GetAll() ([]domain.AuditLog, error) { return nil, nil }

func TestUpdatePostWithoutStatusPreservesPublishedStatus(t *testing.T) {
	posts := &postUpdateRepository{post: domain.ReconstitutePost(domain.PostData{
		ID: 1, Title: "Original", Slug: "original", Content: "Original content",
		Status: domain.PostStatusPublished, CategoryID: 1, AuthorID: 7,
	})}
	useCase := NewPostUseCase(nil, postUpdateUnitOfWork{repositories: TransactionRepositories{
		Posts: posts, Revisions: postUpdateRevisionRepository{}, Audits: postUpdateAuditRepository{},
	}})

	updated, err := useCase.UpdatePost(1, 7, string(domain.UserRoleAuthor), CreatePostDTO{
		Title: "Updated", Slug: "updated", Content: "Updated content", CategoryID: 1,
	}, "")
	if err != nil {
		t.Fatalf("UpdatePost returned error: %v", err)
	}
	if got := updated.Data().Status; got != domain.PostStatusPublished {
		t.Fatalf("expected status to remain %q, got %q", domain.PostStatusPublished, got)
	}
	if got := updated.Data().Title; got != "Updated" {
		t.Fatalf("expected post content to update, got title %q", got)
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

	updated, err := useCase.UpdatePost(1, 7, string(domain.UserRoleAuthor), CreatePostDTO{
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
