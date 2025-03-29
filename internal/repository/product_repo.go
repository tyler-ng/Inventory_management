package repository

import (
	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/gorm"
)

// ProductRepository handles database operations for products
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// GetAll retrieves all products with optional filtering
func (r *ProductRepository) GetAll(params map[string]interface{}) ([]models.Product, error) {
	var products []models.Product
	
	query := r.db
	
	// Apply filters
	if category, ok := params["category"]; ok && category != "" {
		query = query.Joins("JOIN product_category ON products.id = product_category.product_id").
			Joins("JOIN categories ON product_category.category_id = categories.id").
			Where("categories.name = ?", category)
	}
	
	if search, ok := params["search"]; ok && search != "" {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("products.name LIKE ? OR products.sku LIKE ? OR products.description LIKE ?", 
			searchPattern, searchPattern, searchPattern)
	}
	
	if status, ok := params["status"]; ok && status != "" {
		query = query.Where("products.status = ?", status)
	}
	
	// Apply sorting
	if sort, ok := params["sort"]; ok && sort != "" {
		query = query.Order(sort)
	} else {
		query = query.Order("products.name ASC")
	}
	
	// Apply pagination
	if page, ok := params["page"].(int); ok {
		limit := 10 // Default limit
		if pageLimit, ok := params["limit"].(int); ok {
			limit = pageLimit
		}
		offset := (page - 1) * limit
		query = query.Limit(limit).Offset(offset)
	}
	
	// Execute query
	err := query.Find(&products).Error
	return products, err
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetBySKU retrieves a product by SKU
func (r *ProductRepository) GetBySKU(sku string) (*models.Product, error) {
	var product models.Product
	err := r.db.Where("sku = ?", sku).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Create creates a new product
func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

// Update updates an existing product
func (r *ProductRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

// Delete soft-deletes a product by updating its status
func (r *ProductRepository) Delete(id uint) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Update("status", "inactive").Error
}

// GetLowStock retrieves products with quantity below their reorder level
func (r *ProductRepository) GetLowStock() ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("quantity <= reorder_level AND status = 'active'").Find(&products).Error
	return products, err
}

// UpdateQuantity updates the quantity of a product
func (r *ProductRepository) UpdateQuantity(id uint, quantity int) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Update("quantity", quantity).Error
}

// GetProductsByWarehouse retrieves products in a specific warehouse
func (r *ProductRepository) GetProductsByWarehouse(warehouseID uint) ([]models.ProductWarehouse, error) {
	var productWarehouses []models.ProductWarehouse
	err := r.db.Where("warehouse_id = ?", warehouseID).
		Preload("Product").
		Preload("Warehouse").
		Preload("Location").
		Find(&productWarehouses).Error
	return productWarehouses, err
}

// GetProductVariants retrieves all variants of a product
func (r *ProductRepository) GetProductVariants(productID uint) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	err := r.db.Where("product_id = ?", productID).Find(&variants).Error
	return variants, err
}

// GetProductCategories retrieves all categories of a product
func (r *ProductRepository) GetProductCategories(productID uint) ([]models.Category, error) {
	var product models.Product
	err := r.db.Preload("Categories").First(&product, productID).Error
	if err != nil {
		return nil, err
	}
	return product.Categories, nil
}

// AddProductCategory adds a product to a category
func (r *ProductRepository) AddProductCategory(productID, categoryID uint) error {
	productCategory := models.ProductCategory{
		ProductID:  productID,
		CategoryID: categoryID,
	}
	return r.db.Create(&productCategory).Error
}

// RemoveProductCategory removes a product from a category
func (r *ProductRepository) RemoveProductCategory(productID, categoryID uint) error {
	return r.db.Where("product_id = ? AND category_id = ?", productID, categoryID).Delete(&models.ProductCategory{}).Error
}