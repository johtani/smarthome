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
	Name        string
	Description string
	actions     []action.Action
	ignoreError bool
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

func Map(config Config) map[string]Subcommand {
	return map[string]Subcommand{
		StartMeetingCmd:        NewStartMeetingSubcommand(config),
		FinishMeetingCmd:       NewFinishMeetingSubcommand(config),
		SwitchBotDeviceListCmd: NewSwitchBotDeviceListSubcommand(config),
		SwitchBotSceneListCmd:  NewSwitchBotSceneListSubcommand(config),
		LightOffCmd:            NewLightOffSubcommand(config),
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
