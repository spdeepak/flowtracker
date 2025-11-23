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

	//handler := slog.NewJSONHandler(os.Stdout, nil)
	//logger := slog.New(handler)
	slogExporter := exporters.SlogExporter{}
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(&slogExporter))

	http.ListenAndServe(":8080", mw(mux))
}
