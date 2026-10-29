package usecase

import (
	"errors"
	"testing"

	"api/internal/domain"
)

type postRepositoryStub struct {
	criteria domain.PostListCriteria
	posts    []domain.Post
	post     *domain.Post
}

func (r *postRepositoryStub) GetAll(criteria domain.PostListCriteria) ([]domain.Post, error) {
	r.criteria = criteria
	return r.posts, nil
}

func (r *postRepositoryStub) GetByID(uint) (*domain.Post, error) {
	if r.post == nil {
		return nil, domain.ErrPostNotFound
	}
	return r.post, nil
}

func (r *postRepositoryStub) Create(*domain.Post, []uint) (*domain.Post, error) {
	return nil, nil
}

func (r *postRepositoryStub) Update(*domain.Post, []uint) (*domain.Post, error) {
	return nil, nil
}

func (r *postRepositoryStub) Delete(uint) error {
	return nil
}

func TestGetAllPostsRestrictsAnonymousUsersToPublished(t *testing.T) {
	repository := &postRepositoryStub{}
	useCase := NewPostUseCase(repository, nil)

	if _, err := useCase.GetAllPosts("", 0); err != nil {
		t.Fatalf("GetAllPosts returned error: %v", err)
	}
	if !repository.criteria.PublishedOnly {
		t.Fatal("anonymous post listings must be restricted to published posts")
	}
}

func TestGetAllPostsKeepsAuthenticatedVisibility(t *testing.T) {
	tests := []struct {
		name   string
		role   string
		userID uint
		want   domain.PostListCriteria
	}{
		{
			name:   "author sees own posts and published posts",
			role:   string(domain.UserRoleAuthor),
			userID: 12,
			want: domain.PostListCriteria{
				AuthorID:         12,
				IncludePublished: true,
			},
		},
		{
			name: "editor can see all posts",
			role: string(domain.UserRoleEditor),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &postRepositoryStub{}
			useCase := NewPostUseCase(repository, nil)

			if _, err := useCase.GetAllPosts(test.role, test.userID); err != nil {
				t.Fatalf("GetAllPosts returned error: %v", err)
			}
			if repository.criteria != test.want {
				t.Fatalf("criteria = %+v, want %+v", repository.criteria, test.want)
			}
		})
	}
}

func TestGetPostByIDHidesUnpublishedPostsFromUnauthorizedUsers(t *testing.T) {
	post := domain.ReconstitutePost(domain.PostData{
		ID: 1, Status: domain.PostStatusDraft, AuthorID: 12,
	})
	repository := &postRepositoryStub{post: post}
	useCase := NewPostUseCase(repository, nil)

	_, err := useCase.GetPostByID(1, "", 0)
	if !errors.Is(err, domain.ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound for anonymous draft access, got %v", err)
	}

	if _, err := useCase.GetPostByID(1, string(domain.UserRoleAuthor), 12); err != nil {
		t.Fatalf("author should be able to view their draft: %v", err)
	}
	if _, err := useCase.GetPostByID(1, string(domain.UserRoleEditor), 99); err != nil {
		t.Fatalf("editor should be able to view a draft: %v", err)
	}
}

func TestGetPostByIDAllowsAnonymousPublishedPosts(t *testing.T) {
	post := domain.ReconstitutePost(domain.PostData{
		ID: 1, Status: domain.PostStatusPublished, AuthorID: 12,
	})
	useCase := NewPostUseCase(&postRepositoryStub{post: post}, nil)

	if _, err := useCase.GetPostByID(1, "", 0); err != nil {
		t.Fatalf("anonymous published post access returned error: %v", err)
	}
}
