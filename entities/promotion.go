package entities

import (
	"time"

	"gorm.io/gorm"
)

// Promotion represents a promotional discount campaign that is bounded by start and end dates.
// Contract Traceability:
// - Implements the database mapping for promotion.
// - Service Name: SupermarketService
// - Low-Level Design Reference: LLD_SupermarketService.md#section-21-entity-table-schemas
// - Database Schema Reference: db_schema_report.json (table: promotion)
type Promotion struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	Name               string         `gorm:"type:varchar(255);not null;column:name"`
	DiscountPercentage float64        `gorm:"type:double;not null;column:discount_percentage"`
	StartDate          time.Time      `gorm:"type:datetime;not null;column:start_date"`
	EndDate            time.Time      `gorm:"type:datetime;not null;column:end_date"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at"`
	CreatedAt          time.Time      `gorm:"type:timestamp;not null;column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt          time.Time      `gorm:"type:timestamp;not null;column:updated_at;default:CURRENT_TIMESTAMP"`

	// Relationships
	Products []Product `gorm:"many2many:promotion_products;joinForeignKey:PromotionID;joinReferences:ProductID"`
}

// TableName overrides the table name to "promotion".
func (Promotion) TableName() string {
	return "promotion"
}

// PromotionProduct represents the many-to-many junction table between Promotion and Product.
// Contract Traceability:
// - Low-Level Design Reference: LLD_SupermarketService.md#section-21-entity-table-schemas
// - Database Schema Reference: db_schema_report.json (table: promotion_products)
type PromotionProduct struct {
	PromotionID uint64 `gorm:"primaryKey;column:promotion_id"`
	ProductID   uint64 `gorm:"primaryKey;index:idx_promotion_products_product;column:product_id"`
}

// TableName overrides the table name to "promotion_products".
func (PromotionProduct) TableName() string {
	return "promotion_products"
}