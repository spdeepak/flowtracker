package main

import (
	"net/http"

	"github.com/spdeepak/flowtracker"
	"github.com/spdeepak/flowtracker/examples"
	"github.com/spdeepak/flowtracker/exporters"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", examples.Handler)

	slogExporter := exporters.MermaidExporter{}
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(&slogExporter), flowtracker.WithExporter(&exporters.SlogExporter{}))

	http.ListenAndServe(":8080", mw(mux))
}
