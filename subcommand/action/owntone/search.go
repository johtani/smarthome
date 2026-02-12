package owntone

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
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

func (a SearchAndPlayAction) Run(ctx context.Context, query string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "SearchAndPlayAction.Run")
	defer span.End()
	msg := []string{"Search Results..."}
	searchQuery := Parse(query)
	result, err := a.c.Search(ctx, strings.Join(searchQuery.Terms, " "), searchQuery.TypeArray(), searchQuery.Limit)
	if err != nil {
		return "Something wrong...", fmt.Errorf("error in SearchAndDisplayAction\n %v", err)
	}
	var uris []string
	msg, uris = appendMessage(result.Artists, "Artists", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v", item.Name))
		uris = append(uris, item.URI)
		return msg, uris
	})
	msg, uris = appendMessage(result.Albums, "Albums", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v", item.Name, item.Artist))
		uris = append(uris, item.URI)
		return msg, uris
	})
	msg, uris = appendMessage(result.Tracks, "Tracks", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v ", item.Title, item.Artist))
		uris = append(uris, item.URI)
		return msg, uris
	})
	msg, uris = appendMessage(result.Genres, "Genres", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v ", item.Name))
		return msg, uris
	})

	if len(uris) > 0 || len(result.Genres.Items) > 0 {
		err := a.c.ClearQueue(ctx)
		if err != nil {
			return "", fmt.Errorf("error in ClearQueue\n %v", err)
		}
	}

	if len(uris) > 0 {
		err = a.c.AddItem2QueueAndPlay(ctx, strings.Join(uris, ","), "")
		if err != nil {
			return "", fmt.Errorf("error calling AddItem2QueueAndPlay\n %v", err)
		}
	}

	if len(result.Genres.Items) > 0 {
		err = a.c.AddItem2QueueAndPlay(ctx, "", fmt.Sprintf("genre is \"%s\"", strings.Join(searchQuery.Terms, " ")))
		if err != nil {
			return "", fmt.Errorf("error calling AddItem2QueueAndPlay with expression\n %v", err)
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

func (a SearchAndDisplayAction) Run(ctx context.Context, query string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "SearchAndDisplayAction.Run")
	defer span.End()
	msg := []string{"Search Results..."}
	// fmt.Println("original query... " + query)
	searchQuery := Parse(query)
	// fmt.Println("Terms... " + strings.Join(searchQuery.Terms, " "))
	result, err := a.c.Search(ctx, strings.Join(searchQuery.Terms, " "), searchQuery.TypeArray(), searchQuery.Limit)
	if err != nil {
		return "Something wrong...", fmt.Errorf("error in SearchAndDisplayAction(terms=%v)\n %v", strings.Join(searchQuery.Terms, " "), err)
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
	msg, _ = appendMessage(result.Genres, "Genres", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v ", item.Name))
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
		return []SearchType{artist, album, track, genre}
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
		switch {
		case strings.HasPrefix(term, limitPrefix):
			value := term[len(limitPrefix):]
			i, err := strconv.Atoi(value)
			if err == nil {
				limit = i
			} else {
				queries = append(queries, term)
			}
		case strings.HasPrefix(term, offsetPrefix):
			value := term[len(offsetPrefix):]
			i, err := strconv.Atoi(value)
			if err == nil {
				offset = i
			} else {
				queries = append(queries, term)
			}
		case strings.HasPrefix(term, typePrefix):
			value := term[len(typePrefix):]
			st, err := SearchTypeFromString(value)
			if err == nil {
				types = append(types, st)
			} else {
				queries = append(queries, term)
			}
		default:
			queries = append(queries, term)
		}
	}
	return &SearchQuery{Terms: queries, Limit: limit, Offset: offset, Types: types}
}
