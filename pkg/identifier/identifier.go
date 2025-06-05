package identifier

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IIdentifier interface {
	Equal(field string, value interface{}) IIdentifier
	In(field string, values []interface{}) IIdentifier
	Like(field string, pattern string) IIdentifier
	GreaterThan(field string, value interface{}) IIdentifier
	LessThan(field string, value interface{}) IIdentifier
	Between(field string, start, end interface{}) IIdentifier
	IsNull(field string) IIdentifier
	IsNotNull(field string) IIdentifier

	Add(key string, value interface{}) IIdentifier
	AddIf(condition bool, key string, value interface{}) IIdentifier

	ToBSON() bson.M
	ToObjectID(field string) (primitive.ObjectID, error)

	ToMap() map[string]interface{}
	GetQuery() map[string]interface{}
	Has(key string) bool
	Get(key string) (interface{}, bool)
	String() string
}

type Identifier struct {
	query map[string]interface{}
}

func New() *Identifier {
	return &Identifier{
		query: make(map[string]interface{}),
	}
}

func (i *Identifier) Equal(field string, value interface{}) IIdentifier {
	i.query[field] = value
	return i
}

func (i *Identifier) In(field string, values []interface{}) IIdentifier {
	i.query[field+" IN"] = values
	return i
}

func (i *Identifier) Like(field string, pattern string) IIdentifier {
	i.query[field+" LIKE"] = pattern
	return i
}

func (i *Identifier) GreaterThan(field string, value interface{}) IIdentifier {
	i.query[field+" >"] = value
	return i
}

func (i *Identifier) LessThan(field string, value interface{}) IIdentifier {
	i.query[field+" <"] = value
	return i
}

func (i *Identifier) Between(field string, start, end interface{}) IIdentifier {
	i.query[field+" BETWEEN"] = []interface{}{start, end}
	return i
}

func (i *Identifier) IsNull(field string) IIdentifier {
	i.query[field+" IS NULL"] = true
	return i
}

func (i *Identifier) IsNotNull(field string) IIdentifier {
	i.query[field+" IS NOT NULL"] = true
	return i
}

func (i *Identifier) Add(key string, value interface{}) IIdentifier {
	i.query[key] = value
	return i
}

func (i *Identifier) AddIf(condition bool, key string, value interface{}) IIdentifier {
	if condition {
		i.query[key] = value
	}
	return i
}

func (i *Identifier) ToBSON() bson.M {
	filter := bson.M{}
	for key, value := range i.query {

		if strings.Contains(key, " >") {
			field := strings.TrimSuffix(key, " >")
			filter[field] = bson.M{"$gt": value}
		} else if strings.Contains(key, " <") {
			field := strings.TrimSuffix(key, " <")
			filter[field] = bson.M{"$lt": value}
		} else if strings.Contains(key, " IN") {
			field := strings.TrimSuffix(key, " IN")
			filter[field] = bson.M{"$in": value}
		} else if strings.Contains(key, " LIKE") {
			field := strings.TrimSuffix(key, " LIKE")
			filter[field] = bson.M{"$regex": value, "$options": "i"}
		} else if strings.Contains(key, " BETWEEN") {
			field := strings.TrimSuffix(key, " BETWEEN")
			if vals, ok := value.([]interface{}); ok && len(vals) == 2 {
				filter[field] = bson.M{"$gte": vals[0], "$lte": vals[1]}
			}
		} else if strings.Contains(key, " IS NULL") {
			field := strings.TrimSuffix(key, " IS NULL")
			filter[field] = bson.M{"$exists": false}
		} else if strings.Contains(key, " IS NOT NULL") {
			field := strings.TrimSuffix(key, " IS NOT NULL")
			filter[field] = bson.M{"$exists": true}
		} else {
			filter[key] = value
		}
	}
	return filter
}

func (i *Identifier) ToObjectID(field string) (primitive.ObjectID, error) {
	value, exists := i.query[field]
	if !exists {
		return primitive.NilObjectID, fmt.Errorf("field %s not found", field)
	}

	switch v := value.(type) {
	case primitive.ObjectID:
		return v, nil
	case string:
		return primitive.ObjectIDFromHex(v)
	default:
		return primitive.NilObjectID, fmt.Errorf("cannot convert %T to ObjectID", value)
	}
}

func (i *Identifier) ToMap() map[string]interface{} {
	return i.query
}

func (i *Identifier) GetQuery() map[string]interface{} {
	result := make(map[string]interface{}, len(i.query))
	for k, v := range i.query {
		result[k] = v
	}
	return result
}

func (i *Identifier) String() string {
	if len(i.query) == 0 {
		return "{}"
	}

	var builder strings.Builder
	builder.WriteString("{")

	first := true
	for key, value := range i.query {
		if !first {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%s: %v", key, value))
		first = false
	}

	builder.WriteString("}")
	return builder.String()
}

func (i *Identifier) Has(key string) bool {
	_, exists := i.query[key]
	return exists
}

func (i *Identifier) Get(key string) (interface{}, bool) {
	value, exists := i.query[key]
	return value, exists
}

func ByID(id interface{}) IIdentifier {
	return New().Equal("_id", id)
}

func BySlug(slug string) IIdentifier {
	return New().Equal("slug", slug)
}

func ByEmail(email string) IIdentifier {
	return New().Equal("email", email)
}

func Active() IIdentifier {
	return New().Equal("active", true)
}

func Inactive() IIdentifier {
	return New().Equal("active", false)
}

func NotDeleted() IIdentifier {
	return New().IsNull("deletedAt")
}

func Deleted() IIdentifier {
	return New().IsNotNull("deletedAt")
}
