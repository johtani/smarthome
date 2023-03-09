package owntone

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

type SetVolumeAction struct {
	name   string
	path   string
	volume string
	c      Client
}

func (a SetVolumeAction) Run() error {
	client := http.Client{Timeout: 10 * time.Second}
	u, err := url.Parse(a.c.url)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, a.path)
	q := u.Query()
	q.Set("volume", a.volume)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodPut, u.String(), nil)
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

func NewSetVolumeAction() SetVolumeAction {
	return SetVolumeAction{
		"Set Owntone Volume",
		"api/player/volume",
		"33",
		NewOwntoneClient(),
	}
}
