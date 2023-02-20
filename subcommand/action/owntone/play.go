package owntone

type PlayAction struct {
	name string
	c    Client
}

func (a PlayAction) Run() error {

	return nil
}

func NewPlayAction() PlayAction {
	return PlayAction{
		"Play music on Owntone",
		NewOwntoneClient(),
	}
}
