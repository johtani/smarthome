package owntone

type PauseAction struct {
	name string
	c    Client
}

func (a PauseAction) Run() error {
	err := a.c.Pause()
	if err != nil {
		return err
	}
	println("owntone pause action succeeded.")
	return nil
}

func NewPauseAction() PauseAction {
	return PauseAction{
		"Pause music on Owntone",
		NewOwntoneClient(),
	}
}
