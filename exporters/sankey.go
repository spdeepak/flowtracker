package exporters

import (
	"fmt"

	"github.com/spdeepak/flowtracker"
)

// SankeyExporter formats the trace for tools like SankeyMATIC.
type SankeyExporter struct {
	// CleanOutput ensures only the raw data is printed (good for piping)
	// If false, it adds a header/footer for readability in logs.
	CleanOutput bool
}

func (s *SankeyExporter) Export(tr *flowtracker.Trace) {
	// 1. Map IDs to Names for easy lookup
	// Note: If multiple spans have the exact same name, they will be grouped
	// together in the Sankey diagram, which is usually desired behavior.
	spanNames := make(map[string]string)
	for _, span := range tr.Spans {
		spanNames[span.ID] = span.Name
	}

	var output []string

	// 2. Build the links: Parent [Duration] Child
	for _, span := range tr.Spans {
		// Skip the root node acting as a child (it has no parent)
		if span.ParentID == "" {
			continue
		}

		parentName, ok := spanNames[span.ParentID]
		if !ok {
			parentName = "Unknown"
		}

		// Format: Source [Amount] Target
		// We use Duration as the "Amount" (width of the flow)
		line := fmt.Sprintf("%s [%d] %s", parentName, span.Duration, span.Name)
		output = append(output, line)
	}

	// 3. Print the result
	if !s.CleanOutput {
		fmt.Printf("\n----- COPY BELOW THIS LINE FOR SANKEY (trace id: %s)----\n", tr.TraceID)
	}

	for _, line := range output {
		fmt.Println(line)
	}

	if !s.CleanOutput {
		fmt.Printf("----- END SANKEY DATA (trace id: %s)----\n", tr.TraceID)
	}
}
