package usecase

import (
	"fmt"
	"time"

	"api/internal/domain"
)

type PostUseCase struct {
	postRepo   domain.PostRepository
	unitOfWork UnitOfWork
}

func NewPostUseCase(posts domain.PostRepository, unitOfWork UnitOfWork) *PostUseCase {
	return &PostUseCase{postRepo: posts, unitOfWork: unitOfWork}
}

type CreatePostDTO struct {
	Title      string
	Slug       string
	Content    string
	Status     string
	CategoryID uint
	TagIDs     []uint
}

func (uc *PostUseCase) GetAllPosts(userRole string, userID uint) ([]domain.Post, error) {
	criteria := domain.PostListCriteria{}
	if domain.UserRole(userRole) == domain.UserRoleAuthor && userID > 0 {
		criteria.AuthorID = userID
		criteria.IncludePublished = true
	}
	return uc.postRepo.GetAll(criteria)
}

func (uc *PostUseCase) GetPostByID(id uint) (*domain.Post, error) {
	return uc.postRepo.GetByID(id)
}

func (uc *PostUseCase) CreatePost(userID uint, userRole string, input CreatePostDTO, ipAddress string) (*domain.Post, error) {
	role := domain.UserRole(userRole)
	if role != domain.UserRoleAdmin && role != domain.UserRoleEditor && role != domain.UserRoleAuthor {
		return nil, domain.ErrInvalidRole
	}
	status := domain.PostStatus(input.Status)
	if status == "" {
		status = domain.PostStatusDraft
	}
	if status == domain.PostStatusPublished && !role.CanPublishDirectly() {
		return nil, domain.ErrPublishNotAllowed
	}
	if !role.CanSubmitPostStatus(status) {
		return nil, domain.ErrPostStatusNotAllowed
	}
	post, err := domain.NewPost(input.Title, input.Slug, input.Content, status, input.CategoryID, userID)
	if err != nil {
		return nil, err
	}
	err = uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		created, err := repositories.Posts.Create(post, input.TagIDs)
		if err != nil {
			return err
		}
		post = created
		data := post.Data()
		if err := repositories.Revisions.Create(snapshot(post, userID)); err != nil {
			return fmt.Errorf("không thể lưu revision ban đầu: %w", err)
		}
		if err := repositories.Audits.RecordLog(userID, "CREATE_POST", fmt.Sprintf("Bài viết ID: %d", data.ID), fmt.Sprintf("Đã tạo bài viết: '%s'", data.Title), ipAddress); err != nil {
			return fmt.Errorf("không thể ghi audit log: %w", err)
		}
		post, err = repositories.Posts.GetByID(data.ID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("không thể hoàn tất tạo bài viết: %w", err)
	}
	return post, nil
}

func (uc *PostUseCase) UpdatePost(id, userID uint, userRole string, input CreatePostDTO, ipAddress string) (*domain.Post, error) {
	role := domain.UserRole(userRole)
	var updated *domain.Post
	err := uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		post, err := repositories.Posts.GetByID(id)
		if err != nil {
			return domain.ErrPostNotFound
		}
		if !post.CanBeManagedBy(userID, role) {
			return domain.ErrPostAccessDenied
		}
		requestedStatus := domain.PostStatus(input.Status)
		if requestedStatus != "" {
			if requestedStatus == domain.PostStatusPublished && !role.CanPublishDirectly() {
				return domain.ErrPublishNotAllowed
			}
			if !role.CanSubmitPostStatus(requestedStatus) {
				return domain.ErrPostStatusNotAllowed
			}
		}
		data := post.Data()
		previous := snapshot(post, userID)
		if err := post.Update(input.Title, input.Slug, input.Content, domain.PostStatus(input.Status), input.CategoryID); err != nil {
			return err
		}
		if err := repositories.Revisions.Create(previous); err != nil {
			return fmt.Errorf("không thể lưu revision trước khi cập nhật: %w", err)
		}
		updatedPost, err := repositories.Posts.Update(post, input.TagIDs)
		if err != nil {
			return err
		}
		post = updatedPost
		data = post.Data()
		if err := repositories.Audits.RecordLog(userID, "UPDATE_POST", fmt.Sprintf("Bài viết ID: %d", data.ID), fmt.Sprintf("Đã cập nhật bài viết: '%s'", data.Title), ipAddress); err != nil {
			return fmt.Errorf("không thể ghi audit log: %w", err)
		}
		updated, err = repositories.Posts.GetByID(data.ID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("không thể hoàn tất cập nhật bài viết: %w", err)
	}
	return updated, nil
}

func (uc *PostUseCase) DeletePost(id, userID uint, userRole, ipAddress string) error {
	role := domain.UserRole(userRole)
	return uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		post, err := repositories.Posts.GetByID(id)
		if err != nil {
			return domain.ErrPostNotFound
		}
		if !post.CanBeManagedBy(userID, role) {
			return domain.ErrPostAccessDenied
		}
		data := post.Data()
		if err := repositories.Posts.Delete(id); err != nil {
			return err
		}
		return repositories.Audits.RecordLog(userID, "DELETE_POST", fmt.Sprintf("Bài viết ID: %d", id), fmt.Sprintf("Đã xóa bài viết: '%s'", data.Title), ipAddress)
	})
}

func (uc *PostUseCase) ModeratePost(id, moderatorID uint, status, reason, ipAddress string) (*domain.Post, error) {
	var moderated *domain.Post
	err := uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		post, err := repositories.Posts.GetByID(id)
		if err != nil {
			return domain.ErrPostNotFound
		}
		if err := post.Moderate(domain.PostStatus(status), reason, moderatorID, time.Now()); err != nil {
			return err
		}
		data := post.Data()
		updatedPost, err := repositories.Posts.Update(post, tagIDs(data.Tags))
		if err != nil {
			return err
		}
		post = updatedPost
		if err := repositories.Audits.RecordLog(moderatorID, "MODERATE_POST", fmt.Sprintf("Bài viết ID: %d", id), fmt.Sprintf("Đổi trạng thái bài viết thành '%s'", status), ipAddress); err != nil {
			return fmt.Errorf("không thể ghi audit log: %w", err)
		}
		moderated, err = repositories.Posts.GetByID(id)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("không thể hoàn tất kiểm duyệt: %w", err)
	}
	return moderated, nil
}

func snapshot(post *domain.Post, userID uint) *domain.Revision {
	return domain.NewRevision(post.Data(), userID)
}

func tagIDs(tags []domain.Tag) []uint {
	ids := make([]uint, 0, len(tags))
	for _, tag := range tags {
		ids = append(ids, tag.Data().ID)
	}
	return ids
}
