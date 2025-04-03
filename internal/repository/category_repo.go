package repository

import (
	"errors"
	"fmt"

	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/gorm"
)

// CategoryRepository implements ICategoryRepository
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// GetAll retrieves all categories with optional filtering and pagination
func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	var categories []models.Category
	
	// Preload subcategories and parent category
	result := r.db.Preload("Subcategories").Preload("ParentCategory").Find(&categories)
	
	return categories, result.Error
}

// GetByID retrieves a specific category by its ID
func (r *CategoryRepository) GetByID(id uint) (*models.Category, error) {
	var category models.Category
	
	result := r.db.Preload("Subcategories").Preload("ParentCategory").First(&category, id)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("category with ID %d not found", id)
		}
		return nil, result.Error
	}
	
	return &category, nil
}

// Create adds a new category to the database
func (r *CategoryRepository) Create(category *models.Category) error {
	// Validate required fields
	if category.Name == "" {
		return errors.New("category name is required")
	}
	
	// Check for duplicate names
	var existingCategory models.Category
	if err := r.db.Where("name = ?", category.Name).First(&existingCategory).Error; err == nil {
		return errors.New("category with this name already exists")
	}
	
	return r.db.Create(category).Error
}

// Update modifies an existing category
func (r *CategoryRepository) Update(category *models.Category) error {
	// Validate required fields
	if category.ID == 0 {
		return errors.New("category ID is required")
	}
	
	if category.Name == "" {
		return errors.New("category name is required")
	}
	
	// Check if category exists
	var existingCategory models.Category
	if err := r.db.First(&existingCategory, category.ID).Error; err != nil {
		return fmt.Errorf("category not found: %v", err)
	}
	
	// Update the category
	return r.db.Save(category).Error
}

// Delete removes a category from the database
func (r *CategoryRepository) Delete(id uint) error {
	// Check for existing products in the category
	var productCount int64
	r.db.Model(&models.ProductCategory{}).Where("category_id = ?", id).Count(&productCount)
	
	if productCount > 0 {
		return errors.New("cannot delete category with associated products")
	}
	
	// Check for existing subcategories
	var subcategoryCount int64
	r.db.Model(&models.Category{}).Where("parent_id = ?", id).Count(&subcategoryCount)
	
	if subcategoryCount > 0 {
		return errors.New("cannot delete category with existing subcategories")
	}
	
	// Perform soft delete
	result := r.db.Delete(&models.Category{}, id)
	
	return result.Error
}

// GetSubcategories retrieves all subcategories for a given parent category
func (r *CategoryRepository) GetSubcategories(parentID uint) ([]models.Category, error) {
	var subcategories []models.Category
	
	result := r.db.Where("parent_id = ?", parentID).Find(&subcategories)
	
	return subcategories, result.Error
}

// GetCategoryProducts retrieves all products in a specific category
func (r *CategoryRepository) GetCategoryProducts(categoryID uint) ([]models.Product, error) {
	var products []models.Product
	
	result := r.db.Joins("JOIN product_category ON products.id = product_category.product_id").
		Where("product_category.category_id = ?", categoryID).
		Find(&products)
	
	return products, result.Error
}

// AddProductToCategory associates a product with a category
func (r *CategoryRepository) AddProductToCategory(productID, categoryID uint) error {
	// Check if product exists
	var product models.Product
	if err := r.db.First(&product, productID).Error; err != nil {
		return fmt.Errorf("product not found: %v", err)
	}
	
	// Check if category exists
	var category models.Category
	if err := r.db.First(&category, categoryID).Error; err != nil {
		return fmt.Errorf("category not found: %v", err)
	}
	
	// Create association
	productCategory := models.ProductCategory{
		ProductID:  productID,
		CategoryID: categoryID,
	}
	
	return r.db.Create(&productCategory).Error
}

// RemoveProductFromCategory removes a product from a category
func (r *CategoryRepository) RemoveProductFromCategory(productID, categoryID uint) error {
	return r.db.Where("product_id = ? AND category_id = ?", productID, categoryID).
		Delete(&models.ProductCategory{}).Error
}