package subcommand

import (
	"smart_home/subcommand/action"
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
