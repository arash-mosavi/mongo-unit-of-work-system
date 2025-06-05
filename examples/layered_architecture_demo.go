package main

import (
	"context"
	"fmt"
	"log"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/mongodb"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/services"
)

func main() {
	fmt.Println("MongoDB Layered Architecture Demo")
	fmt.Println("=================================")
	fmt.Println("Demonstrating: Service → Repository → Base Repository → Unit of Work → Database")
	fmt.Println()

	// Step 1: Initialize MongoDB configuration
	config := mongodb.NewConfig()
	config.Database = "layered_architecture_demo"
	config.Host = "localhost"
	config.Port = 27017

	fmt.Printf("Configuration: %s:%d/%s\n", config.Host, config.Port, config.Database)

	// Step 2: Create Unit of Work factories
	userUoWFactory, err := mongodb.NewFactory[*persistence.User](config)
	if err != nil {
		log.Fatalf("Failed to create user UoW factory: %v", err)
	}

	productUoWFactory, err := mongodb.NewFactory[*persistence.Product](config)
	if err != nil {
		log.Fatalf("Failed to create product UoW factory: %v", err)
	}

	fmt.Println("Unit of Work factories created")

	// Step 3: Create Base Repositories (delegates to Unit of Work)
	userBaseRepo := mongodb.NewBaseRepository[*persistence.User](userUoWFactory)
	productBaseRepo := mongodb.NewBaseRepository[*persistence.Product](productUoWFactory)

	fmt.Println("Base repositories created")

	// Step 4: Create Specific Repositories (extends base repositories)
	userRepo := mongodb.NewUserRepository(userBaseRepo)
	productRepo := mongodb.NewProductRepository(productBaseRepo)

	fmt.Println("Specific repositories created")

	// Step 5: Create Services (contains business logic)
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)

	fmt.Println("Service layer created")
	fmt.Println()

	// Test connection and demonstrate operations
	ctx := context.Background()

	fmt.Println("Testing MongoDB Connection:")
	fmt.Println("=============================")

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("MongoDB operation failed (expected if no MongoDB): %v\n", r)
			fmt.Println()
			demonstrateOfflineArchitecture(userService, productService)
			return
		}
	}()

	// Test with a simple operation
	testUser, err := userService.CreateUser(ctx, "test@example.com", 25)
	if err != nil {
		fmt.Printf("MongoDB connection failed: %v\n", err)
		fmt.Println()
		demonstrateOfflineArchitecture(userService, productService)
		return
	}

	fmt.Println("MongoDB connection successful!")
	fmt.Println()

	// Clean up test user
	if testUser != nil {
		userService.DeleteUser(ctx, testUser.GetID())
	}

	// Demonstrate full architecture with MongoDB
	demonstrateWithMongoDB(ctx, userService, productService)
}

func demonstrateWithMongoDB(ctx context.Context, userService services.IUserService, productService services.IProductService) {
	fmt.Println("Demonstrating Layered Architecture with MongoDB:")
	fmt.Println("==================================================")

	// 1. User Management through Service Layer
	fmt.Println("\n1. User Management (Service → Repository → UoW → MongoDB)")
	fmt.Println("-----------------------------------------------------------")

	// Create users through service layer (with validation)
	user1, err := userService.CreateUser(ctx, "alice@example.com", 30)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}
	fmt.Printf("Created user: %s (ID: %s)\n", user1.Email, user1.GetID().Hex())

	user2, err := userService.CreateUser(ctx, "bob@example.com", 25)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}
	fmt.Printf("Created user: %s (ID: %s)\n", user2.Email, user2.GetID().Hex())

	// Try to create duplicate user (business logic validation)
	_, err = userService.CreateUser(ctx, "alice@example.com", 28)
	if err != nil {
		fmt.Printf("Business logic validation: %v\n", err)
	}

	// 2. Product Management through Service Layer
	fmt.Println("\n2. Product Management (Service → Repository → UoW → MongoDB)")
	fmt.Println("-------------------------------------------------------------")

	product1, err := productService.CreateProduct(ctx, "Laptop", "Electronics", 999.99)
	if err != nil {
		fmt.Printf("Error creating product: %v\n", err)
		return
	}
	fmt.Printf("Created product: %s (ID: %s)\n", product1.GetName(), product1.GetID().Hex())

	product2, err := productService.CreateProduct(ctx, "Mouse", "Electronics", 29.99)
	if err != nil {
		fmt.Printf("Error creating product: %v\n", err)
		return
	}
	fmt.Printf("Created product: %s (ID: %s)\n", product2.GetName(), product2.GetID().Hex())

	// 3. Demonstrate Business Logic Operations
	fmt.Println("\n3. Business Logic Operations")
	fmt.Println("----------------------------")

	// Update user through service (with validation)
	user1.Age = 31
	updatedUser, err := userService.UpdateUser(ctx, user1)
	if err != nil {
		fmt.Printf("Error updating user: %v\n", err)
	} else {
		fmt.Printf("Updated user age: %d\n", updatedUser.Age)
	}

	// Deactivate user (business operation)
	err = userService.DeactivateUser(ctx, user2.GetID())
	if err != nil {
		fmt.Printf("Error deactivating user: %v\n", err)
	} else {
		fmt.Println("User deactivated")
	}

	// Update product stock
	updatedProduct, err := productService.SetProductStock(ctx, product2.GetID(), false)
	if err != nil {
		fmt.Printf("Error updating product stock: %v\n", err)
	} else {
		fmt.Printf("Product stock updated: InStock=%t\n", updatedProduct.InStock)
	}

	// 4. Query Operations through Repository Layer
	fmt.Println("\n4. Query Operations (Service → Repository → UoW)")
	fmt.Println("------------------------------------------------")

	// Get active users
	activeUsers, err := userService.GetAllActiveUsers(ctx)
	if err != nil {
		fmt.Printf("Error getting active users: %v\n", err)
	} else {
		fmt.Printf("Found %d active users\n", len(activeUsers))
	}

	// Get products by category
	electronicsProducts, err := productService.GetProductsByCategory(ctx, "Electronics")
	if err != nil {
		fmt.Printf("Error getting products by category: %v\n", err)
	} else {
		fmt.Printf("Found %d electronics products\n", len(electronicsProducts))
	}

	// Get in-stock products
	inStockProducts, err := productService.GetInStockProducts(ctx)
	if err != nil {
		fmt.Printf("Error getting in-stock products: %v\n", err)
	} else {
		fmt.Printf("Found %d in-stock products\n", len(inStockProducts))
	}

	// 5. Bulk Operations
	fmt.Println("\n5. Bulk Operations")
	fmt.Println("------------------")

	// Create multiple users
	bulkUsers := []*persistence.User{
		{Email: "charlie@example.com", Age: 35, Active: true},
		{Email: "diana@example.com", Age: 28, Active: true},
	}

	// Set names and slugs for bulk users
	for i, user := range bulkUsers {
		user.SetName(fmt.Sprintf("User_%d", i+3))
		user.SetSlug(fmt.Sprintf("user-%d", i+3))
	}

	createdUsers, err := userService.CreateUsers(ctx, bulkUsers)
	if err != nil {
		fmt.Printf("Error creating bulk users: %v\n", err)
	} else {
		fmt.Printf("Created %d users in bulk\n", len(createdUsers))
	}

	// 6. Statistics and Advanced Queries
	fmt.Println("\n6. Statistics and Advanced Queries")
	fmt.Println("----------------------------------")

	userStats, err := userService.GetUserStatistics(ctx)
	if err != nil {
		fmt.Printf("Error getting user statistics: %v\n", err)
	} else {
		fmt.Printf("User statistics: Total=%d, Active=%d, Average Age=%.1f\n",
			userStats.TotalUsers, userStats.ActiveUsers, userStats.AverageAge)
	}

	productStats, err := productService.GetProductStatistics(ctx)
	if err != nil {
		fmt.Printf("Error getting product statistics: %v\n", err)
	} else {
		fmt.Printf("Product statistics: Total=%d, InStock=%d, Average Price=%.2f\n",
			productStats.TotalProducts, productStats.InStockProducts, productStats.AveragePrice)
	}

	// 7. Cleanup
	fmt.Println("\n7. Cleanup")
	fmt.Println("----------")

	// Delete created data
	for _, user := range append(activeUsers, createdUsers...) {
		if user != nil {
			err := userService.DeleteUser(ctx, user.GetID())
			if err != nil {
				fmt.Printf("Error deleting user %s: %v\n", user.GetID().Hex(), err)
			}
		}
	}

	err = productService.DeleteProduct(ctx, product1.GetID())
	if err != nil {
		fmt.Printf("Error deleting product: %v\n", err)
	}

	err = productService.DeleteProduct(ctx, product2.GetID())
	if err != nil {
		fmt.Printf("Error deleting product: %v\n", err)
	}

	fmt.Println("Cleanup completed")
	fmt.Println("\nLayered architecture demonstration completed!")
}

func demonstrateOfflineArchitecture(userService services.IUserService, productService services.IProductService) {
	fmt.Println("Demonstrating Layered Architecture (Offline Mode):")
	fmt.Println("====================================================")
	fmt.Println("This demonstrates the architectural layers without requiring MongoDB:")
	fmt.Println()

	fmt.Println("Architecture Flow:")
	fmt.Println("1. Client calls Service Layer")
	fmt.Println("   └── Service Layer (business logic, validation)")
	fmt.Println("2. Service calls Repository Interface")
	fmt.Println("   └── Specific Repository (user/product specific methods)")
	fmt.Println("3. Repository extends Base Repository")
	fmt.Println("   └── Base Repository (generic CRUD operations)")
	fmt.Println("4. Base Repository delegates to Unit of Work")
	fmt.Println("   └── Unit of Work (transaction management, data operations)")
	fmt.Println("5. Unit of Work interacts with MongoDB")
	fmt.Println("   └── MongoDB Database (data persistence)")
	fmt.Println()

	fmt.Println("Service Interfaces:")
	fmt.Println("   • IUserService - User business logic")
	fmt.Println("   • IProductService - Product business logic")
	fmt.Println()

	fmt.Println("Repository Interfaces:")
	fmt.Println("   • IBaseRepository[T] - Generic CRUD operations")
	fmt.Println("   • IUserRepository - User-specific queries")
	fmt.Println("   • IProductRepository - Product-specific queries")
	fmt.Println()

	fmt.Println("Implementation Layers:")
	fmt.Println("   • UserService/ProductService - Business logic")
	fmt.Println("   • UserRepository/ProductRepository - Data access")
	fmt.Println("   • BaseRepository[T] - Generic operations")
	fmt.Println("   • Unit of Work - Transaction management")
	fmt.Println("   • MongoDB - Data persistence")
	fmt.Println()

	fmt.Println("🔍 Key Benefits:")
	fmt.Println("   • Clear separation of concerns")
	fmt.Println("   • Business logic isolated in services")
	fmt.Println("   • Repository pattern for data access")
	fmt.Println("   • Unit of Work for transaction management")
	fmt.Println("   • Type safety with generics")
	fmt.Println("   • Easy testing with interface mocking")
	fmt.Println("   • Flexible and maintainable architecture")
	fmt.Println()

	// Demonstrate business logic validation
	fmt.Println("Example Business Logic Validation:")
	ctx := context.Background()

	// This will fail due to validation even without MongoDB
	_, err := userService.CreateUser(ctx, "", 25)
	if err != nil {
		fmt.Printf("   Email validation: %v\n", err)
	}

	_, err = userService.CreateUser(ctx, "test@example.com", -5)
	if err != nil {
		fmt.Printf("   Age validation: %v\n", err)
	}

	_, err = productService.CreateProduct(ctx, "", "Electronics", 100)
	if err != nil {
		fmt.Printf("   Product name validation: %v\n", err)
	}

	_, err = productService.CreateProduct(ctx, "Laptop", "Electronics", -100)
	if err != nil {
		fmt.Printf("   Price validation: %v\n", err)
	}

	fmt.Println()
	fmt.Println("The layered architecture is working correctly!")
	fmt.Println("   Connect to MongoDB to see full database operations.")
}
