package usecase

import (
	"errors"
	"testing"

	"api/internal/domain"
)

// permissionPostRepository dùng lại stub của test cập nhật bài viết,
// thêm khả năng giả lập lỗi khi đọc và ghi nhận việc xóa.
type permissionPostRepository struct {
	postUpdateRepository
	getErr  error
	deleted bool
}

func (r *permissionPostRepository) GetByID(id uint) (*domain.Post, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.postUpdateRepository.GetByID(id)
}

func (r *permissionPostRepository) Delete(uint) error {
	r.deleted = true
	return nil
}

func newPermissionUseCase(posts *permissionPostRepository, audits *auditRepositoryStub) *PostUseCase {
	return NewPostUseCase(posts, postUpdateUnitOfWork{repositories: TransactionRepositories{
		Posts: posts, Revisions: postUpdateRevisionRepository{}, Audits: audits,
	}})
}

func permissionTestPost(status domain.PostStatus, authorID uint) *domain.Post {
	return domain.ReconstitutePost(domain.PostData{
		ID: 1, Title: "Original", Slug: "original", Content: "content",
		Status: status, CategoryID: 1, AuthorID: authorID,
	})
}

func TestCreatePostEnforcesRolePermissions(t *testing.T) {
	tests := []struct {
		name    string
		role    domain.UserRole
		status  domain.PostStatus
		wantErr error
	}{
		{"author can save a draft", domain.UserRoleAuthor, domain.PostStatusDraft, nil},
		{"author can submit for review", domain.UserRoleAuthor, domain.PostStatusReview, nil},
		{"author cannot publish", domain.UserRoleAuthor, domain.PostStatusPublished, domain.ErrPublishNotAllowed},
		{"author cannot set rejected", domain.UserRoleAuthor, domain.PostStatusRejected, domain.ErrPostStatusNotAllowed},
		{"editor can publish", domain.UserRoleEditor, domain.PostStatusPublished, nil},
		{"admin can publish", domain.UserRoleAdmin, domain.PostStatusPublished, nil},
		{"unknown role is rejected", domain.UserRole("guest"), domain.PostStatusDraft, domain.ErrInvalidRole},
		{"empty role is rejected", domain.UserRole(""), domain.PostStatusDraft, domain.ErrInvalidRole},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			posts := &permissionPostRepository{}
			useCase := newPermissionUseCase(posts, &auditRepositoryStub{})

			_, err := useCase.CreatePost(7, string(test.role), CreatePostDTO{
				Title: "Title", Slug: "slug", Content: "content",
				Status: string(test.status), CategoryID: 1,
			}, "")

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil && posts.post != nil {
				t.Fatal("a post was stored even though the request was denied")
			}
		})
	}
}

func TestCreatePostWithoutStatusDefaultsToDraft(t *testing.T) {
	posts := &permissionPostRepository{}
	useCase := newPermissionUseCase(posts, &auditRepositoryStub{})

	created, err := useCase.CreatePost(7, string(domain.UserRoleAuthor), CreatePostDTO{
		Title: "Title", Slug: "slug", Content: "content", CategoryID: 1,
	}, "")
	if err != nil {
		t.Fatalf("CreatePost returned error: %v", err)
	}
	if got := created.Data().Status; got != domain.PostStatusDraft {
		t.Fatalf("status = %q, want draft", got)
	}
}

func TestUpdatePostRequiresOwnershipForAuthors(t *testing.T) {
	tests := []struct {
		name    string
		userID  uint
		role    domain.UserRole
		wantErr error
	}{
		{"another author is denied", 8, domain.UserRoleAuthor, domain.ErrPostAccessDenied},
		{"the owner can update", 7, domain.UserRoleAuthor, nil},
		{"an editor can update any post", 8, domain.UserRoleEditor, nil},
		{"an admin can update any post", 9, domain.UserRoleAdmin, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			posts := &permissionPostRepository{postUpdateRepository: postUpdateRepository{
				post: permissionTestPost(domain.PostStatusDraft, 7),
			}}
			useCase := newPermissionUseCase(posts, &auditRepositoryStub{})

			_, err := useCase.UpdatePost(1, test.userID, string(test.role), CreatePostDTO{
				Title: "Updated", Slug: "updated", Content: "content", CategoryID: 1,
			}, "")

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil && posts.post.Data().Title != "Original" {
				t.Fatal("post was modified even though access was denied")
			}
		})
	}
}

func TestUpdatePostRejectsAuthorPublishingOwnPost(t *testing.T) {
	posts := &permissionPostRepository{postUpdateRepository: postUpdateRepository{
		post: permissionTestPost(domain.PostStatusDraft, 7),
	}}
	useCase := newPermissionUseCase(posts, &auditRepositoryStub{})

	_, err := useCase.UpdatePost(1, 7, string(domain.UserRoleAuthor), CreatePostDTO{
		Title: "Updated", Slug: "updated", Content: "content",
		Status: string(domain.PostStatusPublished), CategoryID: 1,
	}, "")

	if !errors.Is(err, domain.ErrPublishNotAllowed) {
		t.Fatalf("error = %v, want ErrPublishNotAllowed", err)
	}
	if got := posts.post.Data().Status; got != domain.PostStatusDraft {
		t.Fatalf("status changed to %q despite denial", got)
	}
}

func TestDeletePostChecksAccessBeforeDeleting(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		role        domain.UserRole
		wantErr     error
		wantDeleted bool
	}{
		{"another author is denied", 8, domain.UserRoleAuthor, domain.ErrPostAccessDenied, false},
		{"the owner can delete", 7, domain.UserRoleAuthor, nil, true},
		{"an admin can delete any post", 9, domain.UserRoleAdmin, nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			posts := &permissionPostRepository{postUpdateRepository: postUpdateRepository{
				post: permissionTestPost(domain.PostStatusPublished, 7),
			}}
			audits := &auditRepositoryStub{}
			useCase := newPermissionUseCase(posts, audits)

			err := useCase.DeletePost(1, test.userID, string(test.role), "")

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if posts.deleted != test.wantDeleted {
				t.Fatalf("deleted = %v, want %v", posts.deleted, test.wantDeleted)
			}
			if test.wantDeleted && (len(audits.actions) != 1 || audits.actions[0] != "DELETE_POST") {
				t.Fatalf("audit actions = %v, want [DELETE_POST]", audits.actions)
			}
			if !test.wantDeleted && len(audits.actions) != 0 {
				t.Fatalf("audit recorded for a denied request: %v", audits.actions)
			}
		})
	}
}

func TestDeletePostDistinguishesMissingPostFromDatabaseFailure(t *testing.T) {
	missing := &permissionPostRepository{getErr: domain.ErrPostNotFound}
	err := newPermissionUseCase(missing, &auditRepositoryStub{}).DeletePost(1, 7, string(domain.UserRoleAdmin), "")
	if !errors.Is(err, domain.ErrPostNotFound) {
		t.Fatalf("error = %v, want ErrPostNotFound", err)
	}

	databaseFailure := errors.New("database is down")
	broken := &permissionPostRepository{getErr: databaseFailure}
	err = newPermissionUseCase(broken, &auditRepositoryStub{}).DeletePost(1, 7, string(domain.UserRoleAdmin), "")
	if !errors.Is(err, databaseFailure) || errors.Is(err, domain.ErrPostNotFound) {
		t.Fatalf("database failure was reported as %v", err)
	}
}

func TestModeratePostRules(t *testing.T) {
	t.Run("only posts in review can be moderated", func(t *testing.T) {
		posts := &permissionPostRepository{postUpdateRepository: postUpdateRepository{
			post: permissionTestPost(domain.PostStatusDraft, 7),
		}}
		useCase := newPermissionUseCase(posts, &auditRepositoryStub{})

		_, err := useCase.ModeratePost(1, 5, string(domain.PostStatusPublished), "", "")
		if !errors.Is(err, domain.ErrPostNotInReview) {
			t.Fatalf("error = %v, want ErrPostNotInReview", err)
		}
	})

	t.Run("rejecting needs a reason", func(t *testing.T) {
		posts := &permissionPostRepository{postUpdateRepository: postUpdateRepository{
			post: permissionTestPost(domain.PostStatusReview, 7),
		}}
		useCase := newPermissionUseCase(posts, &auditRepositoryStub{})

		_, err := useCase.ModeratePost(1, 5, string(domain.PostStatusRejected), "", "")
		if !errors.Is(err, domain.ErrRejectionReasonRequired) {
			t.Fatalf("error = %v, want ErrRejectionReasonRequired", err)
		}
	})

	t.Run("other statuses are not a moderation result", func(t *testing.T) {
		posts := &permissionPostRepository{postUpdateRepository: postUpdateRepository{
			post: permissionTestPost(domain.PostStatusReview, 7),
		}}
		useCase := newPermissionUseCase(posts, &auditRepositoryStub{})

		_, err := useCase.ModeratePost(1, 5, string(domain.PostStatusDraft), "", "")
		if !errors.Is(err, domain.ErrInvalidPostStatus) {
			t.Fatalf("error = %v, want ErrInvalidPostStatus", err)
		}
	})

	t.Run("publishing records the moderator and an audit entry", func(t *testing.T) {
		posts := &permissionPostRepository{postUpdateRepository: postUpdateRepository{
			post: permissionTestPost(domain.PostStatusReview, 7),
		}}
		audits := &auditRepositoryStub{}
		useCase := newPermissionUseCase(posts, audits)

		moderated, err := useCase.ModeratePost(1, 5, string(domain.PostStatusPublished), "", "")
		if err != nil {
			t.Fatalf("ModeratePost returned error: %v", err)
		}
		data := moderated.Data()
		if data.Status != domain.PostStatusPublished {
			t.Fatalf("status = %q, want published", data.Status)
		}
		if data.ModeratorID == nil || *data.ModeratorID != 5 {
			t.Fatalf("moderator = %v, want 5", data.ModeratorID)
		}
		if len(audits.actions) != 1 || audits.actions[0] != "MODERATE_POST" {
			t.Fatalf("audit actions = %v, want [MODERATE_POST]", audits.actions)
		}
	})
}
