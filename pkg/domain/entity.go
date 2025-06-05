package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BaseEntity provides a concrete implementation of BaseModel for MongoDB
type BaseEntity struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Slug      string             `bson:"slug,omitempty" json:"slug,omitempty"`
	Name      string             `bson:"name,omitempty" json:"name,omitempty"`
	CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
	DeletedAt *time.Time         `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
}

// GetID returns the entity ID
func (b *BaseEntity) GetID() primitive.ObjectID {
	return b.ID
}

// SetID sets the entity ID
func (b *BaseEntity) SetID(id primitive.ObjectID) {
	b.ID = id
}

// GetSlug returns the entity slug
func (b *BaseEntity) GetSlug() string {
	return b.Slug
}

// SetSlug sets the entity slug
func (b *BaseEntity) SetSlug(slug string) {
	b.Slug = slug
}

// GetCreatedAt returns the creation timestamp
func (b *BaseEntity) GetCreatedAt() time.Time {
	return b.CreatedAt
}

// GetUpdatedAt returns the last update timestamp
func (b *BaseEntity) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

// GetDeletedAt returns the deletion timestamp
func (b *BaseEntity) GetDeletedAt() *time.Time {
	return b.DeletedAt
}

// SetDeletedAt sets the deletion timestamp
func (b *BaseEntity) SetDeletedAt(deletedAt *time.Time) {
	b.DeletedAt = deletedAt
}

// GetName returns the entity name
func (b *BaseEntity) GetName() string {
	return b.Name
}

// IsDeleted checks if the entity is soft deleted
func (b *BaseEntity) IsDeleted() bool {
	return b.DeletedAt != nil
}

// SetCreatedAt sets the creation timestamp
func (b *BaseEntity) SetCreatedAt(t time.Time) {
	b.CreatedAt = t
}

// SetUpdatedAt sets the update timestamp
func (b *BaseEntity) SetUpdatedAt(t time.Time) {
	b.UpdatedAt = t
}

// SetName sets the entity name
func (b *BaseEntity) SetName(name string) {
	b.Name = name
}
