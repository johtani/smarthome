package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/switchbot"
)

const SwitchBotDeviceListCmd = "switchbot-device-list"

func NewSwitchBotDeviceListSubcommand() Subcommand {
	return Subcommand{
		SwitchBotDeviceListCmd,
		"List devices on SwitchBot",
		[]action.Action{
			switchbot.NewListDevicesAction(),
		},
		checkConfig,
		true,
	}
}
