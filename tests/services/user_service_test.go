package tests

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"supermarket-backend/entities"
	dto "supermarket-backend/dtos"
	"supermarket-backend/services"
)

type MockUserRepository struct {
	CreateFunc         func(ctx context.Context, user *entities.User) error
	FindByUsernameFunc func(ctx context.Context, username string) (*entities.User, error)
	FindByIDFunc       func(ctx context.Context, id uint64) (*entities.User, error)
	UpdateFunc         func(ctx context.Context, user *entities.User) error
	DeleteFunc         func(ctx context.Context, id uint64) error
	FindAllFunc        func(ctx context.Context, offset, limit int) ([]entities.User, int64, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	if m.FindByUsernameFunc != nil {
		return m.FindByUsernameFunc(ctx, username)
	}
	return nil, nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint64) (*entities.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUserRepository) FindAll(ctx context.Context, offset, limit int) ([]entities.User, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, offset, limit)
	}
	return nil, 0, nil
}

type MockAuthConfig struct {
	Expiration time.Duration
	Secret     string
}

func (m *MockAuthConfig) GetJWTExpiration() time.Duration {
	if m.Expiration == 0 {
		return 24 * time.Hour
	}
	return m.Expiration
}

func (m *MockAuthConfig) GetJWTSecret() string {
	if m.Secret == "" {
		return "test-secret"
	}
	return m.Secret
}

type MockTokenBlacklist struct {
	BlacklistFunc func(token string, ttl time.Duration)
}

func (m *MockTokenBlacklist) Blacklist(token string, ttl time.Duration) {
	if m.BlacklistFunc != nil {
		m.BlacklistFunc(token, ttl)
	}
}

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes)
}

func TestUserService_RegisterUser_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		FindByUsernameFunc: func(ctx context.Context, username string) (*entities.User, error) {
			return nil, nil // Unique username
		},
		CreateFunc: func(ctx context.Context, user *entities.User) error {
			user.ID = 1
			return nil
		},
	}
	mockCfg := &MockAuthConfig{}
	mockBl := &MockTokenBlacklist{}

	srv := services.NewUserService(mockRepo, nil, mockCfg, mockBl)

	req := &dto.RegisterRequest{
		FullName: "Test Operator",
		Username: "operator",
		Password: "secure-password",
		Role:     "CASHIER",
	}

	res, err := srv.RegisterUser(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.Username != "operator" || res.Role != "CASHIER" {
		t.Errorf("unexpected response content: %+v", res)
	}
}

func TestUserService_LoginUser_Success(t *testing.T) {
	hashedPassword := hashPassword("correct-password")

	mockRepo := &MockUserRepository{
		FindByUsernameFunc: func(ctx context.Context, username string) (*entities.User, error) {
			return &entities.User{
				ID:       1,
				Username: username,
				Password: hashedPassword,
				Role:     "CASHIER",
			}, nil
		},
	}
	mockCfg := &MockAuthConfig{}
	mockBl := &MockTokenBlacklist{}

	srv := services.NewUserService(mockRepo, nil, mockCfg, mockBl)

	req := &dto.LoginRequest{
		Username: "operator",
		Password: "correct-password",
	}

	res, err := srv.LoginUser(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Token == "" {
		t.Error("expected generated JWT token, got empty string")
	}
}

func TestUserService_LogoutUser_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	mockCfg := &MockAuthConfig{}
	called := false
	mockBl := &MockTokenBlacklist{
		BlacklistFunc: func(token string, ttl time.Duration) {
			called = true
		},
	}

	srv := services.NewUserService(mockRepo, nil, mockCfg, mockBl)

	err := srv.LogoutUser(context.Background(), "mock-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("expected blacklist method to be called")
	}
}

func TestUserService_ListUsers_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		FindAllFunc: func(ctx context.Context, offset, limit int) ([]entities.User, int64, error) {
			return []entities.User{
				{ID: 1, FullName: "John Doe", Username: "john", Role: "CASHIER"},
			}, 1, nil
		},
	}
	mockCfg := &MockAuthConfig{}
	mockBl := &MockTokenBlacklist{}

	srv := services.NewUserService(mockRepo, nil, mockCfg, mockBl)

	res, err := srv.ListUsers(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalCount != 1 || len(res.Items) != 1 || res.Items[0].Username != "john" {
		t.Errorf("unexpected list response: %+v", res)
	}
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.User, error) {
			return &entities.User{
				ID:       id,
				FullName: "John Doe",
				Username: "john",
				Role:     "CASHIER",
			}, nil
		},
	}
	mockCfg := &MockAuthConfig{}
	mockBl := &MockTokenBlacklist{}

	srv := services.NewUserService(mockRepo, nil, mockCfg, mockBl)

	res, err := srv.GetUserByID(context.Background(), 5, 5, "CASHIER")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 5 || res.Username != "john" {
		t.Errorf("unexpected profile: %+v", res)
	}
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.User, error) {
			return &entities.User{
				ID:       id,
				FullName: "John Old",
				Username: "john",
				Role:     "CASHIER",
			}, nil
		},
		UpdateFunc: func(ctx context.Context, user *entities.User) error {
			return nil
		},
	}
	mockCfg := &MockAuthConfig{}
	mockBl := &MockTokenBlacklist{}

	srv := services.NewUserService(mockRepo, nil, mockCfg, mockBl)

	req := &dto.UpdateUserRequest{
		FullName: "John New",
	}

	res, err := srv.UpdateUser(context.Background(), 5, 5, "ADMIN", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// UpdateUser only ever mutates FullName/Password - Role is immutable via this endpoint.
	if res.FullName != "John New" || res.Role != "CASHIER" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.User, error) {
			return &entities.User{
				ID:       id,
				FullName: "John Doe",
				Username: "john",
				Role:     "CASHIER",
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uint64) error {
			return nil
		},
	}
	mockCfg := &MockAuthConfig{}
	mockBl := &MockTokenBlacklist{}

	srv := services.NewUserService(mockRepo, nil, mockCfg, mockBl)

	err := srv.DeleteUser(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}