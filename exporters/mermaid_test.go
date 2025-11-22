package exporters

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spdeepak/flowtracker"
)

func NewServer(orientation Orientation, tags []string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", Handler)
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(&MermaidExporter{Orientation: orientation, IncludeTags: tags}))
	return mw(mux)
}

func TestMermaidExporter_OK(t *testing.T) {
	// capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	server := NewServer(TopDown, []string{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	time.Sleep(1 * time.Second)

	// restore stdout
	w.Close()
	os.Stdout = old

	// read logs
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Error("Error should be nil")
	}

	logs := buf.String()
	output := strings.Split(logs, "\n")
	if !strings.Contains(output[1], "----- MERMAID OUTPUT -----") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.EqualFold(output[2], "```mermaid") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.EqualFold(output[3], "graph TD") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[4], "[\"GET /\"] -->|") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[4], "[\"Process Payment (currency:USD)\"]") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[5], "[\"GET /\"] -->|") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[5], "[\"DB: Select User (db.query:SELECT * FROM users...)\"]") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[6], "[\"GET /\"] -->|") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[6], "[\"HTTP: Shipping Service\"]") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[7], "[\"HTTP: Shipping Service\"] -->|") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[7], "[\"Calculate Weight\"]") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.EqualFold(output[8], "```") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.EqualFold(output[9], "--------------------------") {
		t.Fatalf("expected log not found: %s", logs)
	}
}

func TestMermaidExporter_OK_IncludeTags(t *testing.T) {
	// capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	server := NewServer(LeftRight, []string{"db.query"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	time.Sleep(1 * time.Second)

	// restore stdout
	w.Close()
	os.Stdout = old

	// read logs
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Error("Error should be nil")
	}

	logs := buf.String()
	output := strings.Split(logs, "\n")
	if !strings.Contains(output[1], "----- MERMAID OUTPUT -----") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.EqualFold(output[2], "```mermaid") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.EqualFold(output[3], "graph LR") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[4], "[\"GET /\"] -->|") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[4], "[\"Process Payment\"]") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[5], "[\"GET /\"] -->|") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[5], "[\"DB: Select User (db.query:SELECT * FROM users...)\"]") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[6], "[\"GET /\"] -->|") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[6], "[\"HTTP: Shipping Service\"]") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[7], "[\"HTTP: Shipping Service\"] -->|") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.Contains(output[7], "[\"Calculate Weight\"]") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.EqualFold(output[8], "```") {
		t.Fatalf("expected log not found: %s", logs)
	}
	if !strings.EqualFold(output[9], "--------------------------") {
		t.Fatalf("expected log not found: %s", logs)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Simulate some logic
	processPayment(ctx)

	// 2. Simulate a Database call
	fetchUserProfile(ctx, "user_123")

	// 3. Simulate external Microservice call
	callShippingService(ctx)

	w.Write([]byte("Order Processed"))
}

func processPayment(ctx context.Context) {
	// Start a sub-span
	ctx, finish := flowtracker.StartSpan(ctx, "Process Payment")
	defer finish()

	time.Sleep(50 * time.Millisecond) // Simulate work
	flowtracker.AddTag(ctx, "currency", "USD")
}

func fetchUserProfile(ctx context.Context, userID string) {
	ctx, finish := flowtracker.StartSpan(ctx, "DB: Select User")
	defer finish()

	flowtracker.AddTag(ctx, "db.query", "SELECT * FROM users...")
	time.Sleep(120 * time.Millisecond) // Simulate DB latency
}

func callShippingService(ctx context.Context) {
	ctx, finish := flowtracker.StartSpan(ctx, "HTTP: Shipping Service")
	defer finish()

	// Nested span example: Logic inside the external call preparation
	func(innerCtx context.Context) {
		_, end := flowtracker.StartSpan(innerCtx, "Calculate Weight")
		defer end()
		time.Sleep(10 * time.Millisecond)
	}(ctx)

	time.Sleep(200 * time.Millisecond) // Simulate network call
}
