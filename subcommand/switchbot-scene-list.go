package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/switchbot"
)

const SwitchBotSceneListCmd = "switchbot-scene-list"

func NewSwitchBotSceneListSubcommand() Subcommand {
	return Subcommand{
		SwitchBotSceneListCmd,
		"List scenes on SwitchBot",
		[]action.Action{
			switchbot.NewListScenesAction(),
		},
		checkConfig,
	}
}
