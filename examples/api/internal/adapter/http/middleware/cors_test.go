package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCors_Preflight(t *testing.T) {
	cases := []struct {
		name           string
		allowedOrigins []string
		allowedHeaders []string
		origin         string
		requestMethod  string
		requestHeaders string
		expectedOrigin string
	}{
		{
			name:           "origem permitida retorna a própria origem",
			allowedOrigins: []string{"https://app.example.com"},
			origin:         "https://app.example.com",
			requestMethod:  http.MethodPost,
			expectedOrigin: "https://app.example.com",
		},
		{
			name:           "origem não permitida não retorna Allow-Origin",
			allowedOrigins: []string{"https://app.example.com"},
			origin:         "https://malicioso.com",
			requestMethod:  http.MethodPost,
			expectedOrigin: "",
		},
		{
			name:           "sem origens configuradas permite qualquer origem",
			allowedOrigins: nil,
			origin:         "https://qualquer.com",
			requestMethod:  http.MethodGet,
			expectedOrigin: "*",
		},
		{
			name:           "método não permitido aborta o preflight",
			allowedOrigins: []string{"https://app.example.com"},
			origin:         "https://app.example.com",
			requestMethod:  http.MethodTrace,
			expectedOrigin: "",
		},
		{
			name:           "headers vazios usam wildcard e aceitam qualquer header",
			allowedOrigins: []string{"https://app.example.com"},
			allowedHeaders: nil,
			origin:         "https://app.example.com",
			requestMethod:  http.MethodPost,
			requestHeaders: "X-Custom-Header",
			expectedOrigin: "https://app.example.com",
		},
		{
			name:           "header não permitido aborta o preflight",
			allowedOrigins: []string{"https://app.example.com"},
			allowedHeaders: []string{"Content-Type"},
			origin:         "https://app.example.com",
			requestMethod:  http.MethodPost,
			requestHeaders: "X-Custom-Header",
			expectedOrigin: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			handler := Cors(tc.allowedOrigins, tc.allowedHeaders, nil)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					nextCalled = true
				}),
			)

			req := httptest.NewRequest(http.MethodOptions, "/", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Access-Control-Request-Method", tc.requestMethod)
			if tc.requestHeaders != "" {
				req.Header.Set("Access-Control-Request-Headers", tc.requestHeaders)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.expectedOrigin {
				t.Errorf("Access-Control-Allow-Origin: esperava %q, obteve %q", tc.expectedOrigin, got)
			}
			if nextCalled {
				t.Error("preflight não deveria chamar o próximo handler")
			}
		})
	}
}

func TestCors_PreflightHeaders(t *testing.T) {
	handler := Cors([]string{"https://app.example.com"}, nil, nil)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	expected := map[string]string{
		"Access-Control-Allow-Origin":      "https://app.example.com",
		"Access-Control-Allow-Methods":     http.MethodPut,
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Max-Age":           "300",
	}
	for header, want := range expected {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s: esperava %q, obteve %q", header, want, got)
		}
	}
}

func TestCors_ActualRequest(t *testing.T) {
	cases := []struct {
		name            string
		allowedOrigins  []string
		exposedHeaders  []string
		origin          string
		expectedOrigin  string
		expectedExposed string
	}{
		{
			name:           "origem permitida adiciona os headers de CORS",
			allowedOrigins: []string{"https://app.example.com"},
			origin:         "https://app.example.com",
			expectedOrigin: "https://app.example.com",
		},
		{
			name:           "origem não permitida não adiciona Allow-Origin",
			allowedOrigins: []string{"https://app.example.com"},
			origin:         "https://malicioso.com",
			expectedOrigin: "",
		},
		{
			name:            "expõe os headers configurados",
			allowedOrigins:  []string{"https://app.example.com"},
			exposedHeaders:  []string{"X-Request-Id", "X-Total-Count"},
			origin:          "https://app.example.com",
			expectedOrigin:  "https://app.example.com",
			expectedExposed: "X-Request-Id, X-Total-Count",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			handler := Cors(tc.allowedOrigins, nil, tc.exposedHeaders)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					nextCalled = true
					w.WriteHeader(http.StatusNoContent)
				}),
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Origin", tc.origin)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if !nextCalled {
				t.Error("esperava que o próximo handler fosse chamado")
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.expectedOrigin {
				t.Errorf("Access-Control-Allow-Origin: esperava %q, obteve %q", tc.expectedOrigin, got)
			}
			if got := rec.Header().Get("Access-Control-Expose-Headers"); got != tc.expectedExposed {
				t.Errorf("Access-Control-Expose-Headers: esperava %q, obteve %q", tc.expectedExposed, got)
			}
		})
	}
}
