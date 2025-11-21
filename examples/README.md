# flowtracker examples

This repo has all the examples on how to use the flowtracker library

Notes:

If you get error when you run `go mod tidy` or `go mod vendor` in examples folder related to `otlptracehttp` try running

```
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
```

Then run `go mod tidy` and `go mod vendor`