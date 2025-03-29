package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents the product entity in the inventory system
type Product struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	SKU           string    `json:"sku" gorm:"uniqueIndex;not null"`
	Name          string    `json:"name" gorm:"not null"`
	Description   string    `json:"description"`
	Quantity      int       `json:"quantity" gorm:"not null;default:0"`
	ReorderLevel  int       `json:"reorder_level" gorm:"default:5"`
	Price         float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	CostPrice     float64   `json:"cost_price" gorm:"type:decimal(10,2)"`
	Weight        float64   `json:"weight"`
	Dimensions    string    `json:"dimensions"`
	ImageURL      string    `json:"image_url"`
	Barcode       string    `json:"barcode"`
	Status        string    `json:"status" gorm:"default:'active'"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Categories      []Category      `json:"categories" gorm:"many2many:product_category"`
	Suppliers       []Supplier      `json:"suppliers" gorm:"many2many:product_supplier"`
	Attachments     []ProductAttachment `json:"attachments" gorm:"foreignKey:ProductID"`
	Variants        []ProductVariant    `json:"variants" gorm:"foreignKey:ProductID"`
	ParentBundles   []ProductBundle     `json:"-" gorm:"foreignKey:ChildProductID"`
	ChildBundles    []ProductBundle     `json:"-" gorm:"foreignKey:ParentProductID"`
}

// ProductAttachment represents files attached to a product
type ProductAttachment struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ProductID   uint      `json:"product_id" gorm:"not null"`
	FileName    string    `json:"file_name" gorm:"not null"`
	FilePath    string    `json:"file_path" gorm:"not null"`
	FileType    string    `json:"file_type"`
	FileSize    int       `json:"file_size"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ProductVariant represents variations of a product
type ProductVariant struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ProductID   uint      `json:"product_id" gorm:"not null"`
	SKU         string    `json:"sku" gorm:"uniqueIndex;not null"`
	Attributes  string    `json:"attributes" gorm:"type:jsonb"`
	Quantity    int       `json:"quantity" gorm:"not null;default:0"`
	Price       float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	CostPrice   float64   `json:"cost_price" gorm:"type:decimal(10,2)"`
	Barcode     string    `json:"barcode"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ProductBundle represents bundled products
type ProductBundle struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ParentProductID  uint      `json:"parent_product_id" gorm:"not null"`
	ChildProductID   uint      `json:"child_product_id" gorm:"not null"`
	Quantity         int       `json:"quantity" gorm:"not null;default:1"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	ParentProduct    *Product  `json:"parent_product" gorm:"foreignKey:ParentProductID"`
	ChildProduct     *Product  `json:"child_product" gorm:"foreignKey:ChildProductID"`
}

// ProductSupplier represents the many-to-many relationship between products and suppliers
type ProductSupplier struct {
	ProductID        uint      `json:"product_id" gorm:"primaryKey"`
	SupplierID       uint      `json:"supplier_id" gorm:"primaryKey"`
	UnitCost         float64   `json:"unit_cost" gorm:"type:decimal(10,2)"`
	MinOrderQuantity int       `json:"min_order_quantity" gorm:"default:1"`
	LeadTimeDays     int       `json:"lead_time_days"`
	SupplierSKU      string    `json:"supplier_sku"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Product          *Product  `json:"product" gorm:"foreignKey:ProductID"`
	Supplier         *Supplier `json:"supplier" gorm:"foreignKey:SupplierID"`
}

// ProductWarehouse represents the many-to-many relationship between products and warehouses
type ProductWarehouse struct {
	ProductID      uint      `json:"product_id" gorm:"primaryKey"`
	WarehouseID    uint      `json:"warehouse_id" gorm:"primaryKey"`
	LocationID     uint      `json:"location_id"`
	Quantity       int       `json:"quantity" gorm:"not null;default:0"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Product        *Product          `json:"product" gorm:"foreignKey:ProductID"`
	Warehouse      *Warehouse        `json:"warehouse" gorm:"foreignKey:WarehouseID"`
	Location       *WarehouseLocation `json:"location" gorm:"foreignKey:LocationID"`
}

// ProductCategory represents the many-to-many relationship between products and categories
type ProductCategory struct {
	ProductID      uint      `json:"product_id" gorm:"primaryKey"`
	CategoryID     uint      `json:"category_id" gorm:"primaryKey"`
	
	// Relationships
	Product        *Product  `json:"product" gorm:"foreignKey:ProductID"`
	Category       *Category `json:"category" gorm:"foreignKey:CategoryID"`
}

// BeforeCreate hook for product to generate SKU if not provided
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	// Add SKU generation logic if needed
	return nil
}