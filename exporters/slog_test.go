package exporters

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spdeepak/flowtracker"
)

var buf bytes.Buffer

func NewSlogServer(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", Handler)
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(&SlogExporter{logger: logger}))
	return mw(mux)
}

func TestSlogExporter_OK(t *testing.T) {
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	server := NewSlogServer(logger)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	time.Sleep(500 * time.Millisecond)

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to unmarshal log output: %v", err)
	}
	now := time.Now()
	if duration, err := time.Parse(time.RFC3339Nano, logEntry["time"].(string)); err != nil {
		t.Fatalf("Failed to unmarshal time in logEntry: %v", err)
	} else if now.Sub(duration).Milliseconds() < 0 {
		t.Fatalf("Invalid time in logEntry")
	}
}

func TestSlogExporter_OK_DefaultLogger(t *testing.T) {
	server := NewSlogServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	time.Sleep(500 * time.Millisecond)

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to unmarshal log output: %v", err)
	}
	now := time.Now()
	if duration, err := time.Parse(time.RFC3339Nano, logEntry["time"].(string)); err != nil {
		t.Fatalf("Failed to unmarshal time in logEntry: %v", err)
	} else if now.Sub(duration).Milliseconds() < 0 {
		t.Fatalf("Invalid time in logEntry")
	}
}
