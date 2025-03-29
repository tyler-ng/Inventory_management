package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/gorm"
)

// CategoryHandler handles HTTP requests for category endpoints
type CategoryHandler struct {
	db *gorm.DB
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

// GetCategories handles GET requests to retrieve all categories
func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	var categories []models.Category
	
	if err := h.db.Find(&categories).Error; err != nil {
		http.Error(w, "Failed to retrieve categories: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// GetCategory handles GET requests to retrieve a single category
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	
	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve category: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// CreateCategory handles POST requests to create a new category
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	if category.Name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}
	
	if err := h.db.Create(&category).Error; err != nil {
		http.Error(w, "Failed to create category: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// UpdateCategory handles PUT requests to update an existing category
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	
	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve category: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	var updatedCategory models.Category
	if err := json.NewDecoder(r.Body).Decode(&updatedCategory); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	updatedCategory.ID = uint(id)
	
	if err := h.db.Save(&updatedCategory).Error; err != nil {
		http.Error(w, "Failed to update category: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCategory)
}

// DeleteCategory handles DELETE requests to delete a category
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	
	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve category: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	if err := h.db.Delete(&category).Error; err != nil {
		http.Error(w, "Failed to delete category: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetSubcategories handles GET requests to retrieve subcategories
func (h *CategoryHandler) GetSubcategories(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	
	var subcategories []models.Category
	if err := h.db.Where("parent_id = ?", id).Find(&subcategories).Error; err != nil {
		http.Error(w, "Failed to retrieve subcategories: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subcategories)
}

// GetCategoryProducts handles GET requests to retrieve products in a category
func (h *CategoryHandler) GetCategoryProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	
	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve category: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	var products []models.Product
	if err := h.db.Joins("JOIN product_category ON products.id = product_category.product_id").
		Where("product_category.category_id = ?", id).
		Find(&products).Error; err != nil {
		http.Error(w, "Failed to retrieve products: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}