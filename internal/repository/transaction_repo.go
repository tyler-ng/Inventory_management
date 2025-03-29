package repository

import (
	"time"

	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/gorm"
)

// TransactionRepository handles database operations for inventory transactions
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// GetAll retrieves all inventory transactions with optional filtering
func (r *TransactionRepository) GetAll(params map[string]interface{}) ([]models.InventoryTransaction, error) {
	var transactions []models.InventoryTransaction
	
	query := r.db.Preload("Product").Preload("Warehouse").
		Preload("SourceLocation").Preload("DestinationLocation").Preload("User")
	
	// Apply filters
	if productID, ok := params["product_id"].(uint); ok {
		query = query.Where("product_id = ?", productID)
	}
	
	if warehouseID, ok := params["warehouse_id"].(uint); ok {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	
	if txType, ok := params["type"].(string); ok {
		query = query.Where("type = ?", txType)
	}
	
	if startDate, ok := params["start_date"].(string); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	
	if endDate, ok := params["end_date"].(string); ok {
		query = query.Where("created_at <= ?", endDate)
	}
	
	// Apply sorting
	query = query.Order("created_at DESC")
	
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
	err := query.Find(&transactions).Error
	return transactions, err
}

// GetByID retrieves a transaction by ID
func (r *TransactionRepository) GetByID(id uint) (*models.InventoryTransaction, error) {
	var transaction models.InventoryTransaction
	err := r.db.Preload("Product").Preload("Warehouse").
		Preload("SourceLocation").Preload("DestinationLocation").Preload("User").
		First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// Create creates a new inventory transaction
func (r *TransactionRepository) Create(transaction *models.InventoryTransaction) error {
	// Start a transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the transaction record
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}
		
		// Update the product quantity based on the transaction type
		var product models.Product
		if err := tx.First(&product, transaction.ProductID).Error; err != nil {
			return err
		}
		
		switch transaction.Type {
		case "receive":
			product.Quantity += transaction.Quantity
		case "issue":
			product.Quantity -= transaction.Quantity
		case "transfer":
			// No change in overall quantity for transfers
			// In a real system, we might track quantities by location
		case "adjustment":
			// For adjustments, the quantity might be positive or negative
			// Here we assume the quantity represents the change to make
			product.Quantity += transaction.Quantity
		}
		
		if err := tx.Save(&product).Error; err != nil {
			return err
		}
		
		// If it's a warehouse transfer, update product_warehouse records
		if transaction.Type == "transfer" && transaction.SourceLocationID != nil && transaction.DestinationLocationID != nil {
			// Reduce quantity at source location
			var sourceProductWarehouse models.ProductWarehouse
			err := tx.Where("product_id = ? AND warehouse_id = ? AND location_id = ?",
				transaction.ProductID, transaction.WarehouseID, *transaction.SourceLocationID).
				First(&sourceProductWarehouse).Error
			
			if err == nil {
				sourceProductWarehouse.Quantity -= transaction.Quantity
				if err := tx.Save(&sourceProductWarehouse).Error; err != nil {
					return err
				}
			}
			
			// Increase quantity at destination location
			var destProductWarehouse models.ProductWarehouse
			err = tx.Where("product_id = ? AND warehouse_id = ? AND location_id = ?",
				transaction.ProductID, transaction.WarehouseID, *transaction.DestinationLocationID).
				First(&destProductWarehouse).Error
			
			if err == gorm.ErrRecordNotFound {
				// Create a new record for destination
				destProductWarehouse = models.ProductWarehouse{
					ProductID:   transaction.ProductID,
					WarehouseID: transaction.WarehouseID,
					LocationID:  *transaction.DestinationLocationID,
					Quantity:    transaction.Quantity,
				}
				if err := tx.Create(&destProductWarehouse).Error; err != nil {
					return err
				}
			} else if err == nil {
				// Update existing destination record
				destProductWarehouse.Quantity += transaction.Quantity
				if err := tx.Save(&destProductWarehouse).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		
		return nil
	})
}

// GetProductTransactions retrieves transactions for a specific product
func (r *TransactionRepository) GetProductTransactions(productID uint, startDate, endDate time.Time) ([]models.InventoryTransaction, error) {
	var transactions []models.InventoryTransaction
	
	query := r.db.Where("product_id = ?", productID).
		Preload("Warehouse").Preload("SourceLocation").Preload("DestinationLocation").Preload("User")
	
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}
	
	err := query.Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

// GetProductMovementSummary returns a summary of product movements
func (r *TransactionRepository) GetProductMovementSummary(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// This would typically use SQL GROUP BY for efficient aggregation
	// Here's a simplified approach
	var transactions []models.InventoryTransaction
	
	query := r.db.Preload("Product")
	
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}
	
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	
	// Aggregate by product
	productSummary := make(map[uint]map[string]int)
	
	for _, tx := range transactions {
		if _, exists := productSummary[tx.ProductID]; !exists {
			productSummary[tx.ProductID] = map[string]int{
				"received":   0,
				"issued":     0,
				"adjusted":   0,
				"transferred": 0,
			}
		}
		
		switch tx.Type {
		case "receive":
			productSummary[tx.ProductID]["received"] += tx.Quantity
		case "issue":
			productSummary[tx.ProductID]["issued"] += tx.Quantity
		case "adjustment":
			productSummary[tx.ProductID]["adjusted"] += tx.Quantity
		case "transfer":
			productSummary[tx.ProductID]["transferred"] += tx.Quantity
		}
	}
	
	// Convert to slice for response
	result := make([]map[string]interface{}, 0, len(productSummary))
	
	for productID, summary := range productSummary {
		var product models.Product
		if err := r.db.First(&product, productID).Error; err != nil {
			continue
		}
		
		result = append(result, map[string]interface{}{
			"product_id":   productID,
			"product_name": product.Name,
			"product_sku":  product.SKU,
			"received":     summary["received"],
			"issued":       summary["issued"],
			"adjusted":     summary["adjusted"],
			"transferred":  summary["transferred"],
			"net_change":   summary["received"] - summary["issued"] + summary["adjusted"],
		})
	}
	
	return result, nil
}