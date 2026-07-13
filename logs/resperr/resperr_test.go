package resperr

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/lasfh/goapi/logs/logger"
)

// captureHandler é um slog.Handler que armazena os registros para inspeção nos testes.
type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *captureHandler) WithGroup(string) slog.Handler { return h }

func contextWithCapture() (context.Context, *captureHandler) {
	handler := &captureHandler{}
	ctx := logger.WithLogger(context.Background(), slog.New(handler))

	return ctx, handler
}

func TestCurrentFileAndLine(t *testing.T) {
	funcName, file, line := currentFileAndLine()

	expectedLine := 38 // Linha onde currentFileAndLine é chamada acima.

	if !strings.Contains(funcName, "resperr.TestCurrentFileAndLine") {
		t.Errorf("esperava que o nome da função contivesse 'resperr.TestCurrentFileAndLine', mas obteve: %s", funcName)
	}

	if !strings.HasSuffix(file, "resperr/resperr_test.go") {
		t.Errorf("esperava que o arquivo terminasse com 'resperr/resperr_test.go', mas obteve: %s", file)
	}

	if line != expectedLine {
		t.Errorf("esperava linha %d, mas obteve linha %d", expectedLine, line)
	}
}

func TestCurrentFileAndLineInvalidSkip(t *testing.T) {
	funcName, file, line := currentFileAndLine(1000)

	if funcName != "" || file != "" || line != 0 {
		t.Errorf("esperava valores vazios para skip inválido, mas obteve: %q, %q, %d", funcName, file, line)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{"Erro genérico", http.StatusInternalServerError, "Erro interno do servidor"},
		{"Erro não autorizado", http.StatusUnauthorized, "Acesso não autorizado"},
		{"Erro não encontrado", http.StatusNotFound, "Recurso não encontrado"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.statusCode, tt.message)

			if err.StatusCode() != tt.statusCode {
				t.Errorf("StatusCode() = %d, esperado %d", err.StatusCode(), tt.statusCode)
			}

			if err.Message() != tt.message {
				t.Errorf("Message() = %q, esperado %q", err.Message(), tt.message)
			}

			if err.Error() != tt.message {
				t.Errorf("Error() = %q, esperado %q", err.Error(), tt.message)
			}

			if err.Details() != nil {
				t.Errorf("Details() = %v, esperado nil", err.Details())
			}

			if err.Unwrap() != nil {
				t.Errorf("Unwrap() = %v, esperado nil", err.Unwrap())
			}
		})
	}
}

func TestNewfFormat(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		format       string
		args         []any
		expectedMsg  string
		expectedCode int
	}{
		{
			name:         "Erro com formatação simples",
			status:       http.StatusBadRequest,
			format:       "Campo %s é obrigatório",
			args:         []any{"nome"},
			expectedMsg:  "Campo nome é obrigatório",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Erro com múltiplos argumentos",
			status:       http.StatusForbidden,
			format:       "Acesso negado para %s com ID %d",
			args:         []any{"usuário", 1234},
			expectedMsg:  "Acesso negado para usuário com ID 1234",
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "Erro sem argumentos",
			status:       http.StatusNotFound,
			format:       "Recurso não encontrado",
			args:         []any{},
			expectedMsg:  "Recurso não encontrado",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Newf(tt.status, tt.format).Format(tt.args...)

			if err.StatusCode() != tt.expectedCode {
				t.Errorf("StatusCode() = %d, esperado %d", err.StatusCode(), tt.expectedCode)
			}

			if err.Message() != tt.expectedMsg {
				t.Errorf("Message() = %q, esperado %q", err.Message(), tt.expectedMsg)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name        string
		err         *ResponseError
		expectedMsg string
	}{
		{
			name: "Erro com duas entradas em errs e errs[1] não nulo",
			err: &ResponseError{
				status:  http.StatusInternalServerError,
				message: "Erro interno do servidor",
				errs: []error{
					New(http.StatusInternalServerError, "Erro interno."),
					errors.New("erro específico"),
				},
			},
			expectedMsg: "erro específico",
		},
		{
			name: "Erro com errs vazio",
			err: &ResponseError{
				status:  http.StatusNotFound,
				message: "Recurso não encontrado",
				errs:    []error{},
			},
			expectedMsg: "Recurso não encontrado",
		},
		{
			name: "Erro com uma entrada em errs",
			err: &ResponseError{
				status:  http.StatusBadRequest,
				message: "Requisição inválida",
				errs:    []error{errors.New("detalhe do erro")},
			},
			expectedMsg: "Requisição inválida",
		},
		{
			name: "Erro com duas entradas em errs e errs[1] nulo",
			err: &ResponseError{
				status:  http.StatusBadRequest,
				message: "Requisição inválida",
				errs:    []error{errors.New("detalhe do erro"), nil},
			},
			expectedMsg: "Requisição inválida",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expectedMsg {
				t.Errorf("Error() = %q, esperado %q", got, tt.expectedMsg)
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	tests := []struct {
		name          string
		with          func(e *ResponseError, ctx context.Context, err error) *ResponseError
		expectedLevel slog.Level
	}{
		{
			name: "WithErrorContext",
			with: func(e *ResponseError, ctx context.Context, err error) *ResponseError {
				return e.WithErrorContext(ctx, err)
			},
			expectedLevel: slog.LevelError,
		},
		{
			name: "WithWarnContext",
			with: func(e *ResponseError, ctx context.Context, err error) *ResponseError {
				return e.WithWarnContext(ctx, err)
			},
			expectedLevel: slog.LevelWarn,
		},
		{
			name: "WithInfoContext",
			with: func(e *ResponseError, ctx context.Context, err error) *ResponseError {
				return e.WithInfoContext(ctx, err)
			},
			expectedLevel: slog.LevelInfo,
		},
		{
			name: "WithDebugContext",
			with: func(e *ResponseError, ctx context.Context, err error) *ResponseError {
				return e.WithDebugContext(ctx, err)
			},
			expectedLevel: slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, handler := contextWithCapture()

			base := New(http.StatusInternalServerError, "Erro interno do servidor")
			innerErr := errors.New("falha na conexão com o banco")

			err := tt.with(base, ctx, innerErr)

			if err == base {
				t.Fatal("esperava uma nova instância de ResponseError, mas obteve a original")
			}

			if err.StatusCode() != base.StatusCode() {
				t.Errorf("StatusCode() = %d, esperado %d", err.StatusCode(), base.StatusCode())
			}

			if err.Message() != base.Message() {
				t.Errorf("Message() = %q, esperado %q", err.Message(), base.Message())
			}

			if err.Error() != innerErr.Error() {
				t.Errorf("Error() = %q, esperado %q", err.Error(), innerErr.Error())
			}

			if !errors.Is(err, base) {
				t.Error("errors.Is(err, base) = false, esperado true")
			}

			if !errors.Is(err, innerErr) {
				t.Error("errors.Is(err, innerErr) = false, esperado true")
			}

			if len(handler.records) != 1 {
				t.Fatalf("esperava 1 registro de log, mas obteve %d", len(handler.records))
			}

			record := handler.records[0]

			if record.Level != tt.expectedLevel {
				t.Errorf("nível de log = %v, esperado %v", record.Level, tt.expectedLevel)
			}

			if record.Message != innerErr.Error() {
				t.Errorf("mensagem de log = %q, esperada %q", record.Message, innerErr.Error())
			}

			assertFileGroup(t, record)
		})
	}
}

// assertFileGroup verifica se o registro contém o grupo "file" apontando para este arquivo de teste.
func assertFileGroup(t *testing.T, record slog.Record) {
	t.Helper()

	var found bool

	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key != "file" || attr.Value.Kind() != slog.KindGroup {
			return true
		}

		found = true

		for _, groupAttr := range attr.Value.Group() {
			switch groupAttr.Key {
			case "file":
				if !strings.HasSuffix(groupAttr.Value.String(), "resperr/resperr_test.go") {
					t.Errorf("file = %q, esperava sufixo 'resperr/resperr_test.go'", groupAttr.Value.String())
				}
			case "func":
				if !strings.Contains(groupAttr.Value.String(), "resperr.TestWithContext") {
					t.Errorf("func = %q, esperava conter 'resperr.TestWithContext'", groupAttr.Value.String())
				}
			case "line":
				if groupAttr.Value.Int64() <= 0 {
					t.Errorf("line = %d, esperava valor maior que zero", groupAttr.Value.Int64())
				}
			}
		}

		return false
	})

	if !found {
		t.Error("registro de log não contém o grupo 'file'")
	}
}

func TestWithContextNilError(t *testing.T) {
	ctx, handler := contextWithCapture()

	base := New(http.StatusInternalServerError, "Erro interno do servidor")

	if err := base.WithErrorContext(ctx, nil); err != base {
		t.Error("esperava a instância original quando o erro é nil")
	}

	if len(handler.records) != 0 {
		t.Errorf("esperava nenhum registro de log, mas obteve %d", len(handler.records))
	}
}
