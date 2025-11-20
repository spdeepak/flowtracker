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
