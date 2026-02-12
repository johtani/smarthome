package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

// StartMeetingCmd is the command name for starting a meeting.
const StartMeetingCmd = "start meeting"

// NewStartMeetingDefinition creates the definition for the start meeting command.
func NewStartMeetingDefinition() Definition {
	return Definition{
		Name:        StartMeetingCmd,
		Description: "Actions before starting meeting",
		Factory:     NewStartMeetingSubcommand,
	}
}

// NewStartMeetingSubcommand creates a new Subcommand for the start meeting command.
func NewStartMeetingSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	switchbotClient := switchbot.NewClient(config.Switchbot)
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewPauseAction(owntoneClient),
			switchbot.NewExecuteSceneAction(switchbotClient, config.Switchbot.LightSceneID),
			yamaha.NewPowerOffAction(yamahaClient),
		},
		ignoreError: true,
	}
}
