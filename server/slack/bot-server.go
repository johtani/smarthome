package slack

import (
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
	appToken string
	botToken string
	debug    bool
}

func loadConfig() (Config, error) {
	var errs []string
	appToken := os.Getenv("SLACK_APP_TOKEN")
	if appToken == "" {
		errs = append(errs, fmt.Sprintf("SLACK_APP_TOKEN must be set.\n"))
	}
	if !strings.HasPrefix(appToken, "xapp-") {
		errs = append(errs, fmt.Sprintf("SLACK_APP_TOKEN must have the prefix \"xapp-\"."))
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		errs = append(errs, fmt.Sprintf("SLACK_BOT_TOKEN must be set.\n"))
	}
	if !strings.HasPrefix(botToken, "xoxb-") {
		errs = append(errs, fmt.Sprintf("SLACK_BOT_TOKEN must have the prefix \"xoxb-\"."))
	}

	debugFlag := os.Getenv("DEBUG")

	if len(errs) > 0 {
		return Config{}, fmt.Errorf(strings.Join(errs, "\n"))
	}

	return Config{
		appToken: appToken,
		botToken: botToken,
		debug:    debugFlag == "true",
	}, nil
}

func Run(config subcommand.Config, smap map[string]subcommand.Definition) error {
	slackConfig, err := loadConfig()
	if err != nil {
		return err
	}
	webApi := slack.New(
		slackConfig.botToken,
		slack.OptionAppLevelToken(slackConfig.appToken),
		slack.OptionDebug(slackConfig.debug),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)
	client := socketmode.New(
		webApi,
		socketmode.OptionDebug(slackConfig.debug),
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
			msg, err = findAndExec(config, smap, strings.ReplaceAll(payloadEvent.Text, botUserIdPrefix, ""))
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
	socketModeHandler.RunEventLoop()
	return nil
}

func findAndExec(config subcommand.Config, smap map[string]subcommand.Definition, text string) (string, error) {
	// TODO message取り出し(もうちょっとスマートにできないか？)
	name := strings.TrimSpace(text)
	var msg string
	d, ok := smap[name]
	if ok {
		c := d.Init(config)
		var err error
		msg, err = c.Exec()
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("Sorry, I cannot understand what you want from what you said '%v'...\n", name)
	}
	// 何を実行したかを返したほうがいい？
	return msg, nil
}
