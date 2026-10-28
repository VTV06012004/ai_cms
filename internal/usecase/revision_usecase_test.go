package usecase

import (
	"errors"
	"testing"

	"api/internal/domain"
)

type revisionPostRepositoryStub struct {
	post *domain.Post
	err  error
}

func (r revisionPostRepositoryStub) GetAll(domain.PostListCriteria) ([]domain.Post, error) {
	return nil, nil
}

func (r revisionPostRepositoryStub) GetByID(uint) (*domain.Post, error) {
	return r.post, r.err
}

func (revisionPostRepositoryStub) Create(*domain.Post, []uint) (*domain.Post, error) {
	return nil, nil
}

func (revisionPostRepositoryStub) Update(*domain.Post, []uint) (*domain.Post, error) {
	return nil, nil
}

func (revisionPostRepositoryStub) Delete(uint) error {
	return nil
}

type revisionRepositoryStub struct {
	revisions []domain.Revision
	calls     int
}

func (r *revisionRepositoryStub) GetByPostID(uint) ([]domain.Revision, error) {
	r.calls++
	return r.revisions, nil
}

func (*revisionRepositoryStub) Create(*domain.Revision) error {
	return nil
}

func (*revisionRepositoryStub) GetByID(uint) (*domain.Revision, error) {
	return nil, nil
}

func TestGetByPostIDAllowsPostOwnerAndEditors(t *testing.T) {
	post := domain.ReconstitutePost(domain.PostData{ID: 7, AuthorID: 22})
	revisions := []domain.Revision{*domain.ReconstituteRevision(domain.RevisionData{
		ID: 3, PostID: 7, Content: "private version",
	})}
	tests := []struct {
		name   string
		userID uint
		role   string
	}{
		{name: "owner", userID: 22, role: string(domain.UserRoleAuthor)},
		{name: "editor", userID: 41, role: string(domain.UserRoleEditor)},
		{name: "admin", userID: 42, role: string(domain.UserRoleAdmin)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			revisionRepo := &revisionRepositoryStub{revisions: revisions}
			useCase := NewRevisionUseCase(revisionRepo, revisionPostRepositoryStub{post: post}, nil)

			got, err := useCase.GetByPostID(7, test.userID, test.role)
			if err != nil {
				t.Fatalf("GetByPostID returned error: %v", err)
			}
			if len(got) != 1 || got[0].Data().Content != "private version" {
				t.Fatalf("unexpected revisions returned: %+v", got)
			}
			if revisionRepo.calls != 1 {
				t.Fatalf("revision repository calls = %d, want 1", revisionRepo.calls)
			}
		})
	}
}

func TestGetByPostIDRejectsAuthorChangingPostID(t *testing.T) {
	post := domain.ReconstitutePost(domain.PostData{ID: 7, AuthorID: 22})
	revisionRepo := &revisionRepositoryStub{
		revisions: []domain.Revision{*domain.ReconstituteRevision(domain.RevisionData{
			PostID: 7, Content: "private version",
		})},
	}
	useCase := NewRevisionUseCase(revisionRepo, revisionPostRepositoryStub{post: post}, nil)

	got, err := useCase.GetByPostID(7, 99, string(domain.UserRoleAuthor))
	if !errors.Is(err, domain.ErrPostAccessDenied) {
		t.Fatalf("expected ErrPostAccessDenied, got %v", err)
	}
	if got != nil {
		t.Fatalf("denied request returned revision content: %+v", got)
	}
	if revisionRepo.calls != 0 {
		t.Fatalf("revision repository queried %d times after denied access", revisionRepo.calls)
	}
}

func TestGetByPostIDReturnsNotFoundWithoutQueryingRevisions(t *testing.T) {
	revisionRepo := &revisionRepositoryStub{}
	useCase := NewRevisionUseCase(
		revisionRepo,
		revisionPostRepositoryStub{err: errors.New("database failure")},
		nil,
	)

	got, err := useCase.GetByPostID(7, 99, string(domain.UserRoleAuthor))
	if !errors.Is(err, domain.ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
	if got != nil || revisionRepo.calls != 0 {
		t.Fatal("revision data was returned or queried for a missing post")
	}
}
