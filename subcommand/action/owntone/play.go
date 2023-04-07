package owntone

type PlayAction struct {
	name string
	c    Client
}

func (a PlayAction) Run() error {
	err := a.c.Play()
	if err != nil {
		return err
	}
	println("owntone play action succeeded.")
	return nil
}

func NewPlayAction() PlayAction {
	return PlayAction{
		"Play music on Owntone",
		NewOwntoneClient(),
	}
}
