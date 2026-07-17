package entities

import (
	"time"
)

// Sale represents a purchase transaction initiated by a customer and processed by a cashier.
// Contract Traceability:
// - Implements the database mapping for sale.
// - Service Name: SupermarketService
// - Low-Level Design Reference: LLD_SupermarketService.md#section-21-entity-table-schemas
// - Database Schema Reference: db_schema_report.json (table: sale)
type Sale struct {
	ID         uint64       `gorm:"primaryKey;autoIncrement;column:id"`
	TotalPrice float64      `gorm:"type:double;not null;column:total_price;default:0.0"`
	Status     string       `gorm:"type:varchar(50);not null;column:status"` // OPEN, CLOSED, CANCELED
	ClientID   *uint64      `gorm:"index:idx_sale_client_id;column:client_id"`
	CashierID  uint64       `gorm:"not null;index:idx_sale_cashier_id;column:cashier_id"`
	FinishedAt *time.Time   `gorm:"type:datetime;column:finished_at"`
	CreatedAt  time.Time    `gorm:"type:timestamp;not null;column:created_at;default:CURRENT_TIMESTAMP"`
	Details    []SaleDetail `gorm:"foreignKey:SaleID"`

	// Relationships
	Client  *Client `gorm:"foreignKey:ClientID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Cashier User    `gorm:"foreignKey:CashierID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

// TableName overrides the table name to "sale".
func (Sale) TableName() string {
	return "sale"
}

// SaleDetail represents a specific line-item in a sale.
// Contract Traceability:
// - Implements the database mapping for sale_detail.
// - Service Name: SupermarketService
// - Low-Level Design Reference: LLD_SupermarketService.md#section-21-entity-table-schemas
// - Database Schema Reference: db_schema_report.json (table: sale_detail)
type SaleDetail struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	UnitPrice float64   `gorm:"type:double;not null;column:unit_price"`
	Quantity  int       `gorm:"type:int;not null;column:quantity"`
	SubTotal  float64   `gorm:"type:double;not null;column:sub_total"`
	ProductID uint64    `gorm:"not null;index:idx_sale_detail_product_id;column:product_id"`
	SaleID    uint64    `gorm:"not null;index:idx_sale_detail_sale_id;column:sale_id"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;column:updated_at;default:CURRENT_TIMESTAMP"`

	// Relationships
	Product Product `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

// TableName overrides the table name to "sale_detail".
func (SaleDetail) TableName() string {
	return "sale_detail"
}
