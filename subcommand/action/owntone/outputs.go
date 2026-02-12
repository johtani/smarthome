package owntone

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.opentelemetry.io/otel"
)

// DisplayOutputsAction lists Owntone outputs ordered by Selected=true first, then false.
// It prints Name, Selected, Volume for each output.
type DisplayOutputsAction struct {
	name                string
	c                   *Client
	defaultOnlySelected bool
}

// NewDisplayOutputsAction creates a DisplayOutputsAction.
// Optionally pass a boolean to set defaultOnlySelected (true to show only selected outputs by default).
// If no boolean is provided, the default is false (show both selected and unselected by default).
func NewDisplayOutputsAction(client *Client, opts ...bool) DisplayOutputsAction {
	only := false
	if len(opts) > 0 {
		only = opts[0]
	}
	return DisplayOutputsAction{
		name:                "Display outputs from Owntone",
		c:                   client,
		defaultOnlySelected: only,
	}
}

// Run fetches outputs and returns a formatted string.
// Note: _ is currently ignored; behavior is controlled by defaultOnlySelected.
// Default (no args) shows both selected and unselected unless defaultOnlySelected is true.
func (a DisplayOutputsAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "DisplayOutputsAction.Run")
	defer span.End()
	outputs, err := a.c.GetOutputs(ctx)
	if err != nil {
		return "", fmt.Errorf("error in GetOutputs\n %v", err)
	}

	// Determine if we should show only selected outputs based on default setting.
	onlySelected := a.defaultOnlySelected
	if onlySelected {
		var filtered []Output
		for _, o := range outputs {
			if o.Selected {
				filtered = append(filtered, o)
			}
		}
		outputs = filtered
	}

	// Sort: Selected=true first, then false. Keep stable ordering within groups by Name.
	sort.SliceStable(outputs, func(i, j int) bool {
		if outputs[i].Selected == outputs[j].Selected {
			return strings.ToLower(outputs[i].Name) < strings.ToLower(outputs[j].Name)
		}
		return outputs[i].Selected && !outputs[j].Selected
	})

	header := "Outputs are..."
	if onlySelected {
		header = "Selected outputs are..."
	}
	lines := []string{header}
	if len(outputs) == 0 {
		lines = append(lines, "  (none)")
	} else {
		for _, o := range outputs {
			lines = append(lines, fmt.Sprintf("  Name: %s, Selected: %t, Volume: %d", o.Name, o.Selected, o.Volume))
		}
	}
	return strings.Join(lines, " \n"), nil
}
