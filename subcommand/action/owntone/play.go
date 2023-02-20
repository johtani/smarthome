package owntone

import (
	"fmt"
	"net/http"
	"time"
)

type PlayAction struct {
	name            string
	playPath        string
	defaultPlaylist string
	c               Client
}

func (a PlayAction) Run() error {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPut, a.c.buildUrl(a.playPath), nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	println("owntone play action succeeded.")
	return nil
}

func NewPlayAction() PlayAction {
	return PlayAction{
		"Play music on Owntone",
		"api/player/play",
		"",
		NewOwntoneClient(),
	}
}
