package middleware

import (
	"context"
)

type requestIDCtxKey struct{}

var requestIDKey requestIDCtxKey

// RequestID recupera o identificador da requisição armazenado no contexto.
//
// Parâmetros:
//   - ctx (context.Context): O contexto da execução de onde o identificador é extraído.
//
// Retorna:
//   - string: O identificador da requisição, ou uma string vazia se não estiver presente.
//   - bool: Indica se o identificador foi encontrado no contexto.
func RequestID(ctx context.Context) (string, bool) {
	if l, ok := ctx.Value(requestIDKey).(string); ok {
		return l, true
	}

	return "", false
}
