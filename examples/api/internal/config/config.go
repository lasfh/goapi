package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lasfh/goapi/examples/api/internal/version"
	"github.com/lasfh/goapi/logs/logger"
	"github.com/lasfh/goapi/storage"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type Config struct {
	API      APIConfig
	Auth     AuthConfig
	Log      logger.Options
	CORS     CORSConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Storage  StorageConfig
}

type APIConfig struct {
	Host        string
	Port        int32
	Environment string
	Timezone    string
	MaxBodySize int64
}

func (a APIConfig) HostOrDefault() string {
	if a.Host == "" {
		return "localhost"
	}

	return a.Host
}

type AuthConfig struct {
	Secret     string
	WithLeeway time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedHeaders []string
	ExposedHeaders []string
	Debug          bool
}

type PostgresConfig struct {
	dsn               string
	dsnMigration      string
	Host              string
	Port              uint32
	User              string
	Password          string
	UserMigration     string
	PasswordMigration string
	Database          string
	MaxIdleConns      int
	MaxOpenConns      int
	MaxLifetime       time.Duration
	MaxIdleTime       time.Duration
	Debug             bool
}

func (c PostgresConfig) DSN() string {
	if c.dsn != "" {
		return c.dsn
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=America/Sao_Paulo",
		c.Host, c.User, url.QueryEscape(c.Password), c.Database, c.Port,
	)
}

func (c PostgresConfig) DSNMigration() string {
	if c.dsnMigration != "" {
		return c.dsnMigration
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=America/Sao_Paulo",
		c.Host, c.UserMigration, url.QueryEscape(c.PasswordMigration), c.Database, c.Port,
	)
}

func (c PostgresConfig) ConnectionURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.User, url.QueryEscape(c.Password), c.Host, c.Port, c.Database,
	)
}

func (c PostgresConfig) ConnectionMigrationURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.User, url.QueryEscape(c.Password), c.Host, c.Port, c.Database,
	)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	Database int
}

type StorageConfig struct {
	storage.StorageOptions

	BaseDir string
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

// Load carrega as configurações do sistema a partir de um arquivo de configuração.
//
// Parâmetros:
//   - addConfigPath (...string): Caminho adicional opcional para localizar o arquivo de configuração.
//
// Retorna:
//   - *Config: Estrutura contendo as configurações carregadas.
//   - error: Erro caso ocorra uma falha na leitura do arquivo de configuração.
func Load(addConfigPath ...string) (*Config, error) {
	if addConfigPath != nil {
		viper.AddConfigPath(addConfigPath[0])
	}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	apiPort := viper.GetInt32("api.port")
	if apiPort <= 0 {
		apiPort = 8080
	}

	environment := viper.GetString("api.environment")
	if environment == "" {
		environment = "local"
	}

	timezone := viper.GetString("api.timezone")
	if timezone == "" {
		timezone = "America/Sao_Paulo"
	}

	maxIdleConns := viper.GetInt("postgres.max_idle_conns")
	if maxIdleConns <= 0 {
		maxIdleConns = 5
	}

	maxOpenConns := viper.GetInt("postgres.max_open_conns")
	if maxOpenConns <= 0 {
		maxOpenConns = 20
	}

	maxLifetime := viper.GetDuration("postgres.max_lifetime")
	if maxLifetime <= 0 {
		maxLifetime = 1 * time.Hour
	}

	maxIdleTime := viper.GetDuration("postgres.max_idle_time")
	if maxIdleTime <= 0 {
		maxIdleTime = 5 * time.Minute
	}

	return &Config{
		API: APIConfig{
			Host: strings.TrimSpace(
				viper.GetString("api.host"),
			),
			Port:        apiPort,
			Environment: environment,
			Timezone:    timezone,
			MaxBodySize: 150 << 20, // 150 MB
		},
		Auth: AuthConfig{
			Secret:     viper.GetString("auth.secret"),
			WithLeeway: viper.GetDuration("auth.with_leeway"),
		},
		Log: logger.Options{
			Level:           logLevel(viper.GetString("log.level")),
			Output:          output(viper.GetString("log.output")),
			UseStdoutStderr: viper.GetBool("log.split_output"),
			FileOptions: logger.LogFileOptions{
				Path: viper.GetString("log.path"),
			},
			OpenTelemetryOptions: func() logger.OpenTelemetryOptions {
				exportOptions := []otlploghttp.Option{
					otlploghttp.WithEndpoint(
						viper.GetString("log.loki_endpoint"),
					),
					otlploghttp.WithURLPath("/otlp/v1/logs"),
				}

				if viper.GetBool("log.open_telemetry_insecure") {
					exportOptions = append(exportOptions, otlploghttp.WithInsecure())
				}

				hostname, _ := os.Hostname()

				return logger.OpenTelemetryOptions{
					ExporterOptions: exportOptions,
					ResourceOptions: []resource.Option{
						resource.WithAttributes(
							semconv.DeploymentEnvironmentName(environment),
							semconv.ServiceName(
								viper.GetString("log.open_telemetry_service_name"),
							),
							semconv.ServiceVersion(
								fmt.Sprintf("%s (%s)", version.Version, version.GitCommit),
							),
							semconv.ServiceNamespace(
								viper.GetString("log.open_telemetry_namespace"),
							),
							semconv.ServiceInstanceID(hostname),
						),
					},
				}
			}(),
		},
		CORS: CORSConfig{
			AllowedOrigins: viper.GetStringSlice("cors.allowed_origins"),
			AllowedHeaders: viper.GetStringSlice("cors.allowed_headers"),
			ExposedHeaders: viper.GetStringSlice("cors.exposed_headers"),
			Debug:          viper.GetBool("cors.debug"),
		},
		Postgres: PostgresConfig{
			dsn:               viper.GetString("postgres.dsn"),
			dsnMigration:      viper.GetString("postgres.dsn_migration"),
			Host:              viper.GetString("postgres.host"),
			Port:              viper.GetUint32("postgres.port"),
			User:              viper.GetString("postgres.user"),
			Password:          viper.GetString("postgres.password"),
			UserMigration:     viper.GetString("postgres.user_migration"),
			PasswordMigration: viper.GetString("postgres.password_migration"),
			Database:          viper.GetString("postgres.database"),
			MaxIdleConns:      maxIdleConns,
			MaxOpenConns:      maxOpenConns,
			MaxLifetime:       maxLifetime,
			MaxIdleTime:       maxIdleTime,
			Debug:             viper.GetBool("postgres.debug"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("redis.host"),
			Port:     viper.GetString("redis.port"),
			Password: viper.GetString("redis.password"),
			Database: viper.GetInt("redis.database"),
		},
		Storage: StorageConfig{
			StorageOptions: storage.StorageOptions{
				DriverName: storage.DriverName(
					viper.GetString("storage.driver"),
				),
				Local: storage.LocalOptions{
					DirPerm: os.FileMode(
						viper.GetUint32("storage.local_dir_perm"),
					),
					FilePerm: os.FileMode(
						viper.GetUint32("storage.local_file_perm"),
					),
				},
				S3: storage.S3Options{
					Endpoint:     viper.GetString("storage.s3_endpoint"),
					Bucket:       viper.GetString("storage.s3_bucket"),
					Region:       viper.GetString("storage.s3_region"),
					AccessKeyID:  viper.GetString("storage.s3_access_key"),
					SecretKey:    viper.GetString("storage.s3_secret_key"),
					UsePathStyle: viper.GetBool("storage.s3_use_path_style"),
				},
			},
			BaseDir: fmt.Sprintf("%s_%s", viper.GetString("storage.base_dir"), environment),
		},
	}, nil
}

// logLevel converte um valor em um nível de log do pacote slog.
//
// Parâmetros:
//   - level (string): Nível do log.
//
// Retorna:
//   - slog.Level: Nível correspondente do slog.
func logLevel(level string) slog.Level {
	switch level {
	case "-4", "debug":
		return slog.LevelDebug
	case "0", "info":
		return slog.LevelInfo
	case "4", "warn":
		return slog.LevelWarn
	case "8", "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// output define o destino de saída do log (stdout ou stderr).
//
// Parâmetros:
//   - out (string): Destino da saída ("stdout" ou "stderr").
//
// Retorna:
//   - io.Writer: O writer correspondente (os.Stdout ou os.Stderr).
func output(out string) logger.Output {
	switch strings.ToLower(out) {
	case "loki":
		return logger.OutputOpenTelemetry
	case "file":
		return logger.OutputFile
	default:
		return logger.OutputOnlyStdoutStderr
	}
}
