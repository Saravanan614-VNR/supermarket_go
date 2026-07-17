/*
 * Contract ID: CTR-004 (Repository Integration Tests)
 * Service Name: SupermarketService
 * Description: Integration tests for Client, Product, and Sale repositories using GORM with SQLite in-memory.
 */

package repositories_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"supermarket-backend/entities"
	"supermarket-backend/repositories"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Initialize GORM in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Auto-migrate tables
	err = db.AutoMigrate(
		&entities.Category{},
		&entities.Product{},
		&entities.Client{},
		&entities.Promotion{},
		&entities.PromotionProduct{},
		&entities.Sale{},
		&entities.SaleDetail{},
	)
	if err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return db
}

func TestClientRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewClientRepository(db)
	ctx := context.Background()

	// 1. Create a client
	client := &entities.Client{
		Name:  "Alice Wonderland",
		DNI:   "1798765432",
		Email: "alice@example.com",
	}

	err := repo.Create(ctx, client)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client.ID == 0 {
		t.Errorf("expected client ID to be populated, got 0")
	}

	// 2. Find client by ID
	found, err := repo.FindByID(ctx, client.ID)
	if err != nil {
		t.Fatalf("failed to find client by ID: %v", err)
	}
	if found == nil || found.Name != "Alice Wonderland" {
		t.Errorf("expected to find Alice, got: %+v", found)
	}

	// 3. Find client by DNI
	foundDNI, err := repo.FindByDNI(ctx, "1798765432")
	if err != nil {
		t.Fatalf("failed to find client by DNI: %v", err)
	}
	if foundDNI == nil || foundDNI.ID != client.ID {
		t.Errorf("expected client ID to match by DNI, got %+v", foundDNI)
	}

	// 4. Ensure DNI uniqueness constraints behave as expected
	duplicate := &entities.Client{
		Name:  "Bob Builder",
		DNI:   "1798765432", // Duplicate DNI
		Email: "bob@example.com",
	}
	err = repo.Create(ctx, duplicate)
	if err == nil {
		t.Errorf("expected error on duplicate DNI insertion, but got nil")
	}
}

func TestProductRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	prodRepo := repositories.NewProductRepository(db)
	ctx := context.Background()

	// Seed Category
	category := &entities.Category{
		Name:        "Snacks",
		Description: "Chips, nuts and cookies",
	}
	if err := db.Create(category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	// 1. Create Product
	product := &entities.Product{
		Name:          "Doritos Nacho 150g",
		Brand:         "Doritos",
		BasePrice:     2.50,
		TaxPercentage: 12.0,
		CategoryID:    category.ID,
		Stock:         50,
	}

	err := prodRepo.Create(ctx, product)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	// 2. Update Stock with transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		return prodRepo.UpdateStockWithTx(ctx, tx, product.ID, 45)
	})
	if err != nil {
		t.Fatalf("failed to update stock: %v", err)
	}

	found, err := prodRepo.FindByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("failed to find product: %v", err)
	}
	if found.Stock != 45 {
		t.Errorf("expected stock to be 45, got %d", found.Stock)
	}

	// 3. Setup active and inactive promotions
	promoActive := &entities.Promotion{
		Name:               "Active Snack Promo",
		DiscountPercentage: 20.0,
		StartDate:          time.Now().Add(-1 * time.Hour),
		EndDate:            time.Now().Add(1 * time.Hour),
	}
	promoExpired := &entities.Promotion{
		Name:               "Expired Snack Promo",
		DiscountPercentage: 50.0,
		StartDate:          time.Now().Add(-5 * time.Hour),
		EndDate:            time.Now().Add(-1 * time.Hour),
	}

	if err := db.Create(promoActive).Error; err != nil {
		t.Fatalf("failed to create active promo: %v", err)
	}
	if err := db.Create(promoExpired).Error; err != nil {
		t.Fatalf("failed to create expired promo: %v", err)
	}

	// Link products
	db.Create(&entities.PromotionProduct{PromotionID: promoActive.ID, ProductID: product.ID})
	db.Create(&entities.PromotionProduct{PromotionID: promoExpired.ID, ProductID: product.ID})

	// Find active promotions for product
	promos, err := prodRepo.FindActivePromotionsByProductID(ctx, nil, product.ID, time.Now())
	if err != nil {
		t.Fatalf("failed to fetch active promotions: %v", err)
	}

	if len(promos) != 1 {
		t.Errorf("expected exactly 1 active promotion, got %d", len(promos))
	} else if promos[0].Name != "Active Snack Promo" {
		t.Errorf("expected 'Active Snack Promo', got %q", promos[0].Name)
	}
}

func TestSaleRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	saleRepo := repositories.NewSaleRepository(db)
	ctx := context.Background()

	// 1. Create a Sale
	sale := &entities.Sale{
		Status:     "OPEN",
		TotalPrice: 0.0,
		CashierID:  1,
	}

	err := saleRepo.CreateSale(ctx, nil, sale)
	if err != nil {
		t.Fatalf("failed to create sale: %v", err)
	}

	// 2. Add details
	detail := &entities.SaleDetail{
		SaleID:    sale.ID,
		ProductID: 99, // Dummy product ID
		UnitPrice: 1.50,
		Quantity:  2,
		SubTotal:  3.00,
	}

	err = saleRepo.CreateDetail(ctx, nil, detail)
	if err != nil {
		t.Fatalf("failed to create sale detail: %v", err)
	}

	// Update sale total
	sale.TotalPrice = 3.00
	_ = saleRepo.UpdateSale(ctx, nil, sale)

	// 3. Find Sale with preloaded details
	found, err := saleRepo.FindSaleByID(ctx, sale.ID)
	if err != nil {
		t.Fatalf("failed to find sale: %v", err)
	}

	if found == nil {
		t.Fatalf("expected sale to be found, got nil")
	}

	if len(found.Details) != 1 {
		t.Errorf("expected 1 sale detail preloaded, got %d", len(found.Details))
	} else if found.Details[0].SubTotal != 3.00 {
		t.Errorf("expected subtotal of detail to be 3.00, got %f", found.Details[0].SubTotal)
	}
}
