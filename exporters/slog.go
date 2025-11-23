package exporters

import (
	"encoding/json"
	"log/slog"

	"github.com/spdeepak/flowtracker"
)

// SlogExporter -- log Exporter (JSON to log) --
type SlogExporter struct {
	logger *slog.Logger
}

func (s *SlogExporter) Export(tr *flowtracker.Trace) {
	b, _ := json.Marshal(tr)
	if s.logger != nil {
		s.logger.Info("Flow log", slog.Any("trace", b))
	} else {
		slog.Info("Flow log", slog.Any("trace", b))
	}
}
