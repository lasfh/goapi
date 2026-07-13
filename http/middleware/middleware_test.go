package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lasfh/goapi/logs/logger"
)

func TestClientIP(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		expected   string
	}{
		{
			name:       "usa RemoteAddr quando não há headers",
			remoteAddr: "192.168.1.1:1234",
			expected:   "192.168.1.1",
		},
		{
			name:       "usa RemoteAddr sem porta",
			remoteAddr: "192.168.1.1",
			expected:   "192.168.1.1",
		},
		{
			name:       "usa X-Forwarded-For com um IP",
			remoteAddr: "10.0.0.1:1234",
			xff:        "203.0.113.5",
			expected:   "203.0.113.5",
		},
		{
			name:       "usa o primeiro IP do X-Forwarded-For",
			remoteAddr: "10.0.0.1:1234",
			xff:        "203.0.113.5, 10.0.0.1",
			expected:   "203.0.113.5",
		},
		{
			name:       "usa X-Real-IP quando X-Forwarded-For está ausente",
			remoteAddr: "10.0.0.1:1234",
			xri:        "203.0.113.9",
			expected:   "203.0.113.9",
		},
		{
			name:       "X-Forwarded-For tem prioridade sobre X-Real-IP",
			remoteAddr: "10.0.0.1:1234",
			xff:        "203.0.113.5",
			xri:        "203.0.113.9",
			expected:   "203.0.113.5",
		},
		{
			name:       "X-Forwarded-For com IPv6",
			remoteAddr: "10.0.0.1:1234",
			xff:        "2001:db8::1",
			expected:   "2001:db8::1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remoteAddr
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			if tc.xri != "" {
				req.Header.Set("X-Real-IP", tc.xri)
			}

			got := clientIP(req)
			if got != tc.expected {
				t.Errorf("esperava %q, obteve %q", tc.expected, got)
			}
		})
	}
}

// TestWithLogger_AddsLoggerToContext verifica se o middleware adiciona um logger ao contexto
func TestWithLogger_AddsLoggerToContext(t *testing.T) {
	// Arrange
	var capturedLogger *slog.Logger

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if logger, ok := r.Context().Value(logger.LoggerKey).(*slog.Logger); ok {
			capturedLogger = logger
		}

		w.WriteHeader(http.StatusNoContent)
	})

	handler := WithLogger(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=123", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if capturedLogger == nil {
		t.Error("esperava que o logger fosse adicionado ao contexto, mas não foi encontrado")
	}
}
