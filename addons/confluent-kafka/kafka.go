package confluent_kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spdeepak/flowtracker"
)

// Config holds the setup parameters.
// You must provide EITHER Producer OR KafkaConfigMap.
type Config struct {
	// Topic is the destination Kafka topic name (Required).
	Topic string

	// Producer allows you to pass an existing Kafka Producer.
	// If this is set, KafkaConfigMap is ignored.
	Producer *kafka.Producer

	// KafkaConfigMap allows you to configure a new Producer.
	// Use this if you don't have an existing client.
	// Example: &kafka.ConfigMap{"bootstrap.servers": "localhost:9092"}
	KafkaConfigMap *kafka.ConfigMap
}

// KafkaExporter implements the flowtracker.Exporter interface.
type KafkaExporter struct {
	producer *kafka.Producer
	topic    string
	// isOwned tracks if this exporter created the producer (and thus should close it).
	isOwned bool
}

// New creates a new KafkaExporter.
func New(cfg Config) (*KafkaExporter, error) {
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka exporter: topic is required")
	}

	var p *kafka.Producer
	var err error
	isOwned := false

	if cfg.Producer != nil {
		// Use the user-provided producer
		p = cfg.Producer
	} else if cfg.KafkaConfigMap != nil {
		// Initialize a new producer
		p, err = kafka.NewProducer(cfg.KafkaConfigMap)
		if err != nil {
			return nil, fmt.Errorf("kafka exporter: failed to create producer: %w", err)
		}
		isOwned = true

		// Background goroutine to handle delivery reports.
		// Essential for confluent-kafka-go to prevent local queue filling up.
		go func() {
			for e := range p.Events() {
				switch ev := e.(type) {
				case *kafka.Message:
					if ev.TopicPartition.Error != nil {
						log.Printf("FlowTracker Kafka Error: %v\n", ev.TopicPartition.Error)
					}
				}
			}
		}()
	} else {
		return nil, fmt.Errorf("kafka exporter: must provide either Producer or KafkaConfigMap")
	}

	return &KafkaExporter{
		producer: p,
		topic:    cfg.Topic,
		isOwned:  isOwned,
	}, nil
}

// Export sends the trace to Kafka.
func (k *KafkaExporter) Export(tr *flowtracker.Trace) {
	// Serialize Trace to JSON
	payload, err := json.Marshal(tr)
	if err != nil {
		log.Printf("FlowTracker: failed to marshal trace: %v", err)
		return
	}

	// Construct the Kafka Message
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Value:          payload,
		// We use the TraceID as the Key. This ensures that if you update this logic
		// to stream updates, all spans for the same trace go to the same partition.
		Key: []byte(tr.TraceID),
	}

	// Produce is asynchronous. We rely on the background event loop (started in New) to handle errors.
	err = k.producer.Produce(msg, nil)
	if err != nil {
		log.Printf("FlowTracker: failed to produce message: %v", err)
	}
}

// Close flushes the producer if we own it.
// Note: The main flowtracker library doesn't call Close(), but you can call this
// manually in your main.go shutdown hook.
func (k *KafkaExporter) Close() {
	if k.isOwned && k.producer != nil {
		// Wait up to 5 seconds for outstanding messages to be delivered
		k.producer.Flush(5000)
		k.producer.Close()
	}
}
