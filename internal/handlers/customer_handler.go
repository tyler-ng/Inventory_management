package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/gorm"
)

// CustomerHandler handles HTTP requests for customer endpoints
type CustomerHandler struct {
	db *gorm.DB
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{db: db}
}

// GetCustomers handles GET requests to retrieve all customers
func (h *CustomerHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {
	var customers []models.Customer
	
	// Apply filters if any
	query := h.db
	
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	
	if name := r.URL.Query().Get("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	
	if email := r.URL.Query().Get("email"); email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	
	// Apply pagination
	page := 1
	limit := 10
	
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum > 0 {
			page = pageNum
		}
	}
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 {
			limit = limitNum
		}
	}
	
	offset := (page - 1) * limit
	
	// Execute query
	if err := query.Order("name ASC").Limit(limit).Offset(offset).Find(&customers).Error; err != nil {
		http.Error(w, "Failed to retrieve customers: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

// GetCustomer handles GET requests to retrieve a single customer
func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	
	var customer models.Customer
	if err := h.db.First(&customer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve customer: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

// CreateCustomer handles POST requests to create a new customer
func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer
	
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if customer.Name == "" {
		http.Error(w, "Customer name is required", http.StatusBadRequest)
		return
	}
	
	// Set default status if not provided
	if customer.Status == "" {
		customer.Status = "active"
	}
	
	// Create customer in database
	if err := h.db.Create(&customer).Error; err != nil {
		http.Error(w, "Failed to create customer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(customer)
}

// UpdateCustomer handles PUT requests to update an existing customer
func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	
	// Check if customer exists
	var existingCustomer models.Customer
	if err := h.db.First(&existingCustomer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve customer: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Parse request body
	var updatedCustomer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&updatedCustomer); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Set the ID to ensure we're updating the correct record
	updatedCustomer.ID = uint(id)
	
	// Update in database
	if err := h.db.Model(&updatedCustomer).Updates(updatedCustomer).Error; err != nil {
		http.Error(w, "Failed to update customer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Retrieve updated customer
	var finalCustomer models.Customer
	if err := h.db.First(&finalCustomer, id).Error; err != nil {
		http.Error(w, "Failed to retrieve updated customer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalCustomer)
}

// DeleteCustomer handles DELETE requests to delete a customer
func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	
	// Check if customer exists
	var customer models.Customer
	if err := h.db.First(&customer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve customer: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Check if customer has any sales orders
	var count int64
	if err := h.db.Model(&models.SalesOrder{}).Where("customer_id = ?", id).Count(&count).Error; err != nil {
		http.Error(w, "Failed to check customer usage: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	if count > 0 {
		// Soft delete by updating status
		if err := h.db.Model(&customer).Update("status", "inactive").Error; err != nil {
			http.Error(w, "Failed to deactivate customer: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Hard delete if no orders are associated
		if err := h.db.Delete(&customer).Error; err != nil {
			http.Error(w, "Failed to delete customer: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetCustomerSalesOrders handles GET requests to retrieve sales orders for a customer
func (h *CustomerHandler) GetCustomerSalesOrders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	
	// Check if customer exists
	var customer models.Customer
	if err := h.db.First(&customer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve customer: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Get orders
	var orders []models.SalesOrder
	if err := h.db.Where("customer_id = ?", id).Order("created_at DESC").Find(&orders).Error; err != nil {
		http.Error(w, "Failed to retrieve sales orders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}