-- Drop triggers
DROP TRIGGER IF EXISTS update_user_updated_at ON users;
DROP TRIGGER IF EXISTS update_product_updated_at ON products;
DROP TRIGGER IF EXISTS update_category_updated_at ON categories;
DROP TRIGGER IF EXISTS update_supplier_updated_at ON suppliers;
DROP TRIGGER IF EXISTS update_warehouse_updated_at ON warehouses;
DROP TRIGGER IF EXISTS update_warehouse_location_updated_at ON warehouse_locations;
DROP TRIGGER IF EXISTS update_purchase_order_updated_at ON purchase_orders;
DROP TRIGGER IF EXISTS update_purchase_order_item_updated_at ON purchase_order_items;
DROP TRIGGER IF EXISTS update_customer_updated_at ON customers;
DROP TRIGGER IF EXISTS update_sales_order_updated_at ON sales_orders;
DROP TRIGGER IF EXISTS update_sales_order_item_updated_at ON sales_order_items;
DROP TRIGGER IF EXISTS update_product_supplier_updated_at ON product_supplier;
DROP TRIGGER IF EXISTS update_product_warehouse_updated_at ON product_warehouse;
DROP TRIGGER IF EXISTS update_product_variant_updated_at ON product_variants;
DROP TRIGGER IF EXISTS update_product_bundle_updated_at ON product_bundles;

-- Drop the update timestamp function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_product_sku;
DROP INDEX IF EXISTS idx_category_parent;
DROP INDEX IF EXISTS idx_transaction_product;
DROP INDEX IF EXISTS idx_transaction_warehouse;
DROP INDEX IF EXISTS idx_transaction_type;
DROP INDEX IF EXISTS idx_transaction_date;
DROP INDEX IF EXISTS idx_purchase_order_supplier;
DROP INDEX IF EXISTS idx_purchase_order_status;
DROP INDEX IF EXISTS idx_sales_order_customer;
DROP INDEX IF EXISTS idx_sales_order_status;
DROP INDEX IF EXISTS idx_audit_log_user;
DROP INDEX IF EXISTS idx_audit_log_entity;
DROP INDEX IF EXISTS idx_product_warehouse_location;

-- Drop tables in reverse order of creation (to avoid foreign key constraints)
DROP TABLE IF EXISTS product_category;
DROP TABLE IF EXISTS product_bundles;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS product_attachments;
DROP TABLE IF EXISTS product_warehouse;
DROP TABLE IF EXISTS product_supplier;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS sales_order_items;
DROP TABLE IF EXISTS sales_orders;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS purchase_order_items;
DROP TABLE IF EXISTS purchase_orders;
DROP TABLE IF EXISTS inventory_transactions;
DROP TABLE IF EXISTS warehouse_locations;
DROP TABLE IF EXISTS warehouses;
DROP TABLE IF EXISTS suppliers;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;