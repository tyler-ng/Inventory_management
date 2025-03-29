package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PurchaseOrder represents an order to a supplier
type PurchaseOrder struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	PONumber      string    `json:"po_number" gorm:"uniqueIndex;not null"`
	SupplierID    uint      `json:"supplier_id" gorm:"not null"`
	WarehouseID   uint      `json:"warehouse_id" gorm:"not null"`
	OrderDate     time.Time `json:"order_date" gorm:"not null"`
	ExpectedDate  time.Time `json:"expected_date"`
	Status        string    `json:"status" gorm:"default:'draft'"`
	TotalAmount   float64   `json:"total_amount" gorm:"type:decimal(10,2);default:0"`
	PaymentTerms  string    `json:"payment_terms"`
	ShippingTerms string    `json:"shipping_terms"`
	UserID        uint      `json:"user_id" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Supplier      *Supplier         `json:"supplier" gorm:"foreignKey:SupplierID"`
	Warehouse     *Warehouse        `json:"warehouse" gorm:"foreignKey:WarehouseID"`
	User          *User             `json:"user" gorm:"foreignKey:UserID"`
	Items         []PurchaseOrderItem `json:"items" gorm:"foreignKey:PurchaseOrderID"`
}

// PurchaseOrderItem represents an item in a purchase order
type PurchaseOrderItem struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	PurchaseOrderID uint      `json:"purchase_order_id" gorm:"not null"`
	ProductID       uint      `json:"product_id" gorm:"not null"`
	Quantity        int       `json:"quantity" gorm:"not null"`
	UnitPrice       float64   `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	TotalPrice      float64   `json:"total_price" gorm:"type:decimal(10,2);not null"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	PurchaseOrder  *PurchaseOrder `json:"purchase_order" gorm:"foreignKey:PurchaseOrderID"`
	Product        *Product      `json:"product" gorm:"foreignKey:ProductID"`
}

// BeforeCreate hook for purchase order to generate PO number if not provided
func (po *PurchaseOrder) BeforeCreate(tx *gorm.DB) error {
	if po.PONumber == "" {
		// Generate a unique PO number
		var lastPO PurchaseOrder
		if err := tx.Order("id desc").First(&lastPO).Error; err == nil {
			po.PONumber = fmt.Sprintf("PO-%06d", lastPO.ID+1)
		} else {
			po.PONumber = "PO-000001"
		}
	}
	return nil
}

// BeforeCreate hook for purchase order item to calculate total price
func (poi *PurchaseOrderItem) BeforeCreate(tx *gorm.DB) error {
	poi.TotalPrice = float64(poi.Quantity) * poi.UnitPrice
	return nil
}

// BeforeSave hook for purchase order item to recalculate total price
func (poi *PurchaseOrderItem) BeforeSave(tx *gorm.DB) error {
	poi.TotalPrice = float64(poi.Quantity) * poi.UnitPrice
	return nil
}

// AfterCreate hook for purchase order item to update purchase order total
func (poi *PurchaseOrderItem) AfterCreate(tx *gorm.DB) error {
	return updatePurchaseOrderTotal(tx, poi.PurchaseOrderID)
}

// AfterUpdate hook for purchase order item to update purchase order total
func (poi *PurchaseOrderItem) AfterUpdate(tx *gorm.DB) error {
	return updatePurchaseOrderTotal(tx, poi.PurchaseOrderID)
}

// AfterDelete hook for purchase order item to update purchase order total
func (poi *PurchaseOrderItem) AfterDelete(tx *gorm.DB) error {
	return updatePurchaseOrderTotal(tx, poi.PurchaseOrderID)
}

// updatePurchaseOrderTotal recalculates the total amount for a purchase order
func updatePurchaseOrderTotal(tx *gorm.DB, poID uint) error {
	var total float64
	if err := tx.Model(&PurchaseOrderItem{}).
		Where("purchase_order_id = ?", poID).
		Select("SUM(total_price)").
		Scan(&total).Error; err != nil {
		return err
	}
	
	return tx.Model(&PurchaseOrder{}).
		Where("id = ?", poID).
		Update("total_amount", total).Error
}