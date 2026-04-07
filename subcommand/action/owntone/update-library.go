package owntone

import (
	"context"
	"fmt"
	"github.com/johtani/smarthome/subcommand/action"
)

// UpdateLibraryAction represents an action to update the Owntone library.
type UpdateLibraryAction struct {
	name string
	c    *Client
}

// Run executes the UpdateLibraryAction.
func (a UpdateLibraryAction) Run(ctx context.Context, args string) (string, error) {
	ctx, span := action.StartRunSpan(ctx, "owntone", "UpdateLibraryAction.Run", args)
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
