package usecase

import (
	"context"
	"fmt"
	"github.com/VTV06012004/ai_cms/internal/domain"
	"time"
)

type RevisionUseCase struct {
	revRepo    domain.RevisionRepository
	postRepo   domain.PostRepository
	unitOfWork UnitOfWork
}

func NewRevisionUseCase(revisions domain.RevisionRepository, posts domain.PostRepository, unitOfWork UnitOfWork) *RevisionUseCase {
	return &RevisionUseCase{revRepo: revisions, postRepo: posts, unitOfWork: unitOfWork}
}

func (uc *RevisionUseCase) GetByPostID(ctx context.Context, postID, userID uint, userRole string) ([]domain.Revision, error) {
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !post.CanBeManagedBy(userID, domain.UserRole(userRole)) {
		return nil, domain.ErrPostAccessDenied
	}
	return uc.revRepo.GetByPostID(ctx, postID)
}

func (uc *RevisionUseCase) RestoreRevision(ctx context.Context, revisionID, userID uint, userRole, ipAddress string) (*domain.Post, error) {
	var restored *domain.Post
	err := uc.unitOfWork.WithinTransaction(ctx, func(repositories TransactionRepositories) error {
		revision, err := repositories.Revisions.GetByID(ctx, revisionID)
		if err != nil {
			return err
		}
		revisionData := revision.Data()
		post, err := repositories.Posts.GetByID(ctx, revisionData.PostID)
		if err != nil {
			return err
		}
		if !post.CanBeManagedBy(userID, domain.UserRole(userRole)) {
			return domain.ErrPostAccessDenied
		}
		previous := snapshot(post, userID)
		postData := post.Data()
		if err := post.Update(revisionData.Title, revisionData.Slug, revisionData.Content, postData.Status, revisionData.CategoryID); err != nil {
			return fmt.Errorf("dữ liệu revision không hợp lệ: %w", err)
		}
		post.SettleStatus(domain.UserRole(userRole), time.Now())
		if err := repositories.Revisions.Create(ctx, previous); err != nil {
			return fmt.Errorf("không thể lưu snapshot trước khi khôi phục: %w", err)
		}
		updatedPost, err := repositories.Posts.Update(ctx, post, tagIDs(postData.Tags))
		if err != nil {
			return fmt.Errorf("không thể khôi phục phiên bản bài viết: %w", err)
		}
		post = updatedPost
		postData = post.Data()
		if err := repositories.Audits.RecordLog(ctx, userID, "RESTORE_REVISION", fmt.Sprintf("Bài viết ID: %d", postData.ID), fmt.Sprintf("Đã khôi phục về Revision ID: %d", revisionData.ID), ipAddress); err != nil {
			return fmt.Errorf("không thể ghi audit log: %w", err)
		}
		restored, err = repositories.Posts.GetByID(ctx, postData.ID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("không thể hoàn tất khôi phục revision: %w", err)
	}
	return restored, nil
}

// CompareRevisions so sánh phiên bản fromID với phiên bản toID; toID = 0 nghĩa là bản hiện tại của bài viết.
func (uc *RevisionUseCase) CompareRevisions(ctx context.Context, fromID, toID, userID uint, userRole string) (*domain.RevisionComparison, error) {
	from, err := uc.revRepo.GetByID(ctx, fromID)
	if err != nil {
		return nil, err
	}
	fromData := from.Data()
	post, err := uc.postRepo.GetByID(ctx, fromData.PostID)
	if err != nil {
		return nil, err
	}
	if !post.CanBeManagedBy(userID, domain.UserRole(userRole)) {
		return nil, domain.ErrPostAccessDenied
	}

	var toData domain.RevisionData
	if toID == 0 {
		toData = domain.RevisionDataFromPost(post.Data())
	} else {
		to, err := uc.revRepo.GetByID(ctx, toID)
		if err != nil {
			return nil, err
		}
		toData = to.Data()
		if toData.PostID != fromData.PostID {
			return nil, domain.ErrRevisionPostMismatch
		}
	}
	comparison := domain.CompareRevisions(fromData, toData)
	return &comparison, nil
}
