package test

import (
	"context"
	"testing"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/mongodb"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/services"
)

// TestLayeredArchitectureIntegration tests the complete layered architecture flow
func TestLayeredArchitectureIntegration(t *testing.T) {
	// Test configuration
	config := mongodb.NewConfig()
	config.Database = "test_layered_architecture"
	config.Host = "localhost"
	config.Port = 27017

	// Create Unit of Work factories
	userUoWFactory, err := mongodb.NewFactory[*persistence.User](config)
	if err != nil {
		t.Logf("Skipping MongoDB integration test (no connection): %v", err)
		testBusinessLogicOnly(t)
		return
	}

	productUoWFactory, err := mongodb.NewFactory[*persistence.Product](config)
	if err != nil {
		t.Logf("Skipping MongoDB integration test (no connection): %v", err)
		testBusinessLogicOnly(t)
		return
	}

	// Create Base Repositories
	userBaseRepo := mongodb.NewBaseRepository[*persistence.User](userUoWFactory)
	productBaseRepo := mongodb.NewBaseRepository[*persistence.Product](productUoWFactory)

	// Create Specific Repositories
	userRepo := mongodb.NewUserRepository(userBaseRepo)
	productRepo := mongodb.NewProductRepository(productBaseRepo)

	// Create Services
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)

	ctx := context.Background()

	// Test User Service Operations
	t.Run("UserService", func(t *testing.T) {
		// Test user creation with validation
		user, err := userService.CreateUser(ctx, "test@example.com", 25)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if user.Email != "test@example.com" {
			t.Errorf("Expected email test@example.com, got %s", user.Email)
		}

		if user.Age != 25 {
			t.Errorf("Expected age 25, got %d", user.Age)
		}

		// Test duplicate email validation
		_, err = userService.CreateUser(ctx, "test@example.com", 30)
		if err == nil {
			t.Error("Expected error for duplicate email")
		}

		// Test user retrieval
		retrievedUser, err := userService.GetUserByID(ctx, user.GetID())
		if err != nil {
			t.Fatalf("Failed to retrieve user: %v", err)
		}

		if retrievedUser.Email != user.Email {
			t.Errorf("Retrieved user email mismatch")
		}

		// Test user update
		retrievedUser.Age = 26
		updatedUser, err := userService.UpdateUser(ctx, retrievedUser)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		if updatedUser.Age != 26 {
			t.Errorf("Expected updated age 26, got %d", updatedUser.Age)
		}

		// Test deactivation
		err = userService.DeactivateUser(ctx, user.GetID())
		if err != nil {
			t.Fatalf("Failed to deactivate user: %v", err)
		}

		// Cleanup
		err = userService.DeleteUser(ctx, user.GetID())
		if err != nil {
			t.Logf("Warning: Failed to cleanup user: %v", err)
		}
	})

	// Test Product Service Operations
	t.Run("ProductService", func(t *testing.T) {
		// Test product creation with validation
		product, err := productService.CreateProduct(ctx, "Test Laptop", "Electronics", 999.99)
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}

		if product.GetName() != "Test Laptop" {
			t.Errorf("Expected name 'Test Laptop', got %s", product.GetName())
		}

		if product.Category != "Electronics" {
			t.Errorf("Expected category 'Electronics', got %s", product.Category)
		}

		if product.Price != 999.99 {
			t.Errorf("Expected price 999.99, got %f", product.Price)
		}

		// Test product retrieval
		retrievedProduct, err := productService.GetProductByID(ctx, product.GetID())
		if err != nil {
			t.Fatalf("Failed to retrieve product: %v", err)
		}

		if retrievedProduct.GetName() != product.GetName() {
			t.Errorf("Retrieved product name mismatch")
		}

		// Test stock update
		updatedProduct, err := productService.SetProductStock(ctx, product.GetID(), false)
		if err != nil {
			t.Fatalf("Failed to update product stock: %v", err)
		}

		if updatedProduct.InStock != false {
			t.Errorf("Expected stock to be false, got %t", updatedProduct.InStock)
		}

		// Cleanup
		err = productService.DeleteProduct(ctx, product.GetID())
		if err != nil {
			t.Logf("Warning: Failed to cleanup product: %v", err)
		}
	})
}

// testBusinessLogicOnly tests the business logic validation without MongoDB
func testBusinessLogicOnly(t *testing.T) {
	t.Log("Testing business logic validation without MongoDB")

	// This would require mock repositories, but for now we just test
	// that the architecture can be constructed without errors
	config := mongodb.NewConfig()
	config.Database = "test_db"

	// The factory creation should work even without MongoDB connection
	// (the connection error only occurs when actually using the factory)
	userUoWFactory, err := mongodb.NewFactory[*persistence.User](config)
	if err != nil {
		t.Fatalf("Failed to create user factory: %v", err)
	}

	productUoWFactory, err := mongodb.NewFactory[*persistence.Product](config)
	if err != nil {
		t.Fatalf("Failed to create product factory: %v", err)
	}

	// Create the layered architecture
	userBaseRepo := mongodb.NewBaseRepository[*persistence.User](userUoWFactory)
	productBaseRepo := mongodb.NewBaseRepository[*persistence.Product](productUoWFactory)

	userRepo := mongodb.NewUserRepository(userBaseRepo)
	productRepo := mongodb.NewProductRepository(productBaseRepo)

	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)

	// Test business logic validation
	ctx := context.Background()

	// Test user validation
	_, err = userService.CreateUser(ctx, "", 25)
	if err == nil || err.Error() != "email is required" {
		t.Errorf("Expected email validation error, got: %v", err)
	}

	_, err = userService.CreateUser(ctx, "test@test.com", -1)
	if err == nil || err.Error() != "age must be between 0 and 150" {
		t.Errorf("Expected age validation error, got: %v", err)
	}

	// Test product validation
	_, err = productService.CreateProduct(ctx, "", "Electronics", 100)
	if err == nil || err.Error() != "product name is required" {
		t.Errorf("Expected product name validation error, got: %v", err)
	}

	_, err = productService.CreateProduct(ctx, "Laptop", "", 100)
	if err == nil || err.Error() != "product category is required" {
		t.Errorf("Expected category validation error, got: %v", err)
	}

	_, err = productService.CreateProduct(ctx, "Laptop", "Electronics", -100)
	if err == nil || err.Error() != "price must be non-negative" {
		t.Errorf("Expected price validation error, got: %v", err)
	}

	t.Log("Business logic validation working correctly")
	t.Log("Layered architecture structure verified")
}
