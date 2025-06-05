# MongoDB Unit of Work System

This is a Go SDK for working with MongoDB using the Unit of Work pattern.

## Installation

```bash
go get github.com/arash-mosavi/mongo-unit-of-work-system
```

## Quick Start

```go
config := mongodb.NewConfig()
config.Host = "localhost"
config.Port = 27017
config.Database = "myapp"

factory, _ := mongodb.NewFactory[*models.Customer](config)
repo := mongodb.NewCustomerRepository(mongodb.NewBaseRepository[*models.Customer](factory))
service := services.NewCustomerService(repo)

ctx := context.Background()
customer, err := service.CreateCustomer(ctx, "John Doe", "john@example.com", "123456", 5000.0)
if err != nil {
    log.Fatal(err)
}
log.Printf("Customer created: %s", customer.Email)
```

## Highlights

- Keeps business logic focused and reusable
- Repository layer handles all DB-related logic
- BaseRepository + Factory make setup easier
- Fully testable structure
- Clean and type-safe

## Folder Layout

```
pkg/
  domain/           // Entity definitions
  mongodb/          // MongoDB logic and factories
  persistence/      // Shared interfaces
  errors/           // Typed errors
  services/         // Business logic layer
examples/           // Usage examples
test/               // Integration tests
```

## Principles

- Keep business logic out of DB layer
- Use services for orchestration
- Repositories are responsible for DB access only
- Unit of Work handles transactions and connection lifecycle

## Testing

```bash
go test ./...
go test -v ./pkg/mongodb
go test -bench=. ./pkg/mongodb
```

## When to Use

- Your app needs well-structured MongoDB operations
- You want to avoid scattering DB calls throughout the code
- You want unit test coverage without needing live Mongo
- You care about clean code and testable services

## Performance

Supports pooling, efficient filtering, and batch writes.

## Error Handling

- Validation errors for invalid input
- Duplicate entry checks
- Not found handling
- Clear wrapped error types

---

