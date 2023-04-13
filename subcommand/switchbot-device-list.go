package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/switchbot"
)

const SwitchBotDeviceListCmd = "switchbot-device-list"

func NewSwitchBotDeviceListSubcommand(config Config) Subcommand {
	switchbotClient := switchbot.NewSwitchBotClient(config.switchbot)
	return Subcommand{
		SwitchBotDeviceListCmd,
		"List devices on SwitchBot",
		[]action.Action{
			switchbot.NewListDevicesAction(switchbotClient),
		},
		true,
	}
}
