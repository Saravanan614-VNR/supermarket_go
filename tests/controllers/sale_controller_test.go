/*
 * Contract ID: CTR-001 (Sale Controller Tests)
 * Service Name: SupermarketService
 * Description: Unit tests for SaleController verifying status code mappings and endpoint behaviors.
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
	"supermarket-backend/exceptions"
)

// MockSaleService implements services.SaleService.
type MockSaleService struct {
	OpenSaleFunc           func(ctx context.Context, cashierID uint64, req *dto.OpenSaleRequest) (*dto.SaleResponse, error)
	ListSalesFunc          func(ctx context.Context, page, size int, status string, clID, csID *uint64) (*dto.PaginatedSales, error)
	GetSaleByIDFunc        func(ctx context.Context, id uint64) (*dto.SaleDetailResponse, error)
	AddItemToSaleFunc      func(ctx context.Context, saleID uint64, req *dto.AddItemRequest) (*dto.SaleDetailResponse, error)
	UpdateItemQuantityFunc func(ctx context.Context, saleID, itemID uint64, req *dto.UpdateItemQtyRequest) (*dto.SaleDetailResponse, error)
	RemoveItemFromSaleFunc func(ctx context.Context, saleID, itemID uint64) (*dto.SaleDetailResponse, error)
	FinalizeSaleFunc       func(ctx context.Context, id uint64) (*dto.SaleResponse, error)
	CancelSaleFunc         func(ctx context.Context, id uint64) (*dto.SaleResponse, error)
}

func (m *MockSaleService) OpenSale(ctx context.Context, cashierID uint64, req *dto.OpenSaleRequest) (*dto.SaleResponse, error) {
	if m.OpenSaleFunc != nil {
		return m.OpenSaleFunc(ctx, cashierID, req)
	}
	return nil, nil
}

func (m *MockSaleService) ListSales(ctx context.Context, page, size int, status string, clID, csID *uint64) (*dto.PaginatedSales, error) {
	if m.ListSalesFunc != nil {
		return m.ListSalesFunc(ctx, page, size, status, clID, csID)
	}
	return nil, nil
}

func (m *MockSaleService) GetSaleByID(ctx context.Context, id uint64) (*dto.SaleDetailResponse, error) {
	if m.GetSaleByIDFunc != nil {
		return m.GetSaleByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSaleService) AddItemToSale(ctx context.Context, saleID uint64, req *dto.AddItemRequest) (*dto.SaleDetailResponse, error) {
	if m.AddItemToSaleFunc != nil {
		return m.AddItemToSaleFunc(ctx, saleID, req)
	}
	return nil, nil
}

func (m *MockSaleService) UpdateItemQuantity(ctx context.Context, saleID, itemID uint64, req *dto.UpdateItemQtyRequest) (*dto.SaleDetailResponse, error) {
	if m.UpdateItemQuantityFunc != nil {
		return m.UpdateItemQuantityFunc(ctx, saleID, itemID, req)
	}
	return nil, nil
}

func (m *MockSaleService) RemoveItemFromSale(ctx context.Context, saleID, itemID uint64) (*dto.SaleDetailResponse, error) {
	if m.RemoveItemFromSaleFunc != nil {
		return m.RemoveItemFromSaleFunc(ctx, saleID, itemID)
	}
	return nil, nil
}

func (m *MockSaleService) FinalizeSale(ctx context.Context, id uint64) (*dto.SaleResponse, error) {
	if m.FinalizeSaleFunc != nil {
		return m.FinalizeSaleFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSaleService) CancelSale(ctx context.Context, id uint64) (*dto.SaleResponse, error) {
	if m.CancelSaleFunc != nil {
		return m.CancelSaleFunc(ctx, id)
	}
	return nil, nil
}

func TestOpenSale_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockSaleService{
		OpenSaleFunc: func(ctx context.Context, cashierID uint64, req *dto.OpenSaleRequest) (*dto.SaleResponse, error) {
			return &dto.SaleResponse{
				ID:         10,
				TotalPrice: 0.0,
				Status:     "OPEN",
				ClientID:   req.ClientID,
				CashierID:  cashierID,
				CreatedAt:  time.Now(),
			}, nil
		},
	}

	ctrl := controllers.NewSaleController(mockService)
	router := gin.New()
	router.POST("/api/v1/sales", func(c *gin.Context) {
		c.Set("operatorID", uint64(5))
		c.Next()
	}, ctrl.OpenSale)

	clientIDVal := uint64(101)
	reqPayload := dto.OpenSaleRequest{
		ClientID: &clientIDVal,
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v1/sales", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp dto.SaleResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.ID != 10 || resp.CashierID != 5 || *resp.ClientID != 101 || resp.Status != "OPEN" {
		t.Errorf("unexpected response body: %+v", resp)
	}
}

func TestOpenSale_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockSaleService{}

	ctrl := controllers.NewSaleController(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			if appErr, ok := err.(*exceptions.AppError); ok {
				c.JSON(appErr.HTTPStatus, gin.H{"error_code": appErr.ErrorCode, "message": appErr.Message})
				return
			}
		}
	})
	router.POST("/api/v1/sales", ctrl.OpenSale) // No operatorID set, so it will fail with unauthorized

	req, _ := http.NewRequest("POST", "/api/v1/sales", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAddItemToSale_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockSaleService{
		AddItemToSaleFunc: func(ctx context.Context, saleID uint64, req *dto.AddItemRequest) (*dto.SaleDetailResponse, error) {
			return &dto.SaleDetailResponse{
				Sale: dto.SaleResponse{
					ID:         saleID,
					TotalPrice: 15.50,
					Status:     "OPEN",
				},
				LineItems: []dto.LineItemResponse{
					{
						ItemID:    1,
						ProductID: req.ProductID,
						Name:      "Mock Product",
						UnitPrice: 5.16,
						Quantity:  req.Quantity,
						SubTotal:  15.50,
					},
				},
			}, nil
		},
	}

	ctrl := controllers.NewSaleController(mockService)
	router := gin.New()
	router.POST("/api/v1/sales/:id/items", ctrl.AddItemToSale)

	reqPayload := dto.AddItemRequest{
		ProductID: 42,
		Quantity:  3,
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v1/sales/10/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp dto.SaleDetailResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Sale.ID != 10 || resp.Sale.TotalPrice != 15.50 || len(resp.LineItems) != 1 || resp.LineItems[0].ProductID != 42 {
		t.Errorf("unexpected response body: %+v", resp)
	}
}

func TestFinalizeSale_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockSaleService{
		FinalizeSaleFunc: func(ctx context.Context, id uint64) (*dto.SaleResponse, error) {
			return &dto.SaleResponse{
				ID:         id,
				TotalPrice: 20.0,
				Status:     "CLOSED",
			}, nil
		},
	}

	ctrl := controllers.NewSaleController(mockService)
	router := gin.New()
	router.POST("/api/v1/sales/:id/finalize", ctrl.FinalizeSale)

	req, _ := http.NewRequest("POST", "/api/v1/sales/10/finalize", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp dto.SaleResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.ID != 10 || resp.Status != "CLOSED" {
		t.Errorf("unexpected response body: %+v", resp)
	}
}
