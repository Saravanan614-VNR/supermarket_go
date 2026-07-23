package entities

import (
	"time"

	"gorm.io/gorm"
)

// Client represents a customer in the supermarket system.
// Contract Traceability:
// - Implements the database mapping for client.
// - Service Name: SupermarketService
// - Low-Level Design Reference: LLD_SupermarketService.md#section-21-entity-table-schemas
// - Database Schema Reference: db_schema_report.json (table: client)
type Client struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	Name      string         `gorm:"type:varchar(255);not null;column:name"`
	DNI       string         `gorm:"type:varchar(255);not null;index:idx_client_dni_deleted_at,unique;column:dni"`
	Email     string         `gorm:"type:varchar(255);not null;index:idx_client_email_deleted_at,unique;column:email"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_client_dni_deleted_at,unique;index:idx_client_email_deleted_at,unique;column:deleted_at"`
	CreatedAt time.Time      `gorm:"type:timestamp;not null;column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:timestamp;not null;column:updated_at;default:CURRENT_TIMESTAMP"`
}

// TableName overrides the table name to "client".
func (Client) TableName() string {
	return "client"
}