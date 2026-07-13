package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/lasfh/goapi/examples/api/internal/adapter/http/handler/healthcheck"
)

func healthcheckRoutes(router chi.Router) {
	handler := healthcheck.NewHandler()

	router.Get("/health", handler.Healthcheck)
	router.Get("/ready", handler.Healthcheck)
}
