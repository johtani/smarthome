package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

// DisplayTemperatureCmd is the command name for displaying temperature and humidity.
const DisplayTemperatureCmd = "display temperature"

// DispTempCmd is a short name for the display temperature command.
const DispTempCmd = "disp temp"

// NewDisplayTemperatureDefinition creates the definition for the display temperature command.
func NewDisplayTemperatureDefinition() Definition {
	return Definition{
		Name:        DisplayTemperatureCmd,
		Description: "Display temperature/humidity",
		Factory:     NewDisplayTemperatureSubcommnad,
		shortnames:  []string{DispTempCmd},
	}
}

// NewDisplayTemperatureSubcommnad creates a new Subcommand for the display temperature command.
func NewDisplayTemperatureSubcommnad(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewGetTemperatureAndHumidityAction(switchbotClient),
		},
		ignoreError: true,
	}
}
