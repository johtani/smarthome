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
// キューに曲が存在する場合は、キューの曲を削除して
func (a ChangePlaylistAction) Run() (string, error) {
	status, err := a.c.GetPlayerStatus()
	msg := []string{"Change playlist to"}
	if err != nil {
		return "", err
	}
	if status.ItemID > 0 {
		err := a.c.ClearQueue()
		if err != nil {
			fmt.Println("error in ClearQueue")
			return "", err
		}
	}
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
		err := a.c.AddItem2Queue(target.Uri)
		if err != nil {
			fmt.Println("error in AddItem2Queue")
			return "", err
		}
	} else {
		fmt.Println("playlists is empty")
	}

	err = a.c.Play()
	if err != nil {
		fmt.Println("error in Play")
		return "", err
	}
	return strings.Join(msg, " "), nil
}

func NewChangePlaylistAction(client *Client) ChangePlaylistAction {
	return ChangePlaylistAction{
		"Change playlist on Owntone",
		client,
	}
}
