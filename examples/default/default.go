package main

import (
	"net/http"

	"github.com/spdeepak/flowtracker"
	"github.com/spdeepak/flowtracker/examples"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", examples.Handler)

	// No arguments = Default ConsoleExporter
	mw := flowtracker.NewMiddleware()

	http.ListenAndServe(":8080", mw(mux))
}
