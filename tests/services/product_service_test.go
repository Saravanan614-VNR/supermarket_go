package tests

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	dto "supermarket-backend/dtos"
	"supermarket-backend/entities"
	"supermarket-backend/repositories"
	"supermarket-backend/services"
)

var _ repositories.ProductRepository = (*MockProductRepository)(nil)

func TestProductService_CreateProduct_Success(t *testing.T) {
	mockRepo := &MockProductRepository{
		CreateFunc: func(ctx context.Context, product *entities.Product) error {
			product.ID = 1
			return nil
		},
	}

	mockCategoryRepo := &MockCategoryRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Category, error) {
			return &entities.Category{ID: id, Name: "Snacks"}, nil
		},
	}

	srv := services.NewProductService(mockRepo, mockCategoryRepo, nil)

	req := &dto.CreateProductReq{
		Name:          "Chips",
		Brand:         "Lays",
		BasePrice:     1.50,
		TaxPercentage: 10,
		CategoryID:    1,
		Stock:         100,
	}

	res, err := srv.CreateProduct(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.Name != "Chips" || res.Stock != 100 {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestProductService_ListProducts_Success(t *testing.T) {
	mockRepo := &MockProductRepository{
		FindAllFunc: func(ctx context.Context, offset, limit int) ([]entities.Product, int64, error) {
			return []entities.Product{
				{ID: 1, Name: "Chips", Brand: "Lays", BasePrice: 1.50, TaxPercentage: 10, CategoryID: 1, Stock: 100},
			}, 1, nil
		},
	}

	srv := services.NewProductService(mockRepo, nil, nil)

	res, err := srv.ListProducts(context.Background(), 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalCount != 1 || len(res.Items) != 1 || res.Items[0].Name != "Chips" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestProductService_GetProductByID_Success(t *testing.T) {
	mockRepo := &MockProductRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Product, error) {
			return &entities.Product{
				ID:            id,
				Name:          "Chips",
				Brand:         "Lays",
				BasePrice:     1.50,
				TaxPercentage: 10,
				CategoryID:    1,
				Stock:         100,
			}, nil
		},
	}

	srv := services.NewProductService(mockRepo, nil, nil)

	res, err := srv.GetProductByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.Name != "Chips" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestProductService_UpdateProduct_Success(t *testing.T) {
	mockRepo := &MockProductRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Product, error) {
			return &entities.Product{
				ID:            id,
				Name:          "Chips",
				Brand:         "Lays",
				BasePrice:     1.50,
				TaxPercentage: 10,
				CategoryID:    1,
				Stock:         100,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, product *entities.Product) error {
			return nil
		},
	}

	mockCategoryRepo := &MockCategoryRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Category, error) {
			return &entities.Category{ID: id, Name: "Snacks"}, nil
		},
	}

	srv := services.NewProductService(mockRepo, mockCategoryRepo, nil)

	req := &dto.UpdateProductReq{
		Name:          "Chips Updated",
		Brand:         "Lays",
		BasePrice:     1.80,
		TaxPercentage: 10,
		CategoryID:    1,
	}

	res, err := srv.UpdateProduct(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Name != "Chips Updated" || res.BasePrice != 1.80 {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestProductService_DeleteProduct_Success(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	// AutoMigrate Product to support db.Delete
	_ = db.AutoMigrate(&entities.Product{})

	// Insert dummy record to SQLite so db.Delete finds it and deletes it
	prod := &entities.Product{ID: 1, Name: "Chips", Brand: "Lays", BasePrice: 1.50, CategoryID: 1}
	_ = db.Create(prod)

	mockRepo := &MockProductRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Product, error) {
			return prod, nil
		},
		UpdateStockWithTxFunc: func(ctx context.Context, tx *gorm.DB, id uint64, stock int) error {
			return nil
		},
		UpdateFunc: func(ctx context.Context, product *entities.Product) error {
			return nil
		},
	}

	srv := services.NewProductService(mockRepo, nil, db)

	err = srv.DeleteProduct(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductService_RestockProduct_Success(t *testing.T) {
	mockRepo := &MockProductRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Product, error) {
			return &entities.Product{ID: id, Name: "Chips", Stock: 100}, nil
		},
		UpdateStockWithTxFunc: func(ctx context.Context, tx *gorm.DB, id uint64, stock int) error {
			return nil
		},
	}

	srv := services.NewProductService(mockRepo, nil, nil)

	res, err := srv.RestockProduct(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Stock != 150 {
		t.Errorf("expected stock 150, got %d", res.Stock)
	}
}

func TestProductService_CalculatePrice_Success(t *testing.T) {
	mockRepo := &MockProductRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Product, error) {
			return &entities.Product{ID: id, Name: "Chips", BasePrice: 2.00, TaxPercentage: 10}, nil
		},
		FindActivePromotionsByProductIDFunc: func(ctx context.Context, tx *gorm.DB, id uint64, t time.Time) ([]entities.Promotion, error) {
			return []entities.Promotion{
				{ID: 1, Name: "Promo 15%", DiscountPercentage: 15.0},
			}, nil
		},
	}

	srv := services.NewProductService(mockRepo, nil, nil)

	res, err := srv.CalculatePrice(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.AppliedPromotion == nil || res.AppliedPromotion.DiscountPercentage != 15 || res.PromotionalFinalPrice != 1.87 {
		t.Errorf("unexpected calculations: %+v", res)
	}
}