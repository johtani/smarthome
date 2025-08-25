package owntone

import (
	"fmt"
	"strings"
)

type DisplayPlaylistsAction struct {
	name string
	c    *Client
}

// Run
// Playlistの一覧を取得して文字列として返す
func (a DisplayPlaylistsAction) Run(category string) (string, error) {
	msg := []string{"Playlists are..."}
	flg := onlySpotify(category)
	playlists, err := a.c.GetPlaylists()
	if err != nil {
		return "", fmt.Errorf("error in GetPlaylists\n %v", err)
	}
	for _, playlist := range playlists {
		if strings.HasPrefix(playlist.Path, "spotify:") {
			msg = append(msg, fmt.Sprintf("  %v (by Spotify)", playlist.Name))
		} else if flg == false {
			msg = append(msg, fmt.Sprintf("  %v", playlist.Name))
		}
	}

	return strings.Join(msg, " \n"), nil
}

func NewDisplayPlaylistsAction(client *Client) DisplayPlaylistsAction {
	return DisplayPlaylistsAction{
		name: "Display playlists from Owntone",
		c:    client,
	}
}

func onlySpotify(category string) bool {
	return category == "spotify"
}
