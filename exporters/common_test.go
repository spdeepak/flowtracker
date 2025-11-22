package exporters

import (
	"context"
	"net/http"
	"time"

	"github.com/spdeepak/flowtracker"
)

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
