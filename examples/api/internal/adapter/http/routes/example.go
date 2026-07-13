package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/lasfh/goapi/examples/api/internal/adapter/http/handler/example"
)

func exampleRoutes(
	router chi.Router,
	exampleService example.ExampleService,
) {
	handler := example.NewHandler(
		exampleService,
	)

	router.Get("/example", handler.Find)
}
