package services

import (
	"context"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	dto "supermarket-backend/dtos"
	"supermarket-backend/entities"
	"supermarket-backend/exceptions"
	"supermarket-backend/repositories"
)

type productServiceImpl struct {
	productRepo  repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
	db           *gorm.DB
}

// NewProductService constructs a new ProductService implementation.
func NewProductService(
	productRepo repositories.ProductRepository,
	categoryRepo repositories.CategoryRepository,
	db *gorm.DB,
) ProductService {
	return &productServiceImpl{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		db:           db,
	}
}

// RoundHalfUpProd performs rounding to 2 decimals using standard HALF_UP rule.
func RoundHalfUpProd(val float64) float64 {
	return math.Round(val*100.0) / 100.0
}

func (s *productServiceImpl) CreateProduct(ctx context.Context, req *dto.CreateProductReq) (*dto.ProductResponse, error) {
	// 1. Verify Category exists and is active
	category, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error checking category: " + err.Error())
	}
	if category == nil {
		return nil, exceptions.NewValidationError("La categoria especificada no existe!")
	}

	product := &entities.Product{
		Name:          req.Name,
		Brand:         req.Brand,
		BasePrice:     req.BasePrice,
		TaxPercentage: req.TaxPercentage,
		CategoryID:    req.CategoryID,
		Stock:         req.Stock,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, exceptions.NewInternalError("Failed to create product: " + err.Error())
	}

	return &dto.ProductResponse{
		ID:            product.ID,
		Name:          product.Name,
		Brand:         product.Brand,
		BasePrice:     product.BasePrice,
		TaxPercentage: product.TaxPercentage,
		CategoryID:    product.CategoryID,
		Stock:         product.Stock,
		CreatedAt:     product.CreatedAt,
	}, nil
}

func (s *productServiceImpl) ListProducts(ctx context.Context, page, size int, name string) (*dto.PaginatedProducts, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size

	products, total, err := s.productRepo.FindAll(ctx, offset, size)
	if err != nil {
		return nil, exceptions.NewInternalError("Failed to list products: " + err.Error())
	}

	// Filter by name inline if requested (basic search behavior)
	var filtered []entities.Product
	if name != "" {
		for _, prod := range products {
			if strings.Contains(strings.ToLower(prod.Name), strings.ToLower(name)) ||
				strings.Contains(strings.ToLower(prod.Brand), strings.ToLower(name)) {
				filtered = append(filtered, prod)
			}
		}
		products = filtered
		total = int64(len(filtered))
	}

	items := make([]dto.ProductResponse, len(products))
	for i, prod := range products {
		items[i] = dto.ProductResponse{
			ID:            prod.ID,
			Name:          prod.Name,
			Brand:         prod.Brand,
			BasePrice:     prod.BasePrice,
			TaxPercentage: prod.TaxPercentage,
			CategoryID:    prod.CategoryID,
			Stock:         prod.Stock,
			CreatedAt:     prod.CreatedAt,
		}
	}

	return &dto.PaginatedProducts{
		Items:      items,
		TotalCount: total,
	}, nil
}

func (s *productServiceImpl) GetProductByID(ctx context.Context, id uint64) (*dto.ProductResponse, error) {
	prod, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error: " + err.Error())
	}
	if prod == nil {
		return nil, exceptions.NewNotFoundError("Product not found")
	}

	return &dto.ProductResponse{
		ID:            prod.ID,
		Name:          prod.Name,
		Brand:         prod.Brand,
		BasePrice:     prod.BasePrice,
		TaxPercentage: prod.TaxPercentage,
		CategoryID:    prod.CategoryID,
		Stock:         prod.Stock,
		CreatedAt:     prod.CreatedAt,
	}, nil
}

func (s *productServiceImpl) UpdateProduct(ctx context.Context, id uint64, req *dto.UpdateProductReq) (*dto.ProductResponse, error) {
	prod, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during product update: " + err.Error())
	}
	if prod == nil {
		return nil, exceptions.NewNotFoundError("Product not found")
	}

	// Verify Category if modified
	if req.CategoryID != 0 && req.CategoryID != prod.CategoryID {
		category, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
		if err != nil {
			return nil, exceptions.NewInternalError("Database error checking category update: " + err.Error())
		}
		if category == nil {
			return nil, exceptions.NewValidationError("La categoria especificada no existe!")
		}
		prod.CategoryID = req.CategoryID
	}

	if req.Name != "" {
		prod.Name = req.Name
	}
	if req.Brand != "" {
		prod.Brand = req.Brand
	}
	if req.BasePrice > 0 {
		prod.BasePrice = req.BasePrice
	}
	if req.TaxPercentage >= 0 {
		prod.TaxPercentage = req.TaxPercentage
	}
	if req.Stock >= 0 {
		prod.Stock = req.Stock
	}

	if err := s.productRepo.Update(ctx, prod); err != nil {
		return nil, exceptions.NewInternalError("Failed to update product: " + err.Error())
	}

	return &dto.ProductResponse{
		ID:            prod.ID,
		Name:          prod.Name,
		Brand:         prod.Brand,
		BasePrice:     prod.BasePrice,
		TaxPercentage: prod.TaxPercentage,
		CategoryID:    prod.CategoryID,
		Stock:         prod.Stock,
		CreatedAt:     prod.CreatedAt,
	}, nil
}

func (s *productServiceImpl) DeleteProduct(ctx context.Context, id uint64) error {
	prod, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return exceptions.NewInternalError("Database error: " + err.Error())
	}
	if prod == nil {
		return exceptions.NewNotFoundError("Product not found")
	}

	if err := s.productRepo.UpdateStockWithTx(ctx, nil, id, 0); err != nil {
		return exceptions.NewInternalError("Failed to reset stock: " + err.Error())
	}

	if err := s.productRepo.Update(ctx, prod); err != nil {
		return exceptions.NewInternalError("Failed to save product status: " + err.Error())
	}

	// Soft-delete product
	return s.db.WithContext(ctx).Delete(&entities.Product{}, id).Error
}

func (s *productServiceImpl) RestockProduct(ctx context.Context, id uint64, quantity int) (*dto.ProductResponse, error) {
	if quantity <= 0 {
		return nil, exceptions.NewValidationError("La cantidad a reabastecer debe ser mayor a cero")
	}

	prod, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during restocking: " + err.Error())
	}
	if prod == nil {
		return nil, exceptions.NewNotFoundError("Product not found")
	}

	newStock := prod.Stock + quantity
	if err := s.productRepo.UpdateStockWithTx(ctx, nil, id, newStock); err != nil {
		return nil, exceptions.NewInternalError("Failed to update product stock: " + err.Error())
	}

	prod.Stock = newStock
	return &dto.ProductResponse{
		ID:            prod.ID,
		Name:          prod.Name,
		Brand:         prod.Brand,
		BasePrice:     prod.BasePrice,
		TaxPercentage: prod.TaxPercentage,
		CategoryID:    prod.CategoryID,
		Stock:         prod.Stock,
		CreatedAt:     prod.CreatedAt,
	}, nil
}

func (s *productServiceImpl) CalculatePrice(ctx context.Context, id uint64) (*dto.PriceBreakdownResponse, error) {
	prod, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error: " + err.Error())
	}
	if prod == nil {
		return nil, exceptions.NewNotFoundError("Product not found")
	}

	// 1. Fetch active promotions
	promos, err := s.productRepo.FindActivePromotionsByProductID(ctx, nil, id, time.Now())
	if err != nil {
		return nil, exceptions.NewInternalError("Failed to retrieve active promotions: " + err.Error())
	}

	var bestPromo *entities.Promotion
	var maxDiscount float64 = 0.0
	for i, promo := range promos {
		if promo.DiscountPercentage > maxDiscount {
			maxDiscount = promo.DiscountPercentage
			bestPromo = &promos[i]
		}
	}

	standardFinalPrice := RoundHalfUpProd(prod.BasePrice * (1.0 + (prod.TaxPercentage / 100.0)))
	promotionalFinalPrice := RoundHalfUpProd(prod.BasePrice * (1.0 - (maxDiscount / 100.0)) * (1.0 + (prod.TaxPercentage / 100.0)))

	var appliedPromo *dto.PromoShortResponse
	if bestPromo != nil {
		appliedPromo = &dto.PromoShortResponse{
			ID:                 bestPromo.ID,
			Name:               bestPromo.Name,
			DiscountPercentage: bestPromo.DiscountPercentage,
		}
	}

	return &dto.PriceBreakdownResponse{
		ProductID:             prod.ID,
		BasePrice:             prod.BasePrice,
		StandardFinalPrice:    standardFinalPrice,
		AppliedPromotion:      appliedPromo,
		PromotionalFinalPrice: promotionalFinalPrice,
	}, nil
}
