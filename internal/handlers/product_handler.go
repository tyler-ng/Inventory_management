package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourusername/inventory-management-system/internal/models"
	"github.com/yourusername/inventory-management-system/internal/repository"
	"gorm.io/gorm"
)

// ProductHandler handles HTTP requests for product endpoints
type ProductHandler struct {
	repo *repository.ProductRepository
	db   *gorm.DB
}

// NewProductHandler creates a new product handler
func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{
		repo: repository.NewProductRepository(db),
		db:   db,
	}
}

// GetProducts handles GET requests to retrieve products
func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	params := make(map[string]interface{})
	
	// Category filter
	if category := r.URL.Query().Get("category"); category != "" {
		params["category"] = category
	}
	
	// Search
	if search := r.URL.Query().Get("search"); search != "" {
		params["search"] = search
	}
	
	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		params["status"] = status
	}
	
	// Sorting
	if sort := r.URL.Query().Get("sort"); sort != "" {
		params["sort"] = sort
	}
	
	// Pagination
	if page := r.URL.Query().Get("page"); page != "" {
		pageNum, err := strconv.Atoi(page)
		if err == nil && pageNum > 0 {
			params["page"] = pageNum
			
			if limit := r.URL.Query().Get("limit"); limit != "" {
				limitNum, err := strconv.Atoi(limit)
				if err == nil && limitNum > 0 {
					params["limit"] = limitNum
				}
			}
		}
	}
	
	// Get products
	products, err := h.repo.GetAll(params)
	if err != nil {
		http.Error(w, "Failed to retrieve products: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetProduct handles GET requests to retrieve a single product
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	
	product, err := h.repo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve product: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// GetProductBySKU handles GET requests to retrieve a product by SKU
func (h *ProductHandler) GetProductBySKU(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sku := vars["sku"]
	
	product, err := h.repo.GetBySKU(sku)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve product: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// CreateProduct handles POST requests to create a new product
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	
	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate product
	if product.Name == "" || product.SKU == "" {
		http.Error(w, "Name and SKU are required", http.StatusBadRequest)
		return
	}
	
	// Check if SKU already exists
	existingProduct, err := h.repo.GetBySKU(product.SKU)
	if err == nil && existingProduct != nil {
		http.Error(w, "Product with this SKU already exists", http.StatusConflict)
		return
	}
	
	// Create product
	err = h.repo.Create(&product)
	if err != nil {
		http.Error(w, "Failed to create product: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// UpdateProduct handles PUT requests to update an existing product
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	
	// Check if product exists
	existingProduct, err := h.repo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve product: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Decode request body
	var updatedProduct models.Product
	err = json.NewDecoder(r.Body).Decode(&updatedProduct)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Set ID to ensure we're updating the correct record
	updatedProduct.ID = uint(id)
	
	// If SKU is being changed, check if new SKU already exists
	if updatedProduct.SKU != existingProduct.SKU {
		product, err := h.repo.GetBySKU(updatedProduct.SKU)
		if err == nil && product != nil && product.ID != uint(id) {
			http.Error(w, "Product with this SKU already exists", http.StatusConflict)
			return
		}
	}
	
	// Update product
	err = h.repo.Update(&updatedProduct)
	if err != nil {
		http.Error(w, "Failed to update product: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedProduct)
}

// DeleteProduct handles DELETE requests to delete a product
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	
	// Check if product exists
	_, err = h.repo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve product: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Delete product (soft delete)
	err = h.repo.Delete(uint(id))
	if err != nil {
		http.Error(w, "Failed to delete product: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// GetLowStockProducts handles GET requests to retrieve products with low stock
func (h *ProductHandler) GetLowStockProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.repo.GetLowStock()
	if err != nil {
		http.Error(w, "Failed to retrieve low stock products: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetProductsByWarehouse handles GET requests to retrieve products in a specific warehouse
func (h *ProductHandler) GetProductsByWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	warehouseID, err := strconv.ParseUint(vars["warehouseId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}
	
	productWarehouses, err := h.repo.GetProductsByWarehouse(uint(warehouseID))
	if err != nil {
		http.Error(w, "Failed to retrieve products: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(productWarehouses)
}

// GetProductCategories handles GET requests to retrieve categories of a product
func (h *ProductHandler) GetProductCategories(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	
	categories, err := h.repo.GetProductCategories(uint(productID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve categories: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}