-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create categories table
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    parent_id INTEGER NULL REFERENCES categories(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create products table
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    quantity INTEGER NOT NULL DEFAULT 0,
    reorder_level INTEGER DEFAULT 5,
    price DECIMAL(10, 2) NOT NULL,
    cost_price DECIMAL(10, 2),
    weight DECIMAL(10, 2),
    dimensions VARCHAR(50),
    image_url VARCHAR(255),
    barcode VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create suppliers table
CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    contact_person VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    tax_id VARCHAR(50),
    payment_terms VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create warehouses table
CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(100),
    address TEXT,
    manager VARCHAR(100),
    phone VARCHAR(20),
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create warehouse_locations table
CREATE TABLE warehouse_locations (
    id SERIAL PRIMARY KEY,
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    zone VARCHAR(20),
    aisle VARCHAR(20),
    rack VARCHAR(20),
    shelf VARCHAR(20),
    bin VARCHAR(20),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create inventory_transactions table
CREATE TABLE inventory_transactions (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    source_location_id INTEGER REFERENCES warehouse_locations(id),
    destination_location_id INTEGER REFERENCES warehouse_locations(id),
    type VARCHAR(20) NOT NULL, -- receive, issue, transfer, adjustment
    quantity INTEGER NOT NULL,
    reference_number VARCHAR(50),
    user_id INTEGER NOT NULL REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create purchase_orders table
CREATE TABLE purchase_orders (
    id SERIAL PRIMARY KEY,
    po_number VARCHAR(20) UNIQUE NOT NULL,
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    order_date DATE NOT NULL,
    expected_date DATE,
    status VARCHAR(20) DEFAULT 'draft',
    total_amount DECIMAL(10, 2) DEFAULT 0,
    payment_terms VARCHAR(100),
    shipping_terms VARCHAR(100),
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create purchase_order_items table
CREATE TABLE purchase_order_items (
    id SERIAL PRIMARY KEY,
    purchase_order_id INTEGER NOT NULL REFERENCES purchase_orders(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create customers table
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    contact_person VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    tax_id VARCHAR(50),
    payment_terms VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create sales_orders table
CREATE TABLE sales_orders (
    id SERIAL PRIMARY KEY,
    so_number VARCHAR(20) UNIQUE NOT NULL,
    customer_id INTEGER NOT NULL REFERENCES customers(id),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    order_date DATE NOT NULL,
    shipping_date DATE,
    status VARCHAR(20) DEFAULT 'draft',
    subtotal DECIMAL(10, 2) DEFAULT 0,
    tax DECIMAL(10, 2) DEFAULT 0,
    shipping_cost DECIMAL(10, 2) DEFAULT 0,
    total_amount DECIMAL(10, 2) DEFAULT 0,
    payment_status VARCHAR(20) DEFAULT 'unpaid',
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create sales_order_items table
CREATE TABLE sales_order_items (
    id SERIAL PRIMARY KEY,
    sales_order_id INTEGER NOT NULL REFERENCES sales_orders(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    discount DECIMAL(10, 2) DEFAULT 0,
    total_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create audit_logs table
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create product_supplier table (many-to-many)
CREATE TABLE product_supplier (
    product_id INTEGER NOT NULL REFERENCES products(id),
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id),
    unit_cost DECIMAL(10, 2),
    min_order_quantity INTEGER DEFAULT 1,
    lead_time_days INTEGER,
    supplier_sku VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (product_id, supplier_id)
);

-- Create product_warehouse table (many-to-many with location)
CREATE TABLE product_warehouse (
    product_id INTEGER NOT NULL REFERENCES products(id),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    location_id INTEGER REFERENCES warehouse_locations(id),
    quantity INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (product_id, warehouse_id)
);

-- Create product_attachment table
CREATE TABLE product_attachments (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    file_type VARCHAR(50),
    file_size INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create product_variant table
CREATE TABLE product_variants (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    sku VARCHAR(50) UNIQUE NOT NULL,
    attributes JSONB,
    quantity INTEGER NOT NULL DEFAULT 0,
    price DECIMAL(10, 2) NOT NULL,
    cost_price DECIMAL(10, 2),
    barcode VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create product_bundle table (for product kits)
CREATE TABLE product_bundles (
    id SERIAL PRIMARY KEY,
    parent_product_id INTEGER NOT NULL REFERENCES products(id),
    child_product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create product_category table (many-to-many)
CREATE TABLE product_category (
    product_id INTEGER NOT NULL REFERENCES products(id),
    category_id INTEGER NOT NULL REFERENCES categories(id),
    PRIMARY KEY (product_id, category_id)
);

-- Add indexes for better performance
CREATE INDEX idx_product_sku ON products(sku);
CREATE INDEX idx_category_parent ON categories(parent_id);
CREATE INDEX idx_transaction_product ON inventory_transactions(product_id);
CREATE INDEX idx_transaction_warehouse ON inventory_transactions(warehouse_id);
CREATE INDEX idx_transaction_type ON inventory_transactions(type);
CREATE INDEX idx_transaction_date ON inventory_transactions(created_at);
CREATE INDEX idx_purchase_order_supplier ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_order_status ON purchase_orders(status);
CREATE INDEX idx_sales_order_customer ON sales_orders(customer_id);
CREATE INDEX idx_sales_order_status ON sales_orders(status);
CREATE INDEX idx_audit_log_user ON audit_logs(user_id);
CREATE INDEX idx_audit_log_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_product_warehouse_location ON product_warehouse(location_id);

-- Create a function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at columns
CREATE TRIGGER update_user_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_product_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_category_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_supplier_updated_at
    BEFORE UPDATE ON suppliers
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_warehouse_updated_at
    BEFORE UPDATE ON warehouses
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_warehouse_location_updated_at
    BEFORE UPDATE ON warehouse_locations
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_purchase_order_updated_at
    BEFORE UPDATE ON purchase_orders
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_purchase_order_item_updated_at
    BEFORE UPDATE ON purchase_order_items
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_customer_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_sales_order_updated_at
    BEFORE UPDATE ON sales_orders
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_sales_order_item_updated_at
    BEFORE UPDATE ON sales_order_items
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_product_supplier_updated_at
    BEFORE UPDATE ON product_supplier
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_product_warehouse_updated_at
    BEFORE UPDATE ON product_warehouse
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_product_variant_updated_at
    BEFORE UPDATE ON product_variants
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_product_bundle_updated_at
    BEFORE UPDATE ON product_bundles
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Insert default admin user (password: admin123)
INSERT INTO users (username, password_hash, email, full_name, role)
VALUES ('admin', '$2a$10$XHRMrAU3MqgXJp3JsHYM0.Cz/lhLZFl7k8bW8/xqyTGcQOW3e87E.', 'admin@example.com', 'System Administrator', 'admin');

-- Insert some default categories
INSERT INTO categories (name, description)
VALUES
    ('Electronics', 'Electronic devices and accessories'),
    ('Office Supplies', 'Office equipment and supplies'),
    ('Furniture', 'Office and home furniture');

-- Insert some default warehouses
INSERT INTO warehouses (name, location, address)
VALUES
    ('Main Warehouse', 'Headquarters', '123 Main St, City, Country'),
    ('East Warehouse', 'East Branch', '456 East St, City, Country');

-- Insert warehouse locations for Main Warehouse
INSERT INTO warehouse_locations (warehouse_id, zone, aisle, rack, shelf, bin)
VALUES
    (1, 'A', '01', 'A', '01', '01'),
    (1, 'A', '01', 'A', '01', '02'),
    (1, 'A', '01', 'A', '02', '01'),
    (1, 'B', '01', 'A', '01', '01');

-- Insert warehouse locations for East Warehouse
INSERT INTO warehouse_locations (warehouse_id, zone, aisle, rack, shelf, bin)
VALUES
    (2, 'A', '01', 'A', '01', '01'),
    (2, 'A', '01', 'A', '01', '02'),
    (2, 'B', '01', 'A', '01', '01');