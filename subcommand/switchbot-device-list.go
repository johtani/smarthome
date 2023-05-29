package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

const SwitchBotDeviceListCmd = "Switchbot-device-list"

func NewSwitchBotDeviceListDefinition() Definition {
	return Definition{

		SwitchBotDeviceListCmd,
		"List devices on SwitchBot",
		NewSwitchBotDeviceListSubcommand,
	}
}

func NewSwitchBotDeviceListSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewListDevicesAction(switchbotClient),
		},
		true,
	}
}
