package owntone

import (
	"fmt"
	"math/rand"
	"time"
)

type PlayAction struct {
	name string
	c    *Client
}

func (a PlayAction) Run() error {
	status, err := a.c.GetPlayerStatus()
	if err != nil {
		return err
	}
	if status.ItemID == 0 {
		//プレイヤーのキューに曲が入っていない状態
		fmt.Print("queue is empty, so playing a randomly selected playlist")
		playlists, err := a.c.GetPlaylists()
		if err != nil {
			fmt.Println("error in GetPlaylists")
			return err
		}
		if len(playlists) > 0 {
			rand.Seed(time.Now().UnixNano())
			index := rand.Intn(len(playlists))
			target := playlists[index]
			fmt.Printf("[%v]\n", target.Name)
			err := a.c.AddItem2Queue(target.Uri)
			if err != nil {
				fmt.Println("error in AddItem2Queue")
				return err
			}
		} else {
			fmt.Println("playlists is empty")
		}
	}
	err = a.c.Play()
	if err != nil {
		fmt.Println("error in Play")
		return err
	}
	fmt.Println("owntone play action succeeded.")
	return nil
}

func NewPlayAction(client *Client) PlayAction {
	return PlayAction{
		"Play music on Owntone",
		client,
	}
}
