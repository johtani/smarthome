package owntone

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
)

// DisplayPlaylistsAction represents an action to displayplaylists from Owntone.
type DisplayPlaylistsAction struct {
	name string
	c    *Client
}

// Run executes the DisplayPlaylistsAction and returns a formatted string of playlists.
func (a DisplayPlaylistsAction) Run(ctx context.Context, category string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "DisplayPlaylistsAction.Run")
	defer span.End()
	msg := []string{"Playlists are..."}
	flg := onlySpotify(category)
	playlists, err := a.c.GetPlaylists(ctx)
	if err != nil {
		return "", fmt.Errorf("error in GetPlaylists\n %v", err)
	}
	for _, playlist := range playlists {
		if strings.HasPrefix(playlist.Path, "spotify:") {
			msg = append(msg, fmt.Sprintf("  %v (by Spotify)", playlist.Name))
		} else if !flg {
			msg = append(msg, fmt.Sprintf("  %v", playlist.Name))
		}
	}

	return strings.Join(msg, " \n"), nil
}

// NewDisplayPlaylistsAction creates a new DisplayPlaylistsAction.
func NewDisplayPlaylistsAction(client *Client) DisplayPlaylistsAction {
	return DisplayPlaylistsAction{
		name: "Display playlists from Owntone",
		c:    client,
	}
}

func onlySpotify(category string) bool {
	return category == "spotify"
}
