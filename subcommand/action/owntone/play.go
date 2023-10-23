package owntone

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type PlayAction struct {
	name string
	c    *Client
}

// Run
// キューに曲がある場合は、そのまま再生
// キューに曲がない場合は、ランダムにプレイリストを選択してからキューに登録して再生
func (a PlayAction) Run() (string, error) {
	status, err := a.c.GetPlayerStatus()
	msg := []string{"Playing music"}
	if err != nil {
		return "", err
	}
	if status.ItemID == 0 {
		//プレイヤーのキューに曲が入っていない状態
		fmt.Print("queue is empty, so playing a randomly selected playlist")
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
			msg = append(msg, fmt.Sprintf("from %v.", target.Name))
			err := a.c.AddItem2Queue(target.Uri)
			if err != nil {
				fmt.Println("error in AddItem2Queue")
				return "", err
			}
		} else {
			fmt.Println("playlists is empty")
		}
	}
	err = a.c.Play()
	if err != nil {
		fmt.Println("error in Play")
		return "", err
	}
	return strings.Join(msg, " "), nil
}

func NewPlayAction(client *Client) PlayAction {
	return PlayAction{
		"Play music on Owntone",
		client,
	}
}
