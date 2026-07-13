package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/lasfh/goapi/examples/api/database"
	"github.com/lasfh/goapi/examples/api/internal/adapter/http/server"
	"github.com/lasfh/goapi/examples/api/internal/config"
	"github.com/lasfh/goapi/examples/api/migrations"
	"github.com/lasfh/goapi/logs/logger"
	"github.com/spf13/cobra"
)

// @title API EXAMPLE
// @version 1.0
// @BasePath /
// @securityDefinitions.apikey Auth
// @in header
// @name Authorization
// @description Digite "Bearer" seguido de um espaço e o token JWT. Exemplo: "Bearer eyJhbGci..."
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}

	loc, err := time.LoadLocation(cfg.API.Timezone)
	if err != nil {
		log.Fatalln(err)
	}

	time.Local = loc

	l, err := logger.NewLogger(context.Background(), cfg.Log)
	if err != nil {
		log.Fatalln(err)
	}

	defer l.Shutdown(context.Background())

	slog.SetDefault(l.Logger())

	rootCmd := runServer(cfg)
	rootCmd.AddCommand(
		migrationCommand(cfg),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Panicln(err)
	}
}

func runServer(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Short: "API",
		Long:  "API HTTP",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			slog.SetDefault(
				slog.With(
					slog.String("scope_name", "api"),
				),
			)

			if err := server.Setup(cfg); err != nil {
				log.Panicln(err)
			}
		},
	}
}

// migrationCommand cria e retorna um comando Cobra para gerenciar migrações.
//
// Parâmetros:
//   - cfg (*config.Config): Configuração da aplicação, incluindo parâmetros de migração.
//
// Retorna:
//   - *cobra.Command: Comando Cobra configurado para execução das migrações.
func migrationCommand(cfg *config.Config) *cobra.Command {
	migrationCmd := &cobra.Command{
		Use:   "migration",
		Short: "Opções para executar migrações",
	}

	var ignoreVersion []uint

	migrationCmd.AddCommand(
		&cobra.Command{
			Use:   "up",
			Short: "Executar apenas migrações Up",
			Run: func(cmd *cobra.Command, args []string) {
				slog.SetDefault(
					slog.With(
						slog.String("scope_name", "migration"),
					),
				)

				ctx, cancel := context.WithTimeout(
					context.Background(),
					30*time.Minute,
				)

				defer cancel()

				db, err := database.NewPostgresClientMigration(cfg.Postgres)
				if err != nil {
					log.Fatal(err)
				}

				migrate, err := migrations.NewMigrationManager(
					db,
					cfg.Postgres.Database,
				)

				slog.Info("Starting migration execution (up)")

				err = migrate.Up(ctx, ignoreVersion...)
				if err != nil {
					log.Fatal(err)
				}
			},
		},
	)

	migrationCmd.PersistentFlags().UintSliceVarP(&ignoreVersion, "ignore", "i", []uint{}, "Versões de migração a serem ignoradas")

	return migrationCmd
}
