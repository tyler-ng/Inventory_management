package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/gorm"
)

// SalesOrderHandler handles HTTP requests for sales order endpoints
type SalesOrderHandler struct {
	db *gorm.DB
}

// NewSalesOrderHandler creates a new sales order handler
func NewSalesOrderHandler(db *gorm.DB) *SalesOrderHandler {
	return &SalesOrderHandler{db: db}
}

// GetSalesOrders handles GET requests to retrieve all sales orders
func (h *SalesOrderHandler) GetSalesOrders(w http.ResponseWriter, r *http.Request) {
	var orders []models.SalesOrder
	
	// Apply filters if any
	query := h.db.Preload("Customer").Preload("Warehouse").Preload("User")
	
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	
	if customerID := r.URL.Query().Get("customer_id"); customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}
	
	if warehouseID := r.URL.Query().Get("warehouse_id"); warehouseID != "" {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		query = query.Where("order_date >= ?", startDate)
	}
	
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		query = query.Where("order_date <= ?", endDate)
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
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		http.Error(w, "Failed to retrieve sales orders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// GetSalesOrder handles GET requests to retrieve a single sales order
func (h *SalesOrderHandler) GetSalesOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid sales order ID", http.StatusBadRequest)
		return
	}
	
	var order models.SalesOrder
	if err := h.db.Preload("Customer").Preload("Warehouse").Preload("User").Preload("Items").
		Preload("Items.Product").First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Sales order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve sales order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// CreateSalesOrder handles POST requests to create a new sales order
func (h *SalesOrderHandler) CreateSalesOrder(w http.ResponseWriter, r *http.Request) {
	var order models.SalesOrder
	
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if order.CustomerID == 0 || order.WarehouseID == 0 {
		http.Error(w, "Customer ID and Warehouse ID are required", http.StatusBadRequest)
		return
	}
	
	// Set default values
	if order.OrderDate.IsZero() {
		order.OrderDate = time.Now()
	}
	
	if order.Status == "" {
		order.Status = "draft"
	}
	
	if order.PaymentStatus == "" {
		order.PaymentStatus = "unpaid"
	}
	
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	order.UserID = userID
	
	// Generate SO number if not provided
	if order.SONumber == "" {
		// Find the last SO to generate a sequential number
		var lastSO models.SalesOrder
		if err := h.db.Order("id desc").First(&lastSO).Error; err == nil {
			order.SONumber = "SO-" + strconv.FormatUint(uint64(lastSO.ID+1), 10)
		} else {
			order.SONumber = "SO-1"
		}
	}
	
	// Create sales order in database
	if err := h.db.Create(&order).Error; err != nil {
		http.Error(w, "Failed to create sales order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// UpdateSalesOrder handles PUT requests to update an existing sales order
func (h *SalesOrderHandler) UpdateSalesOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid sales order ID", http.StatusBadRequest)
		return
	}
	
	// Check if sales order exists
	var existingOrder models.SalesOrder
	if err := h.db.First(&existingOrder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Sales order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve sales order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Only draft orders can be updated
	if existingOrder.Status != "draft" {
		http.Error(w, "Only draft sales orders can be updated", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var updatedOrder models.SalesOrder
	if err := json.NewDecoder(r.Body).Decode(&updatedOrder); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Set the ID to ensure we're updating the correct record
	updatedOrder.ID = uint(id)
	
	// Keep the original SO number
	updatedOrder.SONumber = existingOrder.SONumber
	
	// Update in database
	if err := h.db.Model(&updatedOrder).Updates(updatedOrder).Error; err != nil {
		http.Error(w, "Failed to update sales order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Retrieve updated sales order with relationships
	var finalOrder models.SalesOrder
	if err := h.db.Preload("Customer").Preload("Warehouse").Preload("User").First(&finalOrder, id).Error; err != nil {
		http.Error(w, "Failed to retrieve updated sales order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalOrder)
}

// DeleteSalesOrder handles DELETE requests to delete a sales order
func (h *SalesOrderHandler) DeleteSalesOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid sales order ID", http.StatusBadRequest)
		return
	}
	
	// Check if sales order exists
	var order models.SalesOrder
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Sales order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve sales order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Only draft orders can be deleted
	if order.Status != "draft" {
		http.Error(w, "Only draft sales orders can be deleted", http.StatusBadRequest)
		return
	}
	
	// Delete the order items first
	if err := h.db.Where("sales_order_id = ?", id).Delete(&models.SalesOrderItem{}).Error; err != nil {
		http.Error(w, "Failed to delete sales order items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Delete the order
	if err := h.db.Delete(&order).Error; err != nil {
		http.Error(w, "Failed to delete sales order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetSalesOrderItems handles GET requests to retrieve items for a sales order
func (h *SalesOrderHandler) GetSalesOrderItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid sales order ID", http.StatusBadRequest)
		return
	}
	
	// Check if sales order exists
	var order models.SalesOrder
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Sales order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve sales order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Get items
	var items []models.SalesOrderItem
	if err := h.db.Preload("Product").Where("sales_order_id = ?", id).Find(&items).Error; err != nil {
		http.Error(w, "Failed to retrieve items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// AddSalesOrderItem handles POST requests to add an item to a sales order
func (h *SalesOrderHandler) AddSalesOrderItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid sales order ID", http.StatusBadRequest)
		return
	}
	
	// Check if sales order exists
	var order models.SalesOrder
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Sales order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve sales order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Only draft orders can be modified
	if order.Status != "draft" {
		http.Error(w, "Only draft sales orders can be modified", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var item models.SalesOrderItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate item
	if item.ProductID == 0 || item.Quantity <= 0 || item.UnitPrice <= 0 {
		http.Error(w, "Product ID, quantity, and unit price are required and must be positive", http.StatusBadRequest)
		return
	}
	
	// Check if product exists and has sufficient stock
	var product models.Product
	if err := h.db.First(&product, item.ProductID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve product: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	if product.Quantity < item.Quantity {
		http.Error(w, "Insufficient stock available", http.StatusBadRequest)
		return
	}
	
	// Set sales order ID and calculate total price
	item.SalesOrderID = uint(id)
	item.TotalPrice = float64(item.Quantity) * item.UnitPrice * (1 - item.Discount/100)
	
	// Create item in database
	if err := h.db.Create(&item).Error; err != nil {
		http.Error(w, "Failed to add item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Update the sales order's total amount
	var subtotal float64
	h.db.Model(&models.SalesOrderItem{}).Where("sales_order_id = ?", id).
		Select("SUM(total_price)").Scan(&subtotal)
	
	// Calculate tax and shipping cost
	tax := subtotal * 0.1 // Assuming 10% tax rate
	
	// Update sales order
	h.db.Model(&order).Updates(map[string]interface{}{
		"subtotal": subtotal,
		"tax": tax,
		"total_amount": subtotal + tax + order.ShippingCost,
	})
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// FulfillSalesOrder handles POST requests to fulfill a sales order
func (h *SalesOrderHandler) FulfillSalesOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid sales order ID", http.StatusBadRequest)
		return
	}
	
	// Check if sales order exists
	var order models.SalesOrder
	if err := h.db.Preload("Items").Preload("Items.Product").Preload("Warehouse").
		First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Sales order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve sales order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Only confirmed orders can be fulfilled
	if order.Status != "confirmed" && order.Status != "partial" {
		http.Error(w, "Only confirmed or partially fulfilled sales orders can be fulfilled", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var request struct {
		Items []struct {
			ItemID         uint `json:"item_id"`
			QuantityFulfilled int  `json:"quantity_fulfilled"`
		} `json:"items"`
		ShippingDate *time.Time `json:"shipping_date,omitempty"`
		Notes        string     `json:"notes"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	
	// Start a transaction for database operations
	tx := h.db.Begin()
	
	if tx.Error != nil {
		http.Error(w, "Failed to start database transaction: "+tx.Error.Error(), http.StatusInternalServerError)
		return
	}
	
	// Process each item
	totalFulfilled := 0
	totalOrdered := 0
	
	for _, requestItem := range request.Items {
		// Find the item in the sales order
		var item models.SalesOrderItem
		found := false
		
		for _, orderItem := range order.Items {
			if orderItem.ID == requestItem.ItemID {
				item = orderItem
				found = true
				break
			}
		}
		
		if !found {
			tx.Rollback()
			http.Error(w, "Item not found in sales order", http.StatusBadRequest)
			return
		}
		
		if requestItem.QuantityFulfilled <= 0 || requestItem.QuantityFulfilled > item.Quantity {
			tx.Rollback()
			http.Error(w, "Invalid quantity fulfilled", http.StatusBadRequest)
			return
		}
		
		// Check if enough stock is available
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to retrieve product: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		if product.Quantity < requestItem.QuantityFulfilled {
			tx.Rollback()
			http.Error(w, "Insufficient stock for product: "+product.Name, http.StatusBadRequest)
			return
		}
		
		// Update product quantity
		if err := tx.Model(&product).Update("quantity", product.Quantity-requestItem.QuantityFulfilled).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to update product quantity: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Create inventory transaction
		transaction := models.InventoryTransaction{
			ProductID:         item.ProductID,
			WarehouseID:       order.WarehouseID,
			Type:              "issue",
			Quantity:          requestItem.QuantityFulfilled,
			ReferenceNumber:   order.SONumber,
			UserID:            userID,
			Notes:             "Fulfilled from sales order: " + order.SONumber,
		}
		
		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to create inventory transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		totalFulfilled += requestItem.QuantityFulfilled
		totalOrdered += item.Quantity
	}
	
	// Update sales order status and shipping date
	updates := map[string]interface{}{}
	
	if totalFulfilled == totalOrdered {
		updates["status"] = "fulfilled"
	} else {
		updates["status"] = "partial"
	}
	
	if request.ShippingDate != nil {
		updates["shipping_date"] = request.ShippingDate
	} else {
		updates["shipping_date"] = time.Now()
	}
	
	if err := tx.Model(&order).Updates(updates).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update sales order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return updated sales order
	var updatedOrder models.SalesOrder
	if err := h.db.Preload("Items").Preload("Items.Product").Preload("Customer").
		Preload("Warehouse").Preload("User").First(&updatedOrder, id).Error; err != nil {
		http.Error(w, "Failed to retrieve updated sales order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)
}