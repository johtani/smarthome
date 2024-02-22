package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

const SwitchBotDeviceListCmd = "switchbot device list"
const DeviceListCmd = "device list"

func NewSwitchBotDeviceListDefinition() Definition {
	return Definition{
		Name:        SwitchBotDeviceListCmd,
		Description: "List devices on SwitchBot",
		Factory:     NewSwitchBotDeviceListSubcommand,
		shortnames:  []string{DeviceListCmd},
	}
}

func NewSwitchBotDeviceListSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewListDevicesAction(switchbotClient),
		},
		ignoreError: true,
	}
}
