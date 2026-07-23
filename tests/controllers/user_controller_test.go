/*
 * Description: Unit tests for UserController verifying status code mappings and endpoint behaviors.
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
)

type MockUserService struct {
	RegisterUserFunc func(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	LoginUserFunc    func(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	LogoutUserFunc   func(ctx context.Context, token string) error
	ListUsersFunc    func(ctx context.Context, page, size int) (*dto.PaginatedUsers, error)
	GetUserByIDFunc  func(ctx context.Context, id uint64, reqUserID uint64, reqRole string) (*dto.UserResponse, error)
	UpdateUserFunc   func(ctx context.Context, id uint64, reqUserID uint64, reqRole string, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUserFunc   func(ctx context.Context, id uint64) error
}

func (m *MockUserService) RegisterUser(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	if m.RegisterUserFunc != nil {
		return m.RegisterUserFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockUserService) LoginUser(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	if m.LoginUserFunc != nil {
		return m.LoginUserFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockUserService) LogoutUser(ctx context.Context, token string) error {
	if m.LogoutUserFunc != nil {
		return m.LogoutUserFunc(ctx, token)
	}
	return nil
}

func (m *MockUserService) ListUsers(ctx context.Context, page, size int) (*dto.PaginatedUsers, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, page, size)
	}
	return nil, nil
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uint64, reqUserID uint64, reqRole string) (*dto.UserResponse, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id, reqUserID, reqRole)
	}
	return nil, nil
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uint64, reqUserID uint64, reqRole string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, id, reqUserID, reqRole, req)
	}
	return nil, nil
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint64) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

func TestRegisterUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		RegisterUserFunc: func(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
			return &dto.UserResponse{
				ID:        1,
				FullName:  req.FullName,
				Username:  req.Username,
				Role:      req.Role,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	ctrl := controllers.NewUserController(mockService)
	router := gin.New()
	router.POST("/api/v0.0/auth/register", ctrl.RegisterUser)

	reqPayload := dto.RegisterRequest{
		FullName: "Test Admin",
		Username: "admin",
		Password: "password123",
		Role:     "ADMIN",
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v0.0/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestLoginUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		LoginUserFunc: func(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
			return &dto.LoginResponse{
				Token: "mock-jwt-token",
			}, nil
		},
	}

	ctrl := controllers.NewUserController(mockService)
	router := gin.New()
	router.POST("/api/v0.0/auth/login", ctrl.LoginUser)

	reqPayload := dto.LoginRequest{
		Username: "admin",
		Password: "password123",
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/api/v0.0/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestLogoutUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		LogoutUserFunc: func(ctx context.Context, token string) error {
			return nil
		},
	}

	ctrl := controllers.NewUserController(mockService)
	router := gin.New()
	router.POST("/api/v0.0/auth/logout", ctrl.LogoutUser)

	req, _ := http.NewRequest("POST", "/api/v0.0/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer mock-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestListUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		ListUsersFunc: func(ctx context.Context, page, size int) (*dto.PaginatedUsers, error) {
			return &dto.PaginatedUsers{
				Items: []dto.UserResponse{
					{ID: 1, FullName: "User One", Username: "user1", Role: "CASHIER"},
				},
				TotalCount: 1,
			}, nil
		},
	}

	ctrl := controllers.NewUserController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/users", ctrl.ListUsers)

	req, _ := http.NewRequest("GET", "/api/v0.0/users?page=1&size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetUserByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		GetUserByIDFunc: func(ctx context.Context, id uint64, reqUserID uint64, reqRole string) (*dto.UserResponse, error) {
			return &dto.UserResponse{
				ID:       id,
				FullName: "John Doe",
				Username: "john",
				Role:     "CASHIER",
			}, nil
		},
	}

	ctrl := controllers.NewUserController(mockService)
	router := gin.New()
	router.GET("/api/v0.0/users/:id", func(c *gin.Context) {
		c.Set("operatorID", uint64(5))
		c.Set("role", "ADMIN")
		c.Next()
	}, ctrl.GetUserByID)

	req, _ := http.NewRequest("GET", "/api/v0.0/users/5", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		UpdateUserFunc: func(ctx context.Context, id uint64, reqUserID uint64, reqRole string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
			return &dto.UserResponse{
				ID:       id,
				FullName: req.FullName,
				Username: "john",
				Role:     "CASHIER",
			}, nil
		},
	}

	ctrl := controllers.NewUserController(mockService)
	router := gin.New()
	router.PUT("/api/v0.0/users/:id", func(c *gin.Context) {
		c.Set("operatorID", uint64(5))
		c.Set("role", "ADMIN")
		c.Next()
	}, ctrl.UpdateUser)

	reqPayload := dto.UpdateUserRequest{
		FullName: "John Updated",
	}
	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("PUT", "/api/v0.0/users/5", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		DeleteUserFunc: func(ctx context.Context, id uint64) error {
			return nil
		},
	}

	ctrl := controllers.NewUserController(mockService)
	router := gin.New()
	router.DELETE("/api/v0.0/users/:id", ctrl.DeleteUser)

	req, _ := http.NewRequest("DELETE", "/api/v0.0/users/5", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}