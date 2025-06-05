package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseModel interface {
	GetID() primitive.ObjectID
	SetID(id primitive.ObjectID)
	GetSlug() string
	SetSlug(slug string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	SetDeletedAt(deletedAt *time.Time)
	GetName() string
	IsDeleted() bool
}

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

type SortMap map[string]SortDirection

type QueryParams[E BaseModel] struct {
	Filter  E        `json:"filter,omitempty"`
	Sort    SortMap  `json:"sort,omitempty"`
	Include []string `json:"include,omitempty"`
	Limit   int      `json:"limit,omitempty"`
	Offset  int      `json:"offset,omitempty"`
}

func (q *QueryParams[E]) Validate() error {
	if q.Limit < 0 {
		q.Limit = 10
	}
	if q.Limit > 1000 {
		q.Limit = 1000
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	return nil
}

func (q *QueryParams[E]) GetPageInfo() (page int, size int) {
	size = q.Limit
	if size == 0 {
		size = 10
	}
	page = (q.Offset / size) + 1
	return page, size
}
