package exporters

import (
	"fmt"
	"strings"

	"github.com/spdeepak/flowtracker"
)

// SankeyExporter formats the trace for tools like SankeyMATIC.
type SankeyExporter struct {
	// CleanOutput ensures only the raw data is printed (good for piping)
	// If false, it adds a header/footer for readability in logs.
	CleanOutput bool

	// List of tag keys you want to display in the diagram.
	// Example: []string{"db.table", "http.status_code"}
	IncludeTags []string
}

func (s *SankeyExporter) Export(tr *flowtracker.Trace) {
	// 1. Map IDs to Names for easy lookup
	// Note: If multiple spans have the exact same name, they will be grouped
	// together in the Sankey diagram, which is usually desired behavior.
	spanNames := make(map[string]string)

	// 1. Build names with appended tags
	for _, span := range tr.Spans {
		name := span.Name

		// Check if user wants to see tags for this span
		var tagSuffixes []string
		if len(s.IncludeTags) > 0 && span.Tags != nil {
			for _, key := range s.IncludeTags {
				if val, ok := span.Tags[key]; ok {
					// Format: (key:value)
					tagSuffixes = append(tagSuffixes, fmt.Sprintf("%s:%s", key, val))
				}
			}
		}

		// If we found relevant tags, append them to the name
		// Result: "DB Query (db.table:users)"
		if len(tagSuffixes) > 0 {
			name = fmt.Sprintf("%s (%s)", name, strings.Join(tagSuffixes, ", "))
		}

		spanNames[span.ID] = name
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

		currentName := spanNames[span.ID]

		// Sankey Format: Source [Weight] Target
		line := fmt.Sprintf("%s [%d] %s", parentName, span.Duration, currentName)
		output = append(output, line)
	}

	// 3. Print the result
	if !s.CleanOutput {
		fmt.Printf("\n----- START SANKEY DATA (trace id: %s)----\n", tr.TraceID)
	}

	for _, line := range output {
		fmt.Println(line)
	}

	if !s.CleanOutput {
		fmt.Printf("----- END SANKEY DATA (trace id: %s)----\n", tr.TraceID)
	}
}
