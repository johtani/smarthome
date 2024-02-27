package owntone

import (
	"fmt"
	"strings"
)

type SearchAndPlayAction struct {
	name string
	c    *Client
}

func (a SearchAndPlayAction) Run(query string) (string, error) {
	//TODO resultTypeを複数選択できるようにする or 引数から割り出す or actionを別にする
	msg := []string{"Search Results..."}
	types := []SearchType{track, album, artist}
	result, err := a.c.Search(query, types)
	if err != nil {
		fmt.Println("error in SearchAndDisplayAction")
		return "Something wrong...", err
	}
	var uris []string
	if result.Artists.Total > 0 {
		msg = append(msg, "# Artists")
		for _, item := range result.Artists.Items {
			msg = append(msg, fmt.Sprintf(" %v", item.Name))
			uris = append(uris, item.Uri)
		}
	}
	if result.Albums.Total > 0 {
		msg = append(msg, "# Albums")
		for _, item := range result.Albums.Items {
			msg = append(msg, fmt.Sprintf(" %v / %v", item.Name, item.Artist))
			uris = append(uris, item.Uri)
		}
	}
	if result.Tracks.Total > 0 {
		msg = append(msg, "# Tracks")
		for _, item := range result.Tracks.Items {
			msg = append(msg, fmt.Sprintf(" %v / %v ", item.Title, item.Artist))
			uris = append(uris, item.Uri)
		}
	}
	if len(uris) > 0 {
		err := a.c.ClearQueue()
		if err != nil {
			fmt.Println("error in ClearQueue")
			return "", err
		}
		err = a.c.AddItem2Queue(strings.Join(uris, ","))
		if err != nil {
			fmt.Println("error calling AddItem2Queue")
			return "", err
		}
		err = a.c.Play()
		if err != nil {
			fmt.Println("error in Play")
			return "", err
		}
	}
	if len(msg) > 1 {
		msg = append(msg, "And play these items")
	} else {
		msg = append(msg, "And no play items...")
	}
	return strings.Join(msg, "\n"), nil
}

func NewSearchAndPlayAction(client *Client) SearchAndPlayAction {
	return SearchAndPlayAction{
		name: "Search and Play music on Owntone by keyword",
		c:    client,
	}
}

type SearchAndDisplayAction struct {
	name string
	c    *Client
}

func (a SearchAndDisplayAction) Run(query string) (string, error) {
	msg := []string{"Search Results..."}
	types := []SearchType{track, album, artist}
	result, err := a.c.Search(query, types)
	if err != nil {
		fmt.Println("error in SearchAndDisplayAction")
		return "Something wrong...", err
	}
	if result.Artists.Total > 0 {
		msg = append(msg, "# Artists")
		for _, item := range result.Artists.Items {
			msg = append(msg, fmt.Sprintf(" %v", item.Name))
		}
	}
	if result.Albums.Total > 0 {
		msg = append(msg, "# Albums")
		for _, item := range result.Albums.Items {
			msg = append(msg, fmt.Sprintf(" %v / %v", item.Name, item.Artist))
		}
	}
	if result.Tracks.Total > 0 {
		msg = append(msg, "# Tracks")
		for _, item := range result.Tracks.Items {
			msg = append(msg, fmt.Sprintf(" %v / %v ", item.Title, item.Artist))
		}
	}

	return strings.Join(msg, "\n"), nil
}

func NewSearchAndDisplayAction(client *Client) SearchAndDisplayAction {
	return SearchAndDisplayAction{
		name: "Search music by keyword on Owntone and display results",
		c:    client,
	}

}
