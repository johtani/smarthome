package slack

import (
	"fmt"
	"github.com/johtani/smarthome/subcommand"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"os"
	"strings"
)

func defaultHandler(event *socketmode.Event, client *socketmode.Client) {
	fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", event.Type)
	//client.Debugf("skip event: %v", event.Type)
}

func newMessageSubcommandHandler(config subcommand.Config, botUserIdPrefix string) socketmode.SocketmodeHandlerFunc {
	return func(event *socketmode.Event, client *socketmode.Client) {
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
	}
}

func findAndExec(config subcommand.Config, text string) (string, error) {
	name := strings.TrimSpace(text)
	if len(name) == 0 {
		return config.Commands.Help(), nil
	}
	d, args, dymMsg, err := config.Commands.Find(name)
	if err != nil {
		return "", err
	}
	c := d.Init(config)
	msg, err := c.Exec(args)
	if err != nil {
		return "", err
	}
	if len(dymMsg) > 0 {
		msg = strings.Join([]string{dymMsg, msg}, "\n")
	}
	return msg, nil
}

func newSlashCommandSubcommandHandler(config subcommand.Config) socketmode.SocketmodeHandlerFunc {
	return func(event *socketmode.Event, client *socketmode.Client) {

		ev, ok := event.Data.(slack.SlashCommand)
		if !ok {
			client.Debugf("skipped command: %v", event)
		}
		client.Ack(*event.Request)

		cmd := fmt.Sprintf("%v %v", ev.Command, ev.Text)
		if _, _, err := client.PostMessage(ev.ChannelID, slack.MsgOptionText(cmd, false)); err != nil {
			client.Debugf("failed to post message: %v", err)
			return
		}

		escaped := strings.TrimLeft(ev.Command, "/")
		escaped = strings.ReplaceAll(escaped, "-", " ")
		msg, err := findAndExec(config, escaped+" "+ev.Text)

		if err != nil {
			fmt.Printf("######### : Got error %v\n", err)
			msg = fmt.Sprintf("%v\nError: %v", msg, err.Error())
		}

		if len(msg) == 0 {
			msg = "Yes, master."
		}
		_, _, err = client.PostMessage(ev.ChannelID, slack.MsgOptionText(msg, false))
		if err != nil {
			fmt.Printf("######### : failed posting message: %v\n", err)
			return
		}
	}
}
