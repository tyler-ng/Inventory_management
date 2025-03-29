package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/gorm"
)

// WarehouseHandler handles HTTP requests for warehouse endpoints
type WarehouseHandler struct {
	db *gorm.DB
}

// NewWarehouseHandler creates a new warehouse handler
func NewWarehouseHandler(db *gorm.DB) *WarehouseHandler {
	return &WarehouseHandler{db: db}
}

// GetWarehouses handles GET requests to retrieve all warehouses
func (h *WarehouseHandler) GetWarehouses(w http.ResponseWriter, r *http.Request) {
	var warehouses []models.Warehouse
	
	// Apply filters if any
	query := h.db
	
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	
	if name := r.URL.Query().Get("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	
	if err := query.Find(&warehouses).Error; err != nil {
		http.Error(w, "Failed to retrieve warehouses: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(warehouses)
}

// GetWarehouse handles GET requests to retrieve a single warehouse
func (h *WarehouseHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}
	
	var warehouse models.Warehouse
	if err := h.db.First(&warehouse, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Warehouse not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve warehouse: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(warehouse)
}

// CreateWarehouse handles POST requests to create a new warehouse
func (h *WarehouseHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var warehouse models.Warehouse
	
	if err := json.NewDecoder(r.Body).Decode(&warehouse); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	if warehouse.Name == "" {
		http.Error(w, "Warehouse name is required", http.StatusBadRequest)
		return
	}
	
	if err := h.db.Create(&warehouse).Error; err != nil {
		http.Error(w, "Failed to create warehouse: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(warehouse)
}

// UpdateWarehouse handles PUT requests to update an existing warehouse
func (h *WarehouseHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}
	
	var existingWarehouse models.Warehouse
	if err := h.db.First(&existingWarehouse, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Warehouse not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve warehouse: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	var updatedWarehouse models.Warehouse
	if err := json.NewDecoder(r.Body).Decode(&updatedWarehouse); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Set the ID to ensure we're updating the correct record
	updatedWarehouse.ID = uint(id)
	
	if err := h.db.Save(&updatedWarehouse).Error; err != nil {
		http.Error(w, "Failed to update warehouse: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedWarehouse)
}

// DeleteWarehouse handles DELETE requests to delete a warehouse
func (h *WarehouseHandler) DeleteWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}
	
	var warehouse models.Warehouse
	if err := h.db.First(&warehouse, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Warehouse not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve warehouse: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Soft delete by updating status
	if err := h.db.Model(&warehouse).Update("status", "inactive").Error; err != nil {
		http.Error(w, "Failed to delete warehouse: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetWarehouseProducts handles GET requests to retrieve products in a warehouse
func (h *WarehouseHandler) GetWarehouseProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["warehouseId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}
	
	var products []models.Product
	if err := h.db.Joins("JOIN product_warehouse ON products.id = product_warehouse.product_id").
		Where("product_warehouse.warehouse_id = ?", id).
		Find(&products).Error; err != nil {
		http.Error(w, "Failed to retrieve products: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetWarehouseLocations handles GET requests to retrieve locations within a warehouse
func (h *WarehouseHandler) GetWarehouseLocations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}
	
	var locations []models.WarehouseLocation
	if err := h.db.Where("warehouse_id = ?", id).Find(&locations).Error; err != nil {
		http.Error(w, "Failed to retrieve locations: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

// GetAllLocations handles GET requests to retrieve all warehouse locations
func (h *WarehouseHandler) GetAllLocations(w http.ResponseWriter, r *http.Request) {
	var locations []models.WarehouseLocation
	
	query := h.db.Preload("Warehouse")
	
	// Apply filters
	if warehouseID := r.URL.Query().Get("warehouse_id"); warehouseID != "" {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	
	if zone := r.URL.Query().Get("zone"); zone != "" {
		query = query.Where("zone = ?", zone)
	}
	
	if err := query.Find(&locations).Error; err != nil {
		http.Error(w, "Failed to retrieve locations: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

// GetLocation handles GET requests to retrieve a specific warehouse location
func (h *WarehouseHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}
	
	var location models.WarehouseLocation
	if err := h.db.Preload("Warehouse").First(&location, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Location not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve location: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(location)
}

// CreateLocation handles POST requests to create a new warehouse location
func (h *WarehouseHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var location models.WarehouseLocation
	
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	if location.WarehouseID == 0 {
		http.Error(w, "Warehouse ID is required", http.StatusBadRequest)
		return
	}
	
	if err := h.db.Create(&location).Error; err != nil {
		http.Error(w, "Failed to create location: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(location)
}

// UpdateLocation handles PUT requests to update an existing warehouse location
func (h *WarehouseHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}
	
	var existingLocation models.WarehouseLocation
	if err := h.db.First(&existingLocation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Location not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve location: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	var updatedLocation models.WarehouseLocation
	if err := json.NewDecoder(r.Body).Decode(&updatedLocation); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	updatedLocation.ID = uint(id)
	
	if err := h.db.Save(&updatedLocation).Error; err != nil {
		http.Error(w, "Failed to update location: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedLocation)
}

// DeleteLocation handles DELETE requests to delete a warehouse location
func (h *WarehouseHandler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}
	
	var location models.WarehouseLocation
	if err := h.db.First(&location, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Location not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve location: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	if err := h.db.Delete(&location).Error; err != nil {
		http.Error(w, "Failed to delete location: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}