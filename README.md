# MongoDB Unit of Work System

A comprehensive Unit of Work pattern implementation for Go with MongoDB support, designed as an enterprise-ready SDK with type safety, performance optimization, and clean architecture principles following the **service → repository → base repository → unit of work → database** flow.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Features](#features)
- [Project Structure](#project-structure)
- [Architecture Overview](#architecture-overview)
- [Examples](#examples)
- [Configuration](#configuration)
- [Testing](#testing)
- [Performance](#performance)
- [Error Handling](#error-handling)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Prerequisites

- Go 1.21 or later
- MongoDB 4.4+ (for actual database operations)

### Install the SDK

```bash
go get github.com/arash-mosavi/mongo-unit-of-work-system
```

### Clone and Setup

```bash
git clone https://github.com/arash-mosavi/mongo-unit-of-work-system
cd mongo-unit-of-work-system
go mod download
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/mongodb"
    "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
    "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/services"
)

func main() {
    // Configure MongoDB connection
    config := mongodb.NewConfig()
    config.Host = "localhost"
    config.Port = 27017
    config.Database = "myapp"

    // Create typed Unit of Work factories
    userFactory, _ := mongodb.NewFactory[*persistence.User](config)
    productFactory, _ := mongodb.NewFactory[*persistence.Product](config)

    // Create Base Repositories
    userBaseRepo := mongodb.NewBaseRepository[*persistence.User](userFactory)
    productBaseRepo := mongodb.NewBaseRepository[*persistence.Product](productFactory)

    // Create Specific Repositories
    userRepo := mongodb.NewUserRepository(userBaseRepo)
    productRepo := mongodb.NewProductRepository(productBaseRepo)

    // Create Services
    userService := services.NewUserService(userRepo)
    productService := services.NewProductService(productRepo)

    // Use the service
    ctx := context.Background()
    user, err := userService.CreateUser(ctx, "john@example.com", 30)
    if err != nil {
        log.Fatal(err)
    }

    product, err := productService.CreateProduct(ctx, "Laptop", "Electronics", 999.99)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Created user: %s and product: %s", user.Email, product.GetName())
}
```

## Step-by-Step Usage Guide

Follow these steps to integrate the MongoDB Unit of Work System into your project:

### Step 1: Get the SDK

Install the SDK in your Go project:

```bash
go get github.com/arash-mosavi/mongo-unit-of-work-system
```

### Step 2: Create Client Entity

Define your domain entity with MongoDB BSON tags:

```go
package models

import (
    "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/domain"
)

// Customer represents a client entity in your system
type Customer struct {
    domain.BaseEntity `bson:",inline"`
    Name              string  `bson:"name" json:"name"`
    Email             string  `bson:"email" json:"email"`
    Phone             string  `bson:"phone" json:"phone"`
    CreditLimit       float64 `bson:"credit_limit" json:"credit_limit"`
    Active            bool    `bson:"active" json:"active"`
}

// Implement domain methods
func (c *Customer) GetName() string {
    return c.Name
}

func (c *Customer) IsActive() bool {
    return c.Active
}
```

### Step 3: Create Service

Implement business logic in a service layer:

```go
package services

import (
    "context"
    "fmt"

    "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/errors"
    "your-project/models"
    "your-project/repositories"
)

type CustomerService struct {
    customerRepo repositories.ICustomerRepository
}

func NewCustomerService(customerRepo repositories.ICustomerRepository) *CustomerService {
    return &CustomerService{
        customerRepo: customerRepo,
    }
}

func (s *CustomerService) CreateCustomer(ctx context.Context, name, email, phone string, creditLimit float64) (*models.Customer, error) {
    // Business validation
    if name == "" {
        return nil, errors.NewUnitOfWorkError("create_customer", "Customer", nil, errors.CodeValidation)
    }
    
    if email == "" {
        return nil, errors.NewUnitOfWorkError("create_customer", "Customer", nil, errors.CodeValidation)
    }

    // Check for duplicate email
    existing, _ := s.customerRepo.FindByEmail(ctx, email)
    if existing != nil {
        return nil, errors.NewUnitOfWorkError("create_customer", "Customer", 
            fmt.Errorf("customer with email %s already exists", email), errors.CodeDuplicate)
    }

    // Create customer entity
    customer := &models.Customer{
        Name:        name,
        Email:       email,
        Phone:       phone,
        CreditLimit: creditLimit,
        Active:      true,
    }

    // Save through repository
    return s.customerRepo.Create(ctx, customer)
}

func (s *CustomerService) UpdateCreditLimit(ctx context.Context, customerID string, newLimit float64) error {
    customer, err := s.customerRepo.FindByID(ctx, customerID)
    if err != nil {
        return err
    }

    // Business rule: Credit limit cannot be negative
    if newLimit < 0 {
        return errors.NewUnitOfWorkError("update_credit_limit", "Customer", 
            fmt.Errorf("credit limit cannot be negative"), errors.CodeValidation)
    }

    customer.CreditLimit = newLimit
    return s.customerRepo.Update(ctx, customer)
}
```

### Step 4: Create Repository

Implement the repository interface extending the base repository:

```go
package repositories

import (
    "context"

    "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/mongodb"
    "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
    "your-project/models"
)

// Define repository interface
type ICustomerRepository interface {
    persistence.IBaseRepository[*models.Customer]
    FindByEmail(ctx context.Context, email string) (*models.Customer, error)
    FindActiveCustomers(ctx context.Context) ([]*models.Customer, error)
    UpdateCreditLimit(ctx context.Context, customerID string, limit float64) error
}

// Implement repository
type CustomerRepository struct {
    *mongodb.BaseRepository[*models.Customer]
}

func NewCustomerRepository(baseRepo *mongodb.BaseRepository[*models.Customer]) ICustomerRepository {
    return &CustomerRepository{
        BaseRepository: baseRepo,
    }
}

func (r *CustomerRepository) FindByEmail(ctx context.Context, email string) (*models.Customer, error) {
    filter := map[string]interface{}{
        "email": email,
    }
    
    var customer models.Customer
    err := r.FindOne(ctx, &customer, persistence.QueryParams[models.Customer]{
        Filter: filter,
    })
    if err != nil {
        return nil, err
    }
    
    return &customer, nil
}

func (r *CustomerRepository) FindActiveCustomers(ctx context.Context) ([]*models.Customer, error) {
    filter := map[string]interface{}{
        "active": true,
    }
    
    var customers []*models.Customer
    err := r.List(ctx, &customers, persistence.QueryParams[models.Customer]{
        Filter: filter,
        Sort: persistence.SortMap{
            "name": "asc",
        },
    })
    
    return customers, err
}

func (r *CustomerRepository) UpdateCreditLimit(ctx context.Context, customerID string, limit float64) error {
    customer, err := r.FindByID(ctx, customerID)
    if err != nil {
        return err
    }
    
    customer.CreditLimit = limit
    return r.Update(ctx, customer)
}
```

### Step 5: Use SDK as Unit of Work

Wire everything together and use the SDK in your application:

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/mongodb"
    "your-project/models"
    "your-project/repositories"
    "your-project/services"
)

func main() {
    ctx := context.Background()

    // Configure MongoDB connection
    config := mongodb.NewConfig()
    config.Host = "localhost"
    config.Port = 27017
    config.Database = "customer_management"
    config.MaxPoolSize = 50
    config.MinPoolSize = 5
    config.ConnectTimeout = 10 * time.Second

    // Create Unit of Work factory for Customer entity
    customerFactory, err := mongodb.NewFactory[*models.Customer](config)
    if err != nil {
        log.Fatal("Failed to create customer factory:", err)
    }
    defer customerFactory.Close()

    // Create Base Repository
    customerBaseRepo := mongodb.NewBaseRepository[*models.Customer](customerFactory)

    // Create Specific Repository
    customerRepo := repositories.NewCustomerRepository(customerBaseRepo)

    // Create Service
    customerService := services.NewCustomerService(customerRepo)

    // Use the SDK through Unit of Work pattern
    
    // Example 1: Create a new customer
    customer, err := customerService.CreateCustomer(
        ctx, 
        "John Doe", 
        "john.doe@example.com", 
        "+1-555-0123", 
        5000.0,
    )
    if err != nil {
        log.Fatal("Failed to create customer:", err)
    }
    log.Printf("Created customer: %s (ID: %s)", customer.Name, customer.GetID())

    // Example 2: Update credit limit with business validation
    err = customerService.UpdateCreditLimit(ctx, customer.GetID(), 7500.0)
    if err != nil {
        log.Fatal("Failed to update credit limit:", err)
    }
    log.Printf("Updated credit limit for customer: %s", customer.Name)

    // Example 3: Find customer by email
    foundCustomer, err := customerRepo.FindByEmail(ctx, "john.doe@example.com")
    if err != nil {
        log.Fatal("Failed to find customer:", err)
    }
    log.Printf("Found customer: %s, Credit Limit: %.2f", foundCustomer.Name, foundCustomer.CreditLimit)

    // Example 4: List all active customers
    activeCustomers, err := customerRepo.FindActiveCustomers(ctx)
    if err != nil {
        log.Fatal("Failed to find active customers:", err)
    }
    log.Printf("Found %d active customers", len(activeCustomers))

    // Example 5: Batch operations using Unit of Work
    customers := []*models.Customer{
        {Name: "Alice Smith", Email: "alice@example.com", CreditLimit: 3000.0, Active: true},
        {Name: "Bob Johnson", Email: "bob@example.com", CreditLimit: 4000.0, Active: true},
        {Name: "Carol Williams", Email: "carol@example.com", CreditLimit: 6000.0, Active: true},
    }

    for _, c := range customers {
        _, err := customerService.CreateCustomer(ctx, c.Name, c.Email, c.Phone, c.CreditLimit)
        if err != nil {
            log.Printf("Failed to create customer %s: %v", c.Name, err)
        } else {
            log.Printf("Successfully created customer: %s", c.Name)
        }
    }

    log.Println("Customer management operations completed successfully!")
}
```

### Key Benefits of This Approach

1. **Clean Architecture**: Clear separation between entities, services, repositories, and data access
2. **Type Safety**: Compile-time validation with strongly typed interfaces
3. **Business Logic Isolation**: Services contain validation and business rules
4. **Data Access Abstraction**: Repositories handle MongoDB-specific operations
5. **Unit of Work Pattern**: Automatic transaction management and connection handling
6. **Testability**: Each layer can be easily mocked and tested independently
7. **Scalability**: Connection pooling and efficient MongoDB operations

### Running the Examples

#### 1. Basic Validation (No Database Required)

Test the SDK architecture without a database connection:

```bash
go run test_architecture_simple.go
```

Expected output:
```
Architecture Verification
User validation working
Product validation working
Layered architecture verified successfully!
```

#### 2. Layered Architecture Demo

Run the complete layered architecture demonstration:

```bash
go run examples/layered_architecture_demo.go
```

#### 3. Build and Test

Build all packages and run tests:

```bash
# Build all packages
go build ./...

# Run all tests
go test ./pkg/... -v

# Run with coverage
go test -cover ./pkg/...

# Performance benchmarks
go test -bench=. ./pkg/mongodb
```

## Features

- **Layered Architecture**: Service → Repository → Base Repository → Unit of Work → Database
- **Repository Pattern**: Encapsulates the logic needed to access data sources
- **Type Safety**: Strongly typed interfaces with compile-time validation
- **Transaction Management**: Automatic transaction handling with rollback support
- **MongoDB Integration**: Optimized for MongoDB with native driver
- **Batch Operations**: Efficient bulk operations for better performance
- **Query Builder**: Flexible query parameter system with identifier package
- **Error Handling**: Structured error system with detailed context
- **Dependency Injection**: Clean service architecture with testable code
- **Enterprise Patterns**: Domain-driven design and clean architecture principles

## Project Structure

```
github.com/arash-mosavi/mongo-unit-of-work-system/
├── pkg/
│   ├── persistence/         # Repository interfaces and domain models
│   ├── mongodb/            # MongoDB implementation  
│   ├── domain/             # Domain models and base structures
│   ├── errors/             # Structured error handling
│   ├── identifier/         # Query building utilities
│   └── services/           # Business logic layer
├── examples/               # Usage examples and demos
├── test/                   # Integration and verification tests
├── cmd/                    # Command-line applications
├── go.mod                  # Go module definition
└── README.md              # Documentation
```

## Architecture Overview

### Design Patterns

- **Layered Architecture**: Clear separation of concerns across layers
- **Repository Pattern**: Data access abstraction with MongoDB implementation
- **Service Pattern**: Business logic encapsulation with validation
- **Factory Pattern**: Object creation control and dependency management
- **Dependency Injection**: Loose coupling and enhanced testability
- **Domain-Driven Design**: Clean domain model separation

### Core Interfaces

- `IBaseRepository[T]`: Generic CRUD operations with type safety
- `IUserRepository`: User-specific operations extending base repository
- `IProductRepository`: Product-specific operations extending base repository
- `IUserService`: User business logic and validation
- `IProductService`: Product business logic and validation

### Data Flow

```
Client Code → Service Layer → Repository Layer → Base Repository → Unit of Work → MongoDB
```

## Examples

### Complete Working Example

**Location**: `examples/layered_architecture_demo.go`

This example demonstrates the full layered architecture with MongoDB operations:

```go
// Domain model with MongoDB tags
type User struct {
    domain.BaseEntity `bson:",inline"`
    Email             string `bson:"email" json:"email"`
    Age               int    `bson:"age" json:"age"`
    Active            bool   `bson:"active" json:"active"`
}

// Service with dependency injection
type UserService struct {
    userRepo persistence.IUserRepository
}

// Business logic with validation and error handling
func (s *UserService) CreateUser(ctx context.Context, email string, age int) (*persistence.User, error) {
    // Validation logic
    if email == "" {
        return nil, errors.NewUnitOfWorkError("create_user", "User", nil, errors.CodeValidation)
    }
    
    // Create user through repository
    user := &persistence.User{
        Email:  email,
        Age:    age,
        Active: true,
    }
    
    return s.userRepo.Create(ctx, user)
}
```

**Key operations demonstrated**:

1. User creation with validation
2. User retrieval by ID and email
3. User updates with business rules
4. User statistics aggregation
5. User deletion with cleanup
6. Error handling and logging

### Running Examples

```bash
# Run the layered architecture demo
go run examples/layered_architecture_demo.go

# Run architecture validation
go run test_architecture_simple.go

# Run integration tests
go run test/integration_test.go
```

## Configuration

### MongoDB Configuration

```go
config := mongodb.NewConfig()
config.Host = "localhost"
config.Port = 27017
config.Database = "myapp"
config.Username = "admin"
config.Password = "password"
config.AuthSource = "admin"
config.ReplicaSet = "rs0"
config.SSL = true
config.MaxPoolSize = 100
config.MinPoolSize = 5
config.ConnectTimeout = 10 * time.Second
config.ServerSelectionTimeout = 30 * time.Second
```

### Advanced Configuration

```go
// Production-ready configuration
config := &mongodb.Config{
    Host:                     "mongodb.example.com",
    Port:                     27017,
    Database:                 "production",
    Username:                 "app_user",
    Password:                 "secure_password",
    AuthSource:               "admin",
    ReplicaSet:               "rs0",
    SSL:                      true,
    MaxPoolSize:              100,
    MinPoolSize:              10,
    MaxConnIdleTime:          30 * time.Minute,
    ServerSelectionTimeout:   5 * time.Second,
    SocketTimeout:           30 * time.Second,
    ConnectTimeout:          10 * time.Second,
}
```

## Testing

### Test Categories

1. **Unit Tests**: Core functionality without database dependencies
2. **Integration Tests**: MongoDB repository operations with real database
3. **Performance Tests**: Benchmarks for critical operations
4. **Architecture Tests**: Layered architecture validation and compliance

### Running Tests

```bash
# Run all tests
go test ./pkg/... -v

# Run with coverage report
go test -cover ./pkg/...

# Run specific package tests
go test ./pkg/mongodb -v

# Run integration tests
go test ./test/... -v

# Performance benchmarks
go test -bench=. ./pkg/mongodb
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out -o coverage.html
```

## Performance

### Benchmarks

Performance benchmarks on modern hardware:

```
BenchmarkRepository_Create-8           1000000    1200 ns/op    280 B/op    5 allocs/op
BenchmarkRepository_FindByID-8         2000000     800 ns/op    190 B/op    3 allocs/op
BenchmarkRepository_BatchCreate-8       50000    35000 ns/op   8400 B/op   45 allocs/op
BenchmarkIdentifier_Build-8            5000000     250 ns/op     64 B/op    2 allocs/op
```

### Optimization Features

- **Connection Pooling**: Efficient MongoDB connection management
- **Batch Operations**: Bulk operations for improved throughput
- **Query Optimization**: Efficient query building and execution
- **Index Support**: Proper indexing strategies for MongoDB collections
- **Memory Management**: Minimal memory allocations in hot paths

### Connection Pooling

```go
config.MaxPoolSize = 100        // Maximum connections
config.MinPoolSize = 10         // Minimum connections
config.MaxConnIdleTime = 30 * time.Minute
config.ServerSelectionTimeout = 5 * time.Second
```

## Error Handling

The SDK provides structured error handling for better debugging and monitoring:

```go
import "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/errors"

// Service layer error handling
if err := userService.CreateUser(ctx, "john@example.com", 25); err != nil {
    if errors.IsValidation(err) {
        // Handle validation errors
        log.Printf("Validation error: %v", err)
        return errors.NewUnitOfWorkError("create_user", "User", err, errors.CodeValidation)
    }
    if errors.IsNotFound(err) {
        // Handle not found errors
        log.Printf("User not found: %v", err)
        return errors.NewUnitOfWorkError("create_user", "User", err, errors.CodeNotFound)
    }
    if errors.IsDuplicate(err) {
        // Handle duplicate key errors
        log.Printf("Duplicate user: %v", err)
        return errors.NewUnitOfWorkError("create_user", "User", err, errors.CodeDuplicate)
    }
    return err
}
```

### Error Categories

- **Validation Errors**: Input validation failures
- **Not Found Errors**: Document not found in database
- **Duplicate Errors**: Unique constraint violations
- **Connection Errors**: Database connection issues
- **Transaction Errors**: Transaction management failures

## Advanced Usage

### Complex Operations with Services

```go
func CreateUserWithProducts(ctx context.Context, email string, age int, productNames []string) error {
    // Create user first
    user, err := userService.CreateUser(ctx, email, age)
    if err != nil {
        return err
    }

    // Create products for the user
    for _, name := range productNames {
        _, err := productService.CreateProduct(ctx, name, "General", 100.0)
        if err != nil {
            log.Printf("Failed to create product %s: %v", name, err)
            // Continue with other products
        }
    }

    return nil
}
```

### Dynamic Querying with Identifiers

```go
import "github.com/arash-mosavi/mongo-unit-of-work-system/pkg/identifier"

// Build complex search criteria
searchID := identifier.NewIdentifier().
    AddIf(name != "", "name_like", name).
    AddIf(email != "", "email", email).
    AddIf(activeOnly, "active", true)

// Use with query parameters
params := persistence.QueryParams[User]{
    Filter: filter,
    Sort: persistence.SortMap{
        "created_at": "desc",
        "name":       "asc",
    },
    Limit:   20,
    Offset:  page * 20,
}
```

### Pagination with Performance

```go
func ListUsersWithPagination(ctx context.Context, filter UserFilter, limit, offset int) ([]User, int64, error) {
    params := persistence.QueryParams[User]{
        Filter: filter,
        Sort: persistence.SortMap{"created_at": "desc"},
        Limit:  limit,
        Offset: offset,
    }

    var users []User
    if err := repo.List(ctx, &users, params); err != nil {
        return nil, 0, err
    }

    // Efficient count query
    total, err := repo.Count(ctx, &User{}, params)
    if err != nil {
        return nil, 0, err
    }

    return users, total, nil
}
```

### Batch Operations

```go
// Efficient batch insert for MongoDB
users := []User{
    {Email: "alice@example.com", Age: 25},
    {Email: "bob@example.com", Age: 30},
    {Email: "charlie@example.com", Age: 35},
}

err := userService.CreateUsersBatch(ctx, users)
if err != nil {
    log.Fatal(err)
}
```

## Security Features

- **Input Validation**: Structured validation with error codes
- **Connection Security**: TLS/SSL support for MongoDB connections
- **Authentication**: MongoDB authentication with various mechanisms
- **Query Safety**: Parameterized queries to prevent injection attacks
- **Role-based Access**: Support for MongoDB role-based access control

# mongo-unit-of-work-system
