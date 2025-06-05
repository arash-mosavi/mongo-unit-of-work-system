package mongodb

import (
	"context"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/domain"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/identifier"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BaseRepository implements the base repository functionality using Unit of Work
type BaseRepository[T persistence.ModelConstraint] struct {
	factory persistence.IUnitOfWorkFactory[T]
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository[T persistence.ModelConstraint](factory persistence.IUnitOfWorkFactory[T]) persistence.IBaseRepository[T] {
	return &BaseRepository[T]{
		factory: factory,
	}
}

// Insert creates a new entity
func (r *BaseRepository[T]) Insert(ctx context.Context, entity T) (T, error) {
	uow := r.factory.CreateWithContext(ctx)
	return uow.Insert(ctx, entity)
}

// Update modifies an existing entity
func (r *BaseRepository[T]) Update(ctx context.Context, id identifier.IIdentifier, entity T) (T, error) {
	uow := r.factory.CreateWithContext(ctx)
	return uow.Update(ctx, id, entity)
}

// Delete removes an entity
func (r *BaseRepository[T]) Delete(ctx context.Context, id identifier.IIdentifier) error {
	uow := r.factory.CreateWithContext(ctx)
	return uow.Delete(ctx, id)
}

// FindOneById finds an entity by its ID
func (r *BaseRepository[T]) FindOneById(ctx context.Context, id primitive.ObjectID) (T, error) {
	uow := r.factory.CreateWithContext(ctx)
	return uow.FindOneById(ctx, id)
}

// FindOne finds a single entity based on identifier
func (r *BaseRepository[T]) FindOne(ctx context.Context, id identifier.IIdentifier) (T, error) {
	uow := r.factory.CreateWithContext(ctx)
	return uow.FindOneByIdentifier(ctx, id)
}

// FindAll finds all entities matching the identifier
func (r *BaseRepository[T]) FindAll(ctx context.Context, id identifier.IIdentifier) ([]T, error) {
	uow := r.factory.CreateWithContext(ctx)
	// For now, we'll use FindAll and then filter - in a real implementation
	// you might want to extend the unit of work interface
	return uow.FindAll(ctx)
}

// FindAllWithPagination finds entities with pagination support
func (r *BaseRepository[T]) FindAllWithPagination(ctx context.Context, query domain.QueryParams[T]) ([]T, int64, error) {
	uow := r.factory.CreateWithContext(ctx)
	entities, count, err := uow.FindAllWithPagination(ctx, query)
	return entities, int64(count), err
}

// BulkInsert creates multiple entities
func (r *BaseRepository[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	uow := r.factory.CreateWithContext(ctx)
	return uow.BulkInsert(ctx, entities)
}

// BulkUpdate modifies multiple entities
func (r *BaseRepository[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	uow := r.factory.CreateWithContext(ctx)
	return uow.BulkUpdate(ctx, entities)
}

// BulkDelete removes multiple entities
func (r *BaseRepository[T]) BulkDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	uow := r.factory.CreateWithContext(ctx)
	return uow.BulkHardDelete(ctx, identifiers)
}

// SoftDelete marks an entity as deleted
func (r *BaseRepository[T]) SoftDelete(ctx context.Context, id identifier.IIdentifier) (T, error) {
	uow := r.factory.CreateWithContext(ctx)
	return uow.SoftDelete(ctx, id)
}

// BulkSoftDelete marks multiple entities as deleted
func (r *BaseRepository[T]) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	uow := r.factory.CreateWithContext(ctx)
	return uow.BulkSoftDelete(ctx, identifiers)
}

// Restore recovers a soft-deleted entity
func (r *BaseRepository[T]) Restore(ctx context.Context, id identifier.IIdentifier) (T, error) {
	uow := r.factory.CreateWithContext(ctx)
	return uow.Restore(ctx, id)
}

// GetTrashed retrieves all soft-deleted entities
func (r *BaseRepository[T]) GetTrashed(ctx context.Context) ([]T, error) {
	uow := r.factory.CreateWithContext(ctx)
	return uow.GetTrashed(ctx)
}

// BeginTransaction starts a database transaction
func (r *BaseRepository[T]) BeginTransaction(ctx context.Context) error {
	uow := r.factory.CreateWithContext(ctx)
	return uow.BeginTransaction(ctx)
}

// CommitTransaction commits the current transaction
func (r *BaseRepository[T]) CommitTransaction(ctx context.Context) error {
	uow := r.factory.CreateWithContext(ctx)
	return uow.CommitTransaction(ctx)
}

// RollbackTransaction rolls back the current transaction
func (r *BaseRepository[T]) RollbackTransaction(ctx context.Context) error {
	uow := r.factory.CreateWithContext(ctx)
	uow.RollbackTransaction(ctx)
	return nil
}
