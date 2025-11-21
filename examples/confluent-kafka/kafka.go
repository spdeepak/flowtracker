package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spdeepak/flowtracker"
	confluentkafka "github.com/spdeepak/flowtracker/confluent-kafka"
	"github.com/spdeepak/flowtracker/examples"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", examples.Handler)

	config := confluentkafka.Config{
		Topic: "flowtracker.logs",
		KafkaConfigMap: &kafka.ConfigMap{
			"bootstrap.servers": "localhost:9092",
			"client.id":         "flowtracker-client",
			"acks":              "all",
		},
	}
	exporter, err := confluentkafka.New(config)
	if err != nil {
		slog.Error("Error during kafka exporter creation", slog.Any("error", err))
		os.Exit(1)
	}
	mw := flowtracker.NewMiddleware(flowtracker.WithExporter(exporter))

	http.ListenAndServe(":8080", mw(mux))
}
