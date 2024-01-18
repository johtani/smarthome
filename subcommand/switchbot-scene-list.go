package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

const SwitchBotSceneListCmd = "switchbot-scene-list"
const SceneListCmd = "scene-list"

func NewSwitchBotSceneListDefinition() Definition {
	return Definition{
		Name:        SwitchBotSceneListCmd,
		Description: "List scenes on SwitchBot",
		Factory:     NewSwitchBotSceneListSubcommand,
	}
}

func NewSwitchBotSceneListSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewListScenesAction(switchbotClient),
		},
		ignoreError: true,
	}
}
