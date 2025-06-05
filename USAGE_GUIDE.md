#  Usage Guide - PostgreSQL Unit of Work System

This guide shows you how to run examples and use the PostgreSQL Unit of Work System SDK in your own projects.

##  Table of Contents

1. [Quick Start](#quick-start)
2. [Running Examples](#running-examples)
3. [SDK Usage in Your Project](#sdk-usage-in-your-project)
4. [Complete Examples](#complete-examples)
5. [Testing](#testing)

##  Quick Start

### Prerequisites

- Go 1.21 or later
- PostgreSQL 12+ (for full functionality)
- Git

### Option 1: Run Local Examples

```bash
# Clone the repository
git clone https://github.com/arash-mosavi/postgrs-unit-of-work-system.git
cd postgrs-unit-of-work-system

# Install dependencies
go mod tidy

# Run validation (works without database)
go run validation.go

# Run example (requires PostgreSQL)
go run cmd/main.go
```

### Option 2: Use as SDK in Your Project

```bash
# Create new project
mkdir my-app && cd my-app
go mod init my-app

# Install SDK
go get github.com/arash-mosavi/postgrs-unit-of-work-system

# Create your main.go (see examples below)
```

## 🏃‍♂ Running Examples

### 1. Run Validation Script (No Database Required)

```bash
go run validation.go
```

**Output:**
```
Unit of Work SDK - Validation Example
=====================================
 Configuration created for localhost:5432/testdb
 Unit of Work factories created
 UserService created with dependency injection
 Test Scenarios:
==================
1.  Complex transaction method signature validated
2.  Pagination method signature validated
3.  Batch operations method signature validated
4.  BaseModel interface implementation validated
 All validations passed!
```

### 2. Run Full Example (Requires PostgreSQL)

First, set up PostgreSQL:

```bash
# Using Docker (recommended)
docker run --name postgres-uow \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  -d postgres:15

# Or use docker-compose
docker-compose up -d
```

Then run the example:

```bash
go run cmd/main.go
```

### 3. Run Tests

```bash
# Run all tests (uses in-memory SQLite)
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test -v ./pkg/postgres
go test -v ./examples
```

### 4. Run with Different Configurations

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=testdb

go run cmd/main.go
```

## 💻 SDK Usage in Your Project

### Basic Setup

Create a new Go project and install the SDK:

```bash
mkdir my-uow-app && cd my-uow-app
go mod init my-uow-app
go get github.com/arash-mosavi/postgrs-unit-of-work-system
```

### Example 1: Simple CRUD Operations

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
    "github.com/arash-mosavi/postgrs-unit-of-work-system/examples"
)

func main() {
    // Configure database connection
    config := postgres.NewConfig()
    config.Host = "localhost"
    config.Port = 5432
    config.User = "postgres"
    config.Password = "password"
    config.Database = "testdb"
    config.SSLMode = "disable"

    // Create typed factories
    userFactory := postgres.NewUnitOfWorkFactory[*examples.User](config)
    
    // Create service
    postFactory := postgres.NewUnitOfWorkFactory[*examples.Post](config)
    userService := examples.NewUserService(userFactory, postFactory)

    ctx := context.Background()

    // Create a user
    user := &examples.User{
        Name:  "John Doe",
        Email: "john@example.com",
        Slug:  "john-doe",
    }

    // Create posts
    posts := []*examples.Post{
        {Name: "My First Post", Content: "Hello World!", Slug: "my-first-post"},
        {Name: "Learning Go", Content: "Go is awesome!", Slug: "learning-go"},
    }

    // Service handles the complex transaction
    if err := userService.CreateUserWithPosts(ctx, user, posts); err != nil {
        log.Fatal("Failed to create user with posts:", err)
    }

    fmt.Printf(" Successfully created user '%s' with %d posts\n", 
        user.Name, len(posts))

    // List users with pagination
    users, total, err := userService.ListUsers(ctx, 1, 10)
    if err != nil {
        log.Fatal("Failed to list users:", err)
    }

    fmt.Printf(" Found %d users (total: %d)\n", len(users), total)

    // Search for user by email
    foundUser, err := userService.FindUserByEmail(ctx, "john@example.com")
    if err != nil {
        log.Fatal("Failed to find user:", err)
    }

    fmt.Printf(" Found user: %s (%s)\n", foundUser.Name, foundUser.Email)
}
```

Run it:

```bash
go run main.go
```

### Example 2: Custom Models and Repositories

Create your own models:

```go
// models.go
package main

import (
    "time"
    "gorm.io/gorm"
)

// Product model implementing BaseModel interface
type Product struct {
    ID          int            `gorm:"primaryKey;autoIncrement" json:"id"`
    Slug        string         `gorm:"uniqueIndex;size:100;not null" json:"slug"`
    Name        string         `gorm:"size:255;not null" json:"name"`
    Description string         `gorm:"type:text" json:"description"`
    Price       float64        `gorm:"not null" json:"price"`
    CategoryID  int            `json:"category_id"`
    Active      bool           `gorm:"default:true" json:"active"`
    CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Implement BaseModel interface
func (p *Product) GetID() int                    { return p.ID }
func (p *Product) GetSlug() string               { return p.Slug }
func (p *Product) SetSlug(slug string)           { p.Slug = slug }
func (p *Product) GetCreatedAt() time.Time       { return p.CreatedAt }
func (p *Product) GetUpdatedAt() time.Time       { return p.UpdatedAt }
func (p *Product) GetArchivedAt() gorm.DeletedAt { return p.DeletedAt }
func (p *Product) GetName() string               { return p.Name }

func (Product) TableName() string { return "products" }

// Category model
type Category struct {
    ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
    Slug      string         `gorm:"uniqueIndex;size:100;not null" json:"slug"`
    Name      string         `gorm:"size:255;not null" json:"name"`
    Active    bool           `gorm:"default:true" json:"active"`
    CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Implement BaseModel interface
func (c *Category) GetID() int                    { return c.ID }
func (c *Category) GetSlug() string               { return c.Slug }
func (c *Category) SetSlug(slug string)           { c.Slug = slug }
func (c *Category) GetCreatedAt() time.Time       { return c.CreatedAt }
func (c *Category) GetUpdatedAt() time.Time       { return c.UpdatedAt }
func (c *Category) GetArchivedAt() gorm.DeletedAt { return c.DeletedAt }
func (c *Category) GetName() string               { return c.Name }

func (Category) TableName() string { return "categories" }
```

Create your service:

```go
// service.go
package main

import (
    "context"
    "fmt"

    "github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/persistence"
    "github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/identifier"
)

type ProductService struct {
    productFactory  persistence.IUnitOfWorkFactory[*Product]
    categoryFactory persistence.IUnitOfWorkFactory[*Category]
}

func NewProductService(
    productFactory persistence.IUnitOfWorkFactory[*Product],
    categoryFactory persistence.IUnitOfWorkFactory[*Category],
) *ProductService {
    return &ProductService{
        productFactory:  productFactory,
        categoryFactory: categoryFactory,
    }
}

func (s *ProductService) CreateProductWithCategory(
    ctx context.Context, 
    categoryName string, 
    product *Product,
) error {
    // Create UoW instances
    categoryUow := s.categoryFactory.CreateWithContext(ctx)
    productUow := s.productFactory.CreateWithContext(ctx)

    // Begin transactions
    if err := categoryUow.BeginTransaction(ctx); err != nil {
        return fmt.Errorf("failed to begin category transaction: %w", err)
    }
    defer categoryUow.RollbackTransaction(ctx)

    // Create or find category
    categorySlug := fmt.Sprintf("%s-category", categoryName)
    categoryIdentifier := identifier.NewIdentifier().Equal("slug", categorySlug)
    
    existingCategory, err := categoryUow.FindOneByIdentifier(ctx, categoryIdentifier)
    if err != nil {
        // Category doesn't exist, create it
        category := &Category{
            Name: categoryName,
            Slug: categorySlug,
        }
        
        existingCategory, err = categoryUow.Insert(ctx, category)
        if err != nil {
            return fmt.Errorf("failed to create category: %w", err)
        }
    }

    // Begin product transaction
    if err := productUow.BeginTransaction(ctx); err != nil {
        return fmt.Errorf("failed to begin product transaction: %w", err)
    }
    defer productUow.RollbackTransaction(ctx)

    // Set category ID and create product
    product.CategoryID = existingCategory.GetID()
    _, err = productUow.Insert(ctx, product)
    if err != nil {
        return fmt.Errorf("failed to create product: %w", err)
    }

    // Commit both transactions
    if err := categoryUow.CommitTransaction(ctx); err != nil {
        return fmt.Errorf("failed to commit category transaction: %w", err)
    }

    if err := productUow.CommitTransaction(ctx); err != nil {
        return fmt.Errorf("failed to commit product transaction: %w", err)
    }

    return nil
}

func (s *ProductService) GetProductsByCategory(ctx context.Context, categoryID int) ([]*Product, error) {
    uow := s.productFactory.CreateWithContext(ctx)
    
    identifier := identifier.NewIdentifier().Equal("category_id", categoryID)
    return uow.FindByIdentifier(ctx, identifier)
}
```

Use your custom service:

```go
// main.go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
)

func main() {
    config := postgres.NewConfig()
    config.Host = "localhost"
    config.Port = 5432
    config.User = "postgres"
    config.Password = "password"
    config.Database = "testdb"
    config.SSLMode = "disable"

    // Create factories for your custom models
    productFactory := postgres.NewUnitOfWorkFactory[*Product](config)
    categoryFactory := postgres.NewUnitOfWorkFactory[*Category](config)

    // Create your service
    productService := NewProductService(productFactory, categoryFactory)

    ctx := context.Background()

    // Create product with category
    product := &Product{
        Name:        "Laptop",
        Description: "High-performance laptop",
        Price:       999.99,
        Slug:        "laptop-hp-2024",
    }

    err := productService.CreateProductWithCategory(ctx, "Electronics", product)
    if err != nil {
        log.Fatal("Failed to create product:", err)
    }

    fmt.Printf(" Created product: %s\n", product.Name)

    // Get products by category
    products, err := productService.GetProductsByCategory(ctx, product.CategoryID)
    if err != nil {
        log.Fatal("Failed to get products:", err)
    }

    fmt.Printf(" Found %d products in category\n", len(products))
}
```

### Example 3: Advanced Query Operations

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
    "github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/identifier"
    "github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/domain"
    "github.com/arash-mosavi/postgrs-unit-of-work-system/examples"
)

func advancedQueryExamples() {
    config := postgres.NewConfig()
    config.Host = "localhost"
    config.Port = 5432
    config.User = "postgres"
    config.Password = "password"
    config.Database = "testdb"
    config.SSLMode = "disable"

    userFactory := postgres.NewUnitOfWorkFactory[*examples.User](config)
    uow := userFactory.CreateWithContext(context.Background())
    ctx := context.Background()

    // Example 1: Complex identifier queries
    fmt.Println("1. Complex Query Examples:")
    
    // Find users with specific conditions
    identifier := identifier.NewIdentifier().
        Equal("active", true).
        Like("email", "%@gmail.com").
        Greater("id", 1)
    
    users, err := uow.FindByIdentifier(ctx, identifier)
    if err != nil {
        log.Printf("Query failed: %v", err)
    } else {
        fmt.Printf("   Found %d Gmail users\n", len(users))
    }

    // Example 2: Pagination with sorting
    fmt.Println("2. Pagination Examples:")
    
    params := domain.QueryParams[*examples.User]{
        Limit:  5,
        Offset: 0,
        Sort: domain.SortMap{
            "created_at": domain.SortDesc,
            "name":       domain.SortAsc,
        },
    }

    paginatedUsers, total, err := uow.FindAllWithPagination(ctx, params)
    if err != nil {
        log.Printf("Pagination failed: %v", err)
    } else {
        fmt.Printf("   Page 1: %d users (total: %d)\n", len(paginatedUsers), total)
    }

    // Example 3: Bulk operations
    fmt.Println("3. Bulk Operations:")
    
    newUsers := []*examples.User{
        {Name: "Bulk User 1", Email: "bulk1@test.com", Slug: "bulk-user-1"},
        {Name: "Bulk User 2", Email: "bulk2@test.com", Slug: "bulk-user-2"},
        {Name: "Bulk User 3", Email: "bulk3@test.com", Slug: "bulk-user-3"},
    }

    createdUsers, err := uow.BulkInsert(ctx, newUsers)
    if err != nil {
        log.Printf("Bulk insert failed: %v", err)
    } else {
        fmt.Printf("   Created %d users in bulk\n", len(createdUsers))
    }

    // Example 4: Soft delete and trash management
    fmt.Println("4. Soft Delete Examples:")
    
    if len(createdUsers) > 0 {
        // Soft delete
        deleteIdentifier := identifier.NewIdentifier().Equal("id", createdUsers[0].GetID())
        deletedUser, err := uow.SoftDelete(ctx, deleteIdentifier)
        if err != nil {
            log.Printf("Soft delete failed: %v", err)
        } else {
            fmt.Printf("   Soft deleted user: %s\n", deletedUser.GetName())
        }

        // Find trashed items
        trashedUsers, err := uow.FindTrashed(ctx)
        if err != nil {
            log.Printf("Find trashed failed: %v", err)
        } else {
            fmt.Printf("   Found %d trashed users\n", len(trashedUsers))
        }

        // Restore from trash
        restoreIdentifier := identifier.NewIdentifier().Equal("id", createdUsers[0].GetID())
        restoredUser, err := uow.RestoreFromTrash(ctx, restoreIdentifier)
        if err != nil {
            log.Printf("Restore failed: %v", err)
        } else {
            fmt.Printf("   Restored user: %s\n", restoredUser.GetName())
        }
    }
}

func main() {
    advancedQueryExamples()
}
```

##  Testing

### Run Built-in Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific tests
go test -v ./pkg/postgres -run TestUnitOfWork_Insert
go test -v ./examples -run TestModels
```

### Write Your Own Tests

```go
// my_test.go
package main

import (
    "context"
    "testing"

    "github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestProductService(t *testing.T) {
    // Use in-memory SQLite for testing
    config := postgres.NewConfig()
    config.Database = ":memory:"  // SQLite in-memory

    productFactory := postgres.NewUnitOfWorkFactory[*Product](config)
    categoryFactory := postgres.NewUnitOfWorkFactory[*Category](config)
    
    service := NewProductService(productFactory, categoryFactory)
    
    ctx := context.Background()
    
    product := &Product{
        Name:        "Test Product",
        Description: "Test Description", 
        Price:       99.99,
        Slug:        "test-product",
    }
    
    err := service.CreateProductWithCategory(ctx, "Test Category", product)
    require.NoError(t, err)
    
    assert.Equal(t, "Test Product", product.Name)
    assert.NotZero(t, product.ID)
}
```

##  Configuration Options

### Environment Variables

```bash
# Database configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=testdb
export DB_SSLMODE=disable

# Application configuration  
export LOG_LEVEL=debug
export APP_ENV=development
```

### Programmatic Configuration

```go
config := postgres.NewConfig()
config.Host = "localhost"
config.Port = 5432
config.User = "postgres"
config.Password = "password"
config.Database = "testdb"
config.SSLMode = "disable"
config.MaxIdleConns = 10
config.MaxOpenConns = 100
config.ConnMaxLifetime = time.Hour
```

##  Next Steps

1. **Explore the Examples**: Check out `examples/usage.go` for more patterns
2. **Read the Documentation**: See `README.md` for detailed API docs
3. **Implement Your Models**: Create your own models implementing `BaseModel`
4. **Build Your Services**: Use the repository pattern for clean architecture
5. **Write Tests**: Use the testing patterns shown above

## 🆘 Common Issues

### Database Connection Issues
```go
// Test connection
if err := uow.GetDB().Exec("SELECT 1").Error; err != nil {
    log.Fatal("Database connection failed:", err)
}
```

### Migration Issues
```go
// Auto-migrate your models
db := uow.GetDB()
err := db.AutoMigrate(&Product{}, &Category{})
if err != nil {
    log.Fatal("Migration failed:", err)
}
```

Happy coding! 
