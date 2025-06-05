# MongoDB Layered Architecture Demo

This project demonstrates a proper layered architecture implementation for MongoDB Unit of Work system following the flow:

**Service → Repository → Base Repository → Unit of Work → Database**

## Architecture Overview

```
Client Code
    ↓
📞 Service Layer (Business Logic)
    ↓
📦 Repository Layer (Data Access Abstraction)
    ↓
🔧 Base Repository (Generic CRUD Operations)
    ↓
⚙️  Unit of Work (Transaction Management)
    ↓
🗄️  MongoDB Database (Data Persistence)
```

## Layer Responsibilities

### 1. Service Layer (`pkg/services/`)
- **Purpose**: Contains business logic and validation
- **Files**: `services.go`
- **Interfaces**: `IUserService`, `IProductService`
- **Responsibilities**:
  - Input validation (email format, age ranges, price validation)
  - Business rules enforcement (no duplicate emails, etc.)
  - Complex business operations
  - Error handling and meaningful error messages
  - Orchestrating multiple repository calls

### 2. Repository Layer (`pkg/persistence/repositories.go`)
- **Purpose**: Data access abstraction with domain-specific queries
- **Interfaces**: 
  - `IBaseRepository[T]` - Generic CRUD operations
  - `IUserRepository` - User-specific queries
  - `IProductRepository` - Product-specific queries
- **Responsibilities**:
  - Abstract data access patterns
  - Domain-specific query methods
  - Type-safe operations with generics

### 3. Base Repository Implementation (`pkg/mongodb/base_repository.go`)
- **Purpose**: Generic implementation that delegates to Unit of Work
- **Responsibilities**:
  - Implements `IBaseRepository[T]` interface
  - Delegates all operations to Unit of Work
  - Provides generic CRUD, bulk operations, soft delete
  - Transaction support

### 4. Specific Repository Implementations (`pkg/mongodb/repositories.go`)
- **Purpose**: Domain-specific repository implementations
- **Classes**: `UserRepository`, `ProductRepository`
- **Responsibilities**:
  - Extend base repository functionality
  - Implement domain-specific query methods
  - Use Unit of Work for complex queries

### 5. Unit of Work Layer (existing)
- **Purpose**: Transaction management and database operations
- **Responsibilities**:
  - Database transaction management
  - Actual MongoDB operations
  - Connection management
  - Query execution

## Key Features

### ✅ **Complete Type Safety**
```go
// Generic base repository works with any type
type IBaseRepository[T persistence.IEntity] interface {
    Insert(ctx context.Context, entity T) (T, error)
    Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error)
    // ... other methods
}
```

### ✅ **Business Logic Validation**
```go
func (s *UserService) CreateUser(ctx context.Context, email string, age int) (*persistence.User, error) {
    if email == "" {
        return nil, errors.New("email is required")
    }
    if age < 0 || age > 150 {
        return nil, errors.New("age must be between 0 and 150")
    }
    // ... business logic
}
```

### ✅ **Domain-Specific Queries**
```go
type IUserRepository interface {
    IBaseRepository[*persistence.User]
    FindByEmail(ctx context.Context, email string) (*persistence.User, error)
    FindActiveUsers(ctx context.Context) ([]*persistence.User, error)
    FindUsersByAgeRange(ctx context.Context, minAge, maxAge int) ([]*persistence.User, error)
    GetUserStats(ctx context.Context) (*persistence.UserStats, error)
}
```

### ✅ **Proper Identifier Usage**
```go
func (s *UserService) UpdateUser(ctx context.Context, user *persistence.User) (*persistence.User, error) {
    // Create proper identifier for update criteria
    updateCriteria := identifier.New().Equal("_id", user.GetID())
    return s.userRepo.Update(ctx, updateCriteria, user)
}
```

### ✅ **Repository Delegation Pattern**
```go
func (r *BaseRepository[T]) Insert(ctx context.Context, entity T) (T, error) {
    uow := r.factory.CreateWithContext(ctx)
    return uow.Insert(ctx, entity)
}
```

## Usage Examples

### Running the Demo

```bash
# Run with MongoDB (recommended)
go run examples/layered_architecture_demo.go

# The demo will:
# 1. Test MongoDB connection
# 2. If connected: demonstrate full operations
# 3. If not connected: show architecture explanation
```

### Creating and Using Services

```go
// 1. Setup Unit of Work factories
userUoWFactory, _ := mongodb.NewFactory[*persistence.User](config)
productUoWFactory, _ := mongodb.NewFactory[*persistence.Product](config)

// 2. Create Base Repositories
userBaseRepo := mongodb.NewBaseRepository[*persistence.User](userUoWFactory)
productBaseRepo := mongodb.NewBaseRepository[*persistence.Product](productUoWFactory)

// 3. Create Specific Repositories
userRepo := mongodb.NewUserRepository(userBaseRepo)
productRepo := mongodb.NewProductRepository(productBaseRepo)

// 4. Create Services
userService := services.NewUserService(userRepo)
productService := services.NewProductService(productRepo)

// 5. Use services (business logic layer)
user, err := userService.CreateUser(ctx, "alice@example.com", 30)
if err != nil {
    // Handle validation or business logic errors
}
```

## Service Operations

### User Service Operations
```go
// Basic CRUD with validation
user, err := userService.CreateUser(ctx, "user@email.com", 25)
user, err := userService.GetUserByID(ctx, userID)
user, err := userService.UpdateUser(ctx, user)
err := userService.DeleteUser(ctx, userID)

// Business operations
err := userService.DeactivateUser(ctx, userID)
user, err := userService.ActivateUser(ctx, userID)

// Domain queries
users, err := userService.GetAllActiveUsers(ctx)
users, err := userService.GetUsersByAgeRange(ctx, 18, 65)
stats, err := userService.GetUserStatistics(ctx)

// Bulk operations
users, err := userService.CreateUsers(ctx, userList)
err := userService.BulkDeactivateUsers(ctx, userIDs)
```

### Product Service Operations
```go
// Basic CRUD with validation
product, err := productService.CreateProduct(ctx, "Laptop", "Electronics", 999.99)
product, err := productService.GetProductByID(ctx, productID)
product, err := productService.UpdateProduct(ctx, product)
err := productService.DeleteProduct(ctx, productID)

// Business operations
product, err := productService.SetProductStock(ctx, productID, true)

// Domain queries
products, err := productService.GetProductsByCategory(ctx, "Electronics")
products, err := productService.GetInStockProducts(ctx)
products, err := productService.GetProductsByPriceRange(ctx, 10.0, 100.0)
stats, err := productService.GetProductStatistics(ctx)

// Bulk operations
products, err := productService.CreateProducts(ctx, productList)
err := productService.BulkUpdateStock(ctx, productIDs, true)
```

## Architecture Benefits

### 🎯 **Separation of Concerns**
- Business logic isolated in services
- Data access patterns in repositories
- Generic operations in base repository
- Transaction management in Unit of Work

### 🔒 **Type Safety**
- Generic interfaces ensure compile-time type checking
- No runtime type casting errors
- Clear contracts between layers

### **Testability**
- Each layer can be unit tested independently
- Mock interfaces for isolated testing
- Business logic testing without database

### 🔧 **Maintainability**
- Clear layer boundaries
- Easy to modify business rules
- Repository patterns for data access changes
- Unit of Work for transaction management changes

### 📈 **Scalability**
- Easy to add new entities (extend base repository)
- Service layer handles complex business operations
- Repository layer provides efficient data access

## Testing Strategy

### Unit Testing Services
```go
func TestUserService_CreateUser(t *testing.T) {
    // Mock repository
    mockRepo := &MockUserRepository{}
    service := services.NewUserService(mockRepo)
    
    // Test business logic validation
    _, err := service.CreateUser(ctx, "", 25)
    assert.Error(t, err, "should validate email")
    
    _, err = service.CreateUser(ctx, "test@test.com", -1)
    assert.Error(t, err, "should validate age")
}
```

### Integration Testing
```go
func TestLayeredArchitecture_Integration(t *testing.T) {
    // Setup real repositories and services
    // Test complete flow from service to database
    // Verify data persistence and retrieval
}
```

## Performance Considerations

- **Connection Pooling**: Unit of Work handles MongoDB connection pooling
- **Bulk Operations**: Optimized bulk insert/update operations
- **Query Optimization**: Repository layer provides efficient query patterns
- **Transaction Management**: Unit of Work ensures ACID properties

## Error Handling

- **Service Layer**: Business logic errors with meaningful messages
- **Repository Layer**: Data access errors
- **Base Repository**: Generic operation errors
- **Unit of Work**: Database connection and transaction errors

## Future Enhancements

- [ ] Add caching layer between services and repositories
- [ ] Implement event sourcing for audit trails
- [ ] Add query result pagination
- [ ] Implement soft delete with restore functionality
- [ ] Add comprehensive logging and metrics
- [ ] Implement repository-level caching strategies

## File Structure

```
pkg/
├── services/
│   └── services.go           # Business logic layer
├── persistence/
│   └── repositories.go       # Repository interfaces and domain types
├── mongodb/
│   ├── base_repository.go    # Generic base repository implementation
│   └── repositories.go       # Specific repository implementations
└── (existing packages...)

examples/
└── layered_architecture_demo.go   # Complete demo application
```

This architecture provides a robust, maintainable, and scalable foundation for MongoDB-based applications with clear separation of concerns and type safety throughout all layers.
