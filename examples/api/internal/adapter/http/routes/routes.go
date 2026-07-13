package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lasfh/goapi/examples/api/internal/adapter/repository"
	"github.com/lasfh/goapi/examples/api/internal/config"
	"github.com/lasfh/goapi/examples/api/internal/core/service/example"
	"github.com/lasfh/goapi/examples/api/internal/version"
	"github.com/lasfh/goapi/http/response"
	"github.com/lasfh/goapi/storage"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func Register(
	cfg *config.Config,
	router *chi.Mux,
	db *gorm.DB,
	rdb *redis.Client,
	stg storage.Storage,
) {
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response.EncodeOK(
			w,
			map[string]any{
				"title": "API EXAMPLE",
				"version": map[string]any{
					"version":      version.Version,
					"build_time":   version.BuildTimeOrNow(),
					"git_commit":   version.GitCommit,
					"current_time": time.Now(),
				},
			},
		)
	})

	healthcheckRoutes(router)
	swaggerRoutes(router)

	// Com sessão
	router.Group(func(r chi.Router) {
		// r.Use() // AUTH

		exampleRoutes(
			r,
			example.NewService(
				repository.NewExampleRepo(db),
			),
		)
	})
}
