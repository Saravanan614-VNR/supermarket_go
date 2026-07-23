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

type MockPromotionRepository struct {
	CreateFunc             func(ctx context.Context, promotion *entities.Promotion) error
	FindByIDFunc           func(ctx context.Context, id uint64) (*entities.Promotion, error)
	UpdateFunc             func(ctx context.Context, promotion *entities.Promotion) error
	DeleteFunc             func(ctx context.Context, id uint64) error
	FindAllFunc            func(ctx context.Context, offset, limit int) ([]entities.Promotion, int64, error)
	LinkProductsWithTxFunc func(ctx context.Context, tx *gorm.DB, promotionID uint64, productIDs []uint64) error
	UnlinkProductWithTxFunc func(ctx context.Context, tx *gorm.DB, promotionID uint64, productID uint64) error
}

func (m *MockPromotionRepository) Create(ctx context.Context, promotion *entities.Promotion) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, promotion)
	}
	return nil
}

func (m *MockPromotionRepository) FindByID(ctx context.Context, id uint64) (*entities.Promotion, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockPromotionRepository) Update(ctx context.Context, promotion *entities.Promotion) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, promotion)
	}
	return nil
}

func (m *MockPromotionRepository) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockPromotionRepository) FindAll(ctx context.Context, offset, limit int) ([]entities.Promotion, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockPromotionRepository) LinkProductsWithTx(ctx context.Context, tx *gorm.DB, promotionID uint64, productIDs []uint64) error {
	if m.LinkProductsWithTxFunc != nil {
		return m.LinkProductsWithTxFunc(ctx, tx, promotionID, productIDs)
	}
	return nil
}

func (m *MockPromotionRepository) UnlinkProductWithTx(ctx context.Context, tx *gorm.DB, promotionID uint64, productID uint64) error {
	if m.UnlinkProductWithTxFunc != nil {
		return m.UnlinkProductWithTxFunc(ctx, tx, promotionID, productID)
	}
	return nil
}

var _ repositories.PromotionRepository = (*MockPromotionRepository)(nil)

func TestPromotionService_CreatePromotion_Success(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	mockRepo := &MockPromotionRepository{
		CreateFunc: func(ctx context.Context, promotion *entities.Promotion) error {
			promotion.ID = 1
			return nil
		},
	}

	srv := services.NewPromotionService(mockRepo, nil, db)

	now := time.Now()
	req := &dto.CreatePromotionReq{
		Name:               "Mega Sale",
		DiscountPercentage: 20.0,
		StartDate:          now,
		EndDate:            now.Add(24 * time.Hour),
	}

	res, err := srv.CreatePromotion(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.Name != "Mega Sale" || res.DiscountPercentage != 20.0 {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestPromotionService_ListPromotions_Success(t *testing.T) {
	mockRepo := &MockPromotionRepository{
		FindAllFunc: func(ctx context.Context, offset, limit int) ([]entities.Promotion, int64, error) {
			return []entities.Promotion{
				{ID: 1, Name: "Mega Sale", DiscountPercentage: 20.0, StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour)},
			}, 1, nil
		},
	}

	srv := services.NewPromotionService(mockRepo, nil, nil)

	res, err := srv.ListPromotions(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalCount != 1 || len(res.Items) != 1 || res.Items[0].Name != "Mega Sale" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestPromotionService_GetPromotionByID_Success(t *testing.T) {
	mockRepo := &MockPromotionRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Promotion, error) {
			return &entities.Promotion{
				ID:                 id,
				Name:               "Mega Sale",
				DiscountPercentage: 20.0,
			}, nil
		},
	}

	srv := services.NewPromotionService(mockRepo, nil, nil)

	res, err := srv.GetPromotionByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.Name != "Mega Sale" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestPromotionService_UpdatePromotion_Success(t *testing.T) {
	mockRepo := &MockPromotionRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Promotion, error) {
			return &entities.Promotion{
				ID:                 id,
				Name:               "Mega Sale",
				DiscountPercentage: 20.0,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, promotion *entities.Promotion) error {
			return nil
		},
	}

	srv := services.NewPromotionService(mockRepo, nil, nil)

	req := &dto.UpdatePromotionReq{
		Name:               "Mega Sale Updated",
		DiscountPercentage: 25.0,
	}

	res, err := srv.UpdatePromotion(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Name != "Mega Sale Updated" || res.DiscountPercentage != 25.0 {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestPromotionService_DeletePromotion_Success(t *testing.T) {
	mockRepo := &MockPromotionRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Promotion, error) {
			return &entities.Promotion{
				ID:   id,
				Name: "Mega Sale",
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uint64) error {
			return nil
		},
	}

	srv := services.NewPromotionService(mockRepo, nil, nil)

	err := srv.DeletePromotion(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPromotionService_LinkProductsToPromotion_Success(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	mockRepo := &MockPromotionRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Promotion, error) {
			return &entities.Promotion{ID: id, Name: "Mega Sale"}, nil
		},
		LinkProductsWithTxFunc: func(ctx context.Context, tx *gorm.DB, promotionID uint64, productIDs []uint64) error {
			return nil
		},
	}

	mockProdRepo := &MockProductRepository{
		FindByIDWithLockFunc: func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Product, error) {
			return &entities.Product{ID: id, Name: "Chips"}, nil
		},
	}

	srv := services.NewPromotionService(mockRepo, mockProdRepo, db)

	err = srv.LinkProductsToPromotion(context.Background(), 1, []uint64{101, 102})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPromotionService_UnlinkProductFromPromotion_Success(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	mockRepo := &MockPromotionRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Promotion, error) {
			return &entities.Promotion{ID: id, Name: "Mega Sale"}, nil
		},
		UnlinkProductWithTxFunc: func(ctx context.Context, tx *gorm.DB, promotionID uint64, productID uint64) error {
			return nil
		},
	}

	mockProdRepo := &MockProductRepository{
		FindByIDWithLockFunc: func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Product, error) {
			return &entities.Product{ID: id, Name: "Chips"}, nil
		},
	}

	srv := services.NewPromotionService(mockRepo, mockProdRepo, db)

	err = srv.UnlinkProductFromPromotion(context.Background(), 1, 101)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}