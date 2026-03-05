package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/healthcheck"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

// HealthCmd is the command name for checking the health of the system.
const HealthCmd = "health"

// NewHealthDefinition creates the definition for the health check command.
func NewHealthDefinition() Definition {
	return Definition{
		Name:        HealthCmd,
		Description: "Check the health of the smart home system",
		Factory:     NewHealthSubcommand,
	}
}

// NewHealthSubcommand creates a new Subcommand for the health check command.
func NewHealthSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	owntoneClient := owntone.NewClient(config.Owntone)
	yamahaClient := yamaha.NewClient(config.Yamaha)

	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			healthcheck.NewSwitchBotHealthCheckAction(switchbotClient),
			healthcheck.NewOwnToneHealthCheckAction(owntoneClient),
			healthcheck.NewYamahaHealthCheckAction(yamahaClient),
		},
	}
}
