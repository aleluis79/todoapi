package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "example.com/todoapi/docs"

	"example.com/todoapi/internal/config"
	"example.com/todoapi/internal/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initLogger(config config.Config) {
	var level slog.Level
	var handler slog.Handler

	switch strings.ToUpper(config.LogLevel) {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}

	if strings.EqualFold(config.LogFormat, "json") {
		handler = slog.NewJSONHandler(os.Stdout, &opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Info("Inicializando Logger en formato: " + config.LogFormat + " level: " + config.LogLevel)
}

func initTracer(config config.Config) error {
	var exporter sdktrace.SpanExporter
	var err error

	url := "http://jaeger-collector.observability.svc.cluster.local:9411/api/v2/spans"
	serviceName := fmt.Sprintf("%s.%s", config.PodAPP, config.PodNameSpace)

	switch config.TracingExporter {
	case "disable":
		{
			slog.Info("Tracing disabled")
			return nil
		}
	case "console":
		{
			slog.Info("Inicializando Tracing en console")
			exporter, err = stdouttrace.New(
				stdouttrace.WithPrettyPrint(),
				stdouttrace.WithoutTimestamps())
			if err != nil {
				return err
			}
		}
	default:
		slog.Info("Inicializando Tracing Zipkin")
		exporter, err = zipkin.New(url)
		if err != nil {
			return err
		}
	}

	batcher := sdktrace.NewBatchSpanProcessor(exporter)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(batcher),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

// @title           Todo API
// @version         1.0
// @description     Backend Todo API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
func main() {

	config, err := config.LoadConfig(".")
	if err != nil {
		slog.Error("No se pudo cargar la configuración: " + err.Error())
		os.Exit(1)
	}

	initLogger(config)

	err = initTracer(config) // log err json
	if err != nil {
		slog.Error("No se pudo inicializar las trazas: " + err.Error())
		os.Exit(1)
	}

	s := server.New(config)

	if config.Environment == "local" {
		s.AddSwagger()
	}

	s.AddHealtCheck()

	s.AddRoutes()

	s.Run()

}
