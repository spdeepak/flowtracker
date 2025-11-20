package exporters

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spdeepak/flowtracker"
)

// FileExporter -- File Exporter (Append to a file) --
type FileExporter struct {
	Filename string
}

func (f *FileExporter) Export(tr *flowtracker.Trace) {
	b, _ := json.Marshal(tr)
	// Open file in append mode, create if not exists
	file, err := os.OpenFile(f.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error writing trace to file: %v\n", err)
		return
	}
	defer file.Close()
	if _, err := file.Write(b); err != nil {
		fmt.Printf("Error writing trace to file: %v\n", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		fmt.Printf("Error writing newline to file: %v\n", err)
	}
}
