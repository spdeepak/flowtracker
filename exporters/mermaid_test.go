package exporters

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spdeepak/flowtracker"
)

func NewServer(orientation Orientation, tags []string, includeAllTags bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", Handler)
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(&MermaidExporter{Orientation: orientation, IncludeTags: tags, IncludeAllTags: includeAllTags}))
	return mw(mux)
}

func TestMermaidExporter_OK(t *testing.T) {
	// capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	server := NewServer("", []string{}, true)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	time.Sleep(500 * time.Millisecond)

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

	server := NewServer(LeftRight, []string{"db.query"}, false)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	time.Sleep(500 * time.Millisecond)

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
