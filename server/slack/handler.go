package slack

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/johtani/smarthome/internal/configstore"
	"github.com/johtani/smarthome/internal/resolver"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type executionResult struct {
	Message         string
	ResolvedCommand string
	ResolvedArgs    string
}

func defaultHandler(event *socketmode.Event, _ *socketmode.Client) {
	slog.WarnContext(context.Background(), "Unexpected event type received", "type", event.Type)
}

// PostMessage sends a message to a Slack channel.
func PostMessage(ctx context.Context, client *socketmode.Client, channelID string, options ...slack.MsgOption) (string, string, error) {
	_, span := otel.Tracer("slack").Start(ctx, "PostMessage")
	defer span.End()
	return client.PostMessage(channelID, options...)
}

func newMessageSubcommandHandler(configStore *configstore.Store, botUserIDPrefix string) socketmode.SocketmodeHandlerFunc {
	return func(event *socketmode.Event, client *socketmode.Client) {
		ctx, span := otel.Tracer("slack").Start(context.Background(), "AppMention")
		defer span.End()
		ctx, requestID := resolver.EnsureRequestID(ctx)
		ctx = resolver.WithChannel(ctx, "slack_mention")
		span.SetAttributes(
			attribute.String("resolver.request_id", requestID),
			attribute.String("resolver.channel", "slack_mention"),
		)

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
		var result executionResult

		// とりあえずBotのUserIDが最初にあるメッセージだけ対象とする
		if strings.HasPrefix(payloadEvent.Text, botUserIDPrefix) {
			var err error
			result, err = findAndExec(ctx, configStore, strings.ReplaceAll(payloadEvent.Text, botUserIDPrefix, ""))
			msg = result.Message
			if err != nil {
				slog.ErrorContext(ctx, "Got error in findAndExec", "error", err)
				msg = fmt.Sprintf("%v\nError: %v", msg, err.Error())
			}
		} else {
			slog.DebugContext(ctx, "Skipped message", "text", payloadEvent.Text)
		}

		if len(msg) == 0 {
			msg = "Yes, master."
		}

		requestID, _ = resolver.RequestIDFromContext(ctx)
		options := buildResponseMessageOptions(
			configStore.Get().Resolver.FeedbackEnabled,
			msg,
			requestID,
			result.ResolvedCommand,
			result.ResolvedArgs,
		)
		_, _, err := PostMessage(ctx, client, payloadEvent.Channel, options...)
		if err != nil {
			slog.ErrorContext(ctx, "failed posting message", "error", err)
			return
		}
	}
}

func findAndExec(ctx context.Context, configStore *configstore.Store, text string) (executionResult, error) {
	ctx, span := otel.Tracer("resolver").Start(ctx, "findAndExec")
	defer span.End()
	if requestID, ok := resolver.RequestIDFromContext(ctx); ok {
		span.SetAttributes(attribute.String("resolver.request_id", requestID))
	}
	if channel, ok := resolver.ChannelFromContext(ctx); ok {
		span.SetAttributes(attribute.String("resolver.channel", channel))
	}

	config := configStore.Get()
	name := strings.TrimSpace(text)
	if len(name) == 0 {
		span.SetAttributes(attribute.String("resolver.execution_status", "skipped_empty_input"))
		resolver.RecordExecution(ctx, resolver.ExecutionRecord{
			ExecutionStatus: "skipped_empty_input",
		})
		return executionResult{Message: config.Commands.Help()}, nil
	}
	d, args, dymMsg, err := config.Commands.Find(ctx, config, name)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("resolver.execution_status", "resolve_error"))
		resolver.RecordExecution(ctx, resolver.ExecutionRecord{
			ExecutionStatus: "resolve_error",
		})
		return executionResult{}, err
	}
	c := d.Init(config)
	msg, err := c.Exec(ctx, args)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("resolver.execution_status", "exec_error"),
			attribute.String("resolver.resolved_command", d.Name),
		)
		resolver.RecordExecution(ctx, resolver.ExecutionRecord{
			ExecutionStatus: "exec_error",
			ResolvedCommand: d.Name,
			ResolvedArgs:    args,
		})
		return executionResult{
			ResolvedCommand: d.Name,
			ResolvedArgs:    args,
		}, err
	}
	span.SetAttributes(
		attribute.String("resolver.execution_status", "success"),
		attribute.String("resolver.resolved_command", d.Name),
		attribute.String("resolver.resolved_args", args),
	)
	resolver.RecordExecution(ctx, resolver.ExecutionRecord{
		ExecutionStatus: "success",
		ResolvedCommand: d.Name,
		ResolvedArgs:    args,
	})
	if len(dymMsg) > 0 {
		msg = strings.Join([]string{dymMsg, msg}, "\n")
	}
	return executionResult{
		Message:         msg,
		ResolvedCommand: d.Name,
		ResolvedArgs:    args,
	}, nil
}

func newSlashCommandSubcommandHandler(configStore *configstore.Store) socketmode.SocketmodeHandlerFunc {
	return func(event *socketmode.Event, client *socketmode.Client) {
		ctx, span := otel.Tracer("slack").Start(context.Background(), "SlashCommand")
		defer span.End()
		ctx, requestID := resolver.EnsureRequestID(ctx)
		ctx = resolver.WithChannel(ctx, "slack_slash")
		span.SetAttributes(
			attribute.String("resolver.request_id", requestID),
			attribute.String("resolver.channel", "slack_slash"),
		)

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
		result, err := findAndExec(ctx, configStore, escaped+" "+ev.Text)
		msg := result.Message

		if err != nil {
			slog.ErrorContext(ctx, "Got error in findAndExec for slash command", "error", err)
			msg = fmt.Sprintf("%v\nError: %v", msg, err.Error())
		}

		if len(msg) == 0 {
			msg = "Yes, master."
		}
		requestID, _ = resolver.RequestIDFromContext(ctx)
		options := buildResponseMessageOptions(
			configStore.Get().Resolver.FeedbackEnabled,
			msg,
			requestID,
			result.ResolvedCommand,
			result.ResolvedArgs,
		)
		_, _, err = PostMessage(ctx, client, ev.ChannelID, options...)
		if err != nil {
			slog.ErrorContext(ctx, "failed posting message for slash command", "error", err)
			return
		}
	}
}
