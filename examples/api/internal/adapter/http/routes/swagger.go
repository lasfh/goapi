package routes

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/lasfh/goapi/examples/api/internal/docs"
)

func swaggerRoutes(router chi.Router) {
	router.Get("/swagger/*", httpSwagger.WrapHandler)
}
