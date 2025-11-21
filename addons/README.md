# 🧩 FlowTracker Addons

This directory contains official extensions for the **[FlowTracker](https://github.com/spdeepak/flowtracker)** library.

### Why are these separate?
To keep the core `flowtracker` library lightweight and dependency-free.
*   **Core Library:** Zero heavy dependencies. Just standard Go.
*   **Addons:** May require heavy SDKs (Kafka, OpenTelemetry).

By using a **Multi-Module** architecture, you only download the dependencies for the specific exporters you actually use.

---

## 📦 Available Addons

| Addon                                            | Description | Dependencies |
|:-------------------------------------------------| :--- | :--- |
| **[Kafka Exporter](./confluent-kafka)** | Pushes trace data to an Apache Kafka topic. | `confluent-kafka-go` |
| **[OpenTelemetry Bridge](./otel)**      | Sends traces to Jaeger, Grafana Tempo, Datadog, etc. | `go.opentelemetry.io` |

---

## 🚀 Installation

Because these are separate Go modules, you must `go get` the specific addon you need.

**Example: Installing the Kafka Exporter**
```bash
# 1. Get the core
go get github.com/spdeepak/flowtracker

# 2. Get the addon
go get github.com/spdeepak/flowtracker/kafka-exporter
```

---

## 🛠 General Usage

All addons implement the standard `Exporter` interface defined in the core library. Usage is consistent across all extensions:

```go
package main

import (
    "github.com/spdeepak/flowtracker"
    "github.com/spdeepak/flowtracker/kafka-exporter" // Import the addon
)

func main() {
    // 1. Initialize the Addon
    // (Refer to the specific addon's README for config details)
    myExporter, _ := kafkaexporter.New(...)

    // 2. Inject into Middleware
    mw := flowtracker.NewMiddleware(
        flowtracker.WithExporter(myExporter),
    )

    // 3. Run Server
    http.ListenAndServe(":8080", mw(mux))
}
```

---

## 👩‍💻 Contributing a New Addon

We welcome contributions! If you want to create a new exporter (e.g., for AWS S3, Elasticsearch, or Slack), follow these steps:

1.  **Create a directory:**
    ```bash
    mkdir addons/my-new-exporter
    cd addons/my-new-exporter
    ```

2.  **Initialize the Module:**
    *Important: The module name must include the full path.*
    ```bash
    go mod init github.com/spdeepak/flowtracker/my-new-exporter
    ```

3.  **Develop Locally:**
    In your `go.mod`, use a replace directive to point to the local core library during development:
    ```go
    require github.com/spdeepak/flowtracker v0.0.0
    replace github.com/spdeepak/flowtracker => ../../flowtracker
    ```

4.  **Implement the Interface:**
    Your struct must satisfy:
    ```go
    type Exporter interface {
        Export(trace *flowtracker.Trace)
    }
    ```

5.  **Release:**
    When ready to merge, remove the `replace` directive. Release tags for addons must follow the pattern:
    `addons/my-new-exporter/v0.0.1`