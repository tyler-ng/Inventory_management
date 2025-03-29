package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/gorm"
)

// ReportHandler handles HTTP requests for generating reports
type ReportHandler struct {
	db *gorm.DB
}

// NewReportHandler creates a new report handler
func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

// GetInventoryValueReport generates a report of current inventory value
func (h *ReportHandler) GetInventoryValueReport(w http.ResponseWriter, r *http.Request) {
	type ProductValue struct {
		ID          uint    `json:"id"`
		SKU         string  `json:"sku"`
		Name        string  `json:"name"`
		Category    string  `json:"category"`
		Quantity    int     `json:"quantity"`
		CostPrice   float64 `json:"cost_price"`
		TotalValue  float64 `json:"total_value"`
		LastUpdated time.Time `json:"last_updated"`
	}

	var products []ProductValue
	
	// Get filter parameters
	category := r.URL.Query().Get("category")
	warehouseID := r.URL.Query().Get("warehouse_id")

	// Build query
	query := h.db.Table("products").
		Select("products.id, products.sku, products.name, products.category, products.quantity, products.cost_price, (products.quantity * products.cost_price) as total_value, products.updated_at as last_updated").
		Where("products.status = ?", "active")
		
	// Apply filters
	if category != "" {
		query = query.Where("products.category = ?", category)
	}
	
	if warehouseID != "" {
		query = query.Joins("JOIN product_warehouse ON products.id = product_warehouse.product_id").
			Where("product_warehouse.warehouse_id = ?", warehouseID)
	}
	
	// Execute query
	if err := query.Find(&products).Error; err != nil {
		http.Error(w, "Failed to generate inventory value report: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Calculate total inventory value
	var totalValue float64
	for _, p := range products {
		totalValue += p.TotalValue
	}
	
	// Prepare report response
	report := map[string]interface{}{
		"generated_at": time.Now(),
		"total_items":  len(products),
		"total_value":  totalValue,
		"items":        products,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetLowStockReport generates a report of products with stock below reorder level
func (h *ReportHandler) GetLowStockReport(w http.ResponseWriter, r *http.Request) {
	type LowStockProduct struct {
		ID           uint    `json:"id"`
		SKU          string  `json:"sku"`
		Name         string  `json:"name"`
		Category     string  `json:"category"`
		Quantity     int     `json:"quantity"`
		ReorderLevel int     `json:"reorder_level"`
		Shortage     int     `json:"shortage"`
		Supplier     string  `json:"supplier"`
	}

	var products []LowStockProduct
	
	// Build query
	query := h.db.Table("products").
		Select("products.id, products.sku, products.name, products.category, products.quantity, products.reorder_level, (products.reorder_level - products.quantity) as shortage, suppliers.name as supplier").
		Joins("LEFT JOIN product_supplier ON products.id = product_supplier.product_id").
		Joins("LEFT JOIN suppliers ON product_supplier.supplier_id = suppliers.id").
		Where("products.status = ? AND products.quantity <= products.reorder_level", "active").
		Group("products.id, suppliers.name").
		Order("shortage DESC")
	
	// Execute query
	if err := query.Find(&products).Error; err != nil {
		http.Error(w, "Failed to generate low stock report: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Prepare report response
	report := map[string]interface{}{
		"generated_at": time.Now(),
		"total_items":  len(products),
		"items":        products,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetProductMovementReport generates a report of product movements over a period
func (h *ReportHandler) GetProductMovementReport(w http.ResponseWriter, r *http.Request) {
	// Parse date range parameters
	startDate := time.Now().AddDate(0, -1, 0) // Default to last month
	endDate := time.Now()
	
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsedDate
		}
	}
	
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsedDate.Add(24 * time.Hour) // Include the end date fully
		}
	}
	
	// Get optional product filter
	productID := r.URL.Query().Get("product_id")
	
	type ProductMovement struct {
		ProductID    uint    `json:"product_id"`
		ProductSKU   string  `json:"product_sku"`
		ProductName  string  `json:"product_name"`
		Received     int     `json:"received"`
		Issued       int     `json:"issued"`
		Adjusted     int     `json:"adjusted"`
		NetChange    int     `json:"net_change"`
	}
	
	var movements []ProductMovement
	
	// Build base query
	query := h.db.Table("products").
		Select(`
			products.id as product_id, 
			products.sku as product_sku, 
			products.name as product_name,
			COALESCE(SUM(CASE WHEN inventory_transactions.type = 'receive' THEN inventory_transactions.quantity ELSE 0 END), 0) as received,
			COALESCE(SUM(CASE WHEN inventory_transactions.type = 'issue' THEN inventory_transactions.quantity ELSE 0 END), 0) as issued,
			COALESCE(SUM(CASE WHEN inventory_transactions.type = 'adjustment' THEN inventory_transactions.quantity ELSE 0 END), 0) as adjusted,
			COALESCE(SUM(CASE 
				WHEN inventory_transactions.type = 'receive' THEN inventory_transactions.quantity 
				WHEN inventory_transactions.type = 'issue' THEN -inventory_transactions.quantity 
				ELSE inventory_transactions.quantity END), 0) as net_change
		`).
		Joins("LEFT JOIN inventory_transactions ON products.id = inventory_transactions.product_id AND inventory_transactions.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("products.id, products.sku, products.name").
		Order("net_change DESC")
	
	// Apply product filter if provided
	if productID != "" {
		query = query.Where("products.id = ?", productID)
	}
	
	// Execute query
	if err := query.Find(&movements).Error; err != nil {
		http.Error(w, "Failed to generate product movement report: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Prepare report response
	report := map[string]interface{}{
		"generated_at": time.Now(),
		"start_date":   startDate.Format("2006-01-02"),
		"end_date":     endDate.Add(-24 * time.Hour).Format("2006-01-02"),
		"total_products": len(movements),
		"movements":    movements,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetSalesReport generates a sales report over a period
func (h *ReportHandler) GetSalesReport(w http.ResponseWriter, r *http.Request) {
	// Parse date range parameters
	startDate := time.Now().AddDate(0, -1, 0) // Default to last month
	endDate := time.Now()
	
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsedDate
		}
	}
	
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsedDate.Add(24 * time.Hour) // Include the end date fully
		}
	}
	
	// Get optional filters
	customerID := r.URL.Query().Get("customer_id")
	productID := r.URL.Query().Get("product_id")
	
	// Summary statistics
	var totalSales float64
	var totalOrders int64
	var avgOrderValue float64
	
	// Get total sales amount and order count
	query := h.db.Model(&models.SalesOrder{}).
		Where("order_date BETWEEN ? AND ? AND status NOT IN ('draft', 'cancelled')", startDate, endDate)
	
	if customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}
	
	if err := query.Count(&totalOrders).Error; err != nil {
		http.Error(w, "Failed to count orders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	if err := query.Select("COALESCE(SUM(total_amount), 0)").Scan(&totalSales).Error; err != nil {
		http.Error(w, "Failed to calculate total sales: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	if totalOrders > 0 {
		avgOrderValue = totalSales / float64(totalOrders)
	}
	
	// Get sales by product
	type ProductSales struct {
		ProductID   uint    `json:"product_id"`
		ProductSKU  string  `json:"product_sku"`
		ProductName string  `json:"product_name"`
		Quantity    int     `json:"quantity"`
		Revenue     float64 `json:"revenue"`
	}
	
	var productSales []ProductSales
	
	productQuery := h.db.Table("sales_order_items").
		Select(`
			products.id as product_id,
			products.sku as product_sku,
			products.name as product_name,
			SUM(sales_order_items.quantity) as quantity,
			SUM(sales_order_items.total_price) as revenue
		`).
		Joins("JOIN products ON sales_order_items.product_id = products.id").
		Joins("JOIN sales_orders ON sales_order_items.sales_order_id = sales_orders.id").
		Where("sales_orders.order_date BETWEEN ? AND ? AND sales_orders.status NOT IN ('draft', 'cancelled')", startDate, endDate).
		Group("products.id, products.sku, products.name").
		Order("revenue DESC")
	
	if customerID != "" {
		productQuery = productQuery.Where("sales_orders.customer_id = ?", customerID)
	}
	
	if productID != "" {
		productQuery = productQuery.Where("products.id = ?", productID)
	}
	
	if err := productQuery.Find(&productSales).Error; err != nil {
		http.Error(w, "Failed to retrieve product sales: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get sales by customer
	type CustomerSales struct {
		CustomerID   uint    `json:"customer_id"`
		CustomerName string  `json:"customer_name"`
		OrderCount   int     `json:"order_count"`
		Revenue      float64 `json:"revenue"`
	}
	
	var customerSales []CustomerSales
	
	customerQuery := h.db.Table("sales_orders").
		Select(`
			customers.id as customer_id,
			customers.name as customer_name,
			COUNT(sales_orders.id) as order_count,
			SUM(sales_orders.total_amount) as revenue
		`).
		Joins("JOIN customers ON sales_orders.customer_id = customers.id").
		Where("sales_orders.order_date BETWEEN ? AND ? AND sales_orders.status NOT IN ('draft', 'cancelled')", startDate, endDate).
		Group("customers.id, customers.name").
		Order("revenue DESC")
	
	if customerID != "" {
		customerQuery = customerQuery.Where("customers.id = ?", customerID)
	}
	
	if productID != "" {
		customerQuery = customerQuery.
			Joins("JOIN sales_order_items ON sales_orders.id = sales_order_items.sales_order_id").
			Where("sales_order_items.product_id = ?", productID)
	}
	
	if err := customerQuery.Find(&customerSales).Error; err != nil {
		http.Error(w, "Failed to retrieve customer sales: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Prepare report response
	report := map[string]interface{}{
		"generated_at":    time.Now(),
		"start_date":      startDate.Format("2006-01-02"),
		"end_date":        endDate.Add(-24 * time.Hour).Format("2006-01-02"),
		"total_sales":     totalSales,
		"total_orders":    totalOrders,
		"avg_order_value": avgOrderValue,
		"product_sales":   productSales,
		"customer_sales":  customerSales,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetPurchasesReport generates a purchases report over a period
func (h *ReportHandler) GetPurchasesReport(w http.ResponseWriter, r *http.Request) {
	// Parse date range parameters
	startDate := time.Now().AddDate(0, -1, 0) // Default to last month
	endDate := time.Now()
	
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsedDate
		}
	}
	
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsedDate.Add(24 * time.Hour) // Include the end date fully
		}
	}
	
	// Get optional filters
	supplierID := r.URL.Query().Get("supplier_id")
	productID := r.URL.Query().Get("product_id")
	
	// Summary statistics
	var totalPurchases float64
	var totalOrders int64
	var avgOrderValue float64
	
	// Get total purchases amount and order count
	query := h.db.Model(&models.PurchaseOrder{}).
		Where("order_date BETWEEN ? AND ? AND status NOT IN ('draft', 'cancelled')", startDate, endDate)
	
	if supplierID != "" {
		query = query.Where("supplier_id = ?", supplierID)
	}
	
	if err := query.Count(&totalOrders).Error; err != nil {
		http.Error(w, "Failed to count orders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	if err := query.Select("COALESCE(SUM(total_amount), 0)").Scan(&totalPurchases).Error; err != nil {
		http.Error(w, "Failed to calculate total purchases: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	if totalOrders > 0 {
		avgOrderValue = totalPurchases / float64(totalOrders)
	}
	
	// Get purchases by product
	type ProductPurchases struct {
		ProductID   uint    `json:"product_id"`
		ProductSKU  string  `json:"product_sku"`
		ProductName string  `json:"product_name"`
		Quantity    int     `json:"quantity"`
		Cost        float64 `json:"cost"`
	}
	
	var productPurchases []ProductPurchases
	
	productQuery := h.db.Table("purchase_order_items").
		Select(`
			products.id as product_id,
			products.sku as product_sku,
			products.name as product_name,
			SUM(purchase_order_items.quantity) as quantity,
			SUM(purchase_order_items.total_price) as cost
		`).
		Joins("JOIN products ON purchase_order_items.product_id = products.id").
		Joins("JOIN purchase_orders ON purchase_order_items.purchase_order_id = purchase_orders.id").
		Where("purchase_orders.order_date BETWEEN ? AND ? AND purchase_orders.status NOT IN ('draft', 'cancelled')", startDate, endDate).
		Group("products.id, products.sku, products.name").
		Order("cost DESC")
	
	if supplierID != "" {
		productQuery = productQuery.Where("purchase_orders.supplier_id = ?", supplierID)
	}
	
	if productID != "" {
		productQuery = productQuery.Where("products.id = ?", productID)
	}
	
	if err := productQuery.Find(&productPurchases).Error; err != nil {
		http.Error(w, "Failed to retrieve product purchases: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get purchases by supplier
	type SupplierPurchases struct {
		SupplierID   uint    `json:"supplier_id"`
		SupplierName string  `json:"supplier_name"`
		OrderCount   int     `json:"order_count"`
		Cost         float64 `json:"cost"`
	}
	
	var supplierPurchases []SupplierPurchases
	
	supplierQuery := h.db.Table("purchase_orders").
		Select(`
			suppliers.id as supplier_id,
			suppliers.name as supplier_name,
			COUNT(purchase_orders.id) as order_count,
			SUM(purchase_orders.total_amount) as cost
		`).
		Joins("JOIN suppliers ON purchase_orders.supplier_id = suppliers.id").
		Where("purchase_orders.order_date BETWEEN ? AND ? AND purchase_orders.status NOT IN ('draft', 'cancelled')", startDate, endDate).
		Group("suppliers.id, suppliers.name").
		Order("cost DESC")
	
	if supplierID != "" {
		supplierQuery = supplierQuery.Where("suppliers.id = ?", supplierID)
	}
	
	if productID != "" {
		supplierQuery = supplierQuery.
			Joins("JOIN purchase_order_items ON purchase_orders.id = purchase_order_items.purchase_order_id").
			Where("purchase_order_items.product_id = ?", productID)
	}
	
	if err := supplierQuery.Find(&supplierPurchases).Error; err != nil {
		http.Error(w, "Failed to retrieve supplier purchases: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Prepare report response
	report := map[string]interface{}{
		"generated_at":    time.Now(),
		"start_date":      startDate.Format("2006-01-02"),
		"end_date":        endDate.Add(-24 * time.Hour).Format("2006-01-02"),
		"total_purchases": totalPurchases,
		"total_orders":    totalOrders,
		"avg_order_value": avgOrderValue,
		"product_purchases": productPurchases,
		"supplier_purchases": supplierPurchases,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}