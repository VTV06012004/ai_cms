package persistence

import (
	"api/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) domain.PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) GetAll(ctx context.Context, criteria domain.PostListCriteria, page domain.PageRequest) ([]domain.Post, int64, error) {
	filtered := r.db.WithContext(ctx).Model(&PostModel{})
	if criteria.AuthorID > 0 {
		if criteria.IncludePublished {
			filtered = filtered.Where("author_id = ? OR status = ?", criteria.AuthorID, domain.PostStatusPublished)
		} else {
			filtered = filtered.Where("author_id = ?", criteria.AuthorID)
		}
	} else if criteria.PublishedOnly {
		filtered = filtered.Where("status = ?", domain.PostStatusPublished)
	}
	// Session để dùng lại cùng bộ lọc cho cả Count và Find.
	filtered = filtered.Session(&gorm.Session{})

	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var posts []PostModel
	err := filtered.Preload("Author").Preload("Category").Preload("Tags").
		Order("created_at DESC, id DESC").
		Limit(page.PerPage).Offset(page.Offset()).
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}
	result := make([]domain.Post, 0, len(posts))
	for _, post := range posts {
		result = append(result, post.ToDomain())
	}
	return result, total, nil
}

func (r *PostRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	var post PostModel
	if err := r.db.WithContext(ctx).Preload("Author").Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		return nil, notFoundOr(err, domain.ErrPostNotFound)
	}
	result := post.ToDomain()
	return &result, nil
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post, tagIDs []uint) (*domain.Post, error) {
	var result *domain.Post
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := postModelFromDomain(post)
		if err := tx.Create(&model).Error; err != nil {
			return translatePostWriteError(err)
		}
		if err := replacePostTags(tx, &model, tagIDs); err != nil {
			return err
		}
		loaded := model.ToDomain()
		result = &loaded
		return nil
	})
	return result, err
}

func (r *PostRepository) Update(ctx context.Context, post *domain.Post, tagIDs []uint) (*domain.Post, error) {
	var result *domain.Post
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := postModelFromDomain(post)
		if err := tx.Save(&model).Error; err != nil {
			return translatePostWriteError(err)
		}
		if err := replacePostTags(tx, &model, tagIDs); err != nil {
			return err
		}
		loaded := model.ToDomain()
		result = &loaded
		return nil
	})
	return result, err
}

func (r *PostRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&PostModel{}, id).Error
}

var _ domain.PostRepository = (*PostRepository)(nil)

func translatePostWriteError(err error) error {
	var mysqlError *mysqlDriver.MySQLError
	if errors.As(err, &mysqlError) &&
		mysqlError.Number == 1062 &&
		strings.Contains(mysqlError.Message, "idx_posts_active_slug") {
		return fmt.Errorf("%w: %v", domain.ErrPostSlugAlreadyExists, err)
	}
	return err
}

func replacePostTags(tx *gorm.DB, post *PostModel, tagIDs []uint) error {
	var tags []TagModel
	if len(tagIDs) > 0 {
		if err := tx.Find(&tags, tagIDs).Error; err != nil {
			return err
		}
	}
	return tx.Model(post).Association("Tags").Replace(tags)
}

func (r *PostRepository) ListDue(ctx context.Context, now time.Time, limit int) ([]domain.Post, error) {
	var models []PostModel
	err := r.db.WithContext(ctx).Where("status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?", domain.PostStatusScheduled, now).
		Preload("Tags").
		Order("scheduled_at ASC, id ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	result := make([]domain.Post, 0, len(models))
	for _, model := range models {
		result = append(result, model.ToDomain())
	}
	return result, nil
}

// PublishIfDue đăng bài hẹn giờ đã đến hạn bằng MỘT câu UPDATE có điều kiện, nên khi nhiều
// instance (hoặc nhiều lần quét) cùng thấy một bài thì chỉ một lệnh thắng: các lệnh còn lại
// thấy bài không còn ở trạng thái hẹn giờ và cập nhật 0 dòng.
//
// published_at = scheduled_at phải đứng trước scheduled_at = NULL: MySQL đánh giá các phép
// gán của UPDATE một bảng từ trái sang phải. Quy tắc này trùng với Post.PublishScheduled.
func (r *PostRepository) PublishIfDue(ctx context.Context, id uint, now time.Time) (bool, error) {
	result := r.db.WithContext(ctx).Exec(
		"UPDATE posts SET status = ?, published_at = scheduled_at, scheduled_at = NULL, updated_at = ? "+
			"WHERE id = ? AND status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ? AND deleted_at IS NULL",
		domain.PostStatusPublished, now, id, domain.PostStatusScheduled, now,
	)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}
