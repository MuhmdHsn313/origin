package service

import (
	"fmt"
	"github.com/kataras/iris/v12/core/router"
)

func RegisterHandler[T any](api router.Party, service Service[T]) {
	routerName := structNameToSnake(new(T))
	serviceRouter := api.Party(fmt.Sprintf("/%s", routerName))
	serviceRouter.Get("/", service.GetAll)
	serviceRouter.Get("/{id}", service.GetByID)
	serviceRouter.Post("/", service.Create)
	serviceRouter.Delete("/{id}", service.Delete)
	serviceRouter.Patch("/{id}", service.UpdatePatch)
}
