package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/identifier"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IUserService interface {
	CreateUser(ctx context.Context, email string, age int) (*persistence.User, error)
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*persistence.User, error)
	GetUserByEmail(ctx context.Context, email string) (*persistence.User, error)
	UpdateUser(ctx context.Context, user *persistence.User) (*persistence.User, error)
	DeactivateUser(ctx context.Context, id primitive.ObjectID) error
	ActivateUser(ctx context.Context, id primitive.ObjectID) (*persistence.User, error)
	DeleteUser(ctx context.Context, id primitive.ObjectID) error

	GetAllActiveUsers(ctx context.Context) ([]*persistence.User, error)
	GetUsersByAgeRange(ctx context.Context, minAge, maxAge int) ([]*persistence.User, error)
	GetUserStatistics(ctx context.Context) (*persistence.UserStats, error)

	CreateUsers(ctx context.Context, users []*persistence.User) ([]*persistence.User, error)
	BulkDeactivateUsers(ctx context.Context, userIDs []primitive.ObjectID) error
}

type IProductService interface {
	CreateProduct(ctx context.Context, name, category string, price float64) (*persistence.Product, error)
	GetProductByID(ctx context.Context, id primitive.ObjectID) (*persistence.Product, error)
	UpdateProduct(ctx context.Context, product *persistence.Product) (*persistence.Product, error)
	SetProductStock(ctx context.Context, id primitive.ObjectID, inStock bool) (*persistence.Product, error)
	DeleteProduct(ctx context.Context, id primitive.ObjectID) error

	GetProductsByCategory(ctx context.Context, category string) ([]*persistence.Product, error)
	GetInStockProducts(ctx context.Context) ([]*persistence.Product, error)
	GetProductsByPriceRange(ctx context.Context, minPrice, maxPrice float64) ([]*persistence.Product, error)
	GetProductStatistics(ctx context.Context) (*persistence.ProductStats, error)

	CreateProducts(ctx context.Context, products []*persistence.Product) ([]*persistence.Product, error)
	BulkUpdateStock(ctx context.Context, productIDs []primitive.ObjectID, inStock bool) error
}

type UserService struct {
	userRepo persistence.IUserRepository
}

func NewUserService(userRepo persistence.IUserRepository) IUserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, email string, age int) (*persistence.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if age < 0 || age > 150 {
		return nil, errors.New("age must be between 0 and 150")
	}

	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	user := &persistence.User{
		Email:  email,
		Age:    age,
		Active: true,
	}
	user.SetName(fmt.Sprintf("User_%s", email))
	user.SetSlug(fmt.Sprintf("user-%s", email))

	return s.userRepo.Insert(ctx, user)
}

func (s *UserService) GetUserByID(ctx context.Context, id primitive.ObjectID) (*persistence.User, error) {
	return s.userRepo.FindOneById(ctx, id)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*persistence.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

func (s *UserService) UpdateUser(ctx context.Context, user *persistence.User) (*persistence.User, error) {
	if user.Age < 0 || user.Age > 150 {
		return nil, errors.New("age must be between 0 and 150")
	}

	updateCriteria := identifier.New().Equal("_id", user.GetID())
	return s.userRepo.Update(ctx, updateCriteria, user)
}

func (s *UserService) DeactivateUser(ctx context.Context, id primitive.ObjectID) error {
	user, err := s.userRepo.FindOneById(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Active = false
	updateCriteria := identifier.New().Equal("_id", id)
	_, err = s.userRepo.Update(ctx, updateCriteria, user)
	return err
}

func (s *UserService) ActivateUser(ctx context.Context, id primitive.ObjectID) (*persistence.User, error) {
	user, err := s.userRepo.FindOneById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Active = true
	updateCriteria := identifier.New().Equal("_id", id)
	return s.userRepo.Update(ctx, updateCriteria, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	deleteCriteria := identifier.New().Equal("_id", id)
	return s.userRepo.Delete(ctx, deleteCriteria)
}

func (s *UserService) GetAllActiveUsers(ctx context.Context) ([]*persistence.User, error) {
	return s.userRepo.FindActiveUsers(ctx)
}

func (s *UserService) GetUsersByAgeRange(ctx context.Context, minAge, maxAge int) ([]*persistence.User, error) {
	if minAge < 0 || maxAge > 150 || minAge > maxAge {
		return nil, errors.New("invalid age range")
	}
	return s.userRepo.FindUsersByAgeRange(ctx, minAge, maxAge)
}

func (s *UserService) GetUserStatistics(ctx context.Context) (*persistence.UserStats, error) {
	return s.userRepo.GetUserStats(ctx)
}

func (s *UserService) CreateUsers(ctx context.Context, users []*persistence.User) ([]*persistence.User, error) {

	for i, user := range users {
		if user.Email == "" {
			return nil, fmt.Errorf("user %d: email is required", i)
		}
		if user.Age < 0 || user.Age > 150 {
			return nil, fmt.Errorf("user %d: age must be between 0 and 150", i)
		}
	}

	return s.userRepo.BulkInsert(ctx, users)
}

func (s *UserService) BulkDeactivateUsers(ctx context.Context, userIDs []primitive.ObjectID) error {

	for _, id := range userIDs {
		if err := s.DeactivateUser(ctx, id); err != nil {
			return fmt.Errorf("failed to deactivate user %s: %w", id.Hex(), err)
		}
	}
	return nil
}

type ProductService struct {
	productRepo persistence.IProductRepository
}

func NewProductService(productRepo persistence.IProductRepository) IProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, name, category string, price float64) (*persistence.Product, error) {
	if name == "" {
		return nil, errors.New("product name is required")
	}
	if category == "" {
		return nil, errors.New("product category is required")
	}
	if price < 0 {
		return nil, errors.New("price must be non-negative")
	}

	product := &persistence.Product{
		Price:    price,
		Category: category,
		InStock:  true,
	}
	product.SetName(name)
	product.SetSlug(fmt.Sprintf("%s-%s", category, name))

	return s.productRepo.Insert(ctx, product)
}

func (s *ProductService) GetProductByID(ctx context.Context, id primitive.ObjectID) (*persistence.Product, error) {
	return s.productRepo.FindOneById(ctx, id)
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *persistence.Product) (*persistence.Product, error) {
	if product.Price < 0 {
		return nil, errors.New("price must be non-negative")
	}

	updateCriteria := identifier.New().Equal("_id", product.GetID())
	return s.productRepo.Update(ctx, updateCriteria, product)
}

func (s *ProductService) SetProductStock(ctx context.Context, id primitive.ObjectID, inStock bool) (*persistence.Product, error) {
	product, err := s.productRepo.FindOneById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	product.InStock = inStock
	updateCriteria := identifier.New().Equal("_id", id)
	return s.productRepo.Update(ctx, updateCriteria, product)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id primitive.ObjectID) error {
	deleteCriteria := identifier.New().Equal("_id", id)
	return s.productRepo.Delete(ctx, deleteCriteria)
}

func (s *ProductService) GetProductsByCategory(ctx context.Context, category string) ([]*persistence.Product, error) {
	if category == "" {
		return nil, errors.New("category is required")
	}
	return s.productRepo.FindByCategory(ctx, category)
}

func (s *ProductService) GetInStockProducts(ctx context.Context) ([]*persistence.Product, error) {
	return s.productRepo.FindInStockProducts(ctx)
}

func (s *ProductService) GetProductsByPriceRange(ctx context.Context, minPrice, maxPrice float64) ([]*persistence.Product, error) {
	if minPrice < 0 || maxPrice < 0 || minPrice > maxPrice {
		return nil, errors.New("invalid price range")
	}
	return s.productRepo.FindProductsByPriceRange(ctx, minPrice, maxPrice)
}

func (s *ProductService) GetProductStatistics(ctx context.Context) (*persistence.ProductStats, error) {
	return s.productRepo.GetProductStats(ctx)
}

func (s *ProductService) CreateProducts(ctx context.Context, products []*persistence.Product) ([]*persistence.Product, error) {

	for i, product := range products {
		if product.GetName() == "" {
			return nil, fmt.Errorf("product %d: name is required", i)
		}
		if product.Category == "" {
			return nil, fmt.Errorf("product %d: category is required", i)
		}
		if product.Price < 0 {
			return nil, fmt.Errorf("product %d: price must be non-negative", i)
		}
	}

	return s.productRepo.BulkInsert(ctx, products)
}

func (s *ProductService) BulkUpdateStock(ctx context.Context, productIDs []primitive.ObjectID, inStock bool) error {

	for _, id := range productIDs {
		if _, err := s.SetProductStock(ctx, id, inStock); err != nil {
			return fmt.Errorf("failed to update stock for product %s: %w", id.Hex(), err)
		}
	}
	return nil
}
