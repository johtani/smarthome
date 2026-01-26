package owntone

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type ChangePlaylistAction struct {
	name string
	c    *Client
}

// Run
// キューにプレイリストを追加して再生する
func (a ChangePlaylistAction) Run(ctx context.Context, _ string) (string, error) {
	msg := []string{"Change playlist to"}
	playlists, err := a.c.GetPlaylists(ctx)
	if err != nil {
		return "", fmt.Errorf("error in GetPlaylists\n %v", err)
	}
	if len(playlists) > 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		index := rand.Intn(len(playlists))
		target := playlists[index]
		msg = append(msg, fmt.Sprintf("%v.", target.Name))
		err := a.c.AddItem2QueueAndPlay(ctx, target.Uri, "")
		if err != nil {
			return "", fmt.Errorf("error in AddItem2QueueAndPlay(target=%v)\n %v", target.Name, err)
		}
	} else {
		msg = append(msg, fmt.Sprintf("playlists is empty\n"))
	}
	return strings.Join(msg, " "), nil
}

func NewChangePlaylistAction(client *Client) ChangePlaylistAction {
	return ChangePlaylistAction{
		name: "Change playlist on Owntone",
		c:    client,
	}
}
