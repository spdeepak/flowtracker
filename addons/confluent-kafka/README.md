# FlowTracker Kafka Exporter

This is an addon for the [FlowTracker](https://github.com/spdeepak/flowtracker) library. It implements the `Exporter` interface to asynchronously push trace data to a **Apache Kafka** topic using the [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go) library.

## 📦 Installation

Since this is an addon module, you must install it alongside the core library:

```bash
# Install Core
go get github.com/spdeepak/flowtracker

# Install Kafka Exporter
go get github.com/spdeepak/flowtracker/kafka-exporter
```

*Note: This library uses `confluent-kafka-go` which requires CGO to be enabled.*

## 🚀 Usage

You can configure the exporter in two ways: letting the library manage the connection, or bringing your own producer.

### Option 1: Simple Configuration (Library creates Producer)
Use this if you want FlowTracker to manage the Kafka connection lifecycle.

```go
package main

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spdeepak/flowtracker"
	"github.com/spdeepak/flowtracker/kafka-exporter"
)

func main() {
	// 1. Configure the Exporter
	exporter, err := kafkaexporter.New(kafkaexporter.Config{
		Topic: "microservice-traces",
		KafkaConfigMap: &kafka.ConfigMap{
			"bootstrap.servers": "localhost:9092",
			"client.id":         "flowtracker-exporter",
			"acks":              "all",
		},
	})
	if err != nil {
		panic(err)
	}
	// Ensure outstanding messages are flushed on shutdown
	defer exporter.Close()

	// 2. Apply Middleware
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(exporter))
	
	// ... start your server
}
```

### Option 2: Reuse Existing Producer
Use this if your application already has a Kafka Producer instance and you want to reuse the connection.

```go
func main() {
    // Assuming 'myAppProducer' is your existing *kafka.Producer
    
    exporter, _ := kafkaexporter.New(kafkaexporter.Config{
        Topic:    "microservice-traces",
        Producer: myAppProducer,
    })

    mw := flowtracker.NewMiddleware(flowtracker.WithExporter(exporter))
}
```

## 📝 Data Format

The exporter sends data to Kafka in the following format:

*   **Key:** The `trace_id` (String). This ensures all spans for a specific trace land on the same Kafka partition.
*   **Value:** JSON String of the `Trace` object.

**Example Payload:**
```json
{
  "trace_id": "trace-1700234-99",
  "spans": [
    {
      "span_id": "101",
      "name": "GET /api/checkout",
      "duration_ms": 150
    },
    {
      "span_id": "202", 
      "parent_id": "101",
      "name": "DB: Process Order",
      "duration_ms": 45
    }
  ]
}
```

## 🔗 Examples

You can find a complete, runnable example of how to use this exporter in the main repository:

👉 **[Click here for the Kafka Example Code](../../examples/confluent-kafka)**

*(Note: If the link above doesn't work, navigate to the `examples/kafka` directory in the repository root).*