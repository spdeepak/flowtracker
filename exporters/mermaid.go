package exporters

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spdeepak/flowtracker"
)

// MermaidExporter outputs the trace in Mermaid.js Flowchart syntax.
type MermaidExporter struct {
	// Orientation can be "TD" (Top-Down) or "LR" (Left-Right). Default "TD".
	Orientation Orientation

	// IncludeTags defines a specific list of tags to append to the span name.
	// All tags will be included if this is empty.
	IncludeTags []string
}

type Orientation string

var (
	TopDown   Orientation = "TD"
	LeftRight Orientation = "LR"
)

// Export visual representation of output can be seen using https://mermaid.live/
func (m *MermaidExporter) Export(tr *flowtracker.Trace) {
	var sb strings.Builder

	// 1. Header
	sb.WriteString("\n----- MERMAID OUTPUT -----\n")
	sb.WriteString("```mermaid\n")

	orient := m.Orientation
	if orient == "" {
		orient = TopDown
	}
	sb.WriteString(fmt.Sprintf("graph %s\n", orient))

	// 2. Helper to escape strings
	escape := func(s string) string {
		return strings.ReplaceAll(s, "\"", "'")
	}

	// 3. Pre-calculate Node Labels
	nodeLabels := make(map[string]string)

	for _, span := range tr.Spans {
		name := span.Name
		var tagSuffixes []string

		if span.Tags != nil && len(span.Tags) > 0 {
			var keysToDisplay []string

			// Logic: Decide which keys to show
			if len(m.IncludeTags) > 0 {
				// Get only user-specified keys
				for _, k := range m.IncludeTags {
					if _, exists := span.Tags[k]; exists {
						keysToDisplay = append(keysToDisplay, k)
					}
				}
				// No need to sort strict list, user order is preserved
			} else {
				// Get ALL keys from the map
				for k := range span.Tags {
					keysToDisplay = append(keysToDisplay, k)
				}
				// Sort keys to ensure deterministic diagram output
				sort.Strings(keysToDisplay)
			}

			// Build the display string
			for _, key := range keysToDisplay {
				val := span.Tags[key]
				tagSuffixes = append(tagSuffixes, fmt.Sprintf("%s:%s", key, val))
			}
		}

		// Append tags to name: "SpanName (key:val, key2:val)"
		if len(tagSuffixes) > 0 {
			name = fmt.Sprintf("%s (%s)", name, strings.Join(tagSuffixes, ", "))
		}

		nodeLabels[span.ID] = escape(name)
	}

	// 4. Build Links
	for _, span := range tr.Spans {
		if span.ParentID == "" {
			continue
		}

		parentLabel, pOk := nodeLabels[span.ParentID]
		childLabel, cOk := nodeLabels[span.ID]

		if !pOk || !cOk {
			continue
		}

		// Syntax: N<ID>["Label"] -->|Duration| N<ID>["Label"]
		sb.WriteString(fmt.Sprintf("    N%s[\"%s\"] -->|%dms| N%s[\"%s\"]\n",
			span.ParentID, parentLabel,
			span.Duration,
			span.ID, childLabel,
		))
	}

	// 5. Footer
	sb.WriteString("```\n")
	sb.WriteString("--------------------------\n\n")

	fmt.Print(sb.String())
}
