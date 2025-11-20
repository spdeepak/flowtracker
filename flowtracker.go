package flowtracker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// ---------------------------------------------------------
// 1. Data Structures
// ---------------------------------------------------------

type Span struct {
	ID        string            `json:"span_id"`
	ParentID  string            `json:"parent_id,omitempty"`
	Name      string            `json:"name"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Duration  int64             `json:"duration_ms"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type Trace struct {
	TraceID string  `json:"trace_id"`
	Root    *Span   `json:"-"`
	Spans   []*Span `json:"spans"`
	mu      sync.Mutex
}

// ---------------------------------------------------------
// 2. Exporter Interface & Default Implementation
// ---------------------------------------------------------

// Exporter defines how the Trace data is handled when a request finishes.
type Exporter interface {
	Export(trace *Trace)
}

// ConsoleExporter -- Default Impl 1: Console Exporter (JSON to Stdout) --
type ConsoleExporter struct{}

func (c *ConsoleExporter) Export(tr *Trace) {
	b, _ := json.Marshal(tr)
	fmt.Printf("FLOW_LOG: %s\n", string(b))
}

// ---------------------------------------------------------
// 3. Configuration Options
// ---------------------------------------------------------

type config struct {
	exporter Exporter
}

type Option func(*config)

// WithExporter allows the user to inject a custom or default exporter
func WithExporter(e Exporter) Option {
	return func(c *config) {
		c.exporter = e
	}
}

// ---------------------------------------------------------
// 4. Middleware & Logic
// ---------------------------------------------------------

type key int

const (
	traceKey      key = 0
	parentSpanKey key = 1
)

// NewMiddleware creates the handler wrapper with the provided options
func NewMiddleware(opts ...Option) func(http.Handler) http.Handler {
	// Default configuration
	cfg := &config{
		exporter: &ConsoleExporter{}, // Default to console if nothing passed
	}

	// Apply user options
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Initialize Trace
			traceID := fmt.Sprintf("trace-%d-%d", time.Now().UnixNano(), rand.Intn(1000))
			rootSpan := &Span{
				ID:        fmt.Sprintf("%d", rand.Intn(100000)),
				Name:      fmt.Sprintf("%s %s", r.Method, r.URL.Path),
				StartTime: time.Now(),
			}

			tr := &Trace{
				TraceID: traceID,
				Root:    rootSpan,
				Spans:   []*Span{rootSpan},
			}

			// Inject into Context
			ctx := context.WithValue(r.Context(), traceKey, tr)
			ctx = context.WithValue(ctx, parentSpanKey, rootSpan.ID)

			// Serve Request
			next.ServeHTTP(w, r.WithContext(ctx))

			// Finalize Root Span
			rootSpan.EndTime = time.Now()
			rootSpan.Duration = rootSpan.EndTime.Sub(rootSpan.StartTime).Milliseconds()

			// Export Data (Run in goroutine to avoid blocking the API response)
			go cfg.exporter.Export(tr)
		})
	}
}

// StartSpan starts a new step in the flow
func StartSpan(ctx context.Context, name string) (context.Context, func()) {
	trace, ok := ctx.Value(traceKey).(*Trace)
	if !ok {
		return ctx, func() {}
	}

	parentID, _ := ctx.Value(parentSpanKey).(string)

	span := &Span{
		ID:        fmt.Sprintf("%d", rand.Intn(1000000)),
		ParentID:  parentID,
		Name:      name,
		StartTime: time.Now(),
	}

	trace.mu.Lock()
	trace.Spans = append(trace.Spans, span)
	trace.mu.Unlock()

	newCtx := context.WithValue(ctx, parentSpanKey, span.ID)

	return newCtx, func() {
		span.EndTime = time.Now()
		span.Duration = span.EndTime.Sub(span.StartTime).Milliseconds()
	}
}

// AddTag adds metadata to the current span
func AddTag(ctx context.Context, key, value string) {
	trace, ok := ctx.Value(traceKey).(*Trace)
	if !ok {
		return
	}
	currentSpanID, _ := ctx.Value(parentSpanKey).(string)

	trace.mu.Lock()
	defer trace.mu.Unlock()

	for _, s := range trace.Spans {
		if s.ID == currentSpanID {
			if s.Tags == nil {
				s.Tags = make(map[string]string)
			}
			s.Tags[key] = value
			break
		}
	}
}
