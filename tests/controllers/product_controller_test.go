/*
 * Description: Unit tests for ProductController verifying status code mappings and endpoint behaviors.
 */

package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"supermarket-backend/controllers"
	dto "supermarket-backend/dtos"
	"supermarket-backend/services"
)

type MockProductService struct {
	CreateProductFunc  func(ctx context.Context, req *dto.CreateProductReq) (*dto.ProductResponse, error)
	ListProductsFunc   func(ctx context.Context, page, size int, name string) (*dto.PaginatedProducts, error)
	GetProductByIDFunc func(ctx context.Context, id uint64) (*dto.ProductResponse, error)
	UpdateProductFunc  func(ctx context.Context, id uint64, req *dto.UpdateProductReq) (*dto.ProductResponse, error)
	DeleteProductFunc  func(ctx context.Context, id uint64) error
	RestockProductFunc func(ctx context.Context, id uint64, quantity int) (*dto.ProductResponse, error)
	CalculatePriceFunc func(ctx context.Context, id uint64) (*dto.PriceBreakdownResponse, error)
}

func (m *MockProductService) CreateProduct(ctx context.Context, req *dto.CreateProductReq) (*dto.ProductResponse, error) {
	if m.CreateProductFunc != nil {
		return m.CreateProductFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockProductService) ListProducts(ctx context.Context, page, size int, name string) (*dto.PaginatedProducts, error) {
	if m.ListProductsFunc != nil {
		return m.ListProductsFunc(ctx, page, size, name)
	}
	return nil, nil
}

func (m *MockProductService) GetProductByID(ctx context.Context, id uint64) (*dto.ProductResponse, error) {
	if m.GetProductByIDFunc != nil {
		return m.GetProductByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockProductService) UpdateProduct(ctx context.Context, id uint64, req *dto.UpdateProductReq) (*dto.ProductResponse, error) {
	if m.UpdateProductFunc != nil {
		return m.UpdateProductFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockProductService) DeleteProduct(ctx context.Context, id uint64) error {
	if m.DeleteProductFunc != nil {
		return m.DeleteProductFunc(ctx, id)
	}
	return nil
}

func (m *MockProductService) RestockProduct(ctx context.Context, id uint64, quantity int) (*dto.ProductResponse, error) {
	if m.RestockProductFunc != nil {
		return m.RestockProductFunc(ctx, id, quantity)
	}
	return nil, nil
}

func (m *MockProductService) CalculatePrice(ctx context.Context, id uint64) (*dto.PriceBreakdownResponse, error) {
	if m.CalculatePriceFunc != nil {
		return m.CalculatePriceFunc(ctx, id)
	}
	return nil, nil
}

var _ services.ProductService = (*MockProductService)(nil)

func TestCreateProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockProductService{
		CreateProductFunc: func(ctx context.Context, req *dto.CreateProductReq) (*dto.ProductResponse, error) {
			return &dto.ProductResponse{
				ID:            1,
				Name:          req.Name,
				Brand:         req.Brand,
				BasePrice:     req.BasePrice,
				TaxPercentage: req.TaxPercentage,
				CategoryID:    req.CategoryID,
				Stock:         req.Stock,
			}, nil
		},
	}

	ctrl := controllers.NewProductController(mockService)
	router := gin.New()
	router.POST("/api/v0.0/products", ctrl.CreateProduct)

	reqPayload := dto.CreateProductReq{
		Name:          "Coca Cola 1.5L",
		Brand:         "The Coca-Cola Company",
		BasePrice:     1.50,
		TaxPercentage: 18,
		CategoryID:    2,
		Stock:         50,
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v0.0/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestListProducts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockProductService{
		ListProductsFunc: func(ctx context.Context, page, size int, name string) (*dto.PaginatedProducts, error) {
			return &dto.PaginatedProducts{
				Items: []dto.ProductResponse{
					{ID: 1, Name: "Coca Cola 1.5L", Brand: "The Coca-Cola Company", BasePrice: 1.50, TaxPercentage: 18, CategoryID: 2, Stock: 50},
				},
				TotalCount: 1,
			}, nil
		},
	}

	ctrl := controllers.NewProductController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/products", ctrl.ListProducts)

	req, _ := http.NewRequest("GET", "/api/v0.0/products?page=1&size=10&name=Coca", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetProductByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockProductService{
		GetProductByIDFunc: func(ctx context.Context, id uint64) (*dto.ProductResponse, error) {
			return &dto.ProductResponse{
				ID:            id,
				Name:          "Coca Cola 1.5L",
				Brand:         "The Coca-Cola Company",
				BasePrice:     1.50,
				TaxPercentage: 18,
				CategoryID:    2,
				Stock:         50,
			}, nil
		},
	}

	ctrl := controllers.NewProductController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/products/:id", ctrl.GetProductByID)

	req, _ := http.NewRequest("GET", "/api/v0.0/products/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockProductService{
		UpdateProductFunc: func(ctx context.Context, id uint64, req *dto.UpdateProductReq) (*dto.ProductResponse, error) {
			return &dto.ProductResponse{
				ID:            id,
				Name:          req.Name,
				Brand:         req.Brand,
				BasePrice:     req.BasePrice,
				TaxPercentage: req.TaxPercentage,
				CategoryID:    req.CategoryID,
				Stock:         100,
			}, nil
		},
	}

	ctrl := controllers.NewProductController(mockService)
	router := gin.New()
	router.PUT("/api/v0.0/products/:id", ctrl.UpdateProduct)

	reqPayload := dto.UpdateProductReq{
		Name:          "Coca Cola 2L",
		Brand:         "The Coca-Cola Company",
		BasePrice:     1.99,
		TaxPercentage: 18,
		CategoryID:    2,
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("PUT", "/api/v0.0/products/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockProductService{
		DeleteProductFunc: func(ctx context.Context, id uint64) error {
			return nil
		},
	}

	ctrl := controllers.NewProductController(mockService)
	router := gin.New()
	router.DELETE("/api/v0.0/products/:id", ctrl.DeleteProduct)

	req, _ := http.NewRequest("DELETE", "/api/v0.0/products/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestRestockProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockProductService{
		RestockProductFunc: func(ctx context.Context, id uint64, quantity int) (*dto.ProductResponse, error) {
			return &dto.ProductResponse{
				ID:    id,
				Name:  "Coca Cola 1.5L",
				Stock: 50 + quantity,
			}, nil
		},
	}

	ctrl := controllers.NewProductController(mockService)
	router := gin.New()
	router.PATCH("/api/v0.0/products/:id/restock", ctrl.RestockProduct)

	reqPayload := dto.RestockReq{
		Quantity: 20,
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("PATCH", "/api/v0.0/products/1/restock", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetProductPrice_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockProductService{
		CalculatePriceFunc: func(ctx context.Context, id uint64) (*dto.PriceBreakdownResponse, error) {
			return &dto.PriceBreakdownResponse{
				ProductID:          id,
				BasePrice:          1.50,
				StandardFinalPrice: 1.77,
				AppliedPromotion: &dto.PromoShortResponse{
					ID:                 1,
					Name:               "Test Promo",
					DiscountPercentage: 10,
				},
				PromotionalFinalPrice: 1.59,
			}, nil
		},
	}

	ctrl := controllers.NewProductController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/products/:id/price", ctrl.GetProductPrice)

	req, _ := http.NewRequest("GET", "/api/v0.0/products/1/price", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}