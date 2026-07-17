package services

import (
	"context"

	"gorm.io/gorm"

	dto "supermarket-backend/dtos"
	"supermarket-backend/entities"
	"supermarket-backend/exceptions"
	"supermarket-backend/repositories"
)

type clientServiceImpl struct {
	clientRepo repositories.ClientRepository
	db         *gorm.DB
}

// NewClientService constructs a new ClientService implementation.
func NewClientService(clientRepo repositories.ClientRepository, db *gorm.DB) ClientService {
	return &clientServiceImpl{
		clientRepo: clientRepo,
		db:         db,
	}
}

func (s *clientServiceImpl) CreateClient(ctx context.Context, req *dto.CreateClientReq) (*dto.ClientResponse, error) {
	// 1. Verify DNI uniqueness
	existingDNI, err := s.clientRepo.FindByDNI(ctx, req.DNI)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during DNI check: " + err.Error())
	}
	if existingDNI != nil {
		return nil, exceptions.NewConflictError("Ya existe un cliente con este DNI!")
	}

	// 2. Verify Email uniqueness
	existingEmail, err := s.clientRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during Email check: " + err.Error())
	}
	if existingEmail != nil {
		return nil, exceptions.NewConflictError("Ya existe un cliente con este Email!")
	}

	client := &entities.Client{
		Name:  req.Name,
		DNI:   req.DNI,
		Email: req.Email,
	}

	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, exceptions.NewInternalError("Failed to register client: " + err.Error())
	}

	return &dto.ClientResponse{
		ID:        client.ID,
		Name:      client.Name,
		DNI:       client.DNI,
		Email:     client.Email,
		CreatedAt: client.CreatedAt,
	}, nil
}

func (s *clientServiceImpl) ListClients(ctx context.Context, page, size int, dni, email string) (*dto.PaginatedClients, error) {
	// If dni or email are present, look up specifically
	if dni != "" {
		cli, err := s.clientRepo.FindByDNI(ctx, dni)
		if err != nil {
			return nil, exceptions.NewInternalError("Failed to lookup client by DNI: " + err.Error())
		}
		if cli == nil {
			return &dto.PaginatedClients{Items: []dto.ClientResponse{}, TotalCount: 0}, nil
		}
		return &dto.PaginatedClients{
			Items: []dto.ClientResponse{
				{
					ID:        cli.ID,
					Name:      cli.Name,
					DNI:       cli.DNI,
					Email:     cli.Email,
					CreatedAt: cli.CreatedAt,
				},
			},
			TotalCount: 1,
		}, nil
	}

	if email != "" {
		cli, err := s.clientRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, exceptions.NewInternalError("Failed to lookup client by Email: " + err.Error())
		}
		if cli == nil {
			return &dto.PaginatedClients{Items: []dto.ClientResponse{}, TotalCount: 0}, nil
		}
		return &dto.PaginatedClients{
			Items: []dto.ClientResponse{
				{
					ID:        cli.ID,
					Name:      cli.Name,
					DNI:       cli.DNI,
					Email:     cli.Email,
					CreatedAt: cli.CreatedAt,
				},
			},
			TotalCount: 1,
		}, nil
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size

	clients, total, err := s.clientRepo.FindAll(ctx, offset, size)
	if err != nil {
		return nil, exceptions.NewInternalError("Failed to list clients: " + err.Error())
	}

	items := make([]dto.ClientResponse, len(clients))
	for i, cli := range clients {
		items[i] = dto.ClientResponse{
			ID:        cli.ID,
			Name:      cli.Name,
			DNI:       cli.DNI,
			Email:     cli.Email,
			CreatedAt: cli.CreatedAt,
		}
	}

	return &dto.PaginatedClients{
		Items:      items,
		TotalCount: total,
	}, nil
}

func (s *clientServiceImpl) GetClientByID(ctx context.Context, id uint64) (*dto.ClientResponse, error) {
	cli, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error: " + err.Error())
	}
	if cli == nil {
		return nil, exceptions.NewNotFoundError("Client not found")
	}

	return &dto.ClientResponse{
		ID:        cli.ID,
		Name:      cli.Name,
		DNI:       cli.DNI,
		Email:     cli.Email,
		CreatedAt: cli.CreatedAt,
	}, nil
}

func (s *clientServiceImpl) UpdateClient(ctx context.Context, id uint64, req *dto.UpdateClientReq) (*dto.ClientResponse, error) {
	cli, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during client update lookup: " + err.Error())
	}
	if cli == nil {
		return nil, exceptions.NewNotFoundError("Client not found")
	}

	// Unique check for DNI if changed
	if req.DNI != "" && req.DNI != cli.DNI {
		existingDNI, err := s.clientRepo.FindByDNI(ctx, req.DNI)
		if err != nil {
			return nil, exceptions.NewInternalError("Database error during update DNI check: " + err.Error())
		}
		if existingDNI != nil {
			return nil, exceptions.NewConflictError("Ya existe un cliente con este DNI!")
		}
		cli.DNI = req.DNI
	}

	// Unique check for Email if changed
	if req.Email != "" && req.Email != cli.Email {
		existingEmail, err := s.clientRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, exceptions.NewInternalError("Database error during update Email check: " + err.Error())
		}
		if existingEmail != nil {
			return nil, exceptions.NewConflictError("Ya existe un cliente con este Email!")
		}
		cli.Email = req.Email
	}

	if req.Name != "" {
		cli.Name = req.Name
	}

	if err := s.clientRepo.Update(ctx, cli); err != nil {
		return nil, exceptions.NewInternalError("Failed to update client profile: " + err.Error())
	}

	return &dto.ClientResponse{
		ID:        cli.ID,
		Name:      cli.Name,
		DNI:       cli.DNI,
		Email:     cli.Email,
		CreatedAt: cli.CreatedAt,
	}, nil
}

func (s *clientServiceImpl) DeleteClient(ctx context.Context, id uint64) error {
	cli, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return exceptions.NewInternalError("Database error: " + err.Error())
	}
	if cli == nil {
		return exceptions.NewNotFoundError("Client not found")
	}

	if err := s.clientRepo.Delete(ctx, id); err != nil {
		return exceptions.NewInternalError("Failed to delete client: " + err.Error())
	}

	return nil
}

func (s *clientServiceImpl) SearchClient(ctx context.Context, dni, email string) (*dto.ClientResponse, error) {
	if dni != "" {
		cli, err := s.clientRepo.FindByDNI(ctx, dni)
		if err != nil {
			return nil, exceptions.NewInternalError("Failed to find client by DNI: " + err.Error())
		}
		if cli == nil {
			return nil, exceptions.NewNotFoundError("Client with specified DNI not found")
		}
		return &dto.ClientResponse{
			ID:        cli.ID,
			Name:      cli.Name,
			DNI:       cli.DNI,
			Email:     cli.Email,
			CreatedAt: cli.CreatedAt,
		}, nil
	}

	if email != "" {
		cli, err := s.clientRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, exceptions.NewInternalError("Failed to find client by Email: " + err.Error())
		}
		if cli == nil {
			return nil, exceptions.NewNotFoundError("Client with specified Email not found")
		}
		return &dto.ClientResponse{
			ID:        cli.ID,
			Name:      cli.Name,
			DNI:       cli.DNI,
			Email:     cli.Email,
			CreatedAt: cli.CreatedAt,
		}, nil
	}

	return nil, exceptions.NewValidationError("Either DNI or Email must be provided for client search")
}
