package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v2"
)

type ListScenesAction struct {
	name string
	c    *switchbot.Client
}

func (a ListScenesAction) Run() error {
	scenes, err := a.c.Scene().List(context.Background())
	if err != nil {
		return err
	}
	for _, s := range scenes {
		fmt.Printf("%s\t%s\n", s.Name, s.ID)
	}
	return nil
}

func NewListScenesAction() ListScenesAction {
	return ListScenesAction{
		"List scenes on SwitchBot",
		NewSwitchBotClient(),
	}
}
