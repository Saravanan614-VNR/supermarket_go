package tests

import (
	"context"
	"testing"

	dto "supermarket-backend/dtos"
	"supermarket-backend/entities"
	"supermarket-backend/repositories"
	"supermarket-backend/services"
)

type MockClientRepository struct {
	CreateFunc      func(ctx context.Context, client *entities.Client) error
	FindByIDFunc    func(ctx context.Context, id uint64) (*entities.Client, error)
	UpdateFunc      func(ctx context.Context, client *entities.Client) error
	DeleteFunc      func(ctx context.Context, id uint64) error
	FindAllFunc     func(ctx context.Context, offset, limit int) ([]entities.Client, int64, error)
	FindByDNIFunc   func(ctx context.Context, dni string) (*entities.Client, error)
	FindByEmailFunc func(ctx context.Context, email string) (*entities.Client, error)
	SearchFunc      func(ctx context.Context, query string) ([]entities.Client, error)
}

func (m *MockClientRepository) Create(ctx context.Context, client *entities.Client) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, client)
	}
	return nil
}

func (m *MockClientRepository) FindByID(ctx context.Context, id uint64) (*entities.Client, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockClientRepository) Update(ctx context.Context, client *entities.Client) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, client)
	}
	return nil
}

func (m *MockClientRepository) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockClientRepository) FindAll(ctx context.Context, offset, limit int) ([]entities.Client, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockClientRepository) FindByDNI(ctx context.Context, dni string) (*entities.Client, error) {
	if m.FindByDNIFunc != nil {
		return m.FindByDNIFunc(ctx, dni)
	}
	return nil, nil
}

func (m *MockClientRepository) FindByEmail(ctx context.Context, email string) (*entities.Client, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockClientRepository) Search(ctx context.Context, query string) ([]entities.Client, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query)
	}
	return nil, nil
}

var _ repositories.ClientRepository = (*MockClientRepository)(nil)

func TestClientService_CreateClient_Success(t *testing.T) {
	mockRepo := &MockClientRepository{
		FindByDNIFunc: func(ctx context.Context, dni string) (*entities.Client, error) {
			return nil, nil // DNI is unique
		},
		FindByEmailFunc: func(ctx context.Context, email string) (*entities.Client, error) {
			return nil, nil // Email is unique
		},
		CreateFunc: func(ctx context.Context, client *entities.Client) error {
			client.ID = 1
			return nil
		},
	}

	srv := services.NewClientService(mockRepo, nil)

	req := &dto.CreateClientReq{
		Name:  "Jane Doe",
		DNI:   "1712345678",
		Email: "jane.doe@example.com",
	}

	res, err := srv.CreateClient(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.DNI != "1712345678" || res.Name != "Jane Doe" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestClientService_ListClients_Success(t *testing.T) {
	mockRepo := &MockClientRepository{
		FindAllFunc: func(ctx context.Context, offset, limit int) ([]entities.Client, int64, error) {
			return []entities.Client{
				{ID: 1, Name: "Jane Doe", DNI: "1712345678", Email: "jane.doe@example.com"},
			}, 1, nil
		},
	}

	srv := services.NewClientService(mockRepo, nil)

	res, err := srv.ListClients(context.Background(), 1, 10, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalCount != 1 || len(res.Items) != 1 || res.Items[0].Name != "Jane Doe" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestClientService_GetClientByID_Success(t *testing.T) {
	mockRepo := &MockClientRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Client, error) {
			return &entities.Client{
				ID:    id,
				Name:  "Jane Doe",
				DNI:   "1712345678",
				Email: "jane.doe@example.com",
			}, nil
		},
	}

	srv := services.NewClientService(mockRepo, nil)

	res, err := srv.GetClientByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.Name != "Jane Doe" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestClientService_UpdateClient_Success(t *testing.T) {
	mockRepo := &MockClientRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Client, error) {
			return &entities.Client{
				ID:    id,
				Name:  "Jane Doe",
				DNI:   "1712345678",
				Email: "jane.doe@example.com",
			}, nil
		},
		FindByDNIFunc: func(ctx context.Context, dni string) (*entities.Client, error) {
			return nil, nil
		},
		FindByEmailFunc: func(ctx context.Context, email string) (*entities.Client, error) {
			return nil, nil
		},
		UpdateFunc: func(ctx context.Context, client *entities.Client) error {
			return nil
		},
	}

	srv := services.NewClientService(mockRepo, nil)

	req := &dto.UpdateClientReq{
		Name:  "Jane Updated",
		DNI:   "1712345678",
		Email: "jane.updated@example.com",
	}

	res, err := srv.UpdateClient(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Name != "Jane Updated" || res.Email != "jane.updated@example.com" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestClientService_DeleteClient_Success(t *testing.T) {
	mockRepo := &MockClientRepository{
		FindByIDFunc: func(ctx context.Context, id uint64) (*entities.Client, error) {
			return &entities.Client{
				ID:    id,
				Name:  "Jane Doe",
				DNI:   "1712345678",
				Email: "jane.doe@example.com",
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uint64) error {
			return nil
		},
	}

	srv := services.NewClientService(mockRepo, nil)

	err := srv.DeleteClient(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientService_SearchClient_Success(t *testing.T) {
	mockRepo := &MockClientRepository{
		FindByDNIFunc: func(ctx context.Context, dni string) (*entities.Client, error) {
			return &entities.Client{
				ID:    1,
				Name:  "Jane Doe",
				DNI:   dni,
				Email: "jane.doe@example.com",
			}, nil
		},
	}

	srv := services.NewClientService(mockRepo, nil)

	res, err := srv.SearchClient(context.Background(), "1712345678", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 1 || res.DNI != "1712345678" {
		t.Errorf("unexpected response: %+v", res)
	}
}