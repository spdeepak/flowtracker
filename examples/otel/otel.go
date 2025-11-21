package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/spdeepak/flowtracker"
	otelexporter "github.com/spdeepak/flowtracker/otel"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	// Setup OTel
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// --- THE BRIDGE ---
	// Initialize our custom addon exporter
	// It will use the global provider we just set up
	bridgeExporter := otelexporter.New(tp)

	// Register Middleware
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(bridgeExporter))

	// Run Server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, finish := flowtracker.StartSpan(r.Context(), "Business Logic")
		time.Sleep(100 * time.Millisecond)
		finish()
		w.Write([]byte("Check your console for OTel JSON output!"))
	})

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", mw(mux))
}

func initTracer() *sdktrace.TracerProvider {
	// 1. Create an OTel Exporter (e.g., Stdout, OTLP, Jaeger)
	//    Here we use Stdout for demonstration
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	// 2. Create Resource (Metadata about your app)
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("my-flowtracker-app"),
		),
	)

	// 3. Create Tracer Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// 4. Set as Global (Optional, but good practice)
	otel.SetTracerProvider(tp)

	return tp
}
