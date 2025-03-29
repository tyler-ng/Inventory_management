package repository

import (
	"time"

	"github.com/yourusername/inventory-management-system/internal/models"
)

// ProductRepository defines the interface for product database operations
type IProductRepository interface {
	GetAll(params map[string]interface{}) ([]models.Product, error)
	GetByID(id uint) (*models.Product, error)
	GetBySKU(sku string) (*models.Product, error)
	Create(product *models.Product) error
	Update(product *models.Product) error
	Delete(id uint) error
	GetLowStock() ([]models.Product, error)
	UpdateQuantity(id uint, quantity int) error
	GetProductsByWarehouse(warehouseID uint) ([]models.ProductWarehouse, error)
	GetProductVariants(productID uint) ([]models.ProductVariant, error)
	GetProductCategories(productID uint) ([]models.Category, error)
	AddProductCategory(productID, categoryID uint) error
	RemoveProductCategory(productID, categoryID uint) error
}

// CategoryRepository defines the interface for category database operations
type ICategoryRepository interface {
	GetAll() ([]models.Category, error)
	GetByID(id uint) (*models.Category, error)
	Create(category *models.Category) error
	Update(category *models.Category) error
	Delete(id uint) error
	GetSubcategories(parentID uint) ([]models.Category, error)
	GetCategoryProducts(categoryID uint) ([]models.Product, error)
}

// SupplierRepository defines the interface for supplier database operations
type ISupplierRepository interface {
	GetAll() ([]models.Supplier, error)
	GetByID(id uint) (*models.Supplier, error)
	Create(supplier *models.Supplier) error
	Update(supplier *models.Supplier) error
	Delete(id uint) error
	GetSupplierProducts(supplierID uint) ([]models.Product, error)
}

// WarehouseRepository defines the interface for warehouse database operations
type IWarehouseRepository interface {
	GetAll() ([]models.Warehouse, error)
	GetByID(id uint) (*models.Warehouse, error)
	Create(warehouse *models.Warehouse) error
	Update(warehouse *models.Warehouse) error
	Delete(id uint) error
	GetWarehouseLocations(warehouseID uint) ([]models.WarehouseLocation, error)
	GetLocation(id uint) (*models.WarehouseLocation, error)
	CreateLocation(location *models.WarehouseLocation) error
	UpdateLocation(location *models.WarehouseLocation) error
	DeleteLocation(id uint) error
}

// TransactionRepository defines the interface for inventory transaction database operations
type ITransactionRepository interface {
	GetAll(params map[string]interface{}) ([]models.InventoryTransaction, error)
	GetByID(id uint) (*models.InventoryTransaction, error)
	Create(transaction *models.InventoryTransaction) error
	GetProductTransactions(productID uint, startDate, endDate time.Time) ([]models.InventoryTransaction, error)
	GetProductMovementSummary(startDate, endDate time.Time) ([]map[string]interface{}, error)
}

// PurchaseOrderRepository defines the interface for purchase order database operations
type IPurchaseOrderRepository interface {
	GetAll(params map[string]interface{}) ([]models.PurchaseOrder, error)
	GetByID(id uint) (*models.PurchaseOrder, error)
	Create(order *models.PurchaseOrder) error
	Update(order *models.PurchaseOrder) error
	Delete(id uint) error
	GetItems(orderID uint) ([]models.PurchaseOrderItem, error)
	AddItem(item *models.PurchaseOrderItem) error
	UpdateItem(item *models.PurchaseOrderItem) error
	DeleteItem(id uint) error
	ReceiveOrder(orderID uint, receivedItems map[uint]int, notes string, userID uint) error
}

// SalesOrderRepository defines the interface for sales order database operations
type ISalesOrderRepository interface {
	GetAll(params map[string]interface{}) ([]models.SalesOrder, error)
	GetByID(id uint) (*models.SalesOrder, error)
	Create(order *models.SalesOrder) error
	Update(order *models.SalesOrder) error
	Delete(id uint) error
	GetItems(orderID uint) ([]models.SalesOrderItem, error)
	AddItem(item *models.SalesOrderItem) error
	UpdateItem(item *models.SalesOrderItem) error
	DeleteItem(id uint) error
	FulfillOrder(orderID uint, fulfilledItems map[uint]int, notes string, userID uint) error
}

// CustomerRepository defines the interface for customer database operations
type ICustomerRepository interface {
	GetAll() ([]models.Customer, error)
	GetByID(id uint) (*models.Customer, error)
	Create(customer *models.Customer) error
	Update(customer *models.Customer) error
	Delete(id uint) error
	GetCustomerOrders(customerID uint) ([]models.SalesOrder, error)
}

// UserRepository defines the interface for user database operations
type IUserRepository interface {
	GetAll() ([]models.User, error)
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id uint) error
	ChangePassword(id uint, newPassword string) error
}

// ReportRepository defines the interface for report generation operations
type IReportRepository interface {
	GetInventoryValueReport() (map[string]interface{}, error)
	GetLowStockReport() ([]models.Product, error)
	GetProductMovementReport(startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetSalesReport(startDate, endDate time.Time) (map[string]interface{}, error)
	GetPurchasesReport(startDate, endDate time.Time) (map[string]interface{}, error)
}

// AuditLogRepository defines the interface for audit log operations
type IAuditLogRepository interface {
	CreateLog(log *models.AuditLog) error
	GetLogs(params map[string]interface{}) ([]models.AuditLog, error)
	GetUserLogs(userID uint) ([]models.AuditLog, error)
	GetEntityLogs(entityType string, entityID uint) ([]models.AuditLog, error)
}