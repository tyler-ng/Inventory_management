package handlers

import (
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// RegisterPublicRoutes registers all routes that don't require authentication
func RegisterPublicRoutes(router *mux.Router, db *gorm.DB) {
	// Auth handler for login/register
	authHandler := NewAuthHandler(db)
	router.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
}

// RegisterProtectedRoutes registers all routes that require authentication
func RegisterProtectedRoutes(router *mux.Router, db *gorm.DB) {
	// Products
	productHandler := NewProductHandler(db)
	router.HandleFunc("/products", productHandler.GetProducts).Methods("GET")
	router.HandleFunc("/products", productHandler.CreateProduct).Methods("POST")
	router.HandleFunc("/products/{id:[0-9]+}", productHandler.GetProduct).Methods("GET")
	router.HandleFunc("/products/{id:[0-9]+}", productHandler.UpdateProduct).Methods("PUT")
	router.HandleFunc("/products/{id:[0-9]+}", productHandler.DeleteProduct).Methods("DELETE")
	router.HandleFunc("/products/sku/{sku}", productHandler.GetProductBySKU).Methods("GET")
	router.HandleFunc("/products/{id:[0-9]+}/categories", productHandler.GetProductCategories).Methods("GET")
	router.HandleFunc("/products/low-stock", productHandler.GetLowStockProducts).Methods("GET")
	router.HandleFunc("/products/warehouse/{warehouseId:[0-9]+}", productHandler.GetProductsByWarehouse).Methods("GET")
	
	// Categories
	categoryHandler := NewCategoryHandler(db)
	router.HandleFunc("/categories", categoryHandler.GetCategories).Methods("GET")
	router.HandleFunc("/categories", categoryHandler.CreateCategory).Methods("POST")
	router.HandleFunc("/categories/{id:[0-9]+}", categoryHandler.GetCategory).Methods("GET")
	router.HandleFunc("/categories/{id:[0-9]+}", categoryHandler.UpdateCategory).Methods("PUT")
	router.HandleFunc("/categories/{id:[0-9]+}", categoryHandler.DeleteCategory).Methods("DELETE")
	router.HandleFunc("/categories/{id:[0-9]+}/products", categoryHandler.GetCategoryProducts).Methods("GET")
	router.HandleFunc("/categories/{id:[0-9]+}/subcategories", categoryHandler.GetSubcategories).Methods("GET")
	
	// Suppliers
	supplierHandler := NewSupplierHandler(db)
	router.HandleFunc("/suppliers", supplierHandler.GetSuppliers).Methods("GET")
	router.HandleFunc("/suppliers", supplierHandler.CreateSupplier).Methods("POST")
	router.HandleFunc("/suppliers/{id:[0-9]+}", supplierHandler.GetSupplier).Methods("GET")
	router.HandleFunc("/suppliers/{id:[0-9]+}", supplierHandler.UpdateSupplier).Methods("PUT")
	router.HandleFunc("/suppliers/{id:[0-9]+}", supplierHandler.DeleteSupplier).Methods("DELETE")
	router.HandleFunc("/suppliers/{id:[0-9]+}/products", supplierHandler.GetSupplierProducts).Methods("GET")
	
	// Warehouses
	warehouseHandler := NewWarehouseHandler(db)
	router.HandleFunc("/warehouses", warehouseHandler.GetWarehouses).Methods("GET")
	router.HandleFunc("/warehouses", warehouseHandler.CreateWarehouse).Methods("POST")
	router.HandleFunc("/warehouses/{id:[0-9]+}", warehouseHandler.GetWarehouse).Methods("GET")
	router.HandleFunc("/warehouses/{id:[0-9]+}", warehouseHandler.UpdateWarehouse).Methods("PUT")
	router.HandleFunc("/warehouses/{id:[0-9]+}", warehouseHandler.DeleteWarehouse).Methods("DELETE")
	router.HandleFunc("/warehouses/{id:[0-9]+}/locations", warehouseHandler.GetWarehouseLocations).Methods("GET")
	router.HandleFunc("/warehouses/{id:[0-9]+}/products", warehouseHandler.GetWarehouseProducts).Methods("GET")
	
	// Warehouse Locations
	router.HandleFunc("/locations", warehouseHandler.GetAllLocations).Methods("GET")
	router.HandleFunc("/locations", warehouseHandler.CreateLocation).Methods("POST")
	router.HandleFunc("/locations/{id:[0-9]+}", warehouseHandler.GetLocation).Methods("GET")
	router.HandleFunc("/locations/{id:[0-9]+}", warehouseHandler.UpdateLocation).Methods("PUT")
	router.HandleFunc("/locations/{id:[0-9]+}", warehouseHandler.DeleteLocation).Methods("DELETE")
	
	// Inventory Transactions
	transactionHandler := NewTransactionHandler(db)
	router.HandleFunc("/transactions", transactionHandler.GetTransactions).Methods("GET")
	router.HandleFunc("/transactions", transactionHandler.CreateTransaction).Methods("POST")
	router.HandleFunc("/transactions/{id:[0-9]+}", transactionHandler.GetTransaction).Methods("GET")
	router.HandleFunc("/transactions/product/{productId:[0-9]+}", transactionHandler.GetProductTransactions).Methods("GET")
	router.HandleFunc("/transactions/receive", transactionHandler.CreateReceiveTransaction).Methods("POST")
	router.HandleFunc("/transactions/issue", transactionHandler.CreateIssueTransaction).Methods("POST")
	router.HandleFunc("/transactions/transfer", transactionHandler.CreateTransferTransaction).Methods("POST")
	
	// Purchase Orders
	purchaseHandler := NewPurchaseOrderHandler(db)
	router.HandleFunc("/purchase-orders", purchaseHandler.GetPurchaseOrders).Methods("GET")
	router.HandleFunc("/purchase-orders", purchaseHandler.CreatePurchaseOrder).Methods("POST")
	router.HandleFunc("/purchase-orders/{id:[0-9]+}", purchaseHandler.GetPurchaseOrder).Methods("GET")
	router.HandleFunc("/purchase-orders/{id:[0-9]+}", purchaseHandler.UpdatePurchaseOrder).Methods("PUT")
	router.HandleFunc("/purchase-orders/{id:[0-9]+}", purchaseHandler.DeletePurchaseOrder).Methods("DELETE")
	router.HandleFunc("/purchase-orders/{id:[0-9]+}/items", purchaseHandler.GetPurchaseOrderItems).Methods("GET")
	router.HandleFunc("/purchase-orders/{id:[0-9]+}/items", purchaseHandler.AddPurchaseOrderItem).Methods("POST")
	router.HandleFunc("/purchase-orders/{id:[0-9]+}/receive", purchaseHandler.ReceivePurchaseOrder).Methods("POST")
	
	// Sales Orders
	salesHandler := NewSalesOrderHandler(db)
	router.HandleFunc("/sales-orders", salesHandler.GetSalesOrders).Methods("GET")
	router.HandleFunc("/sales-orders", salesHandler.CreateSalesOrder).Methods("POST")
	router.HandleFunc("/sales-orders/{id:[0-9]+}", salesHandler.GetSalesOrder).Methods("GET")
	router.HandleFunc("/sales-orders/{id:[0-9]+}", salesHandler.UpdateSalesOrder).Methods("PUT")
	router.HandleFunc("/sales-orders/{id:[0-9]+}", salesHandler.DeleteSalesOrder).Methods("DELETE")
	router.HandleFunc("/sales-orders/{id:[0-9]+}/items", salesHandler.GetSalesOrderItems).Methods("GET")
	router.HandleFunc("/sales-orders/{id:[0-9]+}/items", salesHandler.AddSalesOrderItem).Methods("POST")
	router.HandleFunc("/sales-orders/{id:[0-9]+}/fulfill", salesHandler.FulfillSalesOrder).Methods("POST")
	
	// Customers
	customerHandler := NewCustomerHandler(db)
	router.HandleFunc("/customers", customerHandler.GetCustomers).Methods("GET")
	router.HandleFunc("/customers", customerHandler.CreateCustomer).Methods("POST")
	router.HandleFunc("/customers/{id:[0-9]+}", customerHandler.GetCustomer).Methods("GET")
	router.HandleFunc("/customers/{id:[0-9]+}", customerHandler.UpdateCustomer).Methods("PUT")
	router.HandleFunc("/customers/{id:[0-9]+}", customerHandler.DeleteCustomer).Methods("DELETE")
	router.HandleFunc("/customers/{id:[0-9]+}/sales-orders", customerHandler.GetCustomerSalesOrders).Methods("GET")
	
	// Users
	userHandler := NewUserHandler(db)
	router.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	router.HandleFunc("/users/{id:[0-9]+}", userHandler.GetUser).Methods("GET")
	router.HandleFunc("/users/{id:[0-9]+}", userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id:[0-9]+}", userHandler.DeleteUser).Methods("DELETE")
	router.HandleFunc("/users/current", userHandler.GetCurrentUser).Methods("GET")
	router.HandleFunc("/users/change-password", userHandler.ChangePassword).Methods("POST")
	
	// Reports
	reportHandler := NewReportHandler(db)
	router.HandleFunc("/reports/inventory-value", reportHandler.GetInventoryValueReport).Methods("GET")
	router.HandleFunc("/reports/product-movement", reportHandler.GetProductMovementReport).Methods("GET")
	router.HandleFunc("/reports/low-stock", reportHandler.GetLowStockReport).Methods("GET")
	router.HandleFunc("/reports/sales", reportHandler.GetSalesReport).Methods("GET")
	router.HandleFunc("/reports/purchases", reportHandler.GetPurchasesReport).Methods("GET")
}