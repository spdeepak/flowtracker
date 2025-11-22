package flowtracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", Handler)
	mw := NewMiddleware(WithExporter(&ConsoleExporter{}))
	return mw(mux)
}

func TestDefaultExporter_OK(t *testing.T) {
	// capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	server := NewServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	time.Sleep(1 * time.Second)

	// restore stdout
	w.Close()
	os.Stdout = old

	// read logs
	var buf bytes.Buffer
	written, err := io.Copy(&buf, r)
	if err != nil {
		t.Error("Error should be nil")
	}
	fmt.Println(written)

	logs := buf.String()

	if !strings.Contains(logs, "FLOW_LOG: {\"trace_id\":\"trace-") {
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
	ctx, finish := StartSpan(ctx, "Process Payment")
	defer finish()

	time.Sleep(50 * time.Millisecond) // Simulate work
	AddTag(ctx, "currency", "USD")
}

func fetchUserProfile(ctx context.Context, userID string) {
	ctx, finish := StartSpan(ctx, "DB: Select User")
	defer finish()

	AddTag(ctx, "db.query", "SELECT * FROM users...")
	time.Sleep(120 * time.Millisecond) // Simulate DB latency
}

func callShippingService(ctx context.Context) {
	ctx, finish := StartSpan(ctx, "HTTP: Shipping Service")
	defer finish()

	// Nested span example: Logic inside the external call preparation
	func(innerCtx context.Context) {
		_, end := StartSpan(innerCtx, "Calculate Weight")
		defer end()
		time.Sleep(10 * time.Millisecond)
	}(ctx)

	time.Sleep(200 * time.Millisecond) // Simulate network call
}

func TestEndToEndFlow_ConsoleExporter(t *testing.T) {
	// 1. Capture Standard Output
	// We replace os.Stdout with a pipe so we can read what the library prints.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Ensure we restore stdout even if the test crashes
	defer func() {
		os.Stdout = oldStdout
	}()

	// 2. Define a simple handler that mimics real work
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Simulate some logic with a span
		ctx, finish := StartSpan(ctx, "BusinessLogic")
		AddTag(ctx, "test.tag", "integration-check")
		time.Sleep(10 * time.Millisecond) // Simulate work
		finish()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 3. Setup the Server with the Middleware
	// We use the default NewMiddleware which uses the ConsoleExporter
	mw := NewMiddleware()
	server := httptest.NewServer(mw(mockHandler))
	defer server.Close()

	// 4. Make the HTTP Request
	resp, err := http.Get(server.URL + "/test-endpoint")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// 5. WAIT for Async Export
	// The middleware runs the export in a 'go func()'.
	// We must sleep briefly to let that goroutine finish printing to stdout.
	time.Sleep(50 * time.Millisecond)

	// 6. Close the Write end of the pipe and read the output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// 7. Verify the Output
	t.Logf("Captured Output: %s", output)

	if !strings.Contains(output, "FLOW_LOG:") {
		t.Fatalf("Expected output to contain 'FLOW_LOG:', but got empty or wrong data.")
	}

	// 8. Validate the JSON structure
	// Extract the JSON part after the prefix
	jsonPart := strings.TrimSpace(strings.TrimPrefix(output, "FLOW_LOG:"))

	var trace Trace
	if err := json.Unmarshal([]byte(jsonPart), &trace); err != nil {
		t.Fatalf("Failed to unmarshal JSON from logs: %v", err)
	}

	// 9. Assert Data Integrity
	if trace.TraceID == "" {
		t.Error("TraceID should not be empty")
	}

	// Expect 2 spans: The Root Span (GET /test-endpoint) + "BusinessLogic"
	if len(trace.Spans) != 2 {
		t.Errorf("Expected 2 spans, got %d", len(trace.Spans))
	}

	// Check if our specific tag exists
	foundTag := false
	for _, s := range trace.Spans {
		if s.Tags != nil && s.Tags["test.tag"] == "integration-check" {
			foundTag = true
			break
		}
	}
	if !foundTag {
		t.Error("Did not find the expected 'test.tag' in the trace output")
	}
}
