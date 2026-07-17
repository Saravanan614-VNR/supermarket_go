package entities

import (
	"time"

	"gorm.io/gorm"
)

// Product represents a product in the supermarket inventory.
// Contract Traceability:
// - Implements design_contracts.json contract CTR-005.
// - Service Name: SupermarketService
// - Low-Level Design Reference: LLD_SupermarketService.md#section-21-entity-table-schemas
// - Database Schema Reference: db_schema_report.json (table: product)
type Product struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	Name          string         `gorm:"type:varchar(255);not null;index:idx_product_name_deleted_at,unique;column:name"`
	Brand         string         `gorm:"type:varchar(255);not null;column:brand"`
	BasePrice     float64        `gorm:"type:double;not null;column:base_price"`
	TaxPercentage float64        `gorm:"type:double;not null;column:tax_percentage"`
	CategoryID    uint64         `gorm:"not null;index:idx_product_category_id;column:category_id"`
	Stock         int            `gorm:"type:int;not null;column:stock"`
	DeletedAt     gorm.DeletedAt `gorm:"index:idx_product_name_deleted_at,unique;column:deleted_at"`
	CreatedAt     time.Time      `gorm:"type:timestamp;not null;column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time      `gorm:"type:timestamp;not null;column:updated_at;default:CURRENT_TIMESTAMP"`

	// Relationships
	Category Category `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

// TableName overrides the table name to "product".
func (Product) TableName() string {
	return "product"
}
