package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otlog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// OpenTelemetryOptions agrupa as configurações do handler OpenTelemetry.
type OpenTelemetryOptions struct {
	// HandlerName identifica a instrumentação no provedor de logs (ex: nome do serviço).
	HandlerName string

	// ExporterOptions configura o exportador OTLP HTTP (ex: endpoint, headers, TLS).
	ExporterOptions []otlploghttp.Option

	// ResourceOptions define os atributos do recurso associado aos logs (ex: nome do serviço, versão, ambiente).
	// Consulte o pacote go.opentelemetry.io/otel/sdk/resource para as opções disponíveis.
	ResourceOptions []resource.Option
}

func newOpenTelemetryHandler(ctx context.Context, opts OpenTelemetryOptions) (slog.Handler, func(ctx context.Context) error, error) {
	exporter, err := otlploghttp.New(ctx, opts.ExporterOptions...)
	if err != nil {
		return nil, nil, err
	}

	res, err := resource.New(ctx, opts.ResourceOptions...)
	if err != nil {
		_ = exporter.Shutdown(ctx)

		return nil, nil, err
	}

	provider := otlog.NewLoggerProvider(
		otlog.WithProcessor(otlog.NewBatchProcessor(exporter)),
		otlog.WithResource(res),
	)

	return otelslog.NewHandler(
		opts.HandlerName,
		otelslog.WithLoggerProvider(provider),
	), provider.Shutdown, nil
}
