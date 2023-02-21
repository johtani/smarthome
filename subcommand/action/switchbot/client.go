package switchbot

import (
	"fmt"
	"github.com/nasa9084/go-switchbot/v2"
	"os"
)

const EnvToken = "SWITCHBOT_TOKEN"
const EnvSecret = "SWITCHBOT_SECRET"

func CheckConfig() error {
	token := os.Getenv(EnvToken)
	secret := os.Getenv(EnvSecret)
	if len(token) == 0 {
		return fmt.Errorf("not found \"%s\". Please set %s via Environment variable", EnvToken, EnvToken)
	}
	if len(secret) == 0 {
		return fmt.Errorf("not found \"%s\". Please set %s via Environment variable", EnvSecret, EnvSecret)
	}
	return nil
}

func NewSwitchBotClient() *switchbot.Client {
	token := os.Getenv(EnvToken)
	secret := os.Getenv(EnvSecret)
	return switchbot.New(token, secret)
}
