package otel

import (
	"context"

	"github.com/spdeepak/flowtracker"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// OTelExporter converts FlowTracker traces into OpenTelemetry Spans.
type OTelExporter struct {
	tracer trace.Tracer
}

// New creates a new OTelExporter.
// You can pass a specific TracerProvider, or nil to use the global global.TracerProvider().
func New(tp trace.TracerProvider) *OTelExporter {
	if tp == nil {
		tp = otel.GetTracerProvider()
	}
	return &OTelExporter{
		tracer: tp.Tracer("github.com/spdeepak/flowtracker"),
	}
}

func (e *OTelExporter) Export(tr *flowtracker.Trace) {
	if tr.Root == nil {
		return
	}

	// 1. Map FlowTracker Span IDs to the corresponding *flowtracker.Span
	//    This helps us traverse the tree easily.
	spanMap := make(map[string]*flowtracker.Span)
	for _, s := range tr.Spans {
		spanMap[s.ID] = s
	}

	// 2. Recursive function to create OTel spans
	//    We use recursion to ensure the Parent OTel Context is created
	//    before the Child OTel Span is started.
	var createSpan func(node *flowtracker.Span, parentCtx context.Context)
	createSpan = func(node *flowtracker.Span, parentCtx context.Context) {

		// A. Convert Tags to OTel Attributes
		var attrs []attribute.KeyValue
		// Add the original FlowTracker IDs as attributes for cross-referencing
		attrs = append(attrs, attribute.String("flowtracker.trace_id", tr.TraceID))
		attrs = append(attrs, attribute.String("flowtracker.span_id", node.ID))

		if node.Tags != nil {
			for k, v := range node.Tags {
				attrs = append(attrs, attribute.String(k, v))
			}
		}

		// B. Start the OTel Span "retroactively"
		//    We use WithTimestamp to tell OTel exactly when this happened in the past.
		ctx, span := e.tracer.Start(parentCtx, node.Name,
			trace.WithTimestamp(node.StartTime),
			trace.WithAttributes(attrs...),
			// You could map span kinds here if you added that to your library later
			trace.WithSpanKind(trace.SpanKindInternal),
		)

		// C. Check for errors (simple convention: if tag "error" exists, mark span as error)
		if val, ok := node.Tags["error"]; ok && val == "true" {
			span.SetStatus(codes.Error, "Error flagged in FlowTracker")
		}

		// D. End the span "retroactively"
		span.End(trace.WithTimestamp(node.EndTime))

		// E. Find children and process them
		//    (Naive search is O(N^2), but N is usually small < 100 per request)
		for _, s := range tr.Spans {
			if s.ParentID == node.ID {
				createSpan(s, ctx) // Pass the NEW OTel context down
			}
		}
	}

	// 3. Kick off with the Root Span
	//    We use a background context because the root has no parent.
	createSpan(tr.Root, context.Background())
}
