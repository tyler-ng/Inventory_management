package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SalesOrder represents a customer order
type SalesOrder struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	SONumber      string    `json:"so_number" gorm:"uniqueIndex;not null"`
	CustomerID    uint      `json:"customer_id" gorm:"not null"`
	WarehouseID   uint      `json:"warehouse_id" gorm:"not null"`
	OrderDate     time.Time `json:"order_date" gorm:"not null"`
	ShippingDate  time.Time `json:"shipping_date"`
	Status        string    `json:"status" gorm:"default:'draft'"`
	Subtotal      float64   `json:"subtotal" gorm:"type:decimal(10,2);default:0"`
	Tax           float64   `json:"tax" gorm:"type:decimal(10,2);default:0"`
	ShippingCost  float64   `json:"shipping_cost" gorm:"type:decimal(10,2);default:0"`
	TotalAmount   float64   `json:"total_amount" gorm:"type:decimal(10,2);default:0"`
	PaymentStatus string    `json:"payment_status" gorm:"default:'unpaid'"`
	UserID        uint      `json:"user_id" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Customer      *Customer      `json:"customer" gorm:"foreignKey:CustomerID"`
	Warehouse     *Warehouse     `json:"warehouse" gorm:"foreignKey:WarehouseID"`
	User          *User          `json:"user" gorm:"foreignKey:UserID"`
	Items         []SalesOrderItem `json:"items" gorm:"foreignKey:SalesOrderID"`
}

// SalesOrderItem represents an item in a sales order
type SalesOrderItem struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	SalesOrderID  uint      `json:"sales_order_id" gorm:"not null"`
	ProductID     uint      `json:"product_id" gorm:"not null"`
	Quantity      int       `json:"quantity" gorm:"not null"`
	UnitPrice     float64   `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	Discount      float64   `json:"discount" gorm:"type:decimal(10,2);default:0"`
	TotalPrice    float64   `json:"total_price" gorm:"type:decimal(10,2);not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	SalesOrder    *SalesOrder    `json:"sales_order" gorm:"foreignKey:SalesOrderID"`
	Product       *Product       `json:"product" gorm:"foreignKey:ProductID"`
}

// BeforeCreate hook for sales order to generate SO number if not provided
func (so *SalesOrder) BeforeCreate(tx *gorm.DB) error {
	if so.SONumber == "" {
		// Generate a unique SO number
		var lastSO SalesOrder
		if err := tx.Order("id desc").First(&lastSO).Error; err == nil {
			so.SONumber = fmt.Sprintf("SO-%06d", lastSO.ID+1)
		} else {
			so.SONumber = "SO-000001"
		}
	}
	return nil
}

// BeforeCreate hook for sales order item to calculate total price
func (soi *SalesOrderItem) BeforeCreate(tx *gorm.DB) error {
	soi.TotalPrice = float64(soi.Quantity) * soi.UnitPrice * (1 - soi.Discount/100)
	return nil
}

// BeforeSave hook for sales order item to recalculate total price
func (soi *SalesOrderItem) BeforeSave(tx *gorm.DB) error {
	soi.TotalPrice = float64(soi.Quantity) * soi.UnitPrice * (1 - soi.Discount/100)
	return nil
}

// AfterCreate hook for sales order item to update sales order total
func (soi *SalesOrderItem) AfterCreate(tx *gorm.DB) error {
	return updateSalesOrderTotal(tx, soi.SalesOrderID)
}

// AfterUpdate hook for sales order item to update sales order total
func (soi *SalesOrderItem) AfterUpdate(tx *gorm.DB) error {
	return updateSalesOrderTotal(tx, soi.SalesOrderID)
}

// AfterDelete hook for sales order item to update sales order total
func (soi *SalesOrderItem) AfterDelete(tx *gorm.DB) error {
	return updateSalesOrderTotal(tx, soi.SalesOrderID)
}

// updateSalesOrderTotal recalculates all totals for a sales order
func updateSalesOrderTotal(tx *gorm.DB, soID uint) error {
	var subtotal float64
	if err := tx.Model(&SalesOrderItem{}).
		Where("sales_order_id = ?", soID).
		Select("SUM(total_price)").
		Scan(&subtotal).Error; err != nil {
		return err
	}
	
	// Get the sales order to calculate tax and total
	var so SalesOrder
	if err := tx.First(&so, soID).Error; err != nil {
		return err
	}
	
	so.Subtotal = subtotal
	// Tax calculation could be more complex in a real system
	so.Tax = subtotal * 0.10 // Assuming 10% tax rate
	so.TotalAmount = so.Subtotal + so.Tax + so.ShippingCost
	
	return tx.Save(&so).Error
}