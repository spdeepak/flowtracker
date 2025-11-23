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

func NewSankeyServer(tags []string, includeAllTags bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", Handler)
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(&SankeyExporter{IncludeTags: tags, IncludeAllTags: includeAllTags}))
	return mw(mux)
}

func TestSankeyExporter_OK(t *testing.T) {
	// capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	server := NewSankeyServer([]string{}, true)

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
	if !strings.Contains(output[1], "----- START SANKEY DATA (trace id: trace-") {
		t.Fatalf("expected log not found: %s", output[1])
	}
	if !strings.Contains(output[2], "GET / [") {
		t.Fatalf("expected log not found: %s", output[2])
	}
	if !strings.Contains(output[2], "] Process Payment (currency:USD)") {
		t.Fatalf("expected log not found: %s", output[2])
	}
	if !strings.Contains(output[3], "GET / [") {
		t.Fatalf("expected log not found: %s", output[3])
	}
	if !strings.Contains(output[3], "] DB: Select User (db.query:SELECT * FROM users...)") {
		t.Fatalf("expected log not found: %s", output[3])
	}
	if !strings.Contains(output[4], "GET / [") {
		t.Fatalf("expected log not found: %s", output[4])
	}
	if !strings.Contains(output[4], "] HTTP: Shipping Service") {
		t.Fatalf("expected log not found: %s", output[4])
	}
	if !strings.Contains(output[5], "HTTP: Shipping Service [") {
		t.Fatalf("expected log not found: %s", output[5])
	}
	if !strings.Contains(output[5], "] Calculate Weight") {
		t.Fatalf("expected log not found: %s", output[5])
	}
	if !strings.Contains(output[6], "----- END SANKEY DATA (trace id: trace-") {
		t.Fatalf("expected log not found: %s", output[6])
	}
}

func TestSankeyExporter_OK_IncludeTags(t *testing.T) {
	// capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	server := NewSankeyServer([]string{"db.query"}, false)

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
	if !strings.Contains(output[1], "----- START SANKEY DATA (trace id") {
		t.Fatalf("expected log not found: %s", output[1])
	}
	if !strings.Contains(output[2], "GET / [") {
		t.Fatalf("expected log not found: %s", output[2])
	}
	if !strings.Contains(output[2], "] Process Payment") {
		t.Fatalf("expected log not found: %s", output[2])
	}
	if !strings.Contains(output[3], "GET / [") {
		t.Fatalf("expected log not found: %s", output[3])
	}
	if !strings.Contains(output[3], "] DB: Select User (db.query:SELECT * FROM users...)") {
		t.Fatalf("expected log not found: %s", output[3])
	}
	if !strings.Contains(output[4], "GET / [") {
		t.Fatalf("expected log not found: %s", output[4])
	}
	if !strings.Contains(output[4], "] HTTP: Shipping Service") {
		t.Fatalf("expected log not found: %s", output[4])
	}
	if !strings.Contains(output[5], "HTTP: Shipping Service [") {
		t.Fatalf("expected log not found: %s", output[5])
	}
	if !strings.Contains(output[5], "] Calculate Weight") {
		t.Fatalf("expected log not found: %s", output[5])
	}
	if !strings.Contains(output[6], "----- END SANKEY DATA (trace id: trace-") {
		t.Fatalf("expected log not found: %s", output[6])
	}
}
