package owntone

type PauseAction struct {
	name string
	c    *Client
}

func (a PauseAction) Run(_ string) (string, error) {
	err := a.c.Pause()
	if err != nil {
		return "", err
	}
	return "Paused the music.", nil
}

func NewPauseAction(client *Client) PauseAction {
	return PauseAction{
		name: "Pause music on Owntone",
		c:    client,
	}
}
