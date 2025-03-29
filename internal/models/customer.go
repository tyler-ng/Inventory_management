package models

import (
	"time"
)

// Customer represents a customer entity
type Customer struct {
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
	SalesOrders   []SalesOrder `json:"sales_orders,omitempty" gorm:"foreignKey:CustomerID"`
}