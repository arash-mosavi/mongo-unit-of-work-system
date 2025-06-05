package mongodb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/domain"
	"github.com/arash-mosavi/mongo-unit-of-work-system/pkg/identifier"
)

type TestUser struct {
	domain.BaseEntity `bson:",inline"`
	Email             string `bson:"email" json:"email"`
	Age               int    `bson:"age" json:"age"`
	Active            bool   `bson:"active" json:"active"`
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Host:     "localhost",
				Port:     27017,
				Database: "test",
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: &Config{
				Host:     "",
				Port:     27017,
				Database: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				Host:     "localhost",
				Port:     0,
				Database: "test",
			},
			wantErr: true,
		},
		{
			name: "empty database",
			config: &Config{
				Host:     "localhost",
				Port:     27017,
				Database: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "basic connection",
			config: &Config{
				Host:     "localhost",
				Port:     27017,
				Database: "test",
			},
			expected: "mongodb://localhost:27017/test",
		},
		{
			name: "with authentication",
			config: &Config{
				Host:       "localhost",
				Port:       27017,
				Database:   "test",
				Username:   "user",
				Password:   "pass",
				AuthSource: "admin",
			},
			expected: "mongodb://user:pass@localhost:27017/test?authSource=admin",
		},
		{
			name: "with SSL",
			config: &Config{
				Host:     "localhost",
				Port:     27017,
				Database: "test",
				SSL:      true,
			},
			expected: "mongodb://localhost:27017/test?ssl=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ConnectionString()
			assert.Contains(t, result, "mongodb://")
			assert.Contains(t, result, tt.config.Database)
		})
	}
}

func TestIdentifier_ToBSON(t *testing.T) {
	tests := []struct {
		name       string
		identifier *identifier.Identifier
		expected   int
	}{
		{
			name:       "equal condition",
			identifier: identifier.New().Equal("name", "test").(*identifier.Identifier),
			expected:   1,
		},
		{
			name:       "greater than condition",
			identifier: identifier.New().GreaterThan("age", 18).(*identifier.Identifier),
			expected:   1,
		},
		{
			name:       "like condition",
			identifier: identifier.New().Like("email", "test@").(*identifier.Identifier),
			expected:   1,
		},
		{
			name: "multiple conditions",
			identifier: identifier.New().
				Equal("active", true).
				GreaterThan("age", 18).(*identifier.Identifier),
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.identifier.ToBSON()
			assert.Len(t, filter, tt.expected)
		})
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 27017, config.Port)
	assert.Equal(t, "test", config.Database)
	assert.Equal(t, "admin", config.AuthSource)
	assert.Equal(t, uint64(100), config.MaxPoolSize)
	assert.Equal(t, uint64(5), config.MinPoolSize)
	assert.Equal(t, 30*time.Second, config.MaxIdleTime)
	assert.Equal(t, 10*time.Second, config.Timeout)
	assert.False(t, config.SSL)
}

func TestFactory_Create(t *testing.T) {

	config := NewConfig()
	config.Database = "test_db"

	factory, err := NewFactory[*TestUser](config)
	require.NoError(t, err)
	assert.NotNil(t, factory)
}

func BenchmarkIdentifier_ToBSON(b *testing.B) {
	id := identifier.New().
		Equal("name", "test").
		GreaterThan("age", 18).
		Like("email", "test@").
		In("status", []interface{}{"active", "pending"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = id.ToBSON()
	}
}

func BenchmarkConfig_ConnectionString(b *testing.B) {
	config := &Config{
		Host:        "localhost",
		Port:        27017,
		Database:    "test",
		Username:    "user",
		Password:    "pass",
		AuthSource:  "admin",
		MaxPoolSize: 100,
		MinPoolSize: 5,
		SSL:         true,
		ReplicaSet:  "rs0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.ConnectionString()
	}
}

func TestUnitOfWork_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Skip("Integration test requires MongoDB instance")
}

func TestUnitOfWork_GetCollectionName(t *testing.T) {
	user := TestUser{}
	name := getCollectionName(user)
	assert.Equal(t, "testusers", name)
}

func TestBaseEntity_Methods(t *testing.T) {
	entity := &domain.BaseEntity{}

	id := primitive.NewObjectID()
	entity.SetID(id)
	assert.Equal(t, id, entity.GetID())

	entity.SetSlug("test-slug")
	assert.Equal(t, "test-slug", entity.GetSlug())

	entity.SetName("Test Entity")
	assert.Equal(t, "Test Entity", entity.GetName())

	assert.False(t, entity.IsDeleted())
	now := time.Now()
	entity.SetDeletedAt(&now)
	assert.True(t, entity.IsDeleted())
	assert.Equal(t, &now, entity.GetDeletedAt())

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	entity.SetCreatedAt(createdAt)
	entity.SetUpdatedAt(updatedAt)

	assert.Equal(t, createdAt, entity.GetCreatedAt())
	assert.Equal(t, updatedAt, entity.GetUpdatedAt())
}

func TestQueryParams_Validate(t *testing.T) {
	query := &domain.QueryParams[*TestUser]{
		Limit:  -1,
		Offset: 0,
	}

	err := query.Validate()
	assert.NoError(t, err)
	assert.Equal(t, 10, query.Limit)

	query.Limit = 2000
	err = query.Validate()
	assert.NoError(t, err)
	assert.Equal(t, 1000, query.Limit)
}
