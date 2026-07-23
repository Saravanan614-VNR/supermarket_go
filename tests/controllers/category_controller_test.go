/*
 * Description: Unit tests for CategoryController verifying status code mappings and endpoint behaviors.
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

type MockCategoryService struct {
	CreateCategoryFunc       func(ctx context.Context, req *dto.CreateCategoryReq) (*dto.CategoryResponse, error)
	ListCategoriesFunc       func(ctx context.Context, page, size int, filter string) (*dto.PaginatedCategories, error)
	GetCategoryByIDFunc      func(ctx context.Context, id uint64) (*dto.CategoryResponse, error)
	UpdateCategoryFunc       func(ctx context.Context, id uint64, req *dto.UpdateCategoryReq) (*dto.CategoryResponse, error)
	DeleteCategoryFunc       func(ctx context.Context, id uint64) error
	ListCategoryProductsFunc func(ctx context.Context, categoryID uint64, page, size int) (*dto.PaginatedProducts, error)
}

func (m *MockCategoryService) CreateCategory(ctx context.Context, req *dto.CreateCategoryReq) (*dto.CategoryResponse, error) {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockCategoryService) ListCategories(ctx context.Context, page, size int, filter string) (*dto.PaginatedCategories, error) {
	if m.ListCategoriesFunc != nil {
		return m.ListCategoriesFunc(ctx, page, size, filter)
	}
	return nil, nil
}

func (m *MockCategoryService) GetCategoryByID(ctx context.Context, id uint64) (*dto.CategoryResponse, error) {
	if m.GetCategoryByIDFunc != nil {
		return m.GetCategoryByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockCategoryService) UpdateCategory(ctx context.Context, id uint64, req *dto.UpdateCategoryReq) (*dto.CategoryResponse, error) {
	if m.UpdateCategoryFunc != nil {
		return m.UpdateCategoryFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockCategoryService) DeleteCategory(ctx context.Context, id uint64) error {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, id)
	}
	return nil
}

func (m *MockCategoryService) ListCategoryProducts(ctx context.Context, categoryID uint64, page, size int) (*dto.PaginatedProducts, error) {
	if m.ListCategoryProductsFunc != nil {
		return m.ListCategoryProductsFunc(ctx, categoryID, page, size)
	}
	return nil, nil
}

var _ services.CategoryService = (*MockCategoryService)(nil)

func TestCreateCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockCategoryService{
		CreateCategoryFunc: func(ctx context.Context, req *dto.CreateCategoryReq) (*dto.CategoryResponse, error) {
			return &dto.CategoryResponse{
				ID:          1,
				Name:        req.Name,
				Description: req.Description,
			}, nil
		},
	}

	ctrl := controllers.NewCategoryController(mockService)
	router := gin.New()
	router.POST("/api/v0.0/categories", ctrl.CreateCategory)

	reqPayload := dto.CreateCategoryReq{
		Name:        "Snacks",
		Description: "Chips, nuts and savory items",
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v0.0/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestListCategories_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockCategoryService{
		ListCategoriesFunc: func(ctx context.Context, page, size int, filter string) (*dto.PaginatedCategories, error) {
			return &dto.PaginatedCategories{
				Items: []dto.CategoryResponse{
					{ID: 1, Name: "Snacks", Description: "Chips, nuts and savory items"},
				},
				TotalCount: 1,
			}, nil
		},
	}

	ctrl := controllers.NewCategoryController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/categories", ctrl.ListCategories)

	req, _ := http.NewRequest("GET", "/api/v0.0/categories?page=1&size=10&filter=snacks", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetCategoryByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockCategoryService{
		GetCategoryByIDFunc: func(ctx context.Context, id uint64) (*dto.CategoryResponse, error) {
			return &dto.CategoryResponse{
				ID:          id,
				Name:        "Snacks",
				Description: "Chips, nuts and savory items",
			}, nil
		},
	}

	ctrl := controllers.NewCategoryController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/categories/:id", ctrl.GetCategoryByID)

	req, _ := http.NewRequest("GET", "/api/v0.0/categories/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockCategoryService{
		UpdateCategoryFunc: func(ctx context.Context, id uint64, req *dto.UpdateCategoryReq) (*dto.CategoryResponse, error) {
			return &dto.CategoryResponse{
				ID:          id,
				Name:        req.Name,
				Description: req.Description,
			}, nil
		},
	}

	ctrl := controllers.NewCategoryController(mockService)
	router := gin.New()
	router.PUT("/api/v0.0/categories/:id", ctrl.UpdateCategory)

	reqPayload := dto.UpdateCategoryReq{
		Name:        "Snacks Updated",
		Description: "Updated snacks description",
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("PUT", "/api/v0.0/categories/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockCategoryService{
		DeleteCategoryFunc: func(ctx context.Context, id uint64) error {
			return nil
		},
	}

	ctrl := controllers.NewCategoryController(mockService)
	router := gin.New()
	router.DELETE("/api/v0.0/categories/:id", ctrl.DeleteCategory)

	req, _ := http.NewRequest("DELETE", "/api/v0.0/categories/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestListCategoryProducts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockCategoryService{
		ListCategoryProductsFunc: func(ctx context.Context, categoryID uint64, page, size int) (*dto.PaginatedProducts, error) {
			return &dto.PaginatedProducts{
				Items: []dto.ProductResponse{
					{ID: 1, Name: "Potato Chips", Brand: "Lays", BasePrice: 1.99, TaxPercentage: 10, CategoryID: categoryID, Stock: 100},
				},
				TotalCount: 1,
			}, nil
		},
	}

	ctrl := controllers.NewCategoryController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/categories/:id/products", ctrl.ListCategoryProducts)

	req, _ := http.NewRequest("GET", "/api/v0.0/categories/1/products?page=1&size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}