package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthcheck(t *testing.T) {
	h := NewHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Healthcheck(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status: esperava %d, obteve %d", http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type: esperava %q, obteve %q", "application/json", got)
	}

	var body struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("erro ao decodificar a resposta: %v", err)
	}

	if body.Message != "Ok!" {
		t.Errorf("message: esperava %q, obteve %q", "Ok!", body.Message)
	}
}
