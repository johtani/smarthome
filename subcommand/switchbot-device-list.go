package subcommand

import (
	"smarthome/subcommand/action"
	"smarthome/subcommand/action/switchbot"
)

const SwitchBotDeviceListCmd = "switchbot-device-list"

func NewSwitchBotDeviceListDefinition() Definition {
	return Definition{

		SwitchBotDeviceListCmd,
		"List devices on SwitchBot",
		NewSwitchBotDeviceListSubcommand,
	}
}

func NewSwitchBotDeviceListSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewListDevicesAction(switchbotClient),
		},
		true,
	}
}
