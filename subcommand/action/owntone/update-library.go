package owntone

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

// UpdateLibraryAction represents an action to update the Owntone library.
type UpdateLibraryAction struct {
	name string
	c    *Client
}

// Run executes the UpdateLibraryAction.
func (a UpdateLibraryAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "UpdateLibraryAction.Run")
	defer span.End()
	err := a.c.UpdateLibrary(ctx)
	if err != nil {
		return "", fmt.Errorf("error in ClearQueue\n %v", err)
	}
	return "Updated library", nil
}

// NewUpdateLibraryAction creates a new UpdateLibraryAction.
func NewUpdateLibraryAction(client *Client) UpdateLibraryAction {
	return UpdateLibraryAction{
		name: "Update library on Owntone",
		c:    client,
	}
}
