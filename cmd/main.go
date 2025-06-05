package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/domain"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/identifier"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/mongodb"
)

// User represents a user entity
type User struct {
	domain.BaseEntity `bson:",inline"`
	Email             string `bson:"email" json:"email"`
	Age               int    `bson:"age" json:"age"`
	Active            bool   `bson:"active" json:"active"`
}

// Product represents a product entity
type Product struct {
	domain.BaseEntity `bson:",inline"`
	Price             float64 `bson:"price" json:"price"`
	Category          string  `bson:"category" json:"category"`
	InStock           bool    `bson:"inStock" json:"inStock"`
}

func main() {
	fmt.Println("MongoDB Unit of Work System")
	fmt.Println("=====================================")

	// Create MongoDB configuration
	config := mongodb.NewConfig()
	config.Database = "unit_of_work_demo"
	config.Host = "localhost"
	config.Port = 27017

	fmt.Printf("Configuration created for %s:%d/%s\n", config.Host, config.Port, config.Database)

	// Create factory
	userFactory, err := mongodb.NewFactory[*User](config)
	if err != nil {
		log.Fatalf("Failed to create user factory: %v", err)
	}

	productFactory, err := mongodb.NewFactory[*Product](config)
	if err != nil {
		log.Fatalf("Failed to create product factory: %v", err)
	}

	fmt.Println("Unit of Work factories created")

	// Test MongoDB connection
	fmt.Println("\nTesting MongoDB Connection:")
	fmt.Println("=============================")

	mongoConnected := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("MongoDB connection failed (expected if no MongoDB): %v\n", r)
			}
		}()

		ctx := context.Background()
		testUow := userFactory.CreateWithContext(ctx)

		testUser := &User{
			Email:  "connection-test@example.com",
			Age:    25,
			Active: true,
		}
		testUser.SetName("Connection Test")
		testUser.SetSlug("connection-test")

		_, err = testUow.Insert(ctx, testUser)
		if err != nil {
			fmt.Printf("MongoDB operation failed: %v\n", err)
		} else {
			fmt.Println("MongoDB connection successful!")
			mongoConnected = true
		}
	}()

	if mongoConnected {
		// Demonstrate operations with real MongoDB
		demonstrateBasicOperations(userFactory)

		// Demonstrate transactions
		demonstrateTransactions(userFactory, productFactory)

		// Demonstrate bulk operations
		demonstrateBulkOperations(userFactory)

		// Demonstrate soft delete and restore
		demonstrateSoftDeleteRestore(userFactory)
	} else {
		fmt.Println("Running demonstration in offline mode...")
		demonstrateOfflineFeatures()
	}

	fmt.Println("\nAll examples completed successfully!")
	fmt.Println("The MongoDB Unit of Work SDK is ready for use!")
}

func demonstrateBasicOperations(factory *mongodb.Factory[*User]) {
	fmt.Println("\nBasic Operations Demo:")
	fmt.Println("========================")

	ctx := context.Background()
	uow := factory.CreateWithContext(ctx)

	// Create a new user
	user := &User{
		Email:  "john@example.com",
		Age:    30,
		Active: true,
	}
	user.SetName("John Doe")
	user.SetSlug("john-doe")

	// Insert user
	createdUser, err := uow.Insert(ctx, user)
	if err != nil {
		log.Printf("Insert failed (expected if no MongoDB): %v", err)
		return
	}

	fmt.Printf("User created with ID: %s\n", createdUser.GetID().Hex())

	// Find user by ID
	foundUser, err := uow.FindOneById(ctx, createdUser.GetID())
	if err != nil {
		log.Printf("Find failed: %v", err)
		return
	}

	fmt.Printf("User found: %s (%s)\n", foundUser.GetName(), foundUser.Email)

	// Update user
	foundUser.Age = 31
	id := identifier.New().Equal("_id", foundUser.GetID())
	updatedUser, err := uow.Update(ctx, id, foundUser)
	if err != nil {
		log.Printf("Update failed: %v", err)
		return
	}

	fmt.Printf("User updated - new age: %d\n", updatedUser.Age)

	// Find with pagination
	query := domain.QueryParams[*User]{
		Limit:  10,
		Offset: 0,
		Sort: domain.SortMap{
			"createdAt": domain.SortDesc,
		},
	}

	users, total, err := uow.FindAllWithPagination(ctx, query)
	if err != nil {
		log.Printf("Pagination failed: %v", err)
		return
	}

	fmt.Printf("Found %d users (total: %d)\n", len(users), total)
}

func demonstrateTransactions(userFactory *mongodb.Factory[*User], productFactory *mongodb.Factory[*Product]) {
	fmt.Println("\n💳 Transaction Demo:")
	fmt.Println("===================")

	ctx := context.Background()

	// Create unit of work with transaction
	userUow, err := userFactory.CreateWithTransaction(ctx)
	if err != nil {
		log.Printf("Transaction creation failed (expected if no MongoDB): %v", err)
		return
	}

	productUow, err := productFactory.CreateWithTransaction(ctx)
	if err != nil {
		log.Printf("Transaction creation failed: %v", err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			userUow.RollbackTransaction(ctx)
			productUow.RollbackTransaction(ctx)
			fmt.Println("Transaction rolled back due to panic")
		}
	}()

	// Create user and product in transaction
	user := &User{
		Email:  "transactional@example.com",
		Age:    25,
		Active: true,
	}
	user.SetName("Trans User")

	product := &Product{
		Price:    99.99,
		Category: "Electronics",
		InStock:  true,
	}
	product.SetName("Sample Product")

	// Insert both entities
	_, err = userUow.Insert(ctx, user)
	if err != nil {
		log.Printf("User insert failed: %v", err)
		return
	}

	_, err = productUow.Insert(ctx, product)
	if err != nil {
		log.Printf("Product insert failed: %v", err)
		return
	}

	// Commit transactions
	err = userUow.CommitTransaction(ctx)
	if err != nil {
		log.Printf("User transaction commit failed: %v", err)
		return
	}

	err = productUow.CommitTransaction(ctx)
	if err != nil {
		log.Printf("Product transaction commit failed: %v", err)
		return
	}

	fmt.Println("Transaction completed successfully")
}

func demonstrateBulkOperations(factory *mongodb.Factory[*User]) {
	fmt.Println("\nBulk Operations Demo:")
	fmt.Println("=======================")

	ctx := context.Background()
	uow := factory.CreateWithContext(ctx)

	// Create multiple users
	users := []*User{
		{Email: "bulk1@example.com", Age: 25, Active: true},
		{Email: "bulk2@example.com", Age: 30, Active: true},
		{Email: "bulk3@example.com", Age: 35, Active: false},
	}

	for i := range users {
		users[i].SetName(fmt.Sprintf("Bulk User %d", i+1))
		users[i].SetSlug(fmt.Sprintf("bulk-user-%d", i+1))
	}

	// Bulk insert
	createdUsers, err := uow.BulkInsert(ctx, users)
	if err != nil {
		log.Printf("Bulk insert failed (expected if no MongoDB): %v", err)
		return
	}

	fmt.Printf("Bulk inserted %d users\n", len(createdUsers))

	// Update ages
	for i := range createdUsers {
		createdUsers[i].Age += 1
	}

	// Bulk update
	updatedUsers, err := uow.BulkUpdate(ctx, createdUsers)
	if err != nil {
		log.Printf("Bulk update failed: %v", err)
		return
	}

	fmt.Printf("Bulk updated %d users\n", len(updatedUsers))

	// Create identifiers for bulk delete
	var identifiers []identifier.IIdentifier
	for _, user := range updatedUsers {
		id := identifier.New().Equal("_id", user.GetID())
		identifiers = append(identifiers, id)
	}

	// Bulk soft delete
	err = uow.BulkSoftDelete(ctx, identifiers)
	if err != nil {
		log.Printf("Bulk soft delete failed: %v", err)
		return
	}

	fmt.Printf("Bulk soft deleted %d users\n", len(identifiers))
}

func demonstrateSoftDeleteRestore(factory *mongodb.Factory[*User]) {
	fmt.Println("\nSoft Delete & Restore Demo:")
	fmt.Println("==============================")

	ctx := context.Background()
	uow := factory.CreateWithContext(ctx)

	// Create a user
	user := &User{
		Email:  "softdelete@example.com",
		Age:    28,
		Active: true,
	}
	user.SetName("Soft Delete User")
	user.SetSlug("soft-delete-user")

	createdUser, err := uow.Insert(ctx, user)
	if err != nil {
		log.Printf("Insert failed (expected if no MongoDB): %v", err)
		return
	}

	fmt.Printf("Created user: %s\n", createdUser.GetName())

	// Soft delete the user
	id := identifier.New().Equal("_id", createdUser.GetID())
	deletedUser, err := uow.SoftDelete(ctx, id)
	if err != nil {
		log.Printf("Soft delete failed: %v", err)
		return
	}

	fmt.Printf("Soft deleted user: %s (deleted at: %v)\n",
		deletedUser.GetName(), deletedUser.GetDeletedAt())

	// Try to find the user (should not be found in normal queries)
	_, err = uow.FindOneById(ctx, createdUser.GetID())
	if err != nil {
		fmt.Println("User not found in normal queries (expected)")
	}

	// Get trashed users
	trashedUsers, err := uow.GetTrashed(ctx)
	if err != nil {
		log.Printf("Get trashed failed: %v", err)
		return
	}

	fmt.Printf("Found %d trashed users\n", len(trashedUsers))

	// Restore the user
	restoredUser, err := uow.Restore(ctx, id)
	if err != nil {
		log.Printf("Restore failed: %v", err)
		return
	}

	fmt.Printf("Restored user: %s\n", restoredUser.GetName())

	// Verify user can be found again
	foundUser, err := uow.FindOneById(ctx, createdUser.GetID())
	if err != nil {
		log.Printf("Find after restore failed: %v", err)
		return
	}

	fmt.Printf("User found after restore: %s\n", foundUser.GetName())
}

// Performance benchmarking example
func runBenchmarks() {
	fmt.Println("\n⚡ Performance Benchmarks:")
	fmt.Println("=========================")

	// Benchmark identifier creation
	start := time.Now()
	for i := 0; i < 10000; i++ {
		id := identifier.New().
			Equal("name", "test").
			GreaterThan("age", 18).
			Like("email", "test@").
			In("status", []interface{}{"active", "pending"})
		_ = id.ToBSON()
	}
	duration := time.Since(start)
	fmt.Printf("10,000 identifier operations: %v (%.2f ops/ms)\n",
		duration, float64(10000)/float64(duration.Milliseconds()))

	// Benchmark config string generation
	config := &mongodb.Config{
		Host:        "localhost",
		Port:        27017,
		Database:    "test",
		Username:    "user",
		Password:    "pass",
		AuthSource:  "admin",
		MaxPoolSize: 100,
		MinPoolSize: 5,
		SSL:         true,
		ReplicaSet:  "rs0",
	}

	start = time.Now()
	for i := 0; i < 10000; i++ {
		_ = config.ConnectionString()
	}
	duration = time.Since(start)
	fmt.Printf("10,000 connection string generations: %v (%.2f ops/ms)\n",
		duration, float64(10000)/float64(duration.Milliseconds()))
}

func demonstrateOfflineFeatures() {
	fmt.Println("\nMongoDB Unit of Work SDK Features:")
	fmt.Println("====================================")

	fmt.Println("Generic Type System:")
	fmt.Println("  • Type-safe operations with *User and *Product")
	fmt.Println("  • Compile-time type checking")
	fmt.Println("  • Interface-based design")

	fmt.Println("\nQuery Builder (Identifier):")
	fmt.Println("  • Fluent API for building MongoDB queries")
	fmt.Println("  • Type-safe query operations")

	// Demonstrate identifier creation
	id := identifier.New().
		Equal("name", "John Doe").
		GreaterThan("age", 18).
		Like("email", "@example.com").
		In("status", []interface{}{"active", "pending"})

	bsonQuery := id.ToBSON()
	fmt.Printf("  • Sample BSON query: %v\n", bsonQuery)

	fmt.Println("\nConfiguration Management:")
	config := mongodb.NewConfig()
	config.Database = "demo"
	config.Host = "localhost"
	config.Port = 27017
	config.Username = "user"
	config.Password = "pass"
	config.MaxPoolSize = 100
	config.SSL = true

	fmt.Printf("  • Connection string: %s\n", config.ConnectionString())

	fmt.Println("\nEntity System:")
	user := &User{
		Email:  "demo@example.com",
		Age:    30,
		Active: true,
	}
	user.SetName("Demo User")
	user.SetSlug("demo-user")

	fmt.Printf("  • Entity created: %s (%s)\n", user.GetName(), user.GetSlug())
	fmt.Printf("  • Entity ID: %s\n", user.GetID().Hex())

	product := &Product{
		Price:    99.99,
		Category: "Electronics",
		InStock:  true,
	}
	product.SetName("Demo Product")
	product.SetSlug("demo-product")

	fmt.Printf("  • Product created: %s ($%.2f)\n", product.GetName(), product.Price)

	fmt.Println("\nFactory Pattern:")
	fmt.Println("  • Type-safe factory creation")
	fmt.Println("  • Context-aware unit of work creation")
	fmt.Println("  • Transaction support")

	fmt.Println("\nUnit of Work Pattern:")
	fmt.Println("  • CRUD operations: Insert, Update, Delete, Find")
	fmt.Println("  • Bulk operations: BulkInsert, BulkUpdate, BulkDelete")
	fmt.Println("  • Soft delete: SoftDelete, Restore, GetTrashed")
	fmt.Println("  • Pagination: FindAllWithPagination")
	fmt.Println("  • Transactions: CreateWithTransaction, Commit, Rollback")

	// Run performance benchmarks
	runBenchmarks()
}
