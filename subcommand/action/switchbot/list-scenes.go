package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v2"
	"strings"
)

type ListScenesAction struct {
	name string
	*switchbot.Client
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

func NewListScenesAction(client *switchbot.Client) ListScenesAction {
	return ListScenesAction{
		"List scenes on SwitchBot",
		client,
	}
}
