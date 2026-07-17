package services

import (
	"context"

	"gorm.io/gorm"

	dto "supermarket-backend/dtos"
	"supermarket-backend/entities"
	"supermarket-backend/exceptions"
	"supermarket-backend/repositories"
)

type promotionServiceImpl struct {
	promotionRepo repositories.PromotionRepository
	productRepo   repositories.ProductRepository
	db            *gorm.DB
}

// NewPromotionService constructs a new PromotionService implementation.
func NewPromotionService(
	promotionRepo repositories.PromotionRepository,
	productRepo repositories.ProductRepository,
	db *gorm.DB,
) PromotionService {
	return &promotionServiceImpl{
		promotionRepo: promotionRepo,
		productRepo:   productRepo,
		db:            db,
	}
}

func (s *promotionServiceImpl) CreatePromotion(ctx context.Context, req *dto.CreatePromotionReq) (*dto.PromotionResponse, error) {
	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		return nil, exceptions.NewValidationError("La fecha de fin debe ser posterior a la fecha de inicio")
	}

	var promotion *entities.Promotion

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create Promotion
		promotion = &entities.Promotion{
			Name:               req.Name,
			DiscountPercentage: req.DiscountPercentage,
			StartDate:          req.StartDate,
			EndDate:            req.EndDate,
		}

		if err := s.promotionRepo.Create(ctx, promotion); err != nil {
			return exceptions.NewInternalError("Failed to create promotion campaign: " + err.Error())
		}

		// 2. Validate product IDs exist and are active, and link them
		if len(req.ProductIDs) > 0 {
			for _, pid := range req.ProductIDs {
				prod, err := s.productRepo.FindByIDWithLock(ctx, tx, pid)
				if err != nil {
					return exceptions.NewInternalError("Database error checking product: " + err.Error())
				}
				if prod == nil {
					return exceptions.NewValidationError("El producto especificado no existe o esta inactivo")
				}
			}

			if err := s.promotionRepo.LinkProductsWithTx(ctx, tx, promotion.ID, req.ProductIDs); err != nil {
				return exceptions.NewInternalError("Failed to link products to promotion campaign: " + err.Error())
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.PromotionResponse{
		ID:                 promotion.ID,
		Name:               promotion.Name,
		DiscountPercentage: promotion.DiscountPercentage,
		StartDate:          promotion.StartDate,
		EndDate:            promotion.EndDate,
		CreatedAt:          promotion.CreatedAt,
	}, nil
}

func (s *promotionServiceImpl) ListPromotions(ctx context.Context, page, size int) (*dto.PaginatedPromotions, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size

	promos, total, err := s.promotionRepo.FindAll(ctx, offset, size)
	if err != nil {
		return nil, exceptions.NewInternalError("Failed to list promotion campaigns: " + err.Error())
	}

	items := make([]dto.PromotionResponse, len(promos))
	for i, promo := range promos {
		items[i] = dto.PromotionResponse{
			ID:                 promo.ID,
			Name:               promo.Name,
			DiscountPercentage: promo.DiscountPercentage,
			StartDate:          promo.StartDate,
			EndDate:            promo.EndDate,
			CreatedAt:          promo.CreatedAt,
		}
	}

	return &dto.PaginatedPromotions{
		Items:      items,
		TotalCount: total,
	}, nil
}

func (s *promotionServiceImpl) GetPromotionByID(ctx context.Context, id uint64) (*dto.PromotionResponse, error) {
	promo, err := s.promotionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error: " + err.Error())
	}
	if promo == nil {
		return nil, exceptions.NewNotFoundError("Promotion campaign not found")
	}

	return &dto.PromotionResponse{
		ID:                 promo.ID,
		Name:               promo.Name,
		DiscountPercentage: promo.DiscountPercentage,
		StartDate:          promo.StartDate,
		EndDate:            promo.EndDate,
		CreatedAt:          promo.CreatedAt,
	}, nil
}

func (s *promotionServiceImpl) UpdatePromotion(ctx context.Context, id uint64, req *dto.UpdatePromotionReq) (*dto.PromotionResponse, error) {
	promo, err := s.promotionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during promotion lookup: " + err.Error())
	}
	if promo == nil {
		return nil, exceptions.NewNotFoundError("Promotion campaign not found")
	}

	if req.StartDate.After(req.EndDate) {
		return nil, exceptions.NewValidationError("La fecha de fin debe ser posterior a la fecha de inicio")
	}

	if req.Name != "" {
		promo.Name = req.Name
	}
	if req.DiscountPercentage > 0 {
		promo.DiscountPercentage = req.DiscountPercentage
	}
	if !req.StartDate.IsZero() {
		promo.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		promo.EndDate = req.EndDate
	}

	if err := s.promotionRepo.Update(ctx, promo); err != nil {
		return nil, exceptions.NewInternalError("Failed to update promotion campaign: " + err.Error())
	}

	return &dto.PromotionResponse{
		ID:                 promo.ID,
		Name:               promo.Name,
		DiscountPercentage: promo.DiscountPercentage,
		StartDate:          promo.StartDate,
		EndDate:            promo.EndDate,
		CreatedAt:          promo.CreatedAt,
	}, nil
}

func (s *promotionServiceImpl) DeletePromotion(ctx context.Context, id uint64) error {
	promo, err := s.promotionRepo.FindByID(ctx, id)
	if err != nil {
		return exceptions.NewInternalError("Database error: " + err.Error())
	}
	if promo == nil {
		return exceptions.NewNotFoundError("Promotion campaign not found")
	}

	if err := s.promotionRepo.Delete(ctx, id); err != nil {
		return exceptions.NewInternalError("Failed to delete promotion campaign: " + err.Error())
	}

	return nil
}

func (s *promotionServiceImpl) LinkProductsToPromotion(ctx context.Context, id uint64, productIDs []uint64) error {
	promo, err := s.promotionRepo.FindByID(ctx, id)
	if err != nil {
		return exceptions.NewInternalError("Database error during promotion lookup: " + err.Error())
	}
	if promo == nil {
		return exceptions.NewNotFoundError("Promotion campaign not found")
	}

	if len(productIDs) == 0 {
		return exceptions.NewValidationError("At least one product ID must be provided to link")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, pid := range productIDs {
			prod, err := s.productRepo.FindByIDWithLock(ctx, tx, pid)
			if err != nil {
				return exceptions.NewInternalError("Database error checking product link: " + err.Error())
			}
			if prod == nil {
				return exceptions.NewValidationError("El producto especificado no existe o esta inactivo")
			}
		}

		if err := s.promotionRepo.LinkProductsWithTx(ctx, tx, id, productIDs); err != nil {
			return exceptions.NewInternalError("Failed to associate products with campaign: " + err.Error())
		}

		return nil
	})
}

func (s *promotionServiceImpl) UnlinkProductFromPromotion(ctx context.Context, id uint64, productID uint64) error {
	promo, err := s.promotionRepo.FindByID(ctx, id)
	if err != nil {
		return exceptions.NewInternalError("Database error during promotion lookup: " + err.Error())
	}
	if promo == nil {
		return exceptions.NewNotFoundError("Promotion campaign not found")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		prod, err := s.productRepo.FindByIDWithLock(ctx, tx, productID)
		if err != nil {
			return exceptions.NewInternalError("Database error checking product unlink: " + err.Error())
		}
		if prod == nil {
			return exceptions.NewNotFoundError("Product not found")
		}

		if err := s.promotionRepo.UnlinkProductWithTx(ctx, tx, id, productID); err != nil {
			return exceptions.NewInternalError("Failed to remove product from campaign: " + err.Error())
		}

		return nil
	})
}
