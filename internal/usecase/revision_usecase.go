package usecase

import (
	"fmt"

	"api/internal/domain"
)

type RevisionUseCase struct {
	revRepo    domain.RevisionRepository
	postRepo   domain.PostRepository
	unitOfWork UnitOfWork
}

func NewRevisionUseCase(revisions domain.RevisionRepository, posts domain.PostRepository, unitOfWork UnitOfWork) *RevisionUseCase {
	return &RevisionUseCase{revRepo: revisions, postRepo: posts, unitOfWork: unitOfWork}
}

func (uc *RevisionUseCase) GetByPostID(postID, userID uint, userRole string) ([]domain.Revision, error) {
	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return nil, domain.ErrPostNotFound
	}
	if !post.CanBeManagedBy(userID, domain.UserRole(userRole)) {
		return nil, domain.ErrPostAccessDenied
	}
	return uc.revRepo.GetByPostID(postID)
}

func (uc *RevisionUseCase) RestoreRevision(revisionID, userID uint, userRole, ipAddress string) (*domain.Post, error) {
	var restored *domain.Post
	err := uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		revision, err := repositories.Revisions.GetByID(revisionID)
		if err != nil {
			return fmt.Errorf("không tìm thấy bản revision này: %w", err)
		}
		revisionData := revision.Data()
		post, err := repositories.Posts.GetByID(revisionData.PostID)
		if err != nil {
			return domain.ErrPostNotFound
		}
		if !post.CanBeManagedBy(userID, domain.UserRole(userRole)) {
			return domain.ErrPostAccessDenied
		}
		previous := snapshot(post, userID)
		postData := post.Data()
		if err := post.Update(revisionData.Title, revisionData.Slug, revisionData.Content, postData.Status, revisionData.CategoryID); err != nil {
			return fmt.Errorf("dữ liệu revision không hợp lệ: %w", err)
		}
		if err := repositories.Revisions.Create(previous); err != nil {
			return fmt.Errorf("không thể lưu snapshot trước khi khôi phục: %w", err)
		}
		updatedPost, err := repositories.Posts.Update(post, tagIDs(postData.Tags))
		if err != nil {
			return fmt.Errorf("không thể khôi phục phiên bản bài viết: %w", err)
		}
		post = updatedPost
		postData = post.Data()
		if err := repositories.Audits.RecordLog(userID, "RESTORE_REVISION", fmt.Sprintf("Bài viết ID: %d", postData.ID), fmt.Sprintf("Đã khôi phục về Revision ID: %d", revisionData.ID), ipAddress); err != nil {
			return fmt.Errorf("không thể ghi audit log: %w", err)
		}
		restored, err = repositories.Posts.GetByID(postData.ID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("không thể hoàn tất khôi phục revision: %w", err)
	}
	return restored, nil
}
