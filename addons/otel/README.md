# FlowTracker OpenTelemetry (OTel) Bridge

This addon provides a bridge between **FlowTracker** and the **OpenTelemetry** ecosystem.

It allows you to write simple, lightweight instrumentation using FlowTracker, but export the data to industry-standard backend systems like **Jaeger**, **Grafana Tempo**, **Datadog**, **HoneyComb**, and **New Relic**.

## 📦 Installation

This is a separate module. You must install it alongside the core library and the OpenTelemetry SDK.

```bash
# 1. Install FlowTracker Core
go get github.com/spdeepak/flowtracker

# 2. Install the Bridge
go get github.com/spdeepak/flowtracker/otel-exporter

# 3. Install OTel SDK dependencies (Required to configure the destination)
go get go.opentelemetry.io/otel \
       go.opentelemetry.io/otel/sdk \
       go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
```

## 🚀 How it Works

1.  **Instrumentation:** You use `flowtracker` as normal. It captures the execution flow, timestamps, and tags efficiently in memory.
2.  **Export:** When the request finishes, this exporter takes the completed `Trace` object.
3.  **Conversion:** It recursively maps the FlowTracker spans to OpenTelemetry Spans, preserving hierarchy, timestamps, and attributes.
4.  **Push:** It uses the standard OTel SDK to push the data to your configured backend (via HTTP/gRPC).

## 🛠 Usage

You need to configure the **OpenTelemetry SDK** (TracerProvider) in your application startup, and then pass it to the FlowTracker exporter.

### 1. Setup Jaeger or Grafana Tempo (Locally)

If you want to test this locally, run Jaeger via Docker:

```bash
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

or see the [docker compose file in examples](./../../examples/otlp/docker-compose.yaml)

### 2. Integration Code

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/spdeepak/flowtracker"
	otelexporter "github.com/spdeepak/flowtracker/otel-exporter"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func initTracer(ctx context.Context) *sdktrace.TracerProvider {
	// Configure OTLP Exporter to point to Jaeger/Tempo
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("localhost:4318"), // Standard OTLP HTTP port
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	// Define Service Name
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("my-flowtracker-service"),
		),
	)

	// Create Tracer Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp
}

func main() {
	ctx := context.Background()
	tp := initTracer(ctx)
	
	// Ensure data is flushed on shutdown
	defer tp.Shutdown(ctx)

	// --- THE BRIDGE ---
	// Create the exporter using the OTel Provider configured above
	bridgeExporter := otelexporter.New(tp)

	// Register Middleware
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(bridgeExporter))

	http.ListenAndServe(":8080", mw(http.DefaultServeMux))
}
```

## 📝 ID Mapping & Attributes

OpenTelemetry requires strictly formatted 128-bit Trace IDs and 64-bit Span IDs. FlowTracker uses simple strings.

To ensure data integrity, the exporter behaves as follows:

1.  **New IDs:** The exporter generates **new, valid OTel UUIDs** for every trace and span so they are accepted by the backend.
2.  **Cross-Reference:** It automatically adds the original FlowTracker IDs as attributes to every span. You can search for these in your UI:
    *   `flowtracker.trace_id`
    *   `flowtracker.span_id`
3.  **Tags:** All tags added via `flowtracker.AddTag()` are converted to OTel Attributes.

## ⚠️ Limitations

*   **Post-Processing:** FlowTracker traces are exported only *after* the root request finishes. This means you will not see "live" partial traces in Jaeger while the request is still processing (unlike native OTel streaming).
*   **Context Propagation:** If you make an HTTP call to *another* microservice from within your app, you must manually inject the OTel headers if you want distributed tracing across services. This bridge focuses on **internal** process flow.