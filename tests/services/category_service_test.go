package tests

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	dto "supermarket-backend/dtos"
	"supermarket-backend/entities"
	"supermarket-backend/repositories"
	"supermarket-backend/services"
)

type MockCategoryRepository struct {
	CreateFunc                      func(ctx context.Context, category *entities.Category) error
	FindByIDFunc                    func(ctx context.Context, id uint64) (*entities.Category, error)
	FindByIDWithLockFunc            func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Category, error)
	UpdateFunc                      func(ctx context.Context, category *entities.Category) error
	DeleteFunc                      func(ctx context.Context, id uint64) error
	FindAllFunc                     func(ctx context.Context, offset, limit int) ([]entities.Category, int64, error)
	ExistsByNameCaseInsensitiveFunc func(ctx context.Context, name string) (bool, error)
	SoftDeleteWithTxFunc            func(ctx context.Context, tx *gorm.DB, id uint64) error
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *entities.Category) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, category)
	}
	return nil
}

func (m *MockCategoryRepository) FindByID(ctx context.Context, id uint64) (*entities.Category, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockCategoryRepository) FindByIDWithLock(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Category, error) {
	if m.FindByIDWithLockFunc != nil {
		return m.FindByIDWithLockFunc(ctx, tx, id)
	}
	return nil, nil
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *entities.Category) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, category)
	}
	return nil
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockCategoryRepository) FindAll(ctx context.Context, offset, limit int) ([]entities.Category, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockCategoryRepository) ExistsByNameCaseInsensitive(ctx context.Context, name string) (bool, error) {
	if m.ExistsByNameCaseInsensitiveFunc != nil {
		return m.ExistsByNameCaseInsensitiveFunc(ctx, name)
	}
	return false, nil
}

func (m *MockCategoryRepository) SoftDeleteWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	if m.SoftDeleteWithTxFunc != nil {
		return m.SoftDeleteWithTxFunc(ctx, tx, id)
	}
	return nil
}

var _ repositories.CategoryRepository = (*MockCategoryRepository)(nil)

func TestCategoryService_CreateCategory_Success(t *testing.T) {
	mockRepo := &MockCategoryRepository{
		ExistsByNameCaseInsensitiveFunc: func(ctx context.Context, name string) (bool, error) {
			return false, nil // Name is unique
		},
		CreateFunc: func(ctx context.Context, category *entities.Category) error {
			category.ID = 1
			return nil
		},
	}

	srv := services.NewCategoryService(mockRepo, nil, nil)

	req := &dto.CreateCategoryReq{
		Name:        "Produce",
		Description: "Fresh fruits and vegetables",
	}

	res, err := srv.CreateCategory(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.Name != "Produce" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestCategoryService_ListCategories_Success(t *testing.T) {
	mockRepo := &MockCategoryRepository{
		FindAllFunc: func(ctx context.Context, offset, limit int) ([]entities.Category, int64, error) {
			return []entities.Category{
				{ID: 1, Name: "Produce", Description: "Fresh fruits"},
			}, 1, nil
		},
	}

	srv := services.NewCategoryService(mockRepo, nil, nil)

	res, err := srv.ListCategories(context.Background(), 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalCount != 1 || len(res.Items) != 1 || res.Items[0].Name != "Produce" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestCategoryService_GetCategoryByID_Success(t *testing.T) {
	mockRepo := &MockCategoryRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Category, error) {
			return &entities.Category{
				ID:          id,
				Name:        "Produce",
				Description: "Fresh fruits",
			}, nil
		},
	}

	srv := services.NewCategoryService(mockRepo, nil, nil)

	res, err := srv.GetCategoryByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.Name != "Produce" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestCategoryService_UpdateCategory_Success(t *testing.T) {
	mockRepo := &MockCategoryRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Category, error) {
			return &entities.Category{
				ID:          id,
				Name:        "Produce",
				Description: "Fresh fruits",
			}, nil
		},
		ExistsByNameCaseInsensitiveFunc: func(ctx context.Context, name string) (bool, error) {
			return false, nil
		},
		UpdateFunc: func(ctx context.Context, category *entities.Category) error {
			return nil
		},
	}

	srv := services.NewCategoryService(mockRepo, nil, nil)

	req := &dto.UpdateCategoryReq{
		Name:        "Produce Updated",
		Description: "Updated description",
	}

	res, err := srv.UpdateCategory(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Name != "Produce Updated" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestCategoryService_DeleteCategory_Success(t *testing.T) {
	// Setup SQLite GORM in-memory database to support GORM Transaction blocks inside the service
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	mockRepo := &MockCategoryRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Category, error) {
			return &entities.Category{
				ID:          id,
				Name:        "Produce",
				Description: "Fresh fruits",
			}, nil
		},
		SoftDeleteWithTxFunc: func(ctx context.Context, tx *gorm.DB, id uint64) error {
			return nil
		},
	}

	mockProdRepo := &MockProductRepository{
		SoftDeleteByCategoryIDWithTxFunc: func(ctx context.Context, tx *gorm.DB, categoryID uint64) error {
			return nil
		},
	}

	srv := services.NewCategoryService(mockRepo, mockProdRepo, db)

	err = srv.DeleteCategory(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCategoryService_ListCategoryProducts_Success(t *testing.T) {
	mockRepo := &MockCategoryRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Category, error) {
			return &entities.Category{
				ID:   id,
				Name: "Produce",
			}, nil
		},
	}

	mockProdRepo := &MockProductRepository{
		FindAllByCategoryIDFunc: func(ctx context.Context, categoryID uint64, offset, limit int) ([]entities.Product, int64, error) {
			return []entities.Product{
				{ID: 101, Name: "Apple", Stock: 50, CategoryID: categoryID},
			}, 1, nil
		},
	}

	srv := services.NewCategoryService(mockRepo, mockProdRepo, nil)

	res, err := srv.ListCategoryProducts(context.Background(), 1, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalCount != 1 || len(res.Items) != 1 || res.Items[0].Name != "Apple" {
		t.Errorf("unexpected response: %+v", res)
	}
}