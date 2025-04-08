package owntone

import "fmt"

type ClearQueueAction struct {
	name string
	c    *Client
}

func (a ClearQueueAction) Run(_ string) (string, error) {
	err := a.c.ClearQueue()
	if err != nil {
		return "", fmt.Errorf("error in ClearQueue(%v)\n %v", a.c.config.Url, err)
	}
	return "Cleared queue", nil
}

func NewClearQueueAction(client *Client) ClearQueueAction {
	return ClearQueueAction{
		name: "Clear queue on Owntone",
		c:    client,
	}
}
