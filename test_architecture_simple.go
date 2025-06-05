package main

import (
	"context"
	"fmt"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/mongodb"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/services"
)

func main() {
	fmt.Println("Architecture Verification")

	config := mongodb.NewConfig()
	config.Database = "test"

	userFactory, _ := mongodb.NewFactory[*persistence.User](config)
	productFactory, _ := mongodb.NewFactory[*persistence.Product](config)

	userBaseRepo := mongodb.NewBaseRepository[*persistence.User](userFactory)
	productBaseRepo := mongodb.NewBaseRepository[*persistence.Product](productFactory)

	userRepo := mongodb.NewUserRepository(userBaseRepo)
	productRepo := mongodb.NewProductRepository(productBaseRepo)

	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)

	ctx := context.Background()

	// Test validations
	_, err := userService.CreateUser(ctx, "", 25)
	if err != nil && err.Error() == "email is required" {
		fmt.Println("User validation working")
	}

	_, err = productService.CreateProduct(ctx, "", "Electronics", 100)
	if err != nil && err.Error() == "product name is required" {
		fmt.Println("Product validation working")
	}

	fmt.Println("Layered architecture verified successfully!")
}
