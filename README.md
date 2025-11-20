# FlowTracker

**FlowTracker** is a lightweight, thread-safe observability library for Golang microservices. It tracks the lifecycle of an API request as it flows through your application, capturing execution time, hierarchy, and metadata.

Unlike full-blown distributed tracing solutions (like Jaeger/OpenTelemetry) which can be heavy to set up, FlowTracker is designed for **single-service internal flow analysis**. It produces hierarchical JSON data perfectly structured for generating **Sankey Diagrams**, **Gantt Charts**, or **Execution Trees**.

## 🚀 Features

*   **Zero-Config Middleware:** specific `http.Handler` wrapper to start tracking immediately.
*   **Context Propagation:** Uses Go `context` to pass parent/child relationships deep into your call stack.
*   **Pluggable Exporters:** Comes with Console and File exporters, but easily extensible for Databases, Kafka, or external APIs.
*   **Non-Blocking:** Data export runs in a separate goroutine to ensure your API response time isn't affected by logging.
*   **Graph-Ready Data:** outputs flat JSON with `span_id` and `parent_id` relationships, ready for visualization tools.

## 📦 Installation

```bash
go get github.com/spdeepak/flowtracker
```

## ⚡ Quick Start

### 1. Wrap your Router
Initialize the middleware in your `main.go`. By default, this logs traces to the Standard Output.

```go
package main

import (
	"context"
	"net/http"

	"github.com/spdeepak/flowtracker"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", defaultHandler)

	// No arguments = Default ConsoleExporter
	mw := flowtracker.NewMiddleware()

	http.ListenAndServe(":8080", mw(mux))
}
```

### 2. Instrument your Code
Pass `ctx` to your functions and use `StartSpan` to track execution.

```go
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Start a span for a DB call
	// Always defer the finish function!
	ctx, finish := flowtracker.StartSpan(ctx, "FetchFromDB")
	defer finish()
	// Simulate a Database call
	processDatabaseLogic(ctx)
	// Simulate external Microservice call
	w.Write([]byte("Order Processed"))
}

func processDatabaseLogic(ctx context.Context) {
	// Add metadata to the current span
	flowtracker.AddTag(ctx, "db.query", "SELECT * FROM users")
}
```

## 📊 Data Structure & Visualization

The output data is designed to be easily parsed for graphing.

### Example Output (JSON)
```json
{
  "trace_id": "trace-169834-99",
  "spans": [
    {
      "span_id": "100",
      "name": "GET /api/data",
      "start_time": "2023-11-20T10:00:00Z",
      "end_time": "2023-11-20T10:00:01Z",
      "duration_ms": 1000
    },
    {
      "span_id": "200",
      "parent_id": "100", 
      "name": "FetchFromDB",
      "duration_ms": 500,
      "tags": { "db.query": "SELECT..." }
    }
  ]
}
```

### How to visualize?

1.  **Sankey Diagram:**
    *   Use `parent_id` as the **Source**.
    *   Use `span_id` (or Name) as the **Target**.
    *   Use `duration_ms` as the **Weight/Width**.
    *   *This visualizes where the time is going in your flow.*

2.  **Grafana:**
    *   If using the `ConsoleExporter` combined with **Loki**, you can query logs for `{app="myapp"} |= "FLOW_LOG:"`.
    *   If using a custom exporter, you can push directly to **Tempo** or **Jaeger**.

## ⚠️ Best Practices

1.  **Always Defer:** `ctx, finish := flowtracker.StartSpan(...)` followed immediately by `defer finish()`. This ensures the timing is accurate even if functions panic or return early.
2.  **Pass Context:** You must pass `ctx` down your function chain. If you break the context chain, the library cannot link the child span to the parent.
3.  **Tags:** Use `AddTag` sparingly for high-cardinality data (IDs, status codes) to help debugging.