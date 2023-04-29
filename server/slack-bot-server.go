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
	socketModeHandler.HandleEvents(slackevents.AppMention, eventsAppMention)
	socketModeHandler.RunEventLoop()
	return nil
}

func eventsAppMention(event *socketmode.Event, client *socketmode.Client) {
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

	_, _, err := client.PostMessage(payloadEvent.Channel, slack.MsgOptionText("Yes, hello master.", false))
	if err != nil {
		fmt.Printf("failed posting message: %v\n", err)
	}

}
