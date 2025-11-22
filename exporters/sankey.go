package exporters

import (
	"fmt"
	"sort"
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

	// IncludeAllTags overrides IncludeTags. If true, ALL tags present
	IncludeAllTags bool
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
		// Logic: Decide which keys to show
		if s.IncludeAllTags {
			// Get ALL keys from the map
			for key, val := range span.Tags {
				tagSuffixes = append(tagSuffixes, fmt.Sprintf("%s:%s", key, val))
			}
			// Sort keys to ensure deterministic diagram output
			sort.Strings(tagSuffixes)
		} else if len(s.IncludeTags) > 0 && span.Tags != nil {
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

	// 2. Use strings.Builder to construct the output block
	var sb strings.Builder

	if !s.CleanOutput {
		sb.WriteString(fmt.Sprintf("\n----- START SANKEY DATA (trace id: %s)----\n", tr.TraceID))
	}

	for _, span := range tr.Spans {
		if span.ParentID == "" {
			continue
		}

		parentName, ok := spanNames[span.ParentID]
		if !ok {
			parentName = "Unknown"
		}
		currentName := spanNames[span.ID]

		// Format: Source [Weight] Target\n
		sb.WriteString(fmt.Sprintf("%s [%d] %s\n", parentName, span.Duration, currentName))
	}

	if !s.CleanOutput {
		sb.WriteString(fmt.Sprintf("----- END SANKEY DATA (trace id: %s)----\n", tr.TraceID))
	}

	// 3. Print everything in one atomic operation
	fmt.Print(sb.String())
}
