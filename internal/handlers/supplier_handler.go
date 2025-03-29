package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/gorm"
)

// SupplierHandler handles HTTP requests for supplier endpoints
type SupplierHandler struct {
	db *gorm.DB
}

// NewSupplierHandler creates a new supplier handler
func NewSupplierHandler(db *gorm.DB) *SupplierHandler {
	return &SupplierHandler{db: db}
}

// GetSuppliers handles GET requests to retrieve all suppliers
func (h *SupplierHandler) GetSuppliers(w http.ResponseWriter, r *http.Request) {
	var suppliers []models.Supplier
	
	// Apply filters if any
	query := h.db
	
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	
	if name := r.URL.Query().Get("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	
	if err := query.Find(&suppliers).Error; err != nil {
		http.Error(w, "Failed to retrieve suppliers: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

// GetSupplier handles GET requests to retrieve a single supplier
func (h *SupplierHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid supplier ID", http.StatusBadRequest)
		return
	}
	
	var supplier models.Supplier
	if err := h.db.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Supplier not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve supplier: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(supplier)
}

// CreateSupplier handles POST requests to create a new supplier
func (h *SupplierHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var supplier models.Supplier
	
	if err := json.NewDecoder(r.Body).Decode(&supplier); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	if supplier.Name == "" {
		http.Error(w, "Supplier name is required", http.StatusBadRequest)
		return
	}
	
	if err := h.db.Create(&supplier).Error; err != nil {
		http.Error(w, "Failed to create supplier: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(supplier)
}

// UpdateSupplier handles PUT requests to update an existing supplier
func (h *SupplierHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid supplier ID", http.StatusBadRequest)
		return
	}
	
	// Check if supplier exists
	var existingSupplier models.Supplier
	if err := h.db.First(&existingSupplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Supplier not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve supplier: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Decode the request body
	var updatedSupplier models.Supplier
	if err := json.NewDecoder(r.Body).Decode(&updatedSupplier); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Set the ID to ensure we're updating the correct record
	updatedSupplier.ID = uint(id)
	
	// Update the supplier
	if err := h.db.Save(&updatedSupplier).Error; err != nil {
		http.Error(w, "Failed to update supplier: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSupplier)
}

// DeleteSupplier handles DELETE requests to delete a supplier
func (h *SupplierHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid supplier ID", http.StatusBadRequest)
		return
	}
	
	// Check if supplier exists
	var supplier models.Supplier
	if err := h.db.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Supplier not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve supplier: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Soft delete by updating status instead of actually deleting
	if err := h.db.Model(&supplier).Update("status", "inactive").Error; err != nil {
		http.Error(w, "Failed to delete supplier: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetSupplierProducts handles GET requests to retrieve products from a specific supplier
func (h *SupplierHandler) GetSupplierProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid supplier ID", http.StatusBadRequest)
		return
	}
	
	// Check if supplier exists
	var supplier models.Supplier
	if err := h.db.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Supplier not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve supplier: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Get products from the supplier
	var products []models.Product
	if err := h.db.Joins("JOIN product_supplier ON products.id = product_supplier.product_id").
		Where("product_supplier.supplier_id = ?", id).
		Find(&products).Error; err != nil {
		http.Error(w, "Failed to retrieve products: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}