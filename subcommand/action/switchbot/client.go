package switchbot

import (
	"fmt"
	"github.com/nasa9084/go-switchbot/v2"
	"strings"
)

const EnvToken = "SWITCHBOT_TOKEN"
const EnvSecret = "SWITCHBOT_SECRET"

type Config struct {
	token  string
	secret string
}

func NewConfig(token string, secret string) (Config, error) {
	var errs []string
	if len(token) == 0 {
		errs = append(errs, fmt.Sprintf("not found \"%s\". Please set %s via Environment variable", EnvToken, EnvToken))
	}
	if len(secret) == 0 {
		errs = append(errs, fmt.Sprintf("not found \"%s\". Please set %s via Environment variable", EnvSecret, EnvSecret))
	}
	if len(errs) > 0 {
		return Config{}, fmt.Errorf(strings.Join(errs, "\n"))
	}
	return Config{
		token,
		secret,
	}, nil
}

func NewSwitchBotClient(config Config) *switchbot.Client {
	return switchbot.New(config.token, config.secret)
}
