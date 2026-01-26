package owntone

import (
	"context"
	"fmt"
)

type UpdateLibraryAction struct {
	name string
	c    *Client
}

func (a UpdateLibraryAction) Run(ctx context.Context, _ string) (string, error) {
	err := a.c.UpdateLibrary(ctx)
	if err != nil {
		return "", fmt.Errorf("error in ClearQueue\n %v", err)
	}
	return "Updated library", nil
}

func NewUpdateLibraryAction(client *Client) UpdateLibraryAction {
	return UpdateLibraryAction{
		name: "Update library on Owntone",
		c:    client,
	}
}
