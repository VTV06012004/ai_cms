package usecase

import (
	"api/internal/domain"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type PostUseCase struct {
	postRepo   domain.PostRepository
	unitOfWork UnitOfWork
}

func NewPostUseCase(posts domain.PostRepository, unitOfWork UnitOfWork) *PostUseCase {
	return &PostUseCase{postRepo: posts, unitOfWork: unitOfWork}
}

type CreatePostDTO struct {
	Title       string
	Slug        string
	Content     string
	Status      string
	CategoryID  uint
	TagIDs      []uint
	ScheduledAt *time.Time
}

func (uc *PostUseCase) GetAllPosts(ctx context.Context, userRole string, userID uint, page domain.PageRequest) ([]domain.Post, domain.PageInfo, error) {
	criteria := domain.PostListCriteria{PublishedOnly: true}
	switch domain.UserRole(userRole) {
	case domain.UserRoleAdmin, domain.UserRoleEditor:
		criteria.PublishedOnly = false
	case domain.UserRoleAuthor:
		if userID == 0 {
			break
		}
		criteria.AuthorID = userID
		criteria.IncludePublished = true
		criteria.PublishedOnly = false
	}
	posts, total, err := uc.postRepo.GetAll(ctx, criteria, page)
	if err != nil {
		return nil, domain.PageInfo{}, err
	}
	return posts, page.Info(total), nil
}

func (uc *PostUseCase) GetPostByID(ctx context.Context, id uint, userRole string, userID uint) (*domain.Post, error) {
	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.Data().Status != domain.PostStatusPublished &&
		!post.CanBeManagedBy(userID, domain.UserRole(userRole)) {
		return nil, domain.ErrPostNotFound
	}
	return post, nil
}

func (uc *PostUseCase) CreatePost(ctx context.Context, userID uint, userRole string, input CreatePostDTO, ipAddress string) (*domain.Post, error) {
	role := domain.UserRole(userRole)
	if !role.IsValid() {
		return nil, domain.ErrInvalidRole
	}
	status := domain.PostStatus(input.Status)
	if status == "" {
		status = domain.PostStatusDraft
	}
	if err := role.EnsureCanSetPostStatus(status); err != nil {
		return nil, err
	}
	post, err := domain.NewPost(input.Title, input.Slug, input.Content, status, input.CategoryID, userID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if err := post.ApplySchedule(input.ScheduledAt, now); err != nil {
		return nil, err
	}
	post.SettleStatus(role, now)
	err = uc.unitOfWork.WithinTransaction(ctx, func(repositories TransactionRepositories) error {
		created, err := repositories.Posts.Create(ctx, post, input.TagIDs)
		if err != nil {
			return err
		}
		post = created
		data := post.Data()
		if err := repositories.Revisions.Create(ctx, snapshot(post, userID)); err != nil {
			return fmt.Errorf("không thể lưu revision ban đầu: %w", err)
		}
		if err := repositories.Audits.RecordLog(ctx, userID, "CREATE_POST", fmt.Sprintf("Bài viết ID: %d", data.ID), fmt.Sprintf("Đã tạo bài viết: '%s'", data.Title), ipAddress); err != nil {
			return fmt.Errorf("không thể ghi audit log: %w", err)
		}
		post, err = repositories.Posts.GetByID(ctx, data.ID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("không thể hoàn tất tạo bài viết: %w", err)
	}
	return post, nil
}

func (uc *PostUseCase) UpdatePost(ctx context.Context, id, userID uint, userRole string, input CreatePostDTO, ipAddress string) (*domain.Post, error) {
	role := domain.UserRole(userRole)
	var updated *domain.Post
	err := uc.unitOfWork.WithinTransaction(ctx, func(repositories TransactionRepositories) error {
		post, err := repositories.Posts.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if !post.CanBeManagedBy(userID, role) {
			return domain.ErrPostAccessDenied
		}
		requestedStatus := domain.PostStatus(input.Status)
		// Chỉ kiểm quyền khi trạng thái thực sự thay đổi; form sửa bài luôn gửi lại trạng thái hiện tại.
		if requestedStatus != "" && requestedStatus != post.Data().Status {
			if err := role.EnsureCanSetPostStatus(requestedStatus); err != nil {
				return err
			}
		}
		data := post.Data()
		previous := snapshot(post, userID)
		if err := post.Update(input.Title, input.Slug, input.Content, domain.PostStatus(input.Status), input.CategoryID); err != nil {
			return err
		}
		now := time.Now()
		if err := post.ApplySchedule(input.ScheduledAt, now); err != nil {
			return err
		}
		post.SettleStatus(role, now)
		if err := repositories.Revisions.Create(ctx, previous); err != nil {
			return fmt.Errorf("không thể lưu revision trước khi cập nhật: %w", err)
		}
		updatedPost, err := repositories.Posts.Update(ctx, post, input.TagIDs)
		if err != nil {
			return err
		}
		post = updatedPost
		data = post.Data()
		if err := repositories.Audits.RecordLog(ctx, userID, "UPDATE_POST", fmt.Sprintf("Bài viết ID: %d", data.ID), fmt.Sprintf("Đã cập nhật bài viết: '%s'", data.Title), ipAddress); err != nil {
			return fmt.Errorf("không thể ghi audit log: %w", err)
		}
		updated, err = repositories.Posts.GetByID(ctx, data.ID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("không thể hoàn tất cập nhật bài viết: %w", err)
	}
	return updated, nil
}

func (uc *PostUseCase) DeletePost(ctx context.Context, id, userID uint, userRole, ipAddress string) error {
	role := domain.UserRole(userRole)
	return uc.unitOfWork.WithinTransaction(ctx, func(repositories TransactionRepositories) error {
		post, err := repositories.Posts.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if !post.CanBeManagedBy(userID, role) {
			return domain.ErrPostAccessDenied
		}
		data := post.Data()
		if err := repositories.Posts.Delete(ctx, id); err != nil {
			return err
		}
		return repositories.Audits.RecordLog(ctx, userID, "DELETE_POST", fmt.Sprintf("Bài viết ID: %d", id), fmt.Sprintf("Đã xóa bài viết: '%s'", data.Title), ipAddress)
	})
}

func (uc *PostUseCase) ModeratePost(ctx context.Context, id, moderatorID uint, status, reason, ipAddress string) (*domain.Post, error) {
	var moderated *domain.Post
	err := uc.unitOfWork.WithinTransaction(ctx, func(repositories TransactionRepositories) error {
		post, err := repositories.Posts.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := post.Moderate(domain.PostStatus(status), reason, moderatorID, time.Now()); err != nil {
			return err
		}
		data := post.Data()
		if _, err := repositories.Posts.Update(ctx, post, tagIDs(data.Tags)); err != nil {
			return err
		}
		if err := repositories.Audits.RecordLog(ctx, moderatorID, "MODERATE_POST", fmt.Sprintf("Bài viết ID: %d", id), fmt.Sprintf("Đổi trạng thái bài viết thành '%s'", status), ipAddress); err != nil {
			return fmt.Errorf("không thể ghi audit log: %w", err)
		}
		moderated, err = repositories.Posts.GetByID(ctx, id)
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

// PublishDuePosts đăng các bài hẹn giờ đã đến hạn; trả về số bài do chính lần gọi này đăng.
//
// Danh sách ListDue chỉ là danh sách ứng viên: có thể hai instance cùng thấy một bài. Người
// thắng được quyết định bởi PublishIfDue (một UPDATE có điều kiện); chỉ khi nó trả về true
// mới ghi audit, nên bài không bao giờ bị đăng hay ghi log hai lần.
func (uc *PostUseCase) PublishDuePosts(ctx context.Context, now time.Time) (int, error) {
	due, err := uc.postRepo.ListDue(ctx, now, 100)
	if err != nil {
		return 0, err
	}
	published := 0
	for _, candidate := range due {
		data := candidate.Data()
		won := false
		err := uc.unitOfWork.WithinTransaction(ctx, func(repositories TransactionRepositories) error {
			ok, err := repositories.Posts.PublishIfDue(ctx, data.ID, now)
			if err != nil || !ok {
				return err
			}
			won = true
			return repositories.Audits.RecordLog(ctx, data.AuthorID, "AUTO_PUBLISH_POST",
				fmt.Sprintf("Bài viết ID: %d", data.ID),
				fmt.Sprintf("Tự động đăng bài hẹn giờ: '%s'", data.Title), "system")
		})
		if err != nil {
			slog.Error("scheduled post publish failed", "post_id", data.ID, "error", err)
			continue
		}
		if won {
			published++
		}
	}
	return published, nil
}
