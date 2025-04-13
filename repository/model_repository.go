package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"reflect"
)

// GenericRepository is a GORM-based implementation of the Repository interface.
type GenericRepository[T any] struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewGenericRepository creates a new GenericRepository instance using the provided GORM DB.
func NewGenericRepository[T any](db *gorm.DB, logger *logrus.Logger) *GenericRepository[T] {
	return &GenericRepository[T]{db: db, logger: logger}
}

// GetByID retrieves a model instance by its identifier.
func (r *GenericRepository[T]) GetByID(id interface{}) (T, error) {
	var model T
	r.logger.WithFields(logrus.Fields{
		"operation": "GetByID",
		"model_id":  id,
	}).Info("Fetching model by ID")

	tx := r.db.Begin()
	if tx.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"operation": "GetByID",
			"model_id":  id,
		}).Error(tx.Error.Error())
		return model, tx.Error
	}

	if hasContents(model) {
		tx = tx.Preload("Contents")
	}

	result := tx.First(&model, id)
	if result.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"operation": "GetByID",
			"model_id":  id,
			"error":     result.Error.Error(),
		}).Error("Failed to fetch model by ID")
		return model, result.Error
	}

	r.logger.WithFields(logrus.Fields{
		"operation": "GetByID",
		"model_id":  id,
	}).Info("Model fetched successfully")
	return model, nil
}

// GetAll returns all model instances.
func (r *GenericRepository[T]) GetAll(scopes ...ScopeWithLog) ([]T, error) {
	var models []T
	r.logger.WithFields(logrus.Fields{
		"operation": "GetAll",
	}).Info("Fetching all models with filter")

	// Define a scope function to apply the filter.
	filterScopes := make([]func(db *gorm.DB) *gorm.DB, 0)

	for _, scope := range scopes {
		filterScopes = append(filterScopes, func(db *gorm.DB) *gorm.DB {
			scope(db, r.logger)
			return db
		})
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"operation": "GetAll",
		}).Error(tx.Error.Error())
		return models, tx.Error
	}

	if hasContents(models) {
		tx = tx.Preload("Contents")
	}

	result := tx.Scopes(filterScopes...).Find(&models)
	if result.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"operation": "GetAll",
			"error":     result.Error.Error(),
		}).Error("Failed to fetch models with filter")
		return nil, result.Error
	}

	r.logger.WithFields(logrus.Fields{
		"operation": "GetAll",
		"count":     len(models),
	}).Info("Fetched models successfully")
	return models, nil
}

// Create inserts a new model instance into the database within a transaction.
// It automatically sets the CreatedAt and UpdatedAt fields.
func (r *GenericRepository[T]) Create(model *T) error {
	r.logger.WithField("operation", "Create").Info("Creating a new model")

	tx := r.db.Begin()
	if tx.Error != nil {
		r.logger.WithField("operation", "Create").Error(tx.Error.Error())
		return tx.Error
	}

	result := tx.Create(model)
	if result.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"operation": "Create",
			"error":     result.Error.Error(),
		}).Error("Failed to create model, rolling back transaction")
		if rbErr := tx.Rollback().Error; rbErr != nil {
			r.logger.WithField("operation", "Create").Error("Rollback error: " + rbErr.Error())
			return rbErr
		}
		return result.Error
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.WithField("operation", "Create").Error("Commit error: " + err.Error())
		return err
	}

	r.logger.WithField("operation", "Create").Info("Model created successfully")
	return nil
}

// Update modifies an existing model instance in the database within a transaction.
// It automatically sets the UpdatedAt field.
func (r *GenericRepository[T]) Update(model *T) error {
	// Use reflection to extract the model's ID from the embedded Model struct.
	v := reflect.ValueOf(model).Elem()
	idField := v.FieldByName("Model").FieldByName("ID").Uint()

	r.logger.WithFields(logrus.Fields{
		"operation": "Update",
		"model_id":  idField,
	}).Info("Updating model")

	tx := r.db.Begin()
	if tx.Error != nil {
		r.logger.WithField("operation", "Update").Error(tx.Error.Error())
		return tx.Error
	}

	result := tx.Save(model)
	if result.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"operation": "Update",
			"model_id":  idField,
			"error":     result.Error.Error(),
		}).Error("Failed to update model, rolling back transaction")
		if rbErr := tx.Rollback().Error; rbErr != nil {
			r.logger.WithField("operation", "Update").Error("Rollback error: " + rbErr.Error())
			return rbErr
		}
		return result.Error
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"operation": "Update",
			"model_id":  idField,
			"error":     err.Error(),
		}).Error("Commit error during update")
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"operation": "Update",
		"model_id":  idField,
	}).Info("Model updated successfully")
	return nil
}

// Delete removes a model instance identified by id within a transaction.
func (r *GenericRepository[T]) Delete(id interface{}) error {
	var model T
	r.logger.WithFields(logrus.Fields{
		"operation": "Delete",
		"model_id":  id,
	}).Info("Deleting model")

	tx := r.db.Begin()
	if tx.Error != nil {
		r.logger.WithField("operation", "Delete").Error(tx.Error.Error())
		return tx.Error
	}

	if err := tx.First(&model, id).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"operation": "Delete",
			"model_id":  id,
			"error":     err.Error(),
		}).Error("Failed to find model for deletion, rolling back transaction")
		if rbErr := tx.Rollback().Error; rbErr != nil {
			r.logger.WithField("operation", "Delete").Error("Rollback error: " + rbErr.Error())
			return rbErr
		}
		return err
	}

	result := tx.Delete(&model)
	if result.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"operation": "Delete",
			"model_id":  id,
			"error":     result.Error.Error(),
		}).Error("Failed to delete model, rolling back transaction")
		if rbErr := tx.Rollback().Error; rbErr != nil {
			r.logger.WithField("operation", "Delete").Error("Rollback error: " + rbErr.Error())
			return rbErr
		}
		return result.Error
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.WithField("operation", "Delete").Error("Commit error: " + err.Error())
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"operation": "Delete",
		"model_id":  id,
	}).Info("Model deleted successfully")
	return nil
}
