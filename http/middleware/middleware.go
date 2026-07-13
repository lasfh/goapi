package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lasfh/goapi/logs/logger"
)

// WithLogger adiciona um logger ao contexto da requisição HTTP,
// incluindo informações básicas da requisição, como ID, método, caminho,
// host, protocolo, IP de origem e user-agent.
//
// Parâmetros:
//   - next (http.Handler): Próximo handler a ser executado.
//
// Retorna:
//   - http.Handler: Handler com logger configurado no contexto.
func WithLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestID(r)

		ctx := context.WithValue(
			r.Context(),
			logger.LoggerKey,
			slog.With(
				slog.Group(
					"request",
					slog.String("id", id),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("host", r.Host),
					slog.String("proto", r.Proto),
					slog.String("ip", clientIP(r)),
					slog.String("user_agent", r.UserAgent()),
				),
			),
		)

		ctx = context.WithValue(
			ctx,
			requestIDKey,
			id,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requestID extrai o identificador da requisição a partir dos headers
// X-Request-ID ou X-Correlation-ID, permitindo correlacionar os logs de
// uma mesma requisição. Caso nenhum esteja presente, gera um UUID v7 como
// fallback, garantindo que toda requisição tenha um ID.
//
// Parâmetros:
//   - r (*http.Request): Requisição HTTP.
//
// Retorna:
//   - string: ID da requisição.
func requestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}

	if id := r.Header.Get("X-Correlation-ID"); id != "" {
		return id
	}

	if id, err := uuid.NewV7(); err == nil {
		return id.String()
	}

	return ""
}

// clientIP tenta extrair o IP real do cliente, verificando os headers
// X-Forwarded-For e X-Real-IP antes de recorrer ao RemoteAddr.
//
// Parâmetros:
//   - r (*http.Request): Requisição HTTP.
//
// Retorna:
//   - string: IP do cliente.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip, _, err := net.SplitHostPort(strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])); err == nil {
			return ip
		}

		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
