package owntone

import (
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
func (a ChangePlaylistAction) Run(_ string) (string, error) {
	msg := []string{"Change playlist to"}
	playlists, err := a.c.GetPlaylists()
	if err != nil {
		fmt.Println("error in GetPlaylists")
		return "", err
	}
	if len(playlists) > 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		index := rand.Intn(len(playlists))
		target := playlists[index]
		fmt.Printf("[%v]\n", target.Name)
		msg = append(msg, fmt.Sprintf("%v.", target.Name))
		err := a.c.AddItem2QueueAndPlay(target.Uri, "")
		if err != nil {
			fmt.Println("error in AddItem2QueueAndPlay")
			return "", err
		}
	} else {
		fmt.Println("playlists is empty")
	}
	return strings.Join(msg, " "), nil
}

func NewChangePlaylistAction(client *Client) ChangePlaylistAction {
	return ChangePlaylistAction{
		name: "Change playlist on Owntone",
		c:    client,
	}
}
