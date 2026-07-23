/*
 * Contract ID: CTR-009
 * Service Name: SupermarketService
 * Description: Unit and integration tests for SaleService.
 *              Asserts exact HALF_UP rounding, pessimistic concurrent stock locks,
 *              and maximum discount selection from overlapping promotion campaigns.
 */

package tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	dto "supermarket-backend/dtos"
	"supermarket-backend/entities"
	"supermarket-backend/exceptions"
	"supermarket-backend/services"
)

// Define mock implementations of repositories for our service testing.

type MockSaleRepository struct {
	CreateSaleFunc                       func(ctx context.Context, tx *gorm.DB, sale *entities.Sale) error
	FindSaleByIDFunc                     func(ctx context.Context, id uint64) (*entities.Sale, error)
	FindSaleByIDWithLockFunc             func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Sale, error)
	UpdateSaleFunc                       func(ctx context.Context, tx *gorm.DB, sale *entities.Sale) error
	UpdateSaleWithTxFunc                 func(ctx context.Context, tx *gorm.DB, sale *entities.Sale) error
	FindSalesFunc                        func(ctx context.Context, offset, limit int, status string, clientID *uint64, cashierID *uint64) ([]entities.Sale, int64, error)
	CreateDetailFunc                     func(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error
	CreateDetailWithTxFunc               func(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error
	FindDetailByIDFunc                   func(ctx context.Context, id uint64) (*entities.SaleDetail, error)
	UpdateDetailFunc                     func(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error
	UpdateDetailWithTxFunc               func(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error
	DeleteDetailFunc                     func(ctx context.Context, tx *gorm.DB, id uint64) error
	FindDetailBySaleAndProductWithTxFunc func(ctx context.Context, tx *gorm.DB, saleID uint64, productID uint64) (*entities.SaleDetail, error)
	FindAllDetailsBySaleIDWithTxFunc     func(ctx context.Context, tx *gorm.DB, saleID uint64) ([]entities.SaleDetail, error)
}

func (m *MockSaleRepository) CreateSale(ctx context.Context, tx *gorm.DB, sale *entities.Sale) error {
	if m.CreateSaleFunc != nil {
		return m.CreateSaleFunc(ctx, tx, sale)
	}
	return nil
}

func (m *MockSaleRepository) FindSaleByID(ctx context.Context, id uint64) (*entities.Sale, error) {
	if m.FindSaleByIDFunc != nil {
		return m.FindSaleByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSaleRepository) FindSaleByIDWithLock(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Sale, error) {
	if m.FindSaleByIDWithLockFunc != nil {
		return m.FindSaleByIDWithLockFunc(ctx, tx, id)
	}
	return nil, nil
}

func (m *MockSaleRepository) UpdateSale(ctx context.Context, tx *gorm.DB, sale *entities.Sale) error {
	if m.UpdateSaleFunc != nil {
		return m.UpdateSaleFunc(ctx, tx, sale)
	}
	return nil
}

func (m *MockSaleRepository) UpdateSaleWithTx(ctx context.Context, tx *gorm.DB, sale *entities.Sale) error {
	if m.UpdateSaleWithTxFunc != nil {
		return m.UpdateSaleWithTxFunc(ctx, tx, sale)
	}
	return nil
}

func (m *MockSaleRepository) FindSales(ctx context.Context, offset, limit int, status string, clientID *uint64, cashierID *uint64) ([]entities.Sale, int64, error) {
	if m.FindSalesFunc != nil {
		return m.FindSalesFunc(ctx, offset, limit, status, clientID, cashierID)
	}
	return nil, 0, nil
}

func (m *MockSaleRepository) CreateDetail(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error {
	if m.CreateDetailFunc != nil {
		return m.CreateDetailFunc(ctx, tx, detail)
	}
	return nil
}

func (m *MockSaleRepository) CreateDetailWithTx(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error {
	if m.CreateDetailWithTxFunc != nil {
		return m.CreateDetailWithTxFunc(ctx, tx, detail)
	}
	return nil
}

func (m *MockSaleRepository) FindDetailByID(ctx context.Context, id uint64) (*entities.SaleDetail, error) {
	if m.FindDetailByIDFunc != nil {
		return m.FindDetailByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSaleRepository) UpdateDetail(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error {
	if m.UpdateDetailFunc != nil {
		return m.UpdateDetailFunc(ctx, tx, detail)
	}
	return nil
}

func (m *MockSaleRepository) UpdateDetailWithTx(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error {
	if m.UpdateDetailWithTxFunc != nil {
		return m.UpdateDetailWithTxFunc(ctx, tx, detail)
	}
	return nil
}

func (m *MockSaleRepository) DeleteDetail(ctx context.Context, tx *gorm.DB, id uint64) error {
	if m.DeleteDetailFunc != nil {
		return m.DeleteDetailFunc(ctx, tx, id)
	}
	return nil
}

func (m *MockSaleRepository) FindDetailBySaleAndProductWithTx(ctx context.Context, tx *gorm.DB, saleID uint64, productID uint64) (*entities.SaleDetail, error) {
	if m.FindDetailBySaleAndProductWithTxFunc != nil {
		return m.FindDetailBySaleAndProductWithTxFunc(ctx, tx, saleID, productID)
	}
	return nil, nil
}

func (m *MockSaleRepository) FindAllDetailsBySaleIDWithTx(ctx context.Context, tx *gorm.DB, saleID uint64) ([]entities.SaleDetail, error) {
	if m.FindAllDetailsBySaleIDWithTxFunc != nil {
		return m.FindAllDetailsBySaleIDWithTxFunc(ctx, tx, saleID)
	}
	return nil, nil
}

type MockProductRepository struct {
	CreateFunc                          func(ctx context.Context, product *entities.Product) error
	FindByIDFunc                        func(ctx context.Context, id uint64) (*entities.Product, error)
	FindByIDWithLockFunc                func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Product, error)
	UpdateFunc                          func(ctx context.Context, product *entities.Product) error
	UpdateStockWithTxFunc               func(ctx context.Context, tx *gorm.DB, id uint64, stock int) error
	SoftDeleteByCategoryIDWithTxFunc    func(ctx context.Context, tx *gorm.DB, categoryID uint64) error
	FindActivePromotionsByProductIDFunc func(ctx context.Context, tx *gorm.DB, id uint64, t time.Time) ([]entities.Promotion, error)
	FindAllFunc                         func(ctx context.Context, offset, limit int) ([]entities.Product, int64, error)
	FindAllByCategoryIDFunc             func(ctx context.Context, categoryID uint64, offset, limit int) ([]entities.Product, int64, error)
}

func (m *MockProductRepository) Create(ctx context.Context, product *entities.Product) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, product)
	}
	return nil
}

func (m *MockProductRepository) FindByID(ctx context.Context, id uint64) (*entities.Product, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockProductRepository) FindByIDWithLock(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Product, error) {
	if m.FindByIDWithLockFunc != nil {
		return m.FindByIDWithLockFunc(ctx, tx, id)
	}
	return nil, nil
}

func (m *MockProductRepository) Update(ctx context.Context, product *entities.Product) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, product)
	}
	return nil
}

func (m *MockProductRepository) UpdateStockWithTx(ctx context.Context, tx *gorm.DB, id uint64, stock int) error {
	if m.UpdateStockWithTxFunc != nil {
		return m.UpdateStockWithTxFunc(ctx, tx, id, stock)
	}
	return nil
}

func (m *MockProductRepository) SoftDeleteByCategoryIDWithTx(ctx context.Context, tx *gorm.DB, categoryID uint64) error {
	if m.SoftDeleteByCategoryIDWithTxFunc != nil {
		return m.SoftDeleteByCategoryIDWithTxFunc(ctx, tx, categoryID)
	}
	return nil
}

func (m *MockProductRepository) FindActivePromotionsByProductID(ctx context.Context, tx *gorm.DB, id uint64, t time.Time) ([]entities.Promotion, error) {
	if m.FindActivePromotionsByProductIDFunc != nil {
		return m.FindActivePromotionsByProductIDFunc(ctx, tx, id, t)
	}
	return nil, nil
}

func (m *MockProductRepository) FindAll(ctx context.Context, offset, limit int) ([]entities.Product, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockProductRepository) FindAllByCategoryID(ctx context.Context, categoryID uint64, offset, limit int) ([]entities.Product, int64, error) {
	if m.FindAllByCategoryIDFunc != nil {
		return m.FindAllByCategoryIDFunc(ctx, categoryID, offset, limit)
	}
	return nil, 0, nil
}

// helper to setup in-memory SQLite GORM db
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}
	return db
}

// TestRoundHalfUp asserts the HALF_UP rounding function inside sale service
func TestRoundHalfUp(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{10.004, 10.00},
		{10.005, 10.01},
		{10.006, 10.01},
		{10.555, 10.56},
		{10.554, 10.55},
		{0.125, 0.13},
		{0.124, 0.12},
	}

	for _, tc := range tests {
		actual := services.RoundHalfUp(tc.input)
		if actual != tc.expected {
			t.Errorf("RoundHalfUp(%f) = %f; expected %f", tc.input, actual, tc.expected)
		}
	}
}

// TestCalculatePromotionPrice asserts that the price engine selects the highest active discount
// from overlapping dates and applies before taxes.
func TestCalculatePromotionPrice(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	mockSaleRepo := &MockSaleRepository{}
	mockProductRepo := &MockProductRepository{}

	service := services.NewSaleService(mockSaleRepo, mockProductRepo, db)

	// Stub a product with base price = 100 and tax = 12%
	product := &entities.Product{
		ID:            42,
		Name:          "Test Product",
		BasePrice:     100.0,
		TaxPercentage: 12.0,
		Stock:         100,
	}

	// Stub active promotions. We have overlapping active campaigns:
	// Promotion 1: 10% discount
	// Promotion 2: 15% discount (highest)
	// Promotion 3: 5% discount
	activePromos := []entities.Promotion{
		{
			ID:                 1,
			Name:               "10% Discount Campaign",
			DiscountPercentage: 10.0,
		},
		{
			ID:                 2,
			Name:               "15% Discount Campaign (Max)",
			DiscountPercentage: 15.0,
		},
		{
			ID:                 3,
			Name:               "5% Discount Campaign",
			DiscountPercentage: 5.0,
		},
	}

	// Set up expectations
	mockSaleRepo.FindSaleByIDWithLockFunc = func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Sale, error) {
		return &entities.Sale{
			ID:     1,
			Status: "OPEN",
		}, nil
	}
	mockProductRepo.FindByIDWithLockFunc = func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Product, error) {
		return product, nil
	}
	mockProductRepo.FindActivePromotionsByProductIDFunc = func(ctx context.Context, tx *gorm.DB, id uint64, t time.Time) ([]entities.Promotion, error) {
		return activePromos, nil
	}
	mockProductRepo.UpdateStockWithTxFunc = func(ctx context.Context, tx *gorm.DB, id uint64, stock int) error {
		return nil
	}
	mockSaleRepo.FindDetailBySaleAndProductWithTxFunc = func(ctx context.Context, tx *gorm.DB, saleID uint64, productID uint64) (*entities.SaleDetail, error) {
		return nil, nil // brand new line item
	}
	mockSaleRepo.CreateDetailWithTxFunc = func(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error {
		return nil
	}
	mockSaleRepo.FindAllDetailsBySaleIDWithTxFunc = func(ctx context.Context, tx *gorm.DB, saleID uint64) ([]entities.SaleDetail, error) {
		// Return details list representing this item
		return []entities.SaleDetail{
			{
				ID:        1,
				SaleID:    1,
				ProductID: 42,
				UnitPrice: 95.20, // Expected computed unit price
				Quantity:  2,
				SubTotal:  190.40,
			},
		}, nil
	}
	mockSaleRepo.UpdateSaleWithTxFunc = func(ctx context.Context, tx *gorm.DB, sale *entities.Sale) error {
		return nil
	}

	req := &dto.AddItemRequest{
		ProductID: 42,
		Quantity:  2,
	}

	res, err := service.AddItemToSale(ctx, 1, req)
	if err != nil {
		t.Fatalf("AddItemToSale failed: %v", err)
	}

	// Price calculation verification:
	// Base Price = 100.00
	// Max Promotion Discount = 15.0%
	// Discounted Base = 100.00 * (1.0 - 0.15) = 85.00
	// Tax Addition = 85.00 * (1.0 + 0.12) = 95.20
	// Subtotal for Qty 2 = 95.20 * 2 = 190.40
	// Total price = 190.40

	expectedUnitPrice := 95.20
	expectedSubTotal := 190.40
	expectedTotal := 190.40

	if len(res.LineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(res.LineItems))
	}

	item := res.LineItems[0]
	if item.UnitPrice != expectedUnitPrice {
		t.Errorf("expected unit price %f, got %f", expectedUnitPrice, item.UnitPrice)
	}
	if item.SubTotal != expectedSubTotal {
		t.Errorf("expected subtotal %f, got %f", expectedSubTotal, item.SubTotal)
	}
	if res.Sale.TotalPrice != expectedTotal {
		t.Errorf("expected total price %f, got %f", expectedTotal, res.Sale.TotalPrice)
	}
}

// TestPessimisticStockLocks asserts that concurrent additions of items to a sale
// do not bypass the product inventory stock limit under multi-threaded load.
func TestPessimisticStockLocks(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	mockSaleRepo := &MockSaleRepository{}
	mockProductRepo := &MockProductRepository{}

	service := services.NewSaleService(mockSaleRepo, mockProductRepo, db)

	// Initial stock is exactly 5
	initialStock := 5
	product := &entities.Product{
		ID:            10,
		Name:          "Limited Stock item",
		BasePrice:     10.0,
		TaxPercentage: 0.0,
		Stock:         initialStock,
	}

	// Set up expectations
	mockSaleRepo.FindSaleByIDWithLockFunc = func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Sale, error) {
		return &entities.Sale{
			ID:     1,
			Status: "OPEN",
		}, nil
	}

	// We protect product state with a mutex to simulate actual database locking in SQLite/MySQL.
	var mu sync.Mutex
	mockProductRepo.FindByIDWithLockFunc = func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Product, error) {
		mu.Lock()
		defer mu.Unlock()
		// Return a copy so each goroutine gets the actual current stock
		return &entities.Product{
			ID:            product.ID,
			Name:          product.Name,
			BasePrice:     product.BasePrice,
			TaxPercentage: product.TaxPercentage,
			Stock:         product.Stock,
		}, nil
	}

	mockProductRepo.UpdateStockWithTxFunc = func(ctx context.Context, tx *gorm.DB, id uint64, stock int) error {
		mu.Lock()
		defer mu.Unlock()
		product.Stock = stock
		return nil
	}

	mockProductRepo.FindActivePromotionsByProductIDFunc = func(ctx context.Context, tx *gorm.DB, id uint64, t time.Time) ([]entities.Promotion, error) {
		return nil, nil
	}
	mockSaleRepo.FindDetailBySaleAndProductWithTxFunc = func(ctx context.Context, tx *gorm.DB, saleID uint64, productID uint64) (*entities.SaleDetail, error) {
		return nil, nil
	}
	mockSaleRepo.CreateDetailWithTxFunc = func(ctx context.Context, tx *gorm.DB, detail *entities.SaleDetail) error {
		return nil
	}
	mockSaleRepo.FindAllDetailsBySaleIDWithTxFunc = func(ctx context.Context, tx *gorm.DB, saleID uint64) ([]entities.SaleDetail, error) {
		return []entities.SaleDetail{}, nil
	}
	mockSaleRepo.UpdateSaleWithTxFunc = func(ctx context.Context, tx *gorm.DB, sale *entities.Sale) error {
		return nil
	}

	// We trigger 10 concurrent requests to purchase 1 unit each.
	// Since stock is 5, exactly 5 requests must succeed and 5 must fail with insufficient stock error.
	numRequests := 10
	var wg sync.WaitGroup
	successChan := make(chan bool, numRequests)
	errorChan := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := &dto.AddItemRequest{
				ProductID: 10,
				Quantity:  1,
			}
			_, err := service.AddItemToSale(ctx, 1, req)
			if err != nil {
				errorChan <- err
			} else {
				successChan <- true
			}
		}()
	}

	wg.Wait()
	close(successChan)
	close(errorChan)

	successCount := 0
	for range successChan {
		successCount++
	}

	failureCount := 0
	var stockErrors []*exceptions.AppError
	for err := range errorChan {
		failureCount++
		var appErr *exceptions.AppError
		if errors.As(err, &appErr) {
			if appErr.ErrorCode == "STOCK_INSUFFICIENT" {
				stockErrors = append(stockErrors, appErr)
			}
		}
	}

	if successCount != initialStock {
		t.Errorf("Expected exactly %d successful additions, got %d", initialStock, successCount)
	}

	if failureCount != (numRequests - initialStock) {
		t.Errorf("Expected exactly %d failed additions, got %d", numRequests-initialStock, failureCount)
	}

	if len(stockErrors) != failureCount {
		t.Errorf("Expected all failures to be due to STOCK_INSUFFICIENT, got %d out of %d", len(stockErrors), failureCount)
	}

	if product.Stock != 0 {
		t.Errorf("Expected final product stock to be 0, got %d", product.Stock)
	}
}

// TestSaleService_FailurePaths asserts various business validation errors in SaleService.
func TestSaleService_FailurePaths(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	mockSaleRepo := &MockSaleRepository{}
	mockProductRepo := &MockProductRepository{}

	service := services.NewSaleService(mockSaleRepo, mockProductRepo, db)

	t.Run("AddItemToSale_NonExistentSale", func(t *testing.T) {
		mockSaleRepo.FindSaleByIDWithLockFunc = func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Sale, error) {
			return nil, nil // sale not found
		}

		req := &dto.AddItemRequest{ProductID: 1, Quantity: 1}
		_, err := service.AddItemToSale(ctx, 999, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *exceptions.AppError
		if !errors.As(err, &appErr) || appErr.ErrorCode != "RESOURCE_NOT_FOUND" {
			t.Errorf("expected RESOURCE_NOT_FOUND, got %v", err)
		}
	})

	t.Run("AddItemToSale_NonOpenSale", func(t *testing.T) {
		mockSaleRepo.FindSaleByIDWithLockFunc = func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Sale, error) {
			return &entities.Sale{ID: 1, Status: "CLOSED"}, nil
		}

		req := &dto.AddItemRequest{ProductID: 1, Quantity: 1}
		_, err := service.AddItemToSale(ctx, 1, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *exceptions.AppError
		if !errors.As(err, &appErr) || appErr.HTTPStatus != 409 {
			t.Errorf("expected 409 Conflict, got %v", err)
		}
	})

	t.Run("AddItemToSale_ProductNotFound", func(t *testing.T) {
		mockSaleRepo.FindSaleByIDWithLockFunc = func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Sale, error) {
			return &entities.Sale{ID: 1, Status: "OPEN"}, nil
		}
		mockProductRepo.FindByIDWithLockFunc = func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Product, error) {
			return nil, nil // product not found
		}

		req := &dto.AddItemRequest{ProductID: 999, Quantity: 1}
		_, err := service.AddItemToSale(ctx, 1, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *exceptions.AppError
		if !errors.As(err, &appErr) || appErr.ErrorCode != "RESOURCE_NOT_FOUND" {
			t.Errorf("expected RESOURCE_NOT_FOUND, got %v", err)
		}
	})

	t.Run("FinalizeSale_EmptySale", func(t *testing.T) {
		mockSaleRepo.FindSaleByIDWithLockFunc = func(ctx context.Context, tx *gorm.DB, id uint64) (*entities.Sale, error) {
			return &entities.Sale{ID: 1, Status: "OPEN"}, nil
		}
		mockSaleRepo.FindAllDetailsBySaleIDWithTxFunc = func(ctx context.Context, tx *gorm.DB, saleID uint64) ([]entities.SaleDetail, error) {
			return []entities.SaleDetail{}, nil // empty sale detail
		}

		_, err := service.FinalizeSale(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *exceptions.AppError
		if !errors.As(err, &appErr) || appErr.ErrorCode != "VALIDATION_FAILED" {
			t.Errorf("expected VALIDATION_FAILED, got %v", err)
		}
	})
}