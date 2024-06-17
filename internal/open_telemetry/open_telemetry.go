package open_telemetry

import (
	"context"
	"intmax2-node/configs/buildvars"

	"fmt"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

// Init configures an OpenTelemetry exporter and trace provider.
func Init(enable bool) error {
	if !enable {
		return nil
	}

	ctx := context.Background()
	client := otlptracehttp.NewClient()
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return err
	}

	res := resource.NewSchemaless(
		semconv.ServiceNameKey.String(buildvars.BuildName),
		semconv.ServiceVersionKey.String(buildvars.Version))

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// Tracer returns the OpenTelemetry tracer.
func Tracer() trace.Tracer {
	const empty = ""
	return otel.Tracer(empty)
}

// MarkSpanError marks span with error.
func MarkSpanError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	const (
		name           = "MarkSpanError"
		maskErrMessage = "[error] in %s[%s:%d] %v"
		errDescription = "error.description"
		callerNumber   = 1
		errAttribute   = "error"
	)

	_, span := Tracer().Start(ctx, name)
	defer span.End()

	pc, fn, line, _ := runtime.Caller(callerNumber)
	span.SetAttributes(attribute.Key(errDescription).
		String(fmt.Sprintf(maskErrMessage, runtime.FuncForPC(pc).Name(), fn, line, err)))
	span.SetAttributes(attribute.Key(errAttribute).Bool(true))
}
