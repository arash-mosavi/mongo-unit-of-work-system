package mongodb

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/domain"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/identifier"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/persistence"
)

type UnitOfWork[T persistence.ModelConstraint] struct {
	client         *mongo.Client
	database       *mongo.Database
	session        mongo.Session
	ctx            context.Context
	repositories   map[string]interface{}
	mu             sync.RWMutex
	inTx           bool
	collectionName string
}

func NewUnitOfWork[T domain.BaseModel](config *Config) (*UnitOfWork[T], error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.ConnectionString())
	clientOptions.SetMaxPoolSize(config.MaxPoolSize)
	clientOptions.SetMinPoolSize(config.MinPoolSize)
	clientOptions.SetMaxConnIdleTime(config.MaxIdleTime)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	var zero T
	collectionName := getCollectionName(zero)

	return &UnitOfWork[T]{
		client:         client,
		database:       database,
		ctx:            context.Background(),
		repositories:   make(map[string]interface{}),
		collectionName: collectionName,
	}, nil
}

func getCollectionName(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()

	return strings.ToLower(name) + "s"
}

func (uow *UnitOfWork[T]) getCollection() *mongo.Collection {
	return uow.database.Collection(uow.collectionName)
}

func (uow *UnitOfWork[T]) BeginTransaction(ctx context.Context) error {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if uow.inTx {
		return fmt.Errorf("transaction already in progress")
	}

	session, err := uow.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	err = session.StartTransaction()
	if err != nil {
		session.EndSession(ctx)
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	uow.session = session
	uow.ctx = mongo.NewSessionContext(ctx, session)
	uow.inTx = true

	return nil
}

func (uow *UnitOfWork[T]) CommitTransaction(ctx context.Context) error {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if !uow.inTx {
		return fmt.Errorf("no transaction in progress")
	}

	err := uow.session.CommitTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	uow.session.EndSession(ctx)
	uow.session = nil
	uow.ctx = context.Background()
	uow.inTx = false

	return nil
}

func (uow *UnitOfWork[T]) RollbackTransaction(ctx context.Context) {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if !uow.inTx {
		return
	}

	uow.session.AbortTransaction(ctx)
	uow.session.EndSession(ctx)
	uow.session = nil
	uow.ctx = context.Background()
	uow.inTx = false
}

func (uow *UnitOfWork[T]) FindAll(ctx context.Context) ([]T, error) {
	collection := uow.getCollection()

	filter := bson.M{"deletedAt": bson.M{"$exists": false}}

	cursor, err := collection.Find(uow.getContext(ctx), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find all: %w", err)
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	return results, nil
}

func (uow *UnitOfWork[T]) FindAllWithPagination(ctx context.Context, query domain.QueryParams[T]) ([]T, uint, error) {
	collection := uow.getCollection()

	filter := bson.M{"deletedAt": bson.M{"$exists": false}}
	if !isZeroValue(query.Filter) {
		filterBSON := uow.buildFilterFromModel(query.Filter)
		for k, v := range filterBSON {
			filter[k] = v
		}
	}

	total, err := collection.CountDocuments(uow.getContext(ctx), filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
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
		return nil, 0, fmt.Errorf("failed to find with pagination: %w", err)
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to decode results: %w", err)
	}

	return results, uint(total), nil
}

func (uow *UnitOfWork[T]) FindOne(ctx context.Context, filter T) (T, error) {
	var zero T
	collection := uow.getCollection()

	filterBSON := uow.buildFilterFromModel(filter)

	filterBSON["deletedAt"] = bson.M{"$exists": false}

	var result T
	err := collection.FindOne(uow.getContext(ctx), filterBSON).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return zero, fmt.Errorf("entity not found")
		}
		return zero, fmt.Errorf("failed to find one: %w", err)
	}

	return result, nil
}

func (uow *UnitOfWork[T]) FindOneById(ctx context.Context, id primitive.ObjectID) (T, error) {
	var zero T
	collection := uow.getCollection()

	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	var result T
	err := collection.FindOne(uow.getContext(ctx), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return zero, fmt.Errorf("entity not found")
		}
		return zero, fmt.Errorf("failed to find by id: %w", err)
	}

	return result, nil
}

func (uow *UnitOfWork[T]) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var zero T
	collection := uow.getCollection()

	filter := identifier.ToBSON()

	if !identifier.Has("deletedAt") {
		filter["deletedAt"] = bson.M{"$exists": false}
	}

	var result T
	err := collection.FindOne(uow.getContext(ctx), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return zero, fmt.Errorf("entity not found")
		}
		return zero, fmt.Errorf("failed to find by identifier: %w", err)
	}

	return result, nil
}

func (uow *UnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model domain.BaseModel, field string, value interface{}) (primitive.ObjectID, error) {
	collection := uow.getCollection()

	filter := bson.M{
		field:       value,
		"deletedAt": bson.M{"$exists": false},
	}

	var result bson.M
	err := collection.FindOne(uow.getContext(ctx), filter, options.FindOne().SetProjection(bson.M{"_id": 1})).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return primitive.NilObjectID, fmt.Errorf("entity not found")
		}
		return primitive.NilObjectID, fmt.Errorf("failed to resolve ID: %w", err)
	}

	id, ok := result["_id"].(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("invalid ObjectID type")
	}

	return id, nil
}

func (uow *UnitOfWork[T]) Insert(ctx context.Context, entity T) (T, error) {
	collection := uow.getCollection()

	now := time.Now()
	uow.setEntityTimestamp(entity, "createdAt", now)
	uow.setEntityTimestamp(entity, "updatedAt", now)

	if entity.GetID().IsZero() {
		entity.SetID(primitive.NewObjectID())
	}

	_, err := collection.InsertOne(uow.getContext(ctx), entity)
	if err != nil {
		return entity, fmt.Errorf("failed to insert: %w", err)
	}

	return entity, nil
}

func (uow *UnitOfWork[T]) Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error) {
	collection := uow.getCollection()

	filter := identifier.ToBSON()

	filter["deletedAt"] = bson.M{"$exists": false}

	uow.setEntityTimestamp(entity, "updatedAt", time.Now())

	update := bson.M{"$set": entity}

	result := collection.FindOneAndUpdate(
		uow.getContext(ctx),
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var updated T
	if err := result.Decode(&updated); err != nil {
		if err == mongo.ErrNoDocuments {
			return entity, fmt.Errorf("entity not found")
		}
		return entity, fmt.Errorf("failed to update: %w", err)
	}

	return updated, nil
}

func (uow *UnitOfWork[T]) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	collection := uow.getCollection()

	filter := identifier.ToBSON()

	result, err := collection.DeleteOne(uow.getContext(ctx), filter)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("entity not found")
	}

	return nil
}

func (uow *UnitOfWork[T]) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var zero T
	collection := uow.getCollection()

	filter := identifier.ToBSON()
	filter["deletedAt"] = bson.M{"$exists": false}

	update := bson.M{
		"$set": bson.M{
			"deletedAt": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	result := collection.FindOneAndUpdate(
		uow.getContext(ctx),
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var updated T
	if err := result.Decode(&updated); err != nil {
		if err == mongo.ErrNoDocuments {
			return zero, fmt.Errorf("entity not found")
		}
		return zero, fmt.Errorf("failed to soft delete: %w", err)
	}

	return updated, nil
}

func (uow *UnitOfWork[T]) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var zero T
	collection := uow.getCollection()

	filter := identifier.ToBSON()

	var deleted T
	err := collection.FindOneAndDelete(uow.getContext(ctx), filter).Decode(&deleted)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return zero, fmt.Errorf("entity not found")
		}
		return zero, fmt.Errorf("failed to hard delete: %w", err)
	}

	return deleted, nil
}

func (uow *UnitOfWork[T]) getContext(ctx context.Context) context.Context {
	if uow.inTx && uow.session != nil {
		return uow.ctx
	}
	return ctx
}

func (uow *UnitOfWork[T]) buildFilterFromModel(model T) bson.M {
	filter := bson.M{}

	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanInterface() {
			continue
		}

		fieldName := fieldType.Name
		if tag := fieldType.Tag.Get("bson"); tag != "" && tag != "-" {
			fieldName = strings.Split(tag, ",")[0]
		}

		if field.IsZero() {
			continue
		}

		filter[fieldName] = field.Interface()
	}

	return filter
}

func (uow *UnitOfWork[T]) setEntityTimestamp(entity T, fieldName string, timestamp time.Time) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.CanSet() {
		return
	}

	field := v.FieldByName(strings.Title(fieldName))
	if !field.IsValid() || !field.CanSet() {
		return
	}

	if field.Type() == reflect.TypeOf(time.Time{}) {
		field.Set(reflect.ValueOf(timestamp))
	}
}
