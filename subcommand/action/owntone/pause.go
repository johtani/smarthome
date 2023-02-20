package owntone

import (
	"fmt"
	"net/http"
	"time"
)

type PauseAction struct {
	name string
	path string
	c    Client
}

func (a PauseAction) Run() error {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPut, a.c.buildUrl(a.path), nil)
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
	println("owntone pause action succeeded.")
	return nil
}

func NewPauseAction() PauseAction {
	return PauseAction{
		"Pause music on Owntone",
		"api/player/pause",
		NewOwntoneClient(),
	}
}
