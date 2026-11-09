package usecase

import (
	"context"
	"errors"
	"testing"

	"api/internal/domain"
)

type categoryRepositoryStub struct {
	items  map[uint]*domain.Category
	nextID uint
	posts  map[uint]int64
}

func newCategoryRepositoryStub() *categoryRepositoryStub {
	return &categoryRepositoryStub{
		items: map[uint]*domain.Category{}, nextID: 1, posts: map[uint]int64{},
	}
}

func (r *categoryRepositoryStub) add(t *testing.T, name, slug string, parentID *uint) uint {
	t.Helper()
	category, err := domain.NewCategory(name, slug, parentID)
	if err != nil {
		t.Fatal(err)
	}
	created, err := r.Create(context.Background(), category)
	if err != nil {
		t.Fatal(err)
	}
	return created.Data().ID
}

func (r *categoryRepositoryStub) GetAll(context.Context) ([]domain.Category, error) {
	all := make([]domain.Category, 0, len(r.items))
	for _, category := range r.items {
		all = append(all, *category)
	}
	return all, nil
}

func (r *categoryRepositoryStub) GetByID(_ context.Context, id uint) (*domain.Category, error) {
	category, ok := r.items[id]
	if !ok {
		return nil, domain.ErrCategoryNotFound
	}
	return domain.ReconstituteCategory(category.Data()), nil
}

func (r *categoryRepositoryStub) Create(_ context.Context, category *domain.Category) (*domain.Category, error) {
	data := category.Data()
	data.ID = r.nextID
	r.nextID++
	created := domain.ReconstituteCategory(data)
	r.items[data.ID] = created
	return created, nil
}

func (r *categoryRepositoryStub) Update(_ context.Context, category *domain.Category) (*domain.Category, error) {
	r.items[category.Data().ID] = category
	return category, nil
}

func (r *categoryRepositoryStub) Delete(_ context.Context, id uint) error {
	delete(r.items, id)
	return nil
}

func (r *categoryRepositoryStub) CountPosts(_ context.Context, id uint) (int64, error) {
	return r.posts[id], nil
}

func (r *categoryRepositoryStub) CountChildren(_ context.Context, id uint) (int64, error) {
	var count int64
	for _, category := range r.items {
		if parent := category.Data().ParentID; parent != nil && *parent == id {
			count++
		}
	}
	return count, nil
}

type catalogUnitOfWork struct {
	categories domain.CategoryRepository
	audits     domain.AuditRepository
}

func (u catalogUnitOfWork) WithinTransaction(_ context.Context, work func(TransactionRepositories) error) error {
	return work(TransactionRepositories{Categories: u.categories, Audits: u.audits})
}

func newCategoryUseCaseForTest(repo *categoryRepositoryStub, audits *auditRepositoryStub) *CategoryUseCase {
	return NewCategoryUseCase(repo, catalogUnitOfWork{categories: repo, audits: audits})
}

func TestCreateCategoryRecordsAuditAndRejectsUnknownParent(t *testing.T) {
	repo := newCategoryRepositoryStub()
	audits := &auditRepositoryStub{}
	useCase := newCategoryUseCaseForTest(repo, audits)

	created, err := useCase.CreateCategory(context.Background(), 1, CategoryInputDTO{Name: "News", Slug: "news", Description: "Tin tức"}, "")
	if err != nil {
		t.Fatalf("CreateCategory returned error: %v", err)
	}
	if created.Data().Description != "Tin tức" {
		t.Fatalf("description = %q", created.Data().Description)
	}
	if len(audits.actions) != 1 || audits.actions[0] != "CREATE_CATEGORY" {
		t.Fatalf("audit actions = %v, want [CREATE_CATEGORY]", audits.actions)
	}

	missing := uint(99)
	_, err = useCase.CreateCategory(context.Background(), 1, CategoryInputDTO{Name: "X", Slug: "x", ParentID: &missing}, "")
	if !errors.Is(err, domain.ErrInvalidCategoryParent) {
		t.Fatalf("error = %v, want ErrInvalidCategoryParent", err)
	}
	if len(repo.items) != 1 {
		t.Fatalf("a category was stored despite the invalid parent: %d items", len(repo.items))
	}
}

func TestUpdateCategoryRejectsParentCycles(t *testing.T) {
	repo := newCategoryRepositoryStub()
	useCase := newCategoryUseCaseForTest(repo, &auditRepositoryStub{})
	root := repo.add(t, "Root", "root", nil)
	child := repo.add(t, "Child", "child", &root)
	grandchild := repo.add(t, "Grand", "grand", &child)

	_, err := useCase.UpdateCategory(context.Background(), 1, root, CategoryInputDTO{Name: "Root", Slug: "root", ParentID: &root}, "")
	if !errors.Is(err, domain.ErrInvalidCategoryParent) {
		t.Fatalf("own parent: error = %v, want ErrInvalidCategoryParent", err)
	}

	_, err = useCase.UpdateCategory(context.Background(), 1, root, CategoryInputDTO{Name: "Root", Slug: "root", ParentID: &grandchild}, "")
	if !errors.Is(err, domain.ErrInvalidCategoryParent) {
		t.Fatalf("descendant as parent: error = %v, want ErrInvalidCategoryParent", err)
	}
	if repo.items[root].Data().ParentID != nil {
		t.Fatal("root category changed despite the rejected update")
	}
}

func TestUpdateCategoryChangesFieldsAndReportsMissing(t *testing.T) {
	repo := newCategoryRepositoryStub()
	audits := &auditRepositoryStub{}
	useCase := newCategoryUseCaseForTest(repo, audits)
	root := repo.add(t, "Root", "root", nil)
	other := repo.add(t, "Other", "other", nil)

	updated, err := useCase.UpdateCategory(context.Background(), 1, other, CategoryInputDTO{Name: "Renamed", Slug: "renamed", ParentID: &root}, "")
	if err != nil {
		t.Fatalf("UpdateCategory returned error: %v", err)
	}
	data := updated.Data()
	if data.Name != "Renamed" || data.ParentID == nil || *data.ParentID != root {
		t.Fatalf("unexpected category after update: %+v", data)
	}
	if len(audits.actions) != 1 || audits.actions[0] != "UPDATE_CATEGORY" {
		t.Fatalf("audit actions = %v, want [UPDATE_CATEGORY]", audits.actions)
	}

	_, err = useCase.UpdateCategory(context.Background(), 1, 99, CategoryInputDTO{Name: "X", Slug: "x"}, "")
	if !errors.Is(err, domain.ErrCategoryNotFound) {
		t.Fatalf("error = %v, want ErrCategoryNotFound", err)
	}
}

func TestDeleteCategoryProtectsCategoriesInUse(t *testing.T) {
	repo := newCategoryRepositoryStub()
	audits := &auditRepositoryStub{}
	useCase := newCategoryUseCaseForTest(repo, audits)
	withPosts := repo.add(t, "With posts", "with-posts", nil)
	repo.posts[withPosts] = 2
	parent := repo.add(t, "Parent", "parent", nil)
	repo.add(t, "Child", "child", &parent)
	free := repo.add(t, "Free", "free", nil)

	if err := useCase.DeleteCategory(context.Background(), 1, withPosts, ""); !errors.Is(err, domain.ErrCategoryInUse) {
		t.Fatalf("category with posts: error = %v, want ErrCategoryInUse", err)
	}
	if err := useCase.DeleteCategory(context.Background(), 1, parent, ""); !errors.Is(err, domain.ErrCategoryHasChildren) {
		t.Fatalf("category with children: error = %v, want ErrCategoryHasChildren", err)
	}
	if len(audits.actions) != 0 {
		t.Fatalf("audit recorded for rejected deletes: %v", audits.actions)
	}

	if err := useCase.DeleteCategory(context.Background(), 1, free, ""); err != nil {
		t.Fatalf("DeleteCategory returned error: %v", err)
	}
	if _, exists := repo.items[free]; exists {
		t.Fatal("category was not deleted")
	}
	if len(audits.actions) != 1 || audits.actions[0] != "DELETE_CATEGORY" {
		t.Fatalf("audit actions = %v, want [DELETE_CATEGORY]", audits.actions)
	}
}
