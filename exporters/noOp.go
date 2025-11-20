package exporters

import "github.com/spdeepak/flowtracker"

// NoOpExporter -- NoOp Exporter (Does nothing, for testing) --
type NoOpExporter struct{}

func (n *NoOpExporter) Export(tr *flowtracker.Trace) {}
