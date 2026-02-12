package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

// SwitchBotDeviceListCmd is the command name for listing SwitchBot devices.
const SwitchBotDeviceListCmd = "switchbot device list"

// DeviceListCmd is a short name for the SwitchBot device list command.
const DeviceListCmd = "device list"

// NewSwitchBotDeviceListDefinition creates the definition for the SwitchBot device list command.
func NewSwitchBotDeviceListDefinition() Definition {
	return Definition{
		Name:        SwitchBotDeviceListCmd,
		Description: "List devices on SwitchBot",
		Factory:     NewSwitchBotDeviceListSubcommand,
		shortnames:  []string{DeviceListCmd},
	}
}

// NewSwitchBotDeviceListSubcommand creates a new Subcommand for the SwitchBot device list command.
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
