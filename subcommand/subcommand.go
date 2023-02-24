package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/owntone"
	"smart_home/subcommand/action/switchbot"
)

type Subcommand struct {
	Name        string
	Description string
	actions     []action.Action
	checkConfig func() error
}

func (s Subcommand) Exec() error {
	for i := range s.actions {
		err := s.actions[i].Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Subcommand) CheckConfig() error {
	return s.checkConfig()
}

func checkConfig() error {
	// TODO エラーはまとめて返したほうがいいかも
	err := owntone.CheckConfig()
	if err != nil {
		return err
	}
	err = switchbot.CheckConfig()
	if err != nil {
		return err
	}
	return nil
}

func SubcommandMap() map[string]Subcommand {
	return map[string]Subcommand{
		StartMeetingCmd:        NewStartMeetingSubcommand(),
		FinishMeetingCmd:       NewFinishMeetingSubcommand(),
		SwitchBotDeviceListCmd: NewSwitchBotDeviceListSubcommand(),
		SwitchBotSceneListCmd:  NewSwitchBotSceneListSubcommand(),
	}
}
