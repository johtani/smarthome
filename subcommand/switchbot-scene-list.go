package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/switchbot"
)

const SwitchBotSceneListCmd = "switchbot-scene-list"

func NewSwitchBotSceneListDefinition() Definition {
	return Definition{
		SwitchBotSceneListCmd,
		"List scenes on SwitchBot",
		NewSwitchBotSceneListSubcommand,
	}
}

func NewSwitchBotSceneListSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewSwitchBotClient(config.switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewListScenesAction(switchbotClient),
		},
		true,
	}
}
