package action

import "context"

type HelpAction struct {
	help string
}

func (a HelpAction) Run(_ context.Context, _ string) (string, error) {
	return a.help, nil
}

func NewHelpAction(msg string) HelpAction {
	return HelpAction{help: msg}
}
