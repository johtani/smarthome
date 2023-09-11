package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

const DisplayTemperatureCmd = "display-temperature"
const DispTempCmd = "disp-temp"

func NewDisplayTemperatureDefinition() Definition {
	return Definition{
		DisplayTemperatureCmd,
		"Display temperature/humidity",
		NewDisplayTemperatureSubcommnad,
	}
}

func NewDisplayTemperatureSubcommnad(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewGetTemperatureAndHumidityAction(switchbotClient),
		},
		true,
	}
}
