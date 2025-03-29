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

// TransactionHandler handles HTTP requests for inventory transaction endpoints
type TransactionHandler struct {
	repo *repository.TransactionRepository
	db   *gorm.DB
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(db *gorm.DB) *TransactionHandler {
	return &TransactionHandler{
		repo: repository.NewTransactionRepository(db),
		db:   db,
	}
}

// GetTransactions handles GET requests to retrieve inventory transactions
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	params := make(map[string]interface{})
	
	// Type filter
	if txType := r.URL.Query().Get("type"); txType != "" {
		params["type"] = txType
	}
	
	// Date range filter
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		params["start_date"] = startDate
	}
	
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		params["end_date"] = endDate
	}
	
	// Product filter
	if productID := r.URL.Query().Get("product_id"); productID != "" {
		productIDInt, err := strconv.ParseUint(productID, 10, 64)
		if err == nil {
			params["product_id"] = uint(productIDInt)
		}
	}
	
	// Warehouse filter
	if warehouseID := r.URL.Query().Get("warehouse_id"); warehouseID != "" {
		warehouseIDInt, err := strconv.ParseUint(warehouseID, 10, 64)
		if err == nil {
			params["warehouse_id"] = uint(warehouseIDInt)
		}
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
	
	// Get transactions
	transactions, err := h.repo.GetAll(params)
	if err != nil {
		http.Error(w, "Failed to retrieve transactions: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// GetTransaction handles GET requests to retrieve a single transaction
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}
	
	transaction, err := h.repo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Transaction not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve transaction: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

// CreateTransaction handles POST requests to create a new inventory transaction
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction models.InventoryTransaction
	
	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate transaction
	if transaction.ProductID == 0 || transaction.WarehouseID == 0 || transaction.Quantity == 0 {
		http.Error(w, "Product ID, warehouse ID, and quantity are required", http.StatusBadRequest)
		return
	}
	
	// Set user ID from context (would be set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	transaction.UserID = userID
	
	// Create transaction
	err = h.repo.Create(&transaction)
	if err != nil {
		http.Error(w, "Failed to create transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

// GetProductTransactions handles GET requests to retrieve transactions for a specific product
func (h *TransactionHandler) GetProductTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseUint(vars["productId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	
	// Parse query parameters for filtering
	params := make(map[string]interface{})
	params["product_id"] = uint(productID)
	
	// Type filter
	if txType := r.URL.Query().Get("type"); txType != "" {
		params["type"] = txType
	}
	
	// Date range filter
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		params["start_date"] = startDate
	}
	
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		params["end_date"] = endDate
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
	
	transactions, err := h.repo.GetAll(params)
	if err != nil {
		http.Error(w, "Failed to retrieve transactions: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// CreateReceiveTransaction handles POST requests to create a receive transaction
func (h *TransactionHandler) CreateReceiveTransaction(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ProductID      uint   `json:"product_id"`
		WarehouseID    uint   `json:"warehouse_id"`
		LocationID     *uint  `json:"location_id"`
		Quantity       int    `json:"quantity"`
		ReferenceNumber string `json:"reference_number"`
		Notes          string `json:"notes"`
	}
	
	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate request
	if request.ProductID == 0 || request.WarehouseID == 0 || request.Quantity <= 0 {
		http.Error(w, "Product ID, warehouse ID, and quantity > 0 are required", http.StatusBadRequest)
		return
	}
	
	// Set user ID from context (would be set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	
	// Create transaction
	transaction := models.InventoryTransaction{
		ProductID:             request.ProductID,
		WarehouseID:           request.WarehouseID,
		DestinationLocationID: request.LocationID,
		Type:                  "receive",
		Quantity:              request.Quantity,
		ReferenceNumber:       request.ReferenceNumber,
		Notes:                 request.Notes,
		UserID:                userID,
	}
	
	err = h.repo.Create(&transaction)
	if err != nil {
		http.Error(w, "Failed to create transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

// CreateIssueTransaction handles POST requests to create an issue transaction
func (h *TransactionHandler) CreateIssueTransaction(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ProductID      uint   `json:"product_id"`
		WarehouseID    uint   `json:"warehouse_id"`
		LocationID     *uint  `json:"location_id"`
		Quantity       int    `json:"quantity"`
		ReferenceNumber string `json:"reference_number"`
		Notes          string `json:"notes"`
	}
	
	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate request
	if request.ProductID == 0 || request.WarehouseID == 0 || request.Quantity <= 0 {
		http.Error(w, "Product ID, warehouse ID, and quantity > 0 are required", http.StatusBadRequest)
		return
	}
	
	// Check if enough stock is available
	productRepo := repository.NewProductRepository(h.db)
	product, err := productRepo.GetByID(request.ProductID)
	if err != nil {
		http.Error(w, "Failed to retrieve product: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	if product.Quantity < request.Quantity {
		http.Error(w, "Insufficient stock available", http.StatusBadRequest)
		return
	}
	
	// Set user ID from context (would be set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	
	// Create transaction
	transaction := models.InventoryTransaction{
		ProductID:        request.ProductID,
		WarehouseID:      request.WarehouseID,
		SourceLocationID: request.LocationID,
		Type:             "issue",
		Quantity:         request.Quantity,
		ReferenceNumber:  request.ReferenceNumber,
		Notes:            request.Notes,
		UserID:           userID,
	}
	
	err = h.repo.Create(&transaction)
	if err != nil {
		http.Error(w, "Failed to create transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

// CreateTransferTransaction handles POST requests to create a transfer transaction
func (h *TransactionHandler) CreateTransferTransaction(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ProductID             uint   `json:"product_id"`
		WarehouseID           uint   `json:"warehouse_id"`
		SourceLocationID      uint   `json:"source_location_id"`
		DestinationLocationID uint   `json:"destination_location_id"`
		Quantity              int    `json:"quantity"`
		ReferenceNumber       string `json:"reference_number"`
		Notes                 string `json:"notes"`
	}
	
	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate request
	if request.ProductID == 0 || request.WarehouseID == 0 || request.Quantity <= 0 ||
		request.SourceLocationID == 0 || request.DestinationLocationID == 0 {
		http.Error(w, "Product ID, warehouse ID, source location, destination location, and quantity > 0 are required", http.StatusBadRequest)
		return
	}
	
	// Set user ID from context (would be set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	
	// Create transaction
	transaction := models.InventoryTransaction{
		ProductID:             request.ProductID,
		WarehouseID:           request.WarehouseID,
		SourceLocationID:      &request.SourceLocationID,
		DestinationLocationID: &request.DestinationLocationID,
		Type:                  "transfer",
		Quantity:              request.Quantity,
		ReferenceNumber:       request.ReferenceNumber,
		Notes:                 request.Notes,
		UserID:                userID,
	}
	
	err = h.repo.Create(&transaction)
	if err != nil {
		http.Error(w, "Failed to create transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}