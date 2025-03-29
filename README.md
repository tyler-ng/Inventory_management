# Inventory Management System

A comprehensive inventory management system built with Go, GORM, PostgreSQL, and Docker.

## Features

- **Product Management**: Track detailed product information including categories, variants, and bundles
- **Multi-location Inventory**: Manage inventory across multiple warehouses and locations
- **Purchase Order Management**: Create and track purchase orders to suppliers
- **Sales Order Management**: Process customer orders and track fulfillment
- **Inventory Transactions**: Record all stock movements with detailed history
- **User Management**: Role-based access control with secure authentication
- **Reporting**: Generate reports on inventory value, stock levels, and product movement

## Technology Stack

- **Backend**: Go with Gorilla Mux for routing
- **Database**: PostgreSQL
- **ORM**: GORM for database operations
- **Authentication**: JWT-based authentication
- **Containerization**: Docker and Docker Compose for easy deployment

## Getting Started

### Prerequisites

- Go 1.20 or higher
- Docker and Docker Compose
- Git

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/inventory-management-system.git
   cd inventory-management-system
   ```

2. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env file with your configuration
   ```

3. Start the application with Docker Compose:
   ```bash
   docker-compose up -d
   ```

4. The API will be available at http://localhost:8080

### Running Without Docker

1. Set up a PostgreSQL database
2. Configure your environment variables in the `.env` file
3. Run the application:
   ```bash
   go run cmd/api/main.go
   ```

## API Documentation

### Authentication Endpoints

- `POST /api/auth/login`: Authenticate a user and get JWT token
- `POST /api/auth/register`: Register a new user

### Product Endpoints

- `GET /api/products`: Get all products with optional filtering
- `GET /api/products/{id}`: Get a specific product by ID
- `POST /api/products`: Create a new product
- `PUT /api/products/{id}`: Update an existing product
- `DELETE /api/products/{id}`: Delete a product

### Inventory Transaction Endpoints

- `GET /api/transactions`: Get all inventory transactions
- `GET /api/transactions/{id}`: Get a specific transaction
- `POST /api/transactions`: Create a generic transaction
- `POST /api/transactions/receive`: Create a receive transaction
- `POST /api/transactions/issue`: Create an issue transaction
- `POST /api/transactions/transfer`: Create a transfer transaction

### Purchase Order Endpoints

- `GET /api/purchase-orders`: Get all purchase orders
- `GET /api/purchase-orders/{id}`: Get a specific purchase order
- `POST /api/purchase-orders`: Create a new purchase order
- `PUT /api/purchase-orders/{id}`: Update a purchase order
- `POST /api/purchase-orders/{id}/receive`: Receive items from a purchase order

### Sales Order Endpoints

- `GET /api/sales-orders`: Get all sales orders
- `GET /api/sales-orders/{id}`: Get a specific sales order
- `POST /api/sales-orders`: Create a new sales order
- `PUT /api/sales-orders/{id}`: Update a sales order
- `POST /api/sales-orders/{id}/fulfill`: Fulfill a sales order

## Database Structure

The system uses a relational database with the following key entities:

- **Products**: Core inventory items with detailed attributes
- **Categories**: Hierarchical categorization of products
- **Warehouses and Locations**: Physical storage locations
- **Inventory Transactions**: Record of all stock movements
- **Purchase Orders**: Orders to suppliers
- **Sales Orders**: Customer orders
- **Users and Authentication**: User accounts and access control

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Project Structure

inventory-management-system/
├── cmd/
│   └── api/
│       └── main.go              # Entry point for the API server
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration handling
│   ├── database/
│   │   └── database.go          # Database connection setup
│   ├── models/
│   │   ├── product.go           # Product model
│   │   ├── category.go          # Category model
│   │   ├── supplier.go          # Supplier model
│   │   ├── warehouse.go         # Warehouse model
│   │   ├── warehouse_location.go # Warehouse location model
│   │   ├── inventory_transaction.go # Inventory transaction model
│   │   ├── purchase_order.go    # Purchase order model
│   │   ├── sales_order.go       # Sales order model
│   │   ├── customer.go          # Customer model
│   │   ├── user.go              # User model
│   │   └── audit_log.go         # Audit log model
│   ├── handlers/
│   │   ├── product_handler.go   # Product API handlers
│   │   ├── category_handler.go  # Category API handlers
│   │   ├── supplier_handler.go  # Supplier API handlers
│   │   ├── warehouse_handler.go # Warehouse API handlers
│   │   ├── transaction_handler.go # Inventory transaction API handlers
│   │   ├── purchase_handler.go  # Purchase order API handlers
│   │   ├── sales_handler.go     # Sales order API handlers
│   │   ├── customer_handler.go  # Customer API handlers
│   │   └── user_handler.go      # User API handlers
│   ├── middleware/
│   │   ├── auth.go              # Authentication middleware
│   │   └── logging.go           # Logging middleware
│   ├── repository/
│   │   ├── product_repo.go      # Product database operations
│   │   ├── category_repo.go     # Category database operations
│   │   └── ...                  # Other repository files
│   └── services/
│       ├── product_service.go   # Product business logic
│       ├── inventory_service.go # Inventory business logic
│       └── ...                  # Other service files
├── pkg/
│   ├── auth/
│   │   └── jwt.go               # JWT authentication utilities
│   ├── validation/
│   │   └── validator.go         # Input validation utilities
│   └── utils/
│       └── helpers.go           # Utility functions
├── migrations/
│   ├── 001_initial_schema.up.sql # Initial database schema
│   └── 001_initial_schema.down.sql # Rollback for initial schema
├── .env                         # Environment variables
├── .gitignore                   # Git ignore file
├── docker-compose.yml           # Docker Compose for local development
├── Dockerfile                   # Docker configuration
├── go.mod                       # Go modules file
└── README.md                    # Project documentation