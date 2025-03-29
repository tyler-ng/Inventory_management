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

// PurchaseOrderHandler handles HTTP requests for purchase order endpoints
type PurchaseOrderHandler struct {
	db *gorm.DB
}

// NewPurchaseOrderHandler creates a new purchase order handler
func NewPurchaseOrderHandler(db *gorm.DB) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{db: db}
}

// GetPurchaseOrders handles GET requests to retrieve all purchase orders
func (h *PurchaseOrderHandler) GetPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	var orders []models.PurchaseOrder
	
	// Apply filters if any
	query := h.db.Preload("Supplier").Preload("Warehouse").Preload("User")
	
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	
	if supplierID := r.URL.Query().Get("supplier_id"); supplierID != "" {
		query = query.Where("supplier_id = ?", supplierID)
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
		http.Error(w, "Failed to retrieve purchase orders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// GetPurchaseOrder handles GET requests to retrieve a single purchase order
func (h *PurchaseOrderHandler) GetPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid purchase order ID", http.StatusBadRequest)
		return
	}
	
	var order models.PurchaseOrder
	if err := h.db.Preload("Supplier").Preload("Warehouse").Preload("User").Preload("Items").
		Preload("Items.Product").First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Purchase order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve purchase order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// CreatePurchaseOrder handles POST requests to create a new purchase order
func (h *PurchaseOrderHandler) CreatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	var order models.PurchaseOrder
	
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if order.SupplierID == 0 || order.WarehouseID == 0 {
		http.Error(w, "Supplier ID and Warehouse ID are required", http.StatusBadRequest)
		return
	}
	
	// Set default values
	if order.OrderDate.IsZero() {
		order.OrderDate = time.Now()
	}
	
	if order.Status == "" {
		order.Status = "draft"
	}
	
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	order.UserID = userID
	
	// Generate PO number if not provided
	if order.PONumber == "" {
		// Find the last PO to generate a sequential number
		var lastPO models.PurchaseOrder
		if err := h.db.Order("id desc").First(&lastPO).Error; err == nil {
			order.PONumber = "PO-" + strconv.FormatUint(uint64(lastPO.ID+1), 10)
		} else {
			order.PONumber = "PO-1"
		}
	}
	
	// Create purchase order in database
	if err := h.db.Create(&order).Error; err != nil {
		http.Error(w, "Failed to create purchase order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// UpdatePurchaseOrder handles PUT requests to update an existing purchase order
func (h *PurchaseOrderHandler) UpdatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid purchase order ID", http.StatusBadRequest)
		return
	}
	
	// Check if purchase order exists
	var existingOrder models.PurchaseOrder
	if err := h.db.First(&existingOrder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Purchase order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve purchase order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Only draft orders can be updated
	if existingOrder.Status != "draft" {
		http.Error(w, "Only draft purchase orders can be updated", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var updatedOrder models.PurchaseOrder
	if err := json.NewDecoder(r.Body).Decode(&updatedOrder); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Set the ID to ensure we're updating the correct record
	updatedOrder.ID = uint(id)
	
	// Keep the original PO number
	updatedOrder.PONumber = existingOrder.PONumber
	
	// Update in database
	if err := h.db.Model(&updatedOrder).Updates(updatedOrder).Error; err != nil {
		http.Error(w, "Failed to update purchase order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Retrieve updated purchase order with relationships
	var finalOrder models.PurchaseOrder
	if err := h.db.Preload("Supplier").Preload("Warehouse").Preload("User").First(&finalOrder, id).Error; err != nil {
		http.Error(w, "Failed to retrieve updated purchase order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalOrder)
}

// DeletePurchaseOrder handles DELETE requests to delete a purchase order
func (h *PurchaseOrderHandler) DeletePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid purchase order ID", http.StatusBadRequest)
		return
	}
	
	// Check if purchase order exists
	var order models.PurchaseOrder
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Purchase order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve purchase order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Only draft orders can be deleted
	if order.Status != "draft" {
		http.Error(w, "Only draft purchase orders can be deleted", http.StatusBadRequest)
		return
	}
	
	// Delete the order items first
	if err := h.db.Where("purchase_order_id = ?", id).Delete(&models.PurchaseOrderItem{}).Error; err != nil {
		http.Error(w, "Failed to delete purchase order items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Delete the order
	if err := h.db.Delete(&order).Error; err != nil {
		http.Error(w, "Failed to delete purchase order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetPurchaseOrderItems handles GET requests to retrieve items for a purchase order
func (h *PurchaseOrderHandler) GetPurchaseOrderItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid purchase order ID", http.StatusBadRequest)
		return
	}
	
	// Check if purchase order exists
	var order models.PurchaseOrder
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Purchase order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve purchase order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Get items
	var items []models.PurchaseOrderItem
	if err := h.db.Preload("Product").Where("purchase_order_id = ?", id).Find(&items).Error; err != nil {
		http.Error(w, "Failed to retrieve items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// AddPurchaseOrderItem handles POST requests to add an item to a purchase order
func (h *PurchaseOrderHandler) AddPurchaseOrderItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid purchase order ID", http.StatusBadRequest)
		return
	}
	
	// Check if purchase order exists
	var order models.PurchaseOrder
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Purchase order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve purchase order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Only draft orders can be modified
	if order.Status != "draft" {
		http.Error(w, "Only draft purchase orders can be modified", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var item models.PurchaseOrderItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate item
	if item.ProductID == 0 || item.Quantity <= 0 || item.UnitPrice <= 0 {
		http.Error(w, "Product ID, quantity, and unit price are required and must be positive", http.StatusBadRequest)
		return
	}
	
	// Check if product exists
	var product models.Product
	if err := h.db.First(&product, item.ProductID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve product: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Set purchase order ID and calculate total price
	item.PurchaseOrderID = uint(id)
	item.TotalPrice = float64(item.Quantity) * item.UnitPrice
	
	// Create item in database
	if err := h.db.Create(&item).Error; err != nil {
		http.Error(w, "Failed to add item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Update the purchase order's total amount
	var totalAmount float64
	h.db.Model(&models.PurchaseOrderItem{}).Where("purchase_order_id = ?", id).
		Select("SUM(total_price)").Scan(&totalAmount)
	
	h.db.Model(&order).Update("total_amount", totalAmount)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// ReceivePurchaseOrder handles POST requests to receive items from a purchase order
func (h *PurchaseOrderHandler) ReceivePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid purchase order ID", http.StatusBadRequest)
		return
	}
	
	// Check if purchase order exists
	var order models.PurchaseOrder
	if err := h.db.Preload("Items").Preload("Items.Product").Preload("Warehouse").
		First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Purchase order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve purchase order: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Only pending or partially received orders can be received
	if order.Status != "pending" && order.Status != "partial" && order.Status != "approved" {
		http.Error(w, "Only pending, approved, or partially received purchase orders can be received", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var request struct {
		Items []struct {
			ItemID           uint `json:"item_id"`
			QuantityReceived int  `json:"quantity_received"`
		} `json:"items"`
		Notes string `json:"notes"`
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
	totalReceived := 0
	totalOrdered := 0
	
	for _, requestItem := range request.Items {
		// Find the item in the purchase order
		var item models.PurchaseOrderItem
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
			http.Error(w, "Item not found in purchase order", http.StatusBadRequest)
			return
		}
		
		if requestItem.QuantityReceived <= 0 || requestItem.QuantityReceived > item.Quantity {
			tx.Rollback()
			http.Error(w, "Invalid quantity received", http.StatusBadRequest)
			return
		}
		
		// Update product quantity
		if err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).
			UpdateColumn("quantity", gorm.Expr("quantity + ?", requestItem.QuantityReceived)).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to update product quantity: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Create inventory transaction
		transaction := models.InventoryTransaction{
			ProductID:         item.ProductID,
			WarehouseID:       order.WarehouseID,
			Type:              "receive",
			Quantity:          requestItem.QuantityReceived,
			ReferenceNumber:   order.PONumber,
			UserID:            userID,
			Notes:             "Received from purchase order: " + order.PONumber,
		}
		
		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to create inventory transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		totalReceived += requestItem.QuantityReceived
		totalOrdered += item.Quantity
	}
	
	// Update purchase order status
	if totalReceived == totalOrdered {
		if err := tx.Model(&order).Update("status", "received").Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to update purchase order status: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := tx.Model(&order).Update("status", "partial").Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to update purchase order status: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return updated purchase order
	var updatedOrder models.PurchaseOrder
	if err := h.db.Preload("Items").Preload("Items.Product").Preload("Supplier").
		Preload("Warehouse").Preload("User").First(&updatedOrder, id).Error; err != nil {
		http.Error(w, "Failed to retrieve updated purchase order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)
}