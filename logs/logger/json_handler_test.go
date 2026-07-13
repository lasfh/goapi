package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestNewJSONHandler(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	h := NewJSONHandler(stdout, stderr, nil)

	if h == nil {
		t.Fatal("NewJSONHandler returned nil")
	}

	if h.stdout == nil {
		t.Error("jsonHandler.stdout is nil")
	}
	if h.stderr == nil {
		t.Error("jsonHandler.stderr is nil")
	}
}

func TestJsonHandler_Handle(t *testing.T) {
	tests := []struct {
		name         string
		level        slog.Level
		msg          string
		expectStdout bool
		expectStderr bool
	}{
		{
			name:         "Debug level",
			level:        slog.LevelDebug,
			msg:          "debug message",
			expectStdout: true,
			expectStderr: false,
		},
		{
			name:         "Info level",
			level:        slog.LevelInfo,
			msg:          "info message",
			expectStdout: true,
			expectStderr: false,
		},
		{
			name:         "Warn level",
			level:        slog.LevelWarn,
			msg:          "warn message",
			expectStdout: false,
			expectStderr: true,
		},
		{
			name:         "Error level",
			level:        slog.LevelError,
			msg:          "error message",
			expectStdout: false,
			expectStderr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			h := NewJSONHandler(stdout, stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})

			logger := slog.New(h)

			switch tt.level {
			case slog.LevelDebug:
				logger.Debug(tt.msg)
			case slog.LevelInfo:
				logger.Info(tt.msg)
			case slog.LevelWarn:
				logger.Warn(tt.msg)
			case slog.LevelError:
				logger.Error(tt.msg)
			}

			if tt.expectStdout {
				if stdout.Len() == 0 {
					t.Error("Saída esperada em stdout, mas obtida vazia.")
				}
				if !containsMessage(stdout, tt.msg) {
					t.Errorf("Mensagem esperada %q em stdout, recebida %q", tt.msg, stdout.String())
				}
			} else {
				if stdout.Len() > 0 {
					t.Errorf("Não esperava nenhuma saída em stdout, mas obteve %q.", stdout.String())
				}
			}

			if tt.expectStderr {
				if stderr.Len() == 0 {
					t.Error("Saída esperada em stderr, mas obtida vazia.")
				}
				if !containsMessage(stderr, tt.msg) {
					t.Errorf("Mensagem esperada %q em stderr, recebida %q", tt.msg, stderr.String())
				}
			} else {
				if stderr.Len() > 0 {
					t.Errorf("Não esperava nenhuma saída em stderr, mas obteve %q.", stderr.String())
				}
			}
		})
	}
}

func TestJsonHandler_Enabled(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	h := NewJSONHandler(stdout, stderr, nil)

	ctx := context.Background()

	if h.Enabled(ctx, slog.LevelDebug) {
		t.Error("Espera-se que Enabled(Debug) seja falso por padrão.")
	}

	if !h.Enabled(ctx, slog.LevelInfo) {
		t.Error("Espera-se que Enabled(Info) seja verdadeiro por padrão.")
	}
}

func TestJsonHandler_WithAttrs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	h := NewJSONHandler(stdout, stderr, nil)
	hWithAttrs := h.WithAttrs([]slog.Attr{slog.String("key", "value")})

	logger := slog.New(hWithAttrs)
	logger.Info("test")

	if !strings.Contains(stdout.String(), `"key":"value"`) {
		t.Errorf("O resultado esperado deveria conter o atributo, mas foi obtido %s", stdout.String())
	}
}

func TestJsonHandler_WithGroup(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	h := NewJSONHandler(stdout, stderr, nil)
	hWithGroup := h.WithGroup("mygroup")

	logger := slog.New(hWithGroup)
	logger.Info("test", slog.String("key", "value"))

	expected := `"mygroup":{"key":"value"}`
	if !strings.Contains(stdout.String(), expected) {
		t.Errorf("O resultado esperado deveria conter o grupo %q, mas foi obtido %s", expected, stdout.String())
	}
}

func containsMessage(r io.Reader, msg string) bool {
	var entry map[string]any
	if err := json.NewDecoder(r).Decode(&entry); err != nil {
		return false
	}
	return entry["msg"] == msg
}
