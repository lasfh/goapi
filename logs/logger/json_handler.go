package logger

import (
	"context"
	"io"
	"log/slog"
)

// jsonHandler é um handler slog que divide a saída entre stdout e stderr.
type jsonHandler struct {
	stdout slog.Handler
	stderr slog.Handler
}

// NewJSONHandler cria um novo manipulador JSON que divide a saída entre stdout e stderr.
//
// Parâmetros:
//   - stdout: writer para saída padrão (logs com nível menor que Warn).
//   - stderr: writer para saída de erro (logs com nível Warn ou superior).
//   - opts: opções de configuração do handler slog.
//
// Retorno:
//   - *jsonHandler: nova instância do manipulador.
func NewJSONHandler(
	stdout io.Writer,
	stderr io.Writer,
	opts *slog.HandlerOptions,
) *jsonHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	return &jsonHandler{
		stdout: slog.NewJSONHandler(stdout, opts),
		stderr: slog.NewJSONHandler(stderr, opts),
	}
}

// Enabled verifica se o nível de log está habilitado.
//
// Parâmetros:
//   - ctx: contexto da chamada.
//   - level: nível do log a ser verificado.
//
// Retorno:
//   - bool: verdadeiro se o log deve ser processado.
func (h *jsonHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.stdout.Enabled(ctx, level)
}

// Handle processa o registro de log, enviando logs com nível Warn ou superior para stderr e os demais para stdout.
//
// Parâmetros:
//   - ctx: contexto da chamada.
//   - r: registro de log (record).
//
// Retorno:
//   - error: erro se houver falha na escrita.
func (h *jsonHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelWarn {
		return h.stderr.Handle(ctx, r)
	}

	return h.stdout.Handle(ctx, r)
}

// WithAttrs retorna um novo manipulador com os atributos adicionais.
//
// Parâmetros:
//   - attrs: lista de atributos a serem adicionados.
//
// Retorno:
//   - slog.Handler: novo manipulador com os atributos.
func (h *jsonHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &jsonHandler{
		stdout: h.stdout.WithAttrs(attrs),
		stderr: h.stderr.WithAttrs(attrs),
	}
}

// WithGroup retorna um novo manipulador com um grupo de atributos.
//
// Parâmetros:
//   - name: nome do grupo.
//
// Retorno:
//   - slog.Handler: novo manipulador com o grupo.
func (h *jsonHandler) WithGroup(name string) slog.Handler {
	return &jsonHandler{
		stdout: h.stdout.WithGroup(name),
		stderr: h.stderr.WithGroup(name),
	}
}
