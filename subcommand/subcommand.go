package subcommand

import (
	"fmt"
	"os"
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/owntone"
	"smart_home/subcommand/action/switchbot"
	"strings"
)

type Subcommand struct {
	Definition
	actions     []action.Action
	ignoreError bool
}

type Definition struct {
	Name        string
	Description string
	Factory     func(Definition, Config) Subcommand
}

func (s Subcommand) Exec() error {
	for i := range s.actions {
		err := s.actions[i].Run()
		if s.ignoreError && err != nil {
			fmt.Printf("skip error\t %v\n", err)
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (d Definition) Init(config Config) Subcommand {
	return d.Factory(d, config)
}

func Map() map[string]Definition {
	return map[string]Definition{
		StartMeetingCmd:        NewStartMeetingDefinition(),
		FinishMeetingCmd:       NewFinishMeetingDefinition(),
		SwitchBotDeviceListCmd: NewSwitchBotDeviceListDefinition(),
		SwitchBotSceneListCmd:  NewSwitchBotSceneListDefinition(),
		LightOffCmd:            NewLightOffDefinition(),
	}
}

type Config struct {
	owntone   owntone.Config
	switchbot switchbot.Config
}

func NewConfig() (Config, error) {
	var errs []string
	owntoneConfig, err := owntone.NewConfig(os.Getenv(owntone.EnvUrl))
	if err != nil {
		errs = append(errs, err.Error())
	}
	switchbotConfig, err := switchbot.NewConfig(os.Getenv(switchbot.EnvToken), os.Getenv(switchbot.EnvSecret))
	if err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return Config{}, fmt.Errorf(strings.Join(errs, "\n"))
	}
	return Config{owntoneConfig, switchbotConfig}, nil
}
