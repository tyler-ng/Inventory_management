package models

import (
	"time"
)

// Warehouse represents a storage location for inventory
type Warehouse struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Location  string    `json:"location"`
	Address   string    `json:"address"`
	Manager   string    `json:"manager"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Status    string    `json:"status" gorm:"default:'active'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Locations           []WarehouseLocation  `json:"locations,omitempty" gorm:"foreignKey:WarehouseID"`
	InventoryTransactions []InventoryTransaction `json:"inventory_transactions,omitempty" gorm:"foreignKey:WarehouseID"`
	PurchaseOrders      []PurchaseOrder      `json:"purchase_orders,omitempty" gorm:"foreignKey:WarehouseID"`
	SalesOrders         []SalesOrder         `json:"sales_orders,omitempty" gorm:"foreignKey:WarehouseID"`
	Products            []Product            `json:"products,omitempty" gorm:"many2many:product_warehouse"`
}

// WarehouseLocation represents a specific location within a warehouse
type WarehouseLocation struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	WarehouseID uint      `json:"warehouse_id" gorm:"not null"`
	Zone        string    `json:"zone"`
	Aisle       string    `json:"aisle"`
	Rack        string    `json:"rack"`
	Shelf       string    `json:"shelf"`
	Bin         string    `json:"bin"`
	Status      string    `json:"status" gorm:"default:'active'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Warehouse    *Warehouse             `json:"warehouse" gorm:"foreignKey:WarehouseID"`
	Products     []ProductWarehouse     `json:"products,omitempty" gorm:"foreignKey:LocationID"`
	SourceTransactions []InventoryTransaction `json:"source_transactions,omitempty" gorm:"foreignKey:SourceLocationID"`
	DestinationTransactions []InventoryTransaction `json:"destination_transactions,omitempty" gorm:"foreignKey:DestinationLocationID"`
}

// GetFullLocationCode returns a formatted location code (e.g., "A-01-B-03-25")
func (wl *WarehouseLocation) GetFullLocationCode() string {
	return wl.Zone + "-" + wl.Aisle + "-" + wl.Rack + "-" + wl.Shelf + "-" + wl.Bin
}