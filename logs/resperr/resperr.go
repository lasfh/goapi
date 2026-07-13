package resperr

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/lasfh/goapi/logs/logger"
)

type ResponseError struct {
	status  int
	message string
	errs    []error
}

type responseErrorf struct {
	status int
	format string
}

// Newf cria um novo erro HTTP com um código de status específico e uma string de formato.
//
// Parâmetros:
//   - statusCode (int): O código de status HTTP associado ao erro.
//   - format (string): A string de formato usada para criar a mensagem de erro.
//
// Retorna:
//   - (responseErrorf): Uma instância de responseErrorf que contém o código de status e a string de formato.
func Newf(statusCode int, format string) responseErrorf {
	return responseErrorf{statusCode, format}
}

// Format formata a mensagem de erro com base nos argumentos fornecidos e retorna um ResponseError.
//
// Parâmetros:
//   - a (...any): Lista de argumentos a serem inseridos na string de formato.
//
// Retorna:
//   - (*ResponseError): Uma instância de ResponseError contendo o código de status e a mensagem formatada.
func (e responseErrorf) Format(a ...any) *ResponseError {
	return &ResponseError{
		status:  e.status,
		message: fmt.Sprintf(e.format, a...),
	}
}

// New cria um novo ResponseError com o status e mensagem fornecidos.
//
// Parâmetros:
//   - statusCode: Código de status HTTP para o erro.
//   - message: Mensagem pública descritiva do erro.
//
// Retorno:
//   - *ResponseError: um ponteiro para o ResponseError recém-criado.
func New(statusCode int, message string) *ResponseError {
	return &ResponseError{
		status:  statusCode,
		message: message,
	}
}

// Error retorna a mensagem de erro formatada da estrutura ResponseError.
//
// Retorna:
//   - string: A mensagem de erro formatada.
func (e *ResponseError) Error() string {
	if len(e.errs) == 2 && e.errs[1] != nil {
		return e.errs[1].Error()
	}

	return e.message
}

// Status retorna o código de status HTTP da estrutura ResponseError.
//
// Retorna:
//   - int: O código de status HTTP.
func (e *ResponseError) StatusCode() int {
	return e.status
}

// Message retorna a mensagem pública de erro da estrutura ResponseError.
//
// Retorna:
//   - string: A mensagem de erro.
func (e *ResponseError) Message() string {
	return e.message
}

// Details retorna detalhes adicionais sobre o erro HTTP.
//
// Retorna:
//   - map[string]string: Lista de detalhes do erro, se houver.
func (e *ResponseError) Details() map[string]string {
	return nil
}

// Unwrap retorna a causa do erro HTTP, se disponível.
//
// Retorna:
//   - error: a causa do erro, se houver. Caso contrário, retorna nil.
func (e *ResponseError) Unwrap() []error {
	return e.errs
}

// WithErrorContext cria uma instância do ResponseError associada a um erro,
// usando um contexto específico e nível de aviso (Error).
//
// Parâmetros:
//   - ctx (context.Context): Contexto para o registro de log.
//   - err (error): O erro a ser associado.
//
// Retorna:
//   - *ResponseError: Instância do ResponseError com informações adicionais e contexto definido.
func (e *ResponseError) WithErrorContext(ctx context.Context, err error) *ResponseError {
	return e.withLogger(ctx, err, slog.LevelError)
}

// WithWarnContext cria uma instância do ResponseError associada a um erro,
// usando um contexto específico e nível de log de aviso (Warn).
//
// Parâmetros:
//   - ctx (context.Context): Contexto para o registro de log.
//   - err (error): O erro a ser associado.
//
// Retorna:
//   - *ResponseError: Instância do ResponseError com informações adicionais e contexto definido.
func (e *ResponseError) WithWarnContext(ctx context.Context, err error) *ResponseError {
	return e.withLogger(ctx, err, slog.LevelWarn)
}

// WithDebugContext cria uma instância do ResponseError associada a um erro,
// usando um contexto específico e nível de log de depuração (Debug).
//
// Parâmetros:
//   - ctx (context.Context): Contexto para o registro de log.
//   - err (error): O erro a ser associado.
//
// Retorna:
//   - *ResponseError: Instância do ResponseError com informações adicionais e contexto definido.
func (e *ResponseError) WithDebugContext(ctx context.Context, err error) *ResponseError {
	return e.withLogger(ctx, err, slog.LevelDebug)
}

// WithInfoContext cria uma instância do ResponseError associada a um erro,
// usando um contexto específico e nível de log de informação (Info).
//
// Parâmetros:
//   - ctx (context.Context): Contexto para o registro de log.
//   - err (error): O erro a ser associado.
//
// Retorna:
//   - *ResponseError: Instância do ResponseError com informações adicionais e contexto definido.
func (e *ResponseError) WithInfoContext(ctx context.Context, err error) *ResponseError {
	return e.withLogger(ctx, err, slog.LevelInfo)
}

// withLogger cria uma nova instância de ResponseError associada a um erro,
// adicionando informações de arquivo, linha e função ao logger.
//
// Parâmetros:
//   - ctx (context.Context): Contexto para o registro de log.
//   - err (error): Erro a ser associado.
//   - level (slog.Level): Nível de log a ser utilizado.
//
// Retorna:
//   - *ResponseError: Nova instância do ResponseError com logger enriquecido e informações de contexto.
func (e *ResponseError) withLogger(ctx context.Context, err error, level slog.Level) *ResponseError {
	if err == nil {
		return e
	}

	funcName, file, line := currentFileAndLine(3)

	logger.From(ctx).Log(
		ctx,
		level,
		err.Error(),
		slog.Group(
			"file",
			slog.Int("line", line),
			slog.String("file", file),
			slog.String("func", funcName),
		),
	)

	return &ResponseError{
		status:  e.status,
		message: e.message,
		errs:    []error{e, err},
	}
}

// currentFileAndLine retorna o caminho do arquivo e a linha do código onde foi chamada.
//
// Parâmetros:
//   - skip (int, opcional): Quantidade de níveis na pilha para ignorar. Por padrão, usa 1.
//
// Retorna:
//   - string: Caminho do arquivo.
//   - int: Número da linha no arquivo.
func currentFileAndLine(skip ...int) (string, string, int) {
	skipCaller := 1

	if len(skip) > 0 {
		skipCaller = skip[0]
	}

	pc, file, line, ok := runtime.Caller(skipCaller)
	if !ok {
		return "", "", 0
	}

	funcName := runtime.FuncForPC(pc).Name()

	return funcName, file, line
}
