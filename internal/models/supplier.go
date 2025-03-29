package models

import (
	"time"
)

// Supplier represents a supplier entity
type Supplier struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"not null"`
	ContactPerson string    `json:"contact_person"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Address       string    `json:"address"`
	TaxID         string    `json:"tax_id"`
	PaymentTerms  string    `json:"payment_terms"`
	Status        string    `json:"status" gorm:"default:'active'"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Products       []Product       `json:"products,omitempty" gorm:"many2many:product_supplier"`
	PurchaseOrders []PurchaseOrder `json:"purchase_orders,omitempty" gorm:"foreignKey:SupplierID"`
}