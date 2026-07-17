/*
 * Contract ID: CTR-001 (Client Controller Tests)
 * Service Name: SupermarketService
 * Description: Unit tests for ClientController with status code mapping and request validation.
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
	"supermarket-backend/exceptions"
)

// MockClientService implements services.ClientService.
type MockClientService struct {
	CreateClientFunc  func(ctx context.Context, req *dto.CreateClientReq) (*dto.ClientResponse, error)
	ListClientsFunc   func(ctx context.Context, page, size int, dni, email string) (*dto.PaginatedClients, error)
	GetClientByIDFunc func(ctx context.Context, id uint64) (*dto.ClientResponse, error)
	UpdateClientFunc  func(ctx context.Context, id uint64, req *dto.UpdateClientReq) (*dto.ClientResponse, error)
	DeleteClientFunc  func(ctx context.Context, id uint64) error
	SearchClientFunc  func(ctx context.Context, dni, email string) (*dto.ClientResponse, error)
}

func (m *MockClientService) CreateClient(ctx context.Context, req *dto.CreateClientReq) (*dto.ClientResponse, error) {
	if m.CreateClientFunc != nil {
		return m.CreateClientFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockClientService) ListClients(ctx context.Context, page, size int, dni, email string) (*dto.PaginatedClients, error) {
	if m.ListClientsFunc != nil {
		return m.ListClientsFunc(ctx, page, size, dni, email)
	}
	return nil, nil
}

func (m *MockClientService) GetClientByID(ctx context.Context, id uint64) (*dto.ClientResponse, error) {
	if m.GetClientByIDFunc != nil {
		return m.GetClientByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockClientService) UpdateClient(ctx context.Context, id uint64, req *dto.UpdateClientReq) (*dto.ClientResponse, error) {
	if m.UpdateClientFunc != nil {
		return m.UpdateClientFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockClientService) DeleteClient(ctx context.Context, id uint64) error {
	if m.DeleteClientFunc != nil {
		return m.DeleteClientFunc(ctx, id)
	}
	return nil
}

func (m *MockClientService) SearchClient(ctx context.Context, dni, email string) (*dto.ClientResponse, error) {
	if m.SearchClientFunc != nil {
		return m.SearchClientFunc(ctx, dni, email)
	}
	return nil, nil
}

func TestCreateClient_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockClientService{
		CreateClientFunc: func(ctx context.Context, req *dto.CreateClientReq) (*dto.ClientResponse, error) {
			return &dto.ClientResponse{
				ID:    1,
				Name:  req.Name,
				DNI:   req.DNI,
				Email: req.Email,
			}, nil
		},
	}

	ctrl := controllers.NewClientController(mockService)
	router := gin.New()
	router.POST("/api/v1/clients", ctrl.CreateClient)

	reqPayload := dto.CreateClientReq{
		Name:  "John Doe",
		DNI:   "1712345678",
		Email: "john.doe@example.com",
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp dto.ClientResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != 1 || resp.Name != "John Doe" || resp.DNI != "1712345678" {
		t.Errorf("unexpected response body: %+v", resp)
	}
}

func TestCreateClient_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockClientService{}

	ctrl := controllers.NewClientController(mockService)
	router := gin.New()
	router.POST("/api/v1/clients", ctrl.CreateClient)

	// Invalid DNI (not 10 characters, not numeric)
	reqPayload := dto.CreateClientReq{
		Name:  "JD",
		DNI:   "123",
		Email: "invalid-email",
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Gin's model binding returns 400 bad request for validation failures
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for validation failure, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateClient_ConflictError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockClientService{
		CreateClientFunc: func(ctx context.Context, req *dto.CreateClientReq) (*dto.ClientResponse, error) {
			return nil, exceptions.NewConflictError("Client with this DNI already exists")
		},
	}

	ctrl := controllers.NewClientController(mockService)
	router := gin.New()
	
	// Add custom error middleware to map exceptions.AppError HTTP Status
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			if appErr, ok := err.(*exceptions.AppError); ok {
				c.JSON(appErr.HTTPStatus, gin.H{"error_code": appErr.ErrorCode, "message": appErr.Message})
				return
			}
			c.JSON(500, gin.H{"message": err.Error()})
		}
	})
	
	router.POST("/api/v1/clients", ctrl.CreateClient)

	reqPayload := dto.CreateClientReq{
		Name:  "John Doe",
		DNI:   "1712345678",
		Email: "john.doe@example.com",
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestGetClientByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockClientService{
		GetClientByIDFunc: func(ctx context.Context, id uint64) (*dto.ClientResponse, error) {
			return nil, exceptions.NewNotFoundError("Client not found")
		},
	}

	ctrl := controllers.NewClientController(mockService)
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
	router.GET("/api/v1/clients/:id", ctrl.GetClientByID)

	req, _ := http.NewRequest("GET", "/api/v1/clients/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
