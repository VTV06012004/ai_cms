package usecase

import (
	"fmt"

	"api/internal/domain"
)

type TagUseCase struct {
	tagRepo    domain.TagRepository
	unitOfWork UnitOfWork
}

func NewTagUseCase(repo domain.TagRepository, unitOfWork UnitOfWork) *TagUseCase {
	return &TagUseCase{tagRepo: repo, unitOfWork: unitOfWork}
}

func (uc *TagUseCase) GetAllTags() ([]domain.Tag, error) {
	return uc.tagRepo.GetAll()
}

func (uc *TagUseCase) CreateTag(userID uint, name, slug, ipAddress string) (domain.Tag, error) {
	tag, err := domain.NewTag(name, slug)
	if err != nil {
		return domain.Tag{}, err
	}
	var created domain.Tag
	err = uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		saved, err := repositories.Tags.Create(tag)
		if err != nil {
			return err
		}
		created = saved
		data := saved.Data()
		return recordCatalogAudit(repositories, userID, "CREATE_TAG",
			fmt.Sprintf("Thẻ ID: %d", data.ID), fmt.Sprintf("Đã tạo thẻ: '%s'", data.Name), ipAddress)
	})
	if err != nil {
		return domain.Tag{}, err
	}
	return created, nil
}

func (uc *TagUseCase) UpdateTag(userID, id uint, name, slug, ipAddress string) (domain.Tag, error) {
	var updated domain.Tag
	err := uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		current, err := repositories.Tags.GetByID(id)
		if err != nil {
			return err
		}
		renamed, err := current.Rename(name, slug)
		if err != nil {
			return err
		}
		saved, err := repositories.Tags.Update(renamed)
		if err != nil {
			return err
		}
		updated = saved
		return recordCatalogAudit(repositories, userID, "UPDATE_TAG",
			fmt.Sprintf("Thẻ ID: %d", id), fmt.Sprintf("Đã cập nhật thẻ: '%s'", saved.Data().Name), ipAddress)
	})
	if err != nil {
		return domain.Tag{}, err
	}
	return updated, nil
}

func (uc *TagUseCase) DeleteTag(userID, id uint, ipAddress string) error {
	return uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		tag, err := repositories.Tags.GetByID(id)
		if err != nil {
			return err
		}
		if err := repositories.Tags.Delete(id); err != nil {
			return err
		}
		return recordCatalogAudit(repositories, userID, "DELETE_TAG",
			fmt.Sprintf("Thẻ ID: %d", id), fmt.Sprintf("Đã xóa thẻ: '%s'", tag.Data().Name), ipAddress)
	})
}
