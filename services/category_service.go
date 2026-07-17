package services

import (
	"context"
	"strings"

	"gorm.io/gorm"

	dto "supermarket-backend/dtos"
	"supermarket-backend/entities"
	"supermarket-backend/exceptions"
	"supermarket-backend/repositories"
)

type categoryServiceImpl struct {
	categoryRepo repositories.CategoryRepository
	productRepo  repositories.ProductRepository
	db           *gorm.DB
}

// NewCategoryService constructs a new CategoryService implementation.
func NewCategoryService(
	categoryRepo repositories.CategoryRepository,
	productRepo repositories.ProductRepository,
	db *gorm.DB,
) CategoryService {
	return &categoryServiceImpl{
		categoryRepo: categoryRepo,
		productRepo:  productRepo,
		db:           db,
	}
}

func (s *categoryServiceImpl) CreateCategory(ctx context.Context, req *dto.CreateCategoryReq) (*dto.CategoryResponse, error) {
	// 1. Enforce unique category name case-insensitive among active categories
	exists, err := s.categoryRepo.ExistsByNameCaseInsensitive(ctx, req.Name)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during category name uniqueness check: " + err.Error())
	}
	if exists {
		return nil, exceptions.NewConflictError("Ya existe una categoria con ese nombre!")
	}

	category := &entities.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, exceptions.NewInternalError("Failed to create category: " + err.Error())
	}

	return &dto.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt,
	}, nil
}

func (s *categoryServiceImpl) ListCategories(ctx context.Context, page, size int, filter string) (*dto.PaginatedCategories, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size

	categories, total, err := s.categoryRepo.FindAll(ctx, offset, size)
	if err != nil {
		return nil, exceptions.NewInternalError("Failed to list categories: " + err.Error())
	}

	// Filter by name inline if requested (basic search behavior)
	var filtered []entities.Category
	if filter != "" {
		for _, cat := range categories {
			if strings.Contains(strings.ToLower(cat.Name), strings.ToLower(filter)) {
				filtered = append(filtered, cat)
			}
		}
		categories = filtered
		total = int64(len(filtered))
	}

	items := make([]dto.CategoryResponse, len(categories))
	for i, cat := range categories {
		items[i] = dto.CategoryResponse{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description,
			CreatedAt:   cat.CreatedAt,
		}
	}

	return &dto.PaginatedCategories{
		Items:      items,
		TotalCount: total,
	}, nil
}

func (s *categoryServiceImpl) GetCategoryByID(ctx context.Context, id uint64) (*dto.CategoryResponse, error) {
	cat, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error: " + err.Error())
	}
	if cat == nil {
		return nil, exceptions.NewNotFoundError("Category not found")
	}

	return &dto.CategoryResponse{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
		CreatedAt:   cat.CreatedAt,
	}, nil
}

func (s *categoryServiceImpl) UpdateCategory(ctx context.Context, id uint64, req *dto.UpdateCategoryReq) (*dto.CategoryResponse, error) {
	cat, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during category lookup: " + err.Error())
	}
	if cat == nil {
		return nil, exceptions.NewNotFoundError("Category not found")
	}

	// Check name uniqueness if modified
	if req.Name != "" && req.Name != cat.Name {
		exists, err := s.categoryRepo.ExistsByNameCaseInsensitive(ctx, req.Name)
		if err != nil {
			return nil, exceptions.NewInternalError("Database error during category name check: " + err.Error())
		}
		if exists {
			return nil, exceptions.NewConflictError("Ya existe una categoria con ese nombre!")
		}
		cat.Name = req.Name
	}

	if req.Description != "" {
		cat.Description = req.Description
	}

	if err := s.categoryRepo.Update(ctx, cat); err != nil {
		return nil, exceptions.NewInternalError("Failed to update category: " + err.Error())
	}

	return &dto.CategoryResponse{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
		CreatedAt:   cat.CreatedAt,
	}, nil
}

func (s *categoryServiceImpl) DeleteCategory(ctx context.Context, id uint64) error {
	cat, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return exceptions.NewInternalError("Database error: " + err.Error())
	}
	if cat == nil {
		return exceptions.NewNotFoundError("Category not found")
	}

	// Run inside GORM transaction to cascade-soft-delete all linked products
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Soft delete the category
		if err := s.categoryRepo.SoftDeleteWithTx(ctx, tx, id); err != nil {
			return exceptions.NewInternalError("Failed to soft-delete category: " + err.Error())
		}

		// 2. Cascade soft-delete all products associated
		if err := s.productRepo.SoftDeleteByCategoryIDWithTx(ctx, tx, id); err != nil {
			return exceptions.NewInternalError("Failed to cascade-soft-delete product associations: " + err.Error())
		}

		return nil
	})
}

func (s *categoryServiceImpl) ListCategoryProducts(ctx context.Context, categoryID uint64, page, size int) (*dto.PaginatedProducts, error) {
	// Verify category exists first
	cat, err := s.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error: " + err.Error())
	}
	if cat == nil {
		return nil, exceptions.NewNotFoundError("Category not found")
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size

	products, total, err := s.productRepo.FindAllByCategoryID(ctx, categoryID, offset, size)
	if err != nil {
		return nil, exceptions.NewInternalError("Failed to list category products: " + err.Error())
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
