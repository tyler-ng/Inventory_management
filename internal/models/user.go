package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	FullName     string    `json:"full_name" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:'user'"`
	Status       string    `json:"status" gorm:"default:'active'"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	AuditLogs            []AuditLog            `json:"-" gorm:"foreignKey:UserID"`
	InventoryTransactions []InventoryTransaction `json:"-" gorm:"foreignKey:UserID"`
	PurchaseOrders       []PurchaseOrder       `json:"-" gorm:"foreignKey:UserID"`
	SalesOrders          []SalesOrder          `json:"-" gorm:"foreignKey:UserID"`
}

// SetPassword sets a new password for the user
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// BeforeCreate hook for user to hash the password before creating the record
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.PasswordHash != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hashedPassword)
	}
	return nil
}