package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/VTV06012004/ai_cms/internal/domain"
)

type compareRevisionRepository struct {
	byID map[uint]*domain.Revision
}

func (compareRevisionRepository) Create(context.Context, *domain.Revision) error { return nil }

func (compareRevisionRepository) GetByPostID(context.Context, uint) ([]domain.Revision, error) {
	return nil, nil
}

func (r compareRevisionRepository) GetByID(_ context.Context, id uint) (*domain.Revision, error) {
	revision, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrRevisionNotFound
	}
	return revision, nil
}

func newCompareUseCase() *RevisionUseCase {
	post := domain.ReconstitutePost(domain.PostData{ID: 7, AuthorID: 22, Title: "Now", Content: "line1\nline2"})
	repo := compareRevisionRepository{byID: map[uint]*domain.Revision{
		1: domain.ReconstituteRevision(domain.RevisionData{ID: 1, PostID: 7, Title: "Before", Content: "line1"}),
		2: domain.ReconstituteRevision(domain.RevisionData{ID: 2, PostID: 7, Title: "Middle", Content: "line1\nline3"}),
		3: domain.ReconstituteRevision(domain.RevisionData{ID: 3, PostID: 99, Title: "Other post"}),
	}}
	return NewRevisionUseCase(repo, revisionPostRepositoryStub{post: post}, nil)
}

func TestCompareRevisionsWithCurrentPost(t *testing.T) {
	result, err := newCompareUseCase().CompareRevisions(context.Background(), 1, 0, 22, string(domain.UserRoleAuthor))
	if err != nil {
		t.Fatalf("CompareRevisions returned error: %v", err)
	}
	if result.To.Title != "Now" || result.Added != 1 || result.Removed != 0 {
		t.Fatalf("unexpected comparison: %+v", result)
	}
}

func TestCompareRevisionsRejectsOtherPostsAndOtherAuthors(t *testing.T) {
	useCase := newCompareUseCase()

	if _, err := useCase.CompareRevisions(context.Background(), 1, 3, 22, string(domain.UserRoleAuthor)); !errors.Is(err, domain.ErrRevisionPostMismatch) {
		t.Fatalf("expected ErrRevisionPostMismatch, got %v", err)
	}
	if _, err := useCase.CompareRevisions(context.Background(), 1, 2, 99, string(domain.UserRoleAuthor)); !errors.Is(err, domain.ErrPostAccessDenied) {
		t.Fatalf("expected ErrPostAccessDenied, got %v", err)
	}
	if _, err := useCase.CompareRevisions(context.Background(), 1, 404, 22, string(domain.UserRoleAuthor)); !errors.Is(err, domain.ErrRevisionNotFound) {
		t.Fatalf("expected ErrRevisionNotFound, got %v", err)
	}
}
