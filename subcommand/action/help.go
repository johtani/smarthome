package action

type HelpAction struct {
	help string
}

func (a HelpAction) Run() (string, error) {
	return a.help, nil
}

func NewHelpAction(msg string) HelpAction {
	return HelpAction{msg}
}
