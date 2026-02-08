package slack

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/johtani/smarthome/subcommand"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.opentelemetry.io/otel"
)

func defaultHandler(event *socketmode.Event, client *socketmode.Client) {
	slog.Warn("Unexpected event type received", "type", event.Type)
}

func PostMessage(ctx context.Context, client *socketmode.Client, channelID string, options ...slack.MsgOption) (string, string, error) {
	_, span := otel.Tracer("slack").Start(ctx, "PostMessage")
	defer span.End()
	return client.PostMessage(channelID, options...)
}

func newMessageSubcommandHandler(config subcommand.Config, botUserIdPrefix string) socketmode.SocketmodeHandlerFunc {
	return func(event *socketmode.Event, client *socketmode.Client) {
		ctx, span := otel.Tracer("slack").Start(context.Background(), "AppMention")
		defer span.End()

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
			msg, err = findAndExec(ctx, config, strings.ReplaceAll(payloadEvent.Text, botUserIdPrefix, ""))
			if err != nil {
				slog.Error("Got error in findAndExec", "error", err)
				msg = fmt.Sprintf("%v\nError: %v", msg, err.Error())
			}
		} else {
			slog.Debug("Skipped message", "text", payloadEvent.Text)
		}

		if len(msg) == 0 {
			msg = "Yes, master."
		}

		_, _, err := PostMessage(ctx, client, payloadEvent.Channel, slack.MsgOptionText(msg, false))
		if err != nil {
			slog.Error("failed posting message", "error", err)
			return
		}
	}
}

func findAndExec(ctx context.Context, config subcommand.Config, text string) (string, error) {
	name := strings.TrimSpace(text)
	if len(name) == 0 {
		return config.Commands.Help(), nil
	}
	d, args, dymMsg, err := config.Commands.Find(name)
	if err != nil {
		return "", err
	}
	c := d.Init(config)
	msg, err := c.Exec(ctx, args)
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
		ctx, span := otel.Tracer("slack").Start(context.Background(), "SlashCommand")
		defer span.End()

		ev, ok := event.Data.(slack.SlashCommand)
		if !ok {
			client.Debugf("skipped command: %v", event)
		}
		client.Ack(*event.Request)

		cmd := fmt.Sprintf("%v %v", ev.Command, ev.Text)
		if _, _, err := PostMessage(ctx, client, ev.ChannelID, slack.MsgOptionText(cmd, false)); err != nil {
			client.Debugf("failed to post message: %v", err)
			return
		}

		escaped := strings.TrimLeft(ev.Command, "/")
		escaped = strings.ReplaceAll(escaped, "-", " ")
		msg, err := findAndExec(ctx, config, escaped+" "+ev.Text)

		if err != nil {
			slog.Error("Got error in findAndExec for slash command", "error", err)
			msg = fmt.Sprintf("%v\nError: %v", msg, err.Error())
		}

		if len(msg) == 0 {
			msg = "Yes, master."
		}
		_, _, err = PostMessage(ctx, client, ev.ChannelID, slack.MsgOptionText(msg, false))
		if err != nil {
			slog.Error("failed posting message for slash command", "error", err)
			return
		}
	}
}
