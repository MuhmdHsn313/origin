package service

import (
	"github.com/MuhmdHsn313/origin/repository"
	"github.com/kataras/iris/v12"
)

// Service is a generic interface that wraps repository operations and provides additional
// helper functions for extracting payloads (create, update, filter) from a model instance.
type Service[T any] interface {
	// GetByID retrieves a model instance by its identifier.
	GetByID(ctx iris.Context)
	// GetAll returns all model instances.
	GetAll(ctx iris.Context)
	// Create inserts a new model instance into the database.
	Create(ctx iris.Context)
	// UpdatePatch modifies an existing model instance in the database.
	UpdatePatch(ctx iris.Context)
	// Delete removes a model instance identified by id.
	Delete(ctx iris.Context)
}

type modelService[T any] struct {
	eng  Engine[T]
	repo repository.Repository[T]
}

func NewModelService[T any](eng Engine[T], repo repository.Repository[T]) Service[T] {
	return &modelService[T]{
		eng:  eng,
		repo: repo,
	}
}

func (service modelService[T]) GetByID(ctx iris.Context) {
	id, err := ctx.Params().GetUint("id")
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "CANT_READ_ID",
			},
		)
		return
	}

	object, err := service.repo.GetByID(id)
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "FETCH_READ_OBJECT_ERROR",
			},
		)
		return
	}

	_ = ctx.StopWithJSON(iris.StatusOK, object)
}

func (service modelService[T]) GetAll(ctx iris.Context) {
	// Generate filter parameters
	//filter, err := service.eng.GenerateFilterParameters()
	//if err != nil {
	//	_ = ctx.StopWithJSON(
	//		iris.StatusBadRequest,
	//		iris.Map{
	//			"error":      err.Error(),
	//			"error_code": "CANT_GEN_FILTER",
	//		},
	//	)
	//}

	//
	objects, err := service.repo.GetAll()
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "FETCH_ERROR",
			},
		)
		return
	}

	_ = ctx.StopWithJSON(iris.StatusOK, objects)
}

func (service modelService[T]) Create(ctx iris.Context) {
	createParams, err := service.eng.GenerateCreateParameters()
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "GENERATE_CREATE_PARAMS_ERROR",
			},
		)
		return
	}

	err = ctx.ReadBody(&createParams)
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "PARSE_CREATE_PARAMS_ERROR",
			},
		)
		return
	}

	model, err := service.eng.FillModelFromCreateParameters(createParams)
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "GENERATE_CREATE_MODEL_ERROR",
			},
		)
		return
	}

	err = service.repo.Create(model)
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "CREATE_ERROR",
			},
		)
		return
	}

	_ = ctx.StopWithJSON(iris.StatusCreated, model)
}

func (service modelService[T]) UpdatePatch(ctx iris.Context) {
	objId := ctx.Params().Get("id")

	objModel, err := service.repo.GetByID(objId)
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "NOT_FOUND",
			},
		)
		return
	}

	updateParams, err := service.eng.GenerateUpdateParameters()
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "GENERATE_UPDATE_PARAMS_ERROR",
			},
		)
		return
	}

	err = ctx.ReadBody(&updateParams)
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "PARSE_UPDATE_PARAMS_ERROR",
			},
		)
		return
	}

	model, err := service.eng.UpdateModelFromUpdateParameters(&objModel, updateParams)
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "GENERATE_UPDATE_MODEL_ERROR",
			},
		)
		return
	}

	err = service.repo.Update(model)
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "UPDATE_ERROR",
			},
		)
		return
	}

	_ = ctx.StopWithJSON(iris.StatusOK, model)
}

func (service modelService[T]) Delete(ctx iris.Context) {
	objId := ctx.Params().Get("id")

	err := service.repo.Delete(objId)
	if err != nil {
		_ = ctx.StopWithJSON(
			iris.StatusBadRequest,
			iris.Map{
				"error":      err.Error(),
				"error_code": "DELETE_ERROR",
			},
		)
		return
	}

	ctx.StopWithStatus(iris.StatusNoContent)
}
