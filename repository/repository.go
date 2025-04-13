package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ScopeWithLog func(db *gorm.DB, logger *logrus.Logger) *gorm.DB

// Repository is a generic interface that abstracts data storage operations for a model of type T.
type Repository[T any] interface {
	// GetByID retrieves a model instance by its identifier.
	GetByID(id interface{}) (T, error)
	// GetAll returns all model instances that match the provided filter.
	// The filter is a map of field names to their expected values.
	GetAll(scopes ...ScopeWithLog) ([]T, error)
	// Create inserts a new model instance into the database.
	Create(model *T) error
	// Update modifies an existing model instance in the database.
	Update(model *T) error
	// Delete removes a model instance identified by id.
	Delete(id interface{}) error
}
