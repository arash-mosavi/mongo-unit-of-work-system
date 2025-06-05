package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/domain"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/identifier"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
)

func (uow *UnitOfWork[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	collection := uow.getCollection()
	now := time.Now()

	documents := make([]interface{}, len(entities))
	for i, entity := range entities {

		uow.setEntityTimestamp(entity, "createdAt", now)
		uow.setEntityTimestamp(entity, "updatedAt", now)

		if entity.GetID().IsZero() {
			entity.SetID(primitive.NewObjectID())
		}

		documents[i] = entity
		entities[i] = entity
	}

	_, err := collection.InsertMany(uow.getContext(ctx), documents)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk insert: %w", err)
	}

	return entities, nil
}

func (uow *UnitOfWork[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	collection := uow.getCollection()
	now := time.Now()

	var models []mongo.WriteModel
	for _, entity := range entities {
		uow.setEntityTimestamp(entity, "updatedAt", now)

		filter := bson.M{
			"_id":       entity.GetID(),
			"deletedAt": bson.M{"$exists": false},
		}
		update := bson.M{"$set": entity}

		model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	result, err := collection.BulkWrite(uow.getContext(ctx), models, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update: %w", err)
	}

	if result.ModifiedCount != int64(len(entities)) {
		return entities, fmt.Errorf("not all entities were updated: modified %d out of %d", result.ModifiedCount, len(entities))
	}

	return entities, nil
}

func (uow *UnitOfWork[T]) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	collection := uow.getCollection()
	now := time.Now()

	var models []mongo.WriteModel
	for _, id := range identifiers {
		filter := id.ToBSON()
		filter["deletedAt"] = bson.M{"$exists": false}

		update := bson.M{
			"$set": bson.M{
				"deletedAt": now,
				"updatedAt": now,
			},
		}

		model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(uow.getContext(ctx), models, opts)
	if err != nil {
		return fmt.Errorf("failed to bulk soft delete: %w", err)
	}

	return nil
}

func (uow *UnitOfWork[T]) BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	if len(identifiers) == 0 {
		return nil
	}

	collection := uow.getCollection()

	var models []mongo.WriteModel
	for _, id := range identifiers {
		filter := id.ToBSON()
		model := mongo.NewDeleteOneModel().SetFilter(filter)
		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(uow.getContext(ctx), models, opts)
	if err != nil {
		return fmt.Errorf("failed to bulk hard delete: %w", err)
	}

	return nil
}

func (uow *UnitOfWork[T]) GetTrashed(ctx context.Context) ([]T, error) {
	collection := uow.getCollection()

	filter := bson.M{"deletedAt": bson.M{"$exists": true}}

	cursor, err := collection.Find(uow.getContext(ctx), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get trashed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode trashed results: %w", err)
	}

	return results, nil
}

func (uow *UnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, query domain.QueryParams[T]) ([]T, uint, error) {
	collection := uow.getCollection()

	filter := bson.M{"deletedAt": bson.M{"$exists": true}}
	if !isZeroValue(query.Filter) {
		filterBSON := uow.buildFilterFromModel(query.Filter)
		for k, v := range filterBSON {
			if k != "deletedAt" {
				filter[k] = v
			}
		}
	}

	total, err := collection.CountDocuments(uow.getContext(ctx), filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count trashed documents: %w", err)
	}

	opts := options.Find()
	if query.Limit > 0 {
		opts.SetLimit(int64(query.Limit))
	}
	if query.Offset > 0 {
		opts.SetSkip(int64(query.Offset))
	}

	if query.Sort != nil && len(query.Sort) > 0 {
		sort := bson.D{}
		for field, direction := range query.Sort {
			if direction == domain.SortAsc {
				sort = append(sort, bson.E{Key: field, Value: 1})
			} else {
				sort = append(sort, bson.E{Key: field, Value: -1})
			}
		}
		opts.SetSort(sort)
	}

	cursor, err := collection.Find(uow.getContext(ctx), filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find trashed with pagination: %w", err)
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to decode trashed results: %w", err)
	}

	return results, uint(total), nil
}

func (uow *UnitOfWork[T]) Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var zero T
	collection := uow.getCollection()

	filter := identifier.ToBSON()
	filter["deletedAt"] = bson.M{"$exists": true}

	update := bson.M{
		"$unset": bson.M{"deletedAt": ""},
		"$set":   bson.M{"updatedAt": time.Now()},
	}

	result := collection.FindOneAndUpdate(
		uow.getContext(ctx),
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var restored T
	if err := result.Decode(&restored); err != nil {
		if err == mongo.ErrNoDocuments {
			return zero, fmt.Errorf("entity not found in trash")
		}
		return zero, fmt.Errorf("failed to restore: %w", err)
	}

	return restored, nil
}

func (uow *UnitOfWork[T]) RestoreAll(ctx context.Context) error {
	collection := uow.getCollection()

	filter := bson.M{"deletedAt": bson.M{"$exists": true}}
	update := bson.M{
		"$unset": bson.M{"deletedAt": ""},
		"$set":   bson.M{"updatedAt": time.Now()},
	}

	_, err := collection.UpdateMany(uow.getContext(ctx), filter, update)
	if err != nil {
		return fmt.Errorf("failed to restore all: %w", err)
	}

	return nil
}

func (uow *UnitOfWork[T]) GetRepository(entityType string) interface{} {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.repositories[entityType]
}

func (uow *UnitOfWork[T]) RegisterRepository(entityType string, repo interface{}) {
	uow.mu.Lock()
	defer uow.mu.Unlock()
	uow.repositories[entityType] = repo
}

func (uow *UnitOfWork[T]) WithContext(ctx context.Context) persistence.IUnitOfWork[T] {
	newUow := &UnitOfWork[T]{
		client:         uow.client,
		database:       uow.database,
		session:        uow.session,
		ctx:            ctx,
		repositories:   uow.repositories,
		inTx:           uow.inTx,
		collectionName: uow.collectionName,
	}
	return newUow
}

func (uow *UnitOfWork[T]) GetContext() context.Context {
	return uow.ctx
}

func (uow *UnitOfWork[T]) IsInTransaction() bool {
	return uow.inTx
}

func (uow *UnitOfWork[T]) Close(ctx context.Context) error {
	if uow.inTx {
		uow.RollbackTransaction(ctx)
	}
	return uow.client.Disconnect(ctx)
}
