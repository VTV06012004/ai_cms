package usecase

import "api/internal/domain"

type TagUseCase struct {
	tagRepo domain.TagRepository
}

func NewTagUseCase(repo domain.TagRepository) *TagUseCase {
	return &TagUseCase{tagRepo: repo}
}

func (uc *TagUseCase) GetAllTags() ([]domain.Tag, error) {
	return uc.tagRepo.GetAll()
}
