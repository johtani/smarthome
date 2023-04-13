package owntone

import (
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
		println("queue is empty, so playing a randomly selected playlist")
		playlists, err := a.c.GetPlaylists()
		if err != nil {
			println("error in GetPlaylists")
			return err
		}
		if len(playlists) > 0 {
			rand.Seed(time.Now().UnixNano())
			index := rand.Intn(len(playlists))
			target := playlists[index]
			err := a.c.AddItem2Queue(target)
			if err != nil {
				println("error in AddItem2Queue")
				return err
			}
		} else {
			println("playlists is empty")
		}
	}
	err = a.c.Play()
	if err != nil {
		println("error in Play")
		return err
	}
	println("owntone play action succeeded.")
	return nil
}

func NewPlayAction(client *Client) PlayAction {
	return PlayAction{
		"Play music on Owntone",
		client,
	}
}
