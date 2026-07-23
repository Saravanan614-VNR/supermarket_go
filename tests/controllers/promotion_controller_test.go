/*
 * Description: Unit tests for PromotionController verifying status code mappings and endpoint behaviors.
 */

package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"supermarket-backend/controllers"
	dto "supermarket-backend/dtos"
	"supermarket-backend/services"
)

type MockPromotionService struct {
	CreatePromotionFunc           func(ctx context.Context, req *dto.CreatePromotionReq) (*dto.PromotionResponse, error)
	ListPromotionsFunc            func(ctx context.Context, page, size int) (*dto.PaginatedPromotions, error)
	GetPromotionByIDFunc          func(ctx context.Context, id uint64) (*dto.PromotionResponse, error)
	UpdatePromotionFunc           func(ctx context.Context, id uint64, req *dto.UpdatePromotionReq) (*dto.PromotionResponse, error)
	DeletePromotionFunc           func(ctx context.Context, id uint64) error
	LinkProductsToPromotionFunc   func(ctx context.Context, id uint64, productIDs []uint64) error
	UnlinkProductFromPromotionFunc func(ctx context.Context, id uint64, productID uint64) error
}

func (m *MockPromotionService) CreatePromotion(ctx context.Context, req *dto.CreatePromotionReq) (*dto.PromotionResponse, error) {
	if m.CreatePromotionFunc != nil {
		return m.CreatePromotionFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockPromotionService) ListPromotions(ctx context.Context, page, size int) (*dto.PaginatedPromotions, error) {
	if m.ListPromotionsFunc != nil {
		return m.ListPromotionsFunc(ctx, page, size)
	}
	return nil, nil
}

func (m *MockPromotionService) GetPromotionByID(ctx context.Context, id uint64) (*dto.PromotionResponse, error) {
	if m.GetPromotionByIDFunc != nil {
		return m.GetPromotionByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockPromotionService) UpdatePromotion(ctx context.Context, id uint64, req *dto.UpdatePromotionReq) (*dto.PromotionResponse, error) {
	if m.UpdatePromotionFunc != nil {
		return m.UpdatePromotionFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockPromotionService) DeletePromotion(ctx context.Context, id uint64) error {
	if m.DeletePromotionFunc != nil {
		return m.DeletePromotionFunc(ctx, id)
	}
	return nil
}

func (m *MockPromotionService) LinkProductsToPromotion(ctx context.Context, id uint64, productIDs []uint64) error {
	if m.LinkProductsToPromotionFunc != nil {
		return m.LinkProductsToPromotionFunc(ctx, id, productIDs)
	}
	return nil
}

func (m *MockPromotionService) UnlinkProductFromPromotion(ctx context.Context, id uint64, productID uint64) error {
	if m.UnlinkProductFromPromotionFunc != nil {
		return m.UnlinkProductFromPromotionFunc(ctx, id, productID)
	}
	return nil
}

var _ services.PromotionService = (*MockPromotionService)(nil)

func TestCreatePromotion_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockPromotionService{
		CreatePromotionFunc: func(ctx context.Context, req *dto.CreatePromotionReq) (*dto.PromotionResponse, error) {
			return &dto.PromotionResponse{
				ID:                 1,
				Name:               req.Name,
				DiscountPercentage: req.DiscountPercentage,
				StartDate:          req.StartDate,
				EndDate:            req.EndDate,
			}, nil
		},
	}

	ctrl := controllers.NewPromotionController(mockService)
	router := gin.New()
	router.POST("/api/v0.0/promotions", ctrl.CreatePromotion)

	now := time.Now()
	reqPayload := dto.CreatePromotionReq{
		Name:               "Summer Discount",
		DiscountPercentage: 15.0,
		StartDate:          now,
		EndDate:            now.Add(24 * time.Hour),
		ProductIDs:         []uint64{1, 2},
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v0.0/promotions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestListPromotions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockPromotionService{
		ListPromotionsFunc: func(ctx context.Context, page, size int) (*dto.PaginatedPromotions, error) {
			return &dto.PaginatedPromotions{
				Items: []dto.PromotionResponse{
					{ID: 1, Name: "Summer Discount", DiscountPercentage: 15.0, StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour)},
				},
				TotalCount: 1,
			}, nil
		},
	}

	ctrl := controllers.NewPromotionController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/promotions", ctrl.ListPromotions)

	req, _ := http.NewRequest("GET", "/api/v0.0/promotions?page=1&size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetPromotionByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockPromotionService{
		GetPromotionByIDFunc: func(ctx context.Context, id uint64) (*dto.PromotionResponse, error) {
			return &dto.PromotionResponse{
				ID:                 id,
				Name:               "Summer Discount",
				DiscountPercentage: 15.0,
			}, nil
		},
	}

	ctrl := controllers.NewPromotionController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/promotions/:id", ctrl.GetPromotionByID)

	req, _ := http.NewRequest("GET", "/api/v0.0/promotions/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdatePromotion_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockPromotionService{
		UpdatePromotionFunc: func(ctx context.Context, id uint64, req *dto.UpdatePromotionReq) (*dto.PromotionResponse, error) {
			return &dto.PromotionResponse{
				ID:                 id,
				Name:               req.Name,
				DiscountPercentage: req.DiscountPercentage,
			}, nil
		},
	}

	ctrl := controllers.NewPromotionController(mockService)
	router := gin.New()
	router.PUT("/api/v0.0/promotions/:id", ctrl.UpdatePromotion)

	reqPayload := dto.UpdatePromotionReq{
		Name:               "Summer Discount Updated",
		DiscountPercentage: 20.0,
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("PUT", "/api/v0.0/promotions/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeletePromotion_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockPromotionService{
		DeletePromotionFunc: func(ctx context.Context, id uint64) error {
			return nil
		},
	}

	ctrl := controllers.NewPromotionController(mockService)
	router := gin.New()
	router.DELETE("/api/v0.0/promotions/:id", ctrl.DeletePromotion)

	req, _ := http.NewRequest("DELETE", "/api/v0.0/promotions/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestLinkProductsToPromotion_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockPromotionService{
		LinkProductsToPromotionFunc: func(ctx context.Context, id uint64, productIDs []uint64) error {
			return nil
		},
	}

	ctrl := controllers.NewPromotionController(mockService)
	router := gin.New()
	router.POST("/api/v0.0/promotions/:id/products", ctrl.LinkProductsToPromotion)

	reqPayload := dto.LinkProductsReq{
		ProductIDs: []uint64{101, 102},
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v0.0/promotions/1/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUnlinkProductFromPromotion_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockPromotionService{
		UnlinkProductFromPromotionFunc: func(ctx context.Context, id uint64, productID uint64) error {
			return nil
		},
	}

	ctrl := controllers.NewPromotionController(mockService)
	router := gin.New()
	router.DELETE("/api/v0.0/promotions/:id/products/:productId", ctrl.UnlinkProductFromPromotion)

	req, _ := http.NewRequest("DELETE", "/api/v0.0/promotions/1/products/101", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// UnlinkProductFromPromotion returns 200 with a confirmation body, not 204.
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}