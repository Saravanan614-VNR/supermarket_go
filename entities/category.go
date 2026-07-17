package entities

import (
	"time"

	"gorm.io/gorm"
)

// Category represents a product category.
// Contract Traceability:
// - Implements the database mapping for category.
// - Service Name: SupermarketService
// - Low-Level Design Reference: LLD_SupermarketService.md#section-21-entity-table-schemas
// - Database Schema Reference: db_schema_report.json (table: category)
type Category struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	Name        string         `gorm:"type:varchar(255);not null;index:idx_category_name_deleted_at,unique;column:name"`
	Description string         `gorm:"type:varchar(255);not null;column:description"`
	DeletedAt   gorm.DeletedAt `gorm:"index:idx_category_name_deleted_at,unique;column:deleted_at"`
	CreatedAt   time.Time      `gorm:"type:timestamp;not null;column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"type:timestamp;not null;column:updated_at;default:CURRENT_TIMESTAMP"`
	Products    []Product      `gorm:"foreignKey:CategoryID"`
}

// TableName overrides the table name to "category".
func (Category) TableName() string {
	return "category"
}
