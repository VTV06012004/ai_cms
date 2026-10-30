package persistence

import (
	"api/internal/domain"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) domain.PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) GetAll(criteria domain.PostListCriteria) ([]domain.Post, error) {
	var posts []PostModel
	query := r.db.Preload("Author").Preload("Category").Preload("Tags")
	if criteria.AuthorID > 0 {
		if criteria.IncludePublished {
			query = query.Where("author_id = ? OR status = ?", criteria.AuthorID, domain.PostStatusPublished)
		} else {
			query = query.Where("author_id = ?", criteria.AuthorID)
		}
	} else if criteria.PublishedOnly {
		query = query.Where("status = ?", domain.PostStatusPublished)
	}
	if err := query.Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Post, 0, len(posts))
	for _, post := range posts {
		result = append(result, post.ToDomain())
	}
	return result, nil
}

func (r *PostRepository) GetByID(id uint) (*domain.Post, error) {
	var post PostModel
	if err := r.db.Preload("Author").Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		return nil, err
	}
	result := post.ToDomain()
	return &result, nil
}

func (r *PostRepository) Create(post *domain.Post, tagIDs []uint) (*domain.Post, error) {
	var result *domain.Post
	err := r.db.Transaction(func(tx *gorm.DB) error {
		model := postModelFromDomain(post)
		if err := tx.Create(&model).Error; err != nil {
			return err
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
			return err
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

func replacePostTags(tx *gorm.DB, post *PostModel, tagIDs []uint) error {
	var tags []TagModel
	if len(tagIDs) > 0 {
		if err := tx.Find(&tags, tagIDs).Error; err != nil {
			return err
		}
	}
	return tx.Model(post).Association("Tags").Replace(tags)
}
