package mongodb

import (
	"context"
	"fmt"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
)

// Factory implements IUnitOfWorkFactory for MongoDB
type Factory[T persistence.ModelConstraint] struct {
	config *Config
}

// NewFactory creates a new MongoDB unit of work factory
func NewFactory[T persistence.ModelConstraint](config *Config) (*Factory[T], error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Factory[T]{
		config: config,
	}, nil
}

// Create creates a new unit of work instance
func (f *Factory[T]) Create() persistence.IUnitOfWork[T] {
	uow, err := NewUnitOfWork[T](f.config)
	if err != nil {
		// In a real implementation, you might want to handle this differently
		// For now, we'll panic as this indicates a serious configuration error
		panic(fmt.Sprintf("failed to create unit of work: %v", err))
	}
	return uow
}

// CreateWithContext creates a new unit of work instance with context
func (f *Factory[T]) CreateWithContext(ctx context.Context) persistence.IUnitOfWork[T] {
	return f.Create()
}

// CreateWithTransaction creates a new unit of work and starts a transaction
func (f *Factory[T]) CreateWithTransaction(ctx context.Context) (persistence.IUnitOfWork[T], error) {
	uow := f.CreateWithContext(ctx)
	err := uow.BeginTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	return uow, nil
}
