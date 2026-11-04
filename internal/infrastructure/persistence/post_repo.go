package persistence

import (
	"api/internal/domain"
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

func (r *PostRepository) GetAll(criteria domain.PostListCriteria, page domain.PageRequest) ([]domain.Post, int64, error) {
	filtered := r.db.Model(&PostModel{})
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

func (r *PostRepository) GetByID(id uint) (*domain.Post, error) {
	var post PostModel
	if err := r.db.Preload("Author").Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		return nil, notFoundOr(err, domain.ErrPostNotFound)
	}
	result := post.ToDomain()
	return &result, nil
}

func (r *PostRepository) Create(post *domain.Post, tagIDs []uint) (*domain.Post, error) {
	var result *domain.Post
	err := r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *PostRepository) Update(post *domain.Post, tagIDs []uint) (*domain.Post, error) {
	var result *domain.Post
	err := r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *PostRepository) Delete(id uint) error {
	return r.db.Delete(&PostModel{}, id).Error
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

func (r *PostRepository) ListDue(now time.Time, limit int) ([]domain.Post, error) {
	var models []PostModel
	err := r.db.Where("status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?", domain.PostStatusScheduled, now).
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
