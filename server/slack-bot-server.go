package server

import (
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
	"smart_home/subcommand"
	"strings"
)

type SlackConfig struct {
	appToken string
	botToken string
	debug    bool
}

func loadSlackConfig() (SlackConfig, error) {
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
		return SlackConfig{}, fmt.Errorf(strings.Join(errs, "\n"))
	}

	return SlackConfig{
		appToken: appToken,
		botToken: botToken,
		debug:    debugFlag == "true",
	}, nil
}

func Run(config subcommand.Config, smap map[string]subcommand.Definition) error {
	slackConfig, err := loadSlackConfig()
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
	_, authTestErr := webApi.AuthTest()
	if authTestErr != nil {
		return fmt.Errorf("SLACK_BOT_TOKEN is invalid: %v\n", authTestErr)
	}

	socketModeHandler := socketmode.NewSocketmodeHandler(client)
	socketModeHandler.HandleEvents(slackevents.AppMention, func(event *socketmode.Event, client *socketmode.Client) {

		eventPayload, ok := event.Data.(slackevents.EventsAPIEvent)
		if !ok {
			client.Debugf("Skipped Envelope: %v", event)
		}

		client.Ack(*event.Request)

		payloadEvent, ok := eventPayload.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			client.Debugf("Payload Event: %v", payloadEvent)
		}
		fmt.Printf("######### : We have been mentioned in %v\n", payloadEvent.Channel)
		msg, err := findAndRun(config, smap, payloadEvent.Text)
		if err != nil {
			fmt.Printf("######### : Got error %v\n", err)
			msg = fmt.Sprintf("%v\nError: %v", msg, err.Error())
		}
		if len(msg) == 0 {
			msg = "Yes, master."
		}

		_, _, err2 := client.PostMessage(payloadEvent.Channel, slack.MsgOptionText(msg, false))
		if err2 != nil {
			fmt.Printf("failed posting message: %v\n", err2)
		}
	})
	socketModeHandler.RunEventLoop()
	return nil
}

func findAndRun(config subcommand.Config, smap map[string]subcommand.Definition, text string) (string, error) {
	// TODO message取り出し(もうちょっとスマートにできないか？)
	msgs := strings.Split(text, " ")
	name := strings.Join(msgs[1:], " ")
	name = strings.TrimSpace(name)

	d, ok := smap[name]
	if ok {
		c := d.Init(config)
		err := c.Exec()
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("command[%v] is not found.\n", name)
	}
	// 何を実行したかを返したほうがいい？
	return "", nil
}
