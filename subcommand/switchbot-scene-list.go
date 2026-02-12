package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

// SwitchBotSceneListCmd is the command name for listing SwitchBot scenes.
const SwitchBotSceneListCmd = "switchbot scene list"

// SceneListCmd is a short name for the SwitchBot scene list command.
const SceneListCmd = "scene list"

// NewSwitchBotSceneListDefinition creates the definition for the SwitchBot scene list command.
func NewSwitchBotSceneListDefinition() Definition {
	return Definition{
		Name:        SwitchBotSceneListCmd,
		Description: "List scenes on SwitchBot",
		Factory:     NewSwitchBotSceneListSubcommand,
		shortnames:  []string{SceneListCmd},
	}
}

// NewSwitchBotSceneListSubcommand creates a new Subcommand for the SwitchBot scene list command.
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
