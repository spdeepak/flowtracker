package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/spdeepak/flowtracker"
	"github.com/spdeepak/flowtracker/examples"
	otelexporter "github.com/spdeepak/flowtracker/otel"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	ctx := context.Background()

	// Initialize Tracer to Jaeger
	tp := initTracer(ctx)
	defer func() {
		// FLUSH data before exiting
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Initialize the Bridge Addon
	bridgeExporter := otelexporter.New(tp)

	// Use Middleware
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(bridgeExporter))

	// Define Handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Start FlowTracker Span
		ctx, finish := flowtracker.StartSpan(r.Context(), "ProcessRequest")

		// Add tags (these will show up in Jaeger tags!)
		flowtracker.AddTag(ctx, "user.id", "12345")
		flowtracker.AddTag(ctx, "environment", "local-dev")

		// Simulate work
		time.Sleep(50 * time.Millisecond)

		// Nested Span
		_, subFinish := flowtracker.StartSpan(ctx, "DB:FetchUser")
		time.Sleep(200 * time.Millisecond)
		subFinish()

		finish()
		w.Write([]byte("Trace sent to Jaeger! Check http://localhost:16686"))
		examples.Handler(w, r)
	})

	log.Println("Server running on :8080")
	log.Println("Visit localhost:16686 to see the Jaeger UI")
	http.ListenAndServe(":8080", mw(mux))
}

func initTracer(ctx context.Context) *sdktrace.TracerProvider {
	// 1. Create the OTLP HTTP Exporter
	//    Jaeger accepts OTLP over HTTP on port 4318 by default.
	//    We use "Insecure" because we are running locally without TLS.
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	// 2. Define Resource (Service Name appears in Jaeger)
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("my-flowtracker-service"), // <--- Look for this in Jaeger dropdown
		),
	)

	// 3. Create Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp
}
