package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/lasfh/goapi/logs/logfile"
)

// Output define o destino de saída dos logs.
type Output uint8

const (
	// OutputOnlyStdout envia todos os logs para stdout.
	OutputOnlyStdout Output = iota

	// OutputOnlyStderr envia todos os logs para stderr.
	OutputOnlyStderr

	// OutputOnlyStdoutStderr divide a saída entre stdout (abaixo de Warn) e stderr (Warn ou superior).
	OutputOnlyStdoutStderr

	// OutputFile grava os logs em arquivo, opcionalmente dividindo entre stdout e stderr.
	OutputFile

	// OutputOpenTelemetry envia os logs via OpenTelemetry, além de manter saída no terminal.
	OutputOpenTelemetry
)

// ErrInvalidOutput é retornado quando o valor de Output não corresponde a nenhuma opção válida.
var ErrInvalidOutput = errors.New("logger: output inválido")

// Options agrupa as configurações para criação de um logger.
type Options struct {
	// Level define o nível mínimo de severidade dos logs a serem registrados.
	Level slog.Level

	// Output define o destino de saída dos logs.
	Output Output

	// UseStdoutStderr divide a saída entre stdout e stderr quando Output é OutputFile ou OutputOpenTelemetry.
	// Logs com nível abaixo de Warn vão para stdout; Warn ou superior vão para stderr.
	UseStdoutStderr bool

	// FileOptions contém as configurações do arquivo de log.
	// Usado apenas quando Output é OutputFile.
	FileOptions LogFileOptions

	// OpenTelemetryOptions contém as configurações do exportador OpenTelemetry.
	// Usado apenas quando Output é OutputOpenTelemetry.
	OpenTelemetryOptions OpenTelemetryOptions
}

// LogFileOptions agrupa as configurações de gravação em arquivo de log.
type LogFileOptions struct {
	// Path é o caminho para o diretório onde os arquivos de log serão armazenados.
	Path string

	// FormatFileName define a função de formatação do nome do arquivo. Padrão: logfile.NewFileName.
	FormatFileName logfile.FormatFileNameFunc

	// RemoveEmptyFile indica se arquivos de log vazios devem ser removidos ao fechar.
	RemoveEmptyFile bool
}

type logger struct {
	handler  slog.Handler
	shutdown func(ctx context.Context) error
}

// NewLogger cria e retorna uma nova instância de logger configurada de acordo com as opções fornecidas.
//
// Parâmetros:
//   - ctx: contexto usado na inicialização (ex: conexão com exportador OpenTelemetry).
//   - options: configurações do logger.
//
// Retorno:
//   - *logger: instância configurada do logger.
//   - error: erro caso a inicialização falhe ou o Output seja inválido.
func NewLogger(ctx context.Context, options Options) (*logger, error) {
	opts, err := setupOptions(options)
	if err != nil {
		return nil, err
	}

	logLevel := new(slog.LevelVar)
	logLevel.Set(opts.Level)

	handlerOpts := &slog.HandlerOptions{
		Level: logLevel,
	}

	switch opts.Output {
	case OutputOnlyStdout:
		return &logger{
			handler: slog.NewJSONHandler(os.Stdout, handlerOpts),
		}, nil

	case OutputOnlyStderr:
		return &logger{
			handler: slog.NewJSONHandler(os.Stderr, handlerOpts),
		}, nil

	case OutputOnlyStdoutStderr:
		return &logger{
			handler: NewJSONHandler(os.Stdout, os.Stderr, handlerOpts),
		}, nil

	case OutputOpenTelemetry:
		handler, shutdown, err := newOpenTelemetryHandler(ctx, opts.OpenTelemetryOptions)
		if err != nil {
			return nil, fmt.Errorf("logger: falha ao inicializar OpenTelemetry: %w", err)
		}

		return &logger{
			handler: slog.NewMultiHandler(
				stdoutStderrHandler(opts.UseStdoutStderr, os.Stdout, os.Stderr, handlerOpts),
				handler,
			),
			shutdown: shutdown,
		}, nil

	case OutputFile:
		fileWriter, err := logfile.NewFileWriter(
			opts.FileOptions.Path,
			opts.FileOptions.FormatFileName,
			opts.FileOptions.RemoveEmptyFile,
		)
		if err != nil {
			return nil, fmt.Errorf("logger: falha ao criar arquivo de log: %w", err)
		}

		return &logger{
			handler: stdoutStderrHandler(
				opts.UseStdoutStderr,
				io.MultiWriter(os.Stdout, fileWriter),
				io.MultiWriter(os.Stderr, fileWriter),
				handlerOpts,
			),
			shutdown: func(_ context.Context) error {
				return fileWriter.Close()
			},
		}, nil

	default:
		return nil, fmt.Errorf("%w: %d", ErrInvalidOutput, opts.Output)
	}
}

// Logger retorna uma instância de *slog.Logger usando o handler configurado.
func (l *logger) Logger() *slog.Logger {
	return slog.New(l.handler)
}

// Handler retorna o slog.Handler subjacente do logger.
func (l *logger) Handler() slog.Handler {
	return l.handler
}

// Shutdown encerra o logger e libera os recursos associados (ex: fecha arquivos, flush de exportadores).
// Deve ser chamado quando o logger não for mais necessário.
func (l *logger) Shutdown(ctx context.Context) error {
	if l.shutdown == nil {
		return nil
	}

	return l.shutdown(ctx)
}

func setupOptions(opts Options) (Options, error) {
	if opts.Output == OutputFile {
		if opts.FileOptions.Path == "" {
			return opts, errors.New("logger: LogFileOptions.LogsPath é obrigatório para OutputFile")
		}

		if opts.FileOptions.FormatFileName == nil {
			opts.FileOptions.FormatFileName = logfile.NewFileName
		}
	}

	if opts.Output == OutputOpenTelemetry {
		// if opts.OpenTelemetryOptions.HandlerName == "" {
		// 	return opts, errors.New("logger: OpenTelemetryOptions.HandlerName é obrigatório para OutputOpenTelemetry")
		// }

		if len(opts.OpenTelemetryOptions.ExporterOptions) == 0 {
			return opts, errors.New("logger: OpenTelemetryOptions.ExporterOptions é obrigatório para OutputOpenTelemetry")
		}
	}

	return opts, nil
}

func stdoutStderrHandler(
	splitOutput bool,
	stdout io.Writer,
	stderr io.Writer,
	opts *slog.HandlerOptions,
) slog.Handler {
	if splitOutput {
		return NewJSONHandler(stdout, stderr, opts)
	}

	return slog.NewJSONHandler(stdout, opts)
}
