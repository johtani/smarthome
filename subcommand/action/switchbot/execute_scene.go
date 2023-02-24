package switchbot

import "github.com/nasa9084/go-switchbot/v2"

type ExecuteSceneAction struct {
	name  string
	scene string
	c     *switchbot.Client
}

func (a ExecuteSceneAction) Run() error {
	return nil
}

func NewExecuteSceneAction(scene string) ExecuteSceneAction {
	return ExecuteSceneAction{
		"Turn off the target device",
		scene,
		NewSwitchBotClient(),
	}
}
