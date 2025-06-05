package mongodb

import (
	"context"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/identifier"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
)

type UserRepository struct {
	persistence.IBaseRepository[*persistence.User]
}

func NewUserRepository(baseRepo persistence.IBaseRepository[*persistence.User]) persistence.IUserRepository {
	return &UserRepository{
		IBaseRepository: baseRepo,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*persistence.User, error) {
	id := identifier.New().Equal("email", email)
	return r.FindOne(ctx, id)
}

func (r *UserRepository) FindActiveUsers(ctx context.Context) ([]*persistence.User, error) {
	id := identifier.New().Equal("active", true).Equal("deletedAt", nil)
	return r.FindAll(ctx, id)
}

func (r *UserRepository) FindUsersByAgeRange(ctx context.Context, minAge, maxAge int) ([]*persistence.User, error) {

	id := identifier.New().Between("age", minAge, maxAge).Equal("deletedAt", nil)
	return r.FindAll(ctx, id)
}

func (r *UserRepository) GetUserStats(ctx context.Context) (*persistence.UserStats, error) {

	allUsers, err := r.FindAll(ctx, identifier.New().Equal("deletedAt", nil))
	if err != nil {
		return nil, err
	}

	activeUsers, err := r.FindActiveUsers(ctx)
	if err != nil {
		return nil, err
	}

	var totalAge int64
	for _, user := range allUsers {
		totalAge += int64(user.Age)
	}

	var averageAge float64
	if len(allUsers) > 0 {
		averageAge = float64(totalAge) / float64(len(allUsers))
	}

	return &persistence.UserStats{
		TotalUsers:  int64(len(allUsers)),
		ActiveUsers: int64(len(activeUsers)),
		AverageAge:  averageAge,
	}, nil
}

type ProductRepository struct {
	persistence.IBaseRepository[*persistence.Product]
}

func NewProductRepository(baseRepo persistence.IBaseRepository[*persistence.Product]) persistence.IProductRepository {
	return &ProductRepository{
		IBaseRepository: baseRepo,
	}
}

func (r *ProductRepository) FindByCategory(ctx context.Context, category string) ([]*persistence.Product, error) {
	id := identifier.New().Equal("category", category).Equal("deletedAt", nil)
	return r.FindAll(ctx, id)
}

func (r *ProductRepository) FindInStockProducts(ctx context.Context) ([]*persistence.Product, error) {
	id := identifier.New().Equal("inStock", true).Equal("deletedAt", nil)
	return r.FindAll(ctx, id)
}

func (r *ProductRepository) FindProductsByPriceRange(ctx context.Context, minPrice, maxPrice float64) ([]*persistence.Product, error) {

	id := identifier.New().Between("price", minPrice, maxPrice).Equal("deletedAt", nil)
	return r.FindAll(ctx, id)
}

func (r *ProductRepository) GetProductStats(ctx context.Context) (*persistence.ProductStats, error) {

	allProducts, err := r.FindAll(ctx, identifier.New().Equal("deletedAt", nil))
	if err != nil {
		return nil, err
	}

	inStockProducts, err := r.FindInStockProducts(ctx)
	if err != nil {
		return nil, err
	}

	var totalPrice float64
	categorySet := make(map[string]bool)
	for _, product := range allProducts {
		totalPrice += product.Price
		categorySet[product.Category] = true
	}

	var averagePrice float64
	if len(allProducts) > 0 {
		averagePrice = totalPrice / float64(len(allProducts))
	}

	var categories []string
	for category := range categorySet {
		categories = append(categories, category)
	}

	return &persistence.ProductStats{
		TotalProducts:   int64(len(allProducts)),
		InStockProducts: int64(len(inStockProducts)),
		AveragePrice:    averagePrice,
		Categories:      categories,
	}, nil
}
