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

type Config struct {
	AppToken string `json:"app_token"`
	BotToken string `json:"bot_token"`
	Debug    bool   `json:"debug"`
}

const ConfigFileName = "./config/slack.json"

func (c Config) validate() error {
	var errs []string
	if c.AppToken == "" {
		errs = append(errs, fmt.Sprintf("app_token must be set.\n"))
	}
	if !strings.HasPrefix(c.AppToken, "xapp-") {
		errs = append(errs, fmt.Sprintf("app_token must have the prefix \"xapp-\"."))
	}

	if c.BotToken == "" {
		errs = append(errs, fmt.Sprintf("bot_token must be set.\n"))
	}
	if !strings.HasPrefix(c.BotToken, "xoxb-") {
		errs = append(errs, fmt.Sprintf("bot_token must have the prefix \"xoxb-\"."))
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func loadConfigFromFile() Config {
	file, err := os.Open(ConfigFileName)
	if err != nil {
		panic(fmt.Sprintf("ファイルの読み込みエラー: %v", err))
	}
	// JSONデコード
	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		panic(fmt.Sprintf("JSONデコードエラー: %v", err))
	}
	err = config.validate()
	if err != nil {
		panic(fmt.Sprintf("Validation エラー: %v", err))
	}

	return config
}

func Run(config subcommand.Config) error {
	slackConfig := loadConfigFromFile()
	webApi := slack.New(
		slackConfig.BotToken,
		slack.OptionAppLevelToken(slackConfig.AppToken),
		slack.OptionDebug(slackConfig.Debug),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)
	client := socketmode.New(
		webApi,
		socketmode.OptionDebug(slackConfig.Debug),
		socketmode.OptionLog(log.New(os.Stdout, "sm: ", log.Lshortfile|log.LstdFlags)),
	)
	authTest, authTestErr := webApi.AuthTest()
	if authTestErr != nil {
		return fmt.Errorf("SLACK_BOT_TOKEN is invalid: %v\n", authTestErr)
	}
	botUserId := authTest.UserID
	botUserIdPrefix := fmt.Sprintf("<@%v>", botUserId)

	socketModeHandler := socketmode.NewSocketmodeHandler(client)
	socketModeHandler.HandleEvents(slackevents.AppMention, func(event *socketmode.Event, client *socketmode.Client) {

		eventPayload, ok := event.Data.(slackevents.EventsAPIEvent)
		if !ok {
			client.Debugf("######### : Skipped Envelope: %v", event)
			return
		}

		client.Ack(*event.Request)

		payloadEvent, ok := eventPayload.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			client.Debugf("######### : Payload Event: %v", payloadEvent)
			return
		}
		var msg string

		// とりあえずBotのUserIDが最初にあるメッセージだけ対象とする
		if strings.HasPrefix(payloadEvent.Text, botUserIdPrefix) {
			var err error
			msg, err = findAndExec(config, strings.ReplaceAll(payloadEvent.Text, botUserIdPrefix, ""))
			if err != nil {
				fmt.Printf("######### : Got error %v\n", err)
				msg = fmt.Sprintf("%v\nError: %v", msg, err.Error())
			}
		} else {
			fmt.Printf("######### : Skipped message: %v", payloadEvent.Text)
		}

		if len(msg) == 0 {
			msg = "Yes, master."
		}

		_, _, err := client.PostMessage(payloadEvent.Channel, slack.MsgOptionText(msg, false))
		if err != nil {
			fmt.Printf("######### : failed posting message: %v\n", err)
			return
		}
	})
	err := socketModeHandler.RunEventLoop()
	if err != nil {
		return err
	}
	return nil
}

func findAndExec(config subcommand.Config, text string) (string, error) {
	// TODO message取り出し(もうちょっとスマートにできないか？)
	name := strings.TrimSpace(text)
	var msg string
	dymMsg := ""
	d, err := config.Commands.Find(name, true)
	if err != nil {
		candidates, cmds := config.Commands.DidYouMean(name, true)
		if len(candidates) == 0 {
			return "", fmt.Errorf("Sorry, I cannot understand what you want from what you said '%v'...\n", name)
		} else {
			d = candidates[0]
			dymMsg = fmt.Sprintf("Did you mean \"%v\"?", cmds[0])
		}
	}
	c := d.Init(config)
	msg, err = c.Exec()
	if err != nil {
		return "", err
	}
	if len(dymMsg) > 0 {
		msg = strings.Join([]string{dymMsg, msg}, "\n")
	}
	return msg, nil
}
