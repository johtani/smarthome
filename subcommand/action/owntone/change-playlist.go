package owntone

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
)

// ChangePlaylistAction represents an action to change the current playlist on Owntone.
type ChangePlaylistAction struct {
	name string
	c    *Client
}

// Run executes the ChangePlaylistAction.
// It picks a random playlist and starts playback.
func (a ChangePlaylistAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "ChangePlaylistAction.Run")
	defer span.End()
	msg := []string{"Change playlist to"}
	playlists, err := a.c.GetPlaylists(ctx)
	if err != nil {
		return "", fmt.Errorf("error in GetPlaylists\n %v", err)
	}
	if len(playlists) > 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		index := rand.Intn(len(playlists))
		target := playlists[index]
		msg = append(msg, target.Name+".")
		err := a.c.AddItem2QueueAndPlay(ctx, target.URI, "")
		if err != nil {
			return "", fmt.Errorf("error in AddItem2QueueAndPlay(target=%v)\n %v", target.Name, err)
		}
	} else {
		msg = append(msg, "playlists is empty\n")
	}
	return strings.Join(msg, " "), nil
}

// NewChangePlaylistAction creates a new ChangePlaylistAction.
func NewChangePlaylistAction(client *Client) ChangePlaylistAction {
	return ChangePlaylistAction{
		name: "Change playlist on Owntone",
		c:    client,
	}
}
