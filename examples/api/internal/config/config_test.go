package config

import (
	"log/slog"
	"testing"

	"github.com/lasfh/goapi/logs/logger"
)

func Test_logLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"-4", slog.LevelDebug},
		{"debug", slog.LevelDebug},
		{"0", slog.LevelInfo},
		{"info", slog.LevelInfo},
		{"4", slog.LevelWarn},
		{"warn", slog.LevelWarn},
		{"8", slog.LevelError},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := logLevel(tt.input)
			if got != tt.want {
				t.Errorf("logLevel(%q) = %v, esperado %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_output(t *testing.T) {
	tests := []struct {
		input string
		want  logger.Output
	}{
		{"loki", logger.OutputOpenTelemetry},
		{"LOKI", logger.OutputOpenTelemetry},
		{"Loki", logger.OutputOpenTelemetry},
		{"file", logger.OutputFile},
		{"FILE", logger.OutputFile},
		{"File", logger.OutputFile},
		{"stdout", logger.OutputOnlyStdoutStderr},
		{"stderr", logger.OutputOnlyStdoutStderr},
		{"", logger.OutputOnlyStdoutStderr},
		{"unknown", logger.OutputOnlyStdoutStderr},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := output(tt.input)
			if got != tt.want {
				t.Errorf("output(%q) = %v, esperado %v", tt.input, got, tt.want)
			}
		})
	}
}
