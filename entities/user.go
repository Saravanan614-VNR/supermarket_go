package entities

import (
	"time"

	"gorm.io/gorm"
)

// User represents the operator of the system (ADMIN, CASHIER, INVENTORY).
// Contract Traceability:
// - Implements the database mapping for _user.
// - Service Name: SupermarketService
// - Low-Level Design Reference: LLD_SupermarketService.md#section-21-entity-table-schemas
// - Database Schema Reference: db_schema_report.json (table: _user)
type User struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	FullName  string         `gorm:"type:varchar(255);not null;column:full_name"`
	Username  string         `gorm:"type:varchar(255);not null;index:idx_user_username_deleted_at,unique;column:username"`
	Password  string         `gorm:"type:varchar(255);not null;column:password"`
	Role      string         `gorm:"type:varchar(50);not null;column:role"` // ADMIN, CASHIER, INVENTORY
	DeletedAt gorm.DeletedAt `gorm:"index:idx_user_username_deleted_at,unique;column:deleted_at"`
	CreatedAt time.Time      `gorm:"type:timestamp;not null;column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:timestamp;not null;column:updated_at;default:CURRENT_TIMESTAMP"`
}

// TableName overrides the table name to "_user" to avoid conflicts with SQL reserved keywords.
func (User) TableName() string {
	return "_user"
}