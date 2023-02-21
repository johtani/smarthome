package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/switchbot"
)

const SwitchBotDeviceListCmd = "switchbot-list"

func NewSwitchBotListSubcommand() Subcommand {
	return Subcommand{
		SwitchBotDeviceListCmd,
		"List devices on SwithcBot",
		[]action.Action{
			switchbot.NewDeviceListAction(),
		},
		checkConfig,
	}
}
