package owntone

import (
	"fmt"
	"strconv"
	"strings"
)

type SearchAndPlayAction struct {
	name string
	c    *Client
}

func appendMessage(items Items, label string, msg []string, uris []string, loopFunc func(item SearchItem, msg []string) ([]string, []string)) ([]string, []string) {
	if items.Total > 0 {
		msg = append(msg, fmt.Sprintf("# %s (%d)", label, items.Total))
		for _, item := range items.Items {
			msg, uris = loopFunc(item, msg)
		}
	}
	return msg, uris
}

func (a SearchAndPlayAction) Run(query string) (string, error) {
	msg := []string{"Search Results..."}
	searchQuery := Parse(query)
	result, err := a.c.Search(strings.Join(searchQuery.Terms, " "), searchQuery.TypeArray(), searchQuery.Limit)
	if err != nil {
		fmt.Println("error in SearchAndDisplayAction")
		return "Something wrong...", err
	}
	var uris []string
	msg, uris = appendMessage(result.Artists, "Artists", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v", item.Name))
		uris = append(uris, item.Uri)
		return msg, uris
	})
	msg, uris = appendMessage(result.Albums, "Albums", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v", item.Name, item.Artist))
		uris = append(uris, item.Uri)
		return msg, uris
	})
	msg, uris = appendMessage(result.Tracks, "Tracks", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v ", item.Title, item.Artist))
		uris = append(uris, item.Uri)
		return msg, uris
	})

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
	searchQuery := Parse(query)
	fmt.Println(strings.Join(searchQuery.Terms, " "))
	result, err := a.c.Search(strings.Join(searchQuery.Terms, " "), searchQuery.TypeArray(), searchQuery.Limit)
	if err != nil {
		fmt.Println("error in SearchAndDisplayAction")
		return "Something wrong...", err
	}
	msg, _ = appendMessage(result.Artists, "Artists", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v", item.Name))
		return msg, nil
	})
	msg, _ = appendMessage(result.Albums, "Albums", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v", item.Name, item.Artist))
		return msg, nil
	})
	msg, _ = appendMessage(result.Tracks, "Tracks", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v ", item.Title, item.Artist))
		return msg, nil
	})

	return strings.Join(msg, "\n"), nil
}

func NewSearchAndDisplayAction(client *Client) SearchAndDisplayAction {
	return SearchAndDisplayAction{
		name: "Search music by keyword on Owntone and display results",
		c:    client,
	}

}

type SearchQuery struct {
	Terms  []string
	Types  []SearchType
	Limit  int
	Offset int
}

func (sq SearchQuery) TypeArray() []SearchType {
	if sq.Types == nil {
		return []SearchType{artist, album, track}
	}
	return sq.Types
}

const limitPrefix = "limit:"
const offsetPrefix = "offset:"
const typePrefix = "type:"

func Parse(target string) *SearchQuery {
	split := strings.Fields(target)
	var queries []string
	var types []SearchType
	limit := -1
	offset := -1
	for _, term := range split {
		if strings.HasPrefix(term, limitPrefix) {
			value := term[len(limitPrefix):]
			i, err := strconv.Atoi(value)
			if err == nil {
				limit = i
			} else {
				queries = append(queries, term)
			}
		} else if strings.HasPrefix(term, offsetPrefix) {
			value := term[len(offsetPrefix):]
			i, err := strconv.Atoi(value)
			if err == nil {
				offset = i
			} else {
				queries = append(queries, term)
			}
		} else if strings.HasPrefix(term, typePrefix) {
			value := term[len(typePrefix):]
			st, err := SearchTypeFromString(value)
			if err == nil {
				types = append(types, st)
			} else {
				queries = append(queries, term)
			}
		} else {
			queries = append(queries, term)
		}
	}
	return &SearchQuery{Terms: queries, Limit: limit, Offset: offset, Types: types}
}
