package switchbot

import (
	"context"
	"fmt"
	"strings"
)

type ListScenesAction struct {
	name string
	CachedClient
}

func (a ListScenesAction) Run() (string, error) {
	scenes, err := a.Scene().List(context.Background())
	var msg []string
	if err != nil {
		return "", err
	}
	for _, s := range scenes {
		msg = append(msg, fmt.Sprintf("%s\t%s", s.Name, s.ID))
	}
	return strings.Join(msg, "\n"), nil
}

func NewListScenesAction(client CachedClient) ListScenesAction {
	return ListScenesAction{
		name:         "List scenes on SwitchBot",
		CachedClient: client,
	}
}
