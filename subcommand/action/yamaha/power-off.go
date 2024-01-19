package yamaha

import "fmt"

type PowerOffAction struct {
	name string
	c    *Client
}

func (a PowerOffAction) Run(_ string) (string, error) {
	err := a.c.PowerOff()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Amplifier Power off."), nil
}

func NewPowerOffAction(client *Client) PowerOffAction {
	return PowerOffAction{
		name: "Power off Yamaha Amplifier",
		c:    client,
	}
}
