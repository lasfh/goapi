package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/lasfh/goapi/examples/api/database"
	cors "github.com/lasfh/goapi/examples/api/internal/adapter/http/middleware"
	"github.com/lasfh/goapi/examples/api/internal/adapter/http/routes"
	"github.com/lasfh/goapi/examples/api/internal/config"
	"github.com/lasfh/goapi/examples/api/internal/version"
	"github.com/lasfh/goapi/http/middleware"
	"github.com/lasfh/goapi/storage"
)

func Setup(cfg *config.Config) error {
	ctx := context.Background()

	stg, err := storage.NewStorage(
		ctx,
		cfg.Storage.StorageOptions,
	)
	if err != nil {
		return err
	}

	db, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		return err
	}

	rdb, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		return err
	}

	router := chi.NewRouter()
	router.Use(chiMiddleware.Logger)
	router.Use(chiMiddleware.Recoverer)
	router.Use(chiMiddleware.RequestSize(cfg.API.MaxBodySize))
	router.Use(
		cors.Cors(
			cfg.CORS.AllowedOrigins,
			cfg.CORS.AllowedHeaders,
			cfg.CORS.ExposedHeaders,
			cfg.CORS.Debug,
		),
	)
	router.Use(middleware.WithLogger)

	routes.Register(cfg, router, db, rdb, stg)

	addr := fmt.Sprintf("http://%s:%d", cfg.API.HostOrDefault(), cfg.API.Port)

	slog.Debug(
		"Servidor rodando",
		slog.String("addr", addr),
		slog.Group(
			"version",
			slog.String("version", version.Version),
			slog.String("git_commit", version.GitCommit),
			slog.String("build_time", version.BuildTimeOrNow()),
		),
	)

	slog.Debug(
		"Documentação",
		slog.String(
			"url", fmt.Sprintf("%s/swagger/", addr),
		),
	)

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second, // Protege contra Slowloris
		IdleTimeout:       60 * time.Second,
	}

	// Inicie o servidor em uma goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	// Crie um canal para ouvir sinais de interrupção ou término do sistema operacional
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Bloquear até que um sinal seja recebido
	<-quit

	// Crie um prazo para aguardar a conclusão das solicitações em andamento
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Tentar um desligamento suave
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Servidor forçado a desligar: %v", err)
	}

	return nil
}
