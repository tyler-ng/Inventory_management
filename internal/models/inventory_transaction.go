package models

import (
	"time"

	"gorm.io/gorm"
)

// InventoryTransaction represents a movement of inventory
type InventoryTransaction struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	ProductID             uint      `json:"product_id" gorm:"not null"`
	WarehouseID           uint      `json:"warehouse_id" gorm:"not null"`
	SourceLocationID      *uint     `json:"source_location_id"`
	DestinationLocationID *uint     `json:"destination_location_id"`
	Type                  string    `json:"type" gorm:"not null"` // "receive", "issue", "transfer", "adjustment"
	Quantity              int       `json:"quantity" gorm:"not null"`
	ReferenceNumber       string    `json:"reference_number"`
	UserID                uint      `json:"user_id" gorm:"not null"`
	Notes                 string    `json:"notes"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
	
	// Relationships
	Product             *Product          `json:"product" gorm:"foreignKey:ProductID"`
	Warehouse           *Warehouse        `json:"warehouse" gorm:"foreignKey:WarehouseID"`
	SourceLocation      *WarehouseLocation `json:"source_location,omitempty" gorm:"foreignKey:SourceLocationID"`
	DestinationLocation *WarehouseLocation `json:"destination_location,omitempty" gorm:"foreignKey:DestinationLocationID"`
	User                *User              `json:"user" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook for inventory transaction to update product quantity
func (it *InventoryTransaction) BeforeCreate(tx *gorm.DB) error {
	// This would update the product quantity based on the transaction type
	// In a real implementation, this would be more sophisticated
	
	var product Product
	if err := tx.First(&product, it.ProductID).Error; err != nil {
		return err
	}
	
	switch it.Type {
	case "receive":
		product.Quantity += it.Quantity
	case "issue":
		product.Quantity -= it.Quantity
	case "adjustment":
		// Adjustment logic would go here
	case "transfer":
		// Transfer logic would go here
	}
	
	return tx.Save(&product).Error
}