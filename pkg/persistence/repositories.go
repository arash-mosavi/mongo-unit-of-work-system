package persistence

import (
	"context"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/domain"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/identifier"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IBaseRepository[T ModelConstraint] interface {
	Insert(ctx context.Context, entity T) (T, error)
	Update(ctx context.Context, id identifier.IIdentifier, entity T) (T, error)
	Delete(ctx context.Context, id identifier.IIdentifier) error
	FindOneById(ctx context.Context, id primitive.ObjectID) (T, error)
	FindOne(ctx context.Context, id identifier.IIdentifier) (T, error)
	FindAll(ctx context.Context, id identifier.IIdentifier) ([]T, error)
	FindAllWithPagination(ctx context.Context, query domain.QueryParams[T]) ([]T, int64, error)

	BulkInsert(ctx context.Context, entities []T) ([]T, error)
	BulkUpdate(ctx context.Context, entities []T) ([]T, error)
	BulkDelete(ctx context.Context, identifiers []identifier.IIdentifier) error

	SoftDelete(ctx context.Context, id identifier.IIdentifier) (T, error)
	BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error
	Restore(ctx context.Context, id identifier.IIdentifier) (T, error)
	GetTrashed(ctx context.Context) ([]T, error)

	BeginTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error
}

type IUserRepository interface {
	IBaseRepository[*User]

	FindByEmail(ctx context.Context, email string) (*User, error)
	FindActiveUsers(ctx context.Context) ([]*User, error)
	FindUsersByAgeRange(ctx context.Context, minAge, maxAge int) ([]*User, error)
	GetUserStats(ctx context.Context) (*UserStats, error)
}

type IProductRepository interface {
	IBaseRepository[*Product]

	FindByCategory(ctx context.Context, category string) ([]*Product, error)
	FindInStockProducts(ctx context.Context) ([]*Product, error)
	FindProductsByPriceRange(ctx context.Context, minPrice, maxPrice float64) ([]*Product, error)
	GetProductStats(ctx context.Context) (*ProductStats, error)
}

type User struct {
	domain.BaseEntity `bson:",inline"`
	Email             string `bson:"email" json:"email"`
	Age               int    `bson:"age" json:"age"`
	Active            bool   `bson:"active" json:"active"`
}

type Product struct {
	domain.BaseEntity `bson:",inline"`
	Price             float64 `bson:"price" json:"price"`
	Category          string  `bson:"category" json:"category"`
	InStock           bool    `bson:"inStock" json:"inStock"`
}

type UserStats struct {
	TotalUsers  int64   `json:"totalUsers"`
	ActiveUsers int64   `json:"activeUsers"`
	AverageAge  float64 `json:"averageAge"`
}

type ProductStats struct {
	TotalProducts   int64    `json:"totalProducts"`
	InStockProducts int64    `json:"inStockProducts"`
	AveragePrice    float64  `json:"averagePrice"`
	Categories      []string `json:"categories"`
}
