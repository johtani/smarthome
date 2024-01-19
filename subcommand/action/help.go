package action

type HelpAction struct {
	help string
}

func (a HelpAction) Run(_ string) (string, error) {
	return a.help, nil
}

func NewHelpAction(msg string) HelpAction {
	return HelpAction{help: msg}
}
