package exporters

import (
	"encoding/json"
	"log/slog"

	"github.com/spdeepak/flowtracker"
)

// SlogExporter -- log Exporter (JSON to log) --
type SlogExporter struct{}

func (c *SlogExporter) Export(tr *flowtracker.Trace) {
	b, _ := json.Marshal(tr)
	slog.Info("Flow log", slog.Any("trace", b))
}
