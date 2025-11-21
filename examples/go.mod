module github.com/spdeepak/flowtracker/examples

go 1.24

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/spdeepak/flowtracker v0.0.3
	github.com/spdeepak/flowtracker/confluent-kafka v0.0.1
	github.com/spdeepak/flowtracker/otel v0.0.1
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
)

require (
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
)

replace (
	github.com/spdeepak/flowtracker => ../
	github.com/spdeepak/flowtracker/confluent-kafka => ../addons/confluent-kafka/
	github.com/spdeepak/flowtracker/otel => ../addons/otel
)
