package owntone

import "fmt"

type ClearQueueAction struct {
	name string
	c    *Client
}

func (a ClearQueueAction) Run(_ string) (string, error) {
	err := a.c.ClearQueue()
	if err != nil {
		fmt.Println("error in ClearQueue")
		return "", err
	}
	return "Cleared queue", nil
}

func NewClearQueueAction(client *Client) ClearQueueAction {
	return ClearQueueAction{
		name: "Clear queue on Owntone",
		c:    client,
	}
}
