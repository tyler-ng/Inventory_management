package models

import (
	"time"
)

// Category represents a product category
type Category struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	ParentID    *uint     `json:"parent_id"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	ParentCategory *Category  `json:"parent_category,omitempty" gorm:"foreignKey:ParentID"`
	Subcategories  []Category `json:"subcategories,omitempty" gorm:"foreignKey:ParentID"`
	Products       []Product  `json:"products,omitempty" gorm:"many2many:product_category"`
}

// GetCategoryPath returns the full path of the category
func (c *Category) GetCategoryPath() ([]Category, error) {
	// This would be implemented with recursive SQL query
	// For simplicity, we're just returning a placeholder
	return []Category{*c}, nil
}