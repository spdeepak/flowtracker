package main

import (
	"context"
	"net/http"

	"github.com/spdeepak/flowtracker"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", slogHandler)

	slogExporter := flowtracker.SlogExporter{}
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(&slogExporter))

	http.ListenAndServe(":8080", mw(mux))
}

func slogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Start a span for a DB call
	// Always defer the finish function!
	ctx, finish := flowtracker.StartSpan(ctx, "FetchFromDB")
	defer finish()

	// 2. Simulate a Database call
	processSlogDatabaseLogic(ctx)

	// 3. Simulate external Microservice call
	processSlogServiceCall(ctx)

	w.Write([]byte("Order Processed"))
}

func processSlogDatabaseLogic(ctx context.Context) {
	// Add metadata to the current span
	flowtracker.AddTag(ctx, "db.query", "SELECT * FROM users")
}

func processSlogServiceCall(ctx context.Context) {
	// Add metadata to the current span
	flowtracker.AddTag(ctx, "ext.service", "Calling endpoint /dummy")
}
