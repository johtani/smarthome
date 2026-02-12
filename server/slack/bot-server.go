/*
Package slack provides a Slack bot server that handles commands via Socket Mode and Slash Commands.
*/
package slack

import (
	"encoding/json"
	"fmt"
	"github.com/johtani/smarthome/subcommand"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
	"strings"
)

// Config represents the configuration for the Slack bot.
type Config struct {
	AppToken string `json:"app_token"`
	BotToken string `json:"bot_token"`
	Debug    bool   `json:"debug"`
}

// ConfigFileName is the default path to the Slack configuration file.
const ConfigFileName = "./config/slack.json"

func (c Config) validate() error {
	var errs []string
	if c.AppToken == "" {
		errs = append(errs, "app_token must be set.\n")
	}
	if !strings.HasPrefix(c.AppToken, "xapp-") {
		errs = append(errs, "app_token must have the prefix \"xapp-\".")
	}

	if c.BotToken == "" {
		errs = append(errs, "bot_token must be set.\n")
	}
	if !strings.HasPrefix(c.BotToken, "xoxb-") {
		errs = append(errs, "bot_token must have the prefix \"xoxb-\".")
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

func loadConfigFromFile() (Config, error) {
	file, err := os.Open(ConfigFileName)
	if err != nil {
		return Config{}, fmt.Errorf("Slack設定ファイルの読み込みに失敗しました (%s): %w", ConfigFileName, err)
	}
	defer func() {
		_ = file.Close()
	}()

	// JSONデコード
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("Slack設定ファイルのJSON解析に失敗しました: %w", err)
	}

	if err := config.validate(); err != nil {
		return Config{}, fmt.Errorf("Slack設定のバリデーションに失敗しました: %w", err)
	}

	return config, nil
}

// Run starts the Slack bot server.
func Run(config subcommand.Config) error {
	slackConfig, err := loadConfigFromFile()
	if err != nil {
		return err
	}
	webAPI := slack.New(
		slackConfig.BotToken,
		slack.OptionAppLevelToken(slackConfig.AppToken),
		slack.OptionDebug(slackConfig.Debug),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)
	client := socketmode.New(
		webAPI,
		socketmode.OptionDebug(slackConfig.Debug),
		socketmode.OptionLog(log.New(os.Stdout, "sm: ", log.Lshortfile|log.LstdFlags)),
	)
	authTest, authTestErr := webAPI.AuthTest()
	if authTestErr != nil {
		return fmt.Errorf("SLACK_BOT_TOKEN is invalid: %v", authTestErr)
	}
	botUserID := authTest.UserID
	botUserIDPrefix := fmt.Sprintf("<@%v>", botUserID)

	socketModeHandler := socketmode.NewSocketmodeHandler(client)
	socketModeHandler.HandleEvents(slackevents.AppMention, newMessageSubcommandHandler(config, botUserIDPrefix))
	socketModeHandler.Handle(socketmode.EventTypeSlashCommand, newSlashCommandSubcommandHandler(config))
	socketModeHandler.HandleDefault(defaultHandler)
	err = socketModeHandler.RunEventLoop()
	if err != nil {
		return err
	}
	return nil
}
