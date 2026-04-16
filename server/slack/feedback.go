package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/johtani/smarthome/internal/resolver"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const (
	feedbackCorrectActionID   = "resolver_feedback_correct"
	feedbackIncorrectActionID = "resolver_feedback_incorrect"

	feedbackCorrectionModalID  = "resolver_feedback_modal"
	feedbackCorrectionBlockID  = "resolver_feedback_correction_block"
	feedbackCorrectionActionID = "resolver_feedback_correction_action"
)

type feedbackActionValue struct {
	RequestID string `json:"request_id"`
	Label     string `json:"label"`
	Command   string `json:"command"`
	Args      string `json:"args"`
}

func buildResponseMessageOptions(feedbackEnabled bool, message, requestID, command, args string) []slack.MsgOption {
	if !feedbackEnabled || strings.TrimSpace(command) == "" {
		return []slack.MsgOption{slack.MsgOptionText(message, false)}
	}
	blocks, err := buildFeedbackBlocks(message, requestID, command, args)
	if err != nil {
		slog.Warn("failed to build feedback blocks, fallback to plain text", "error", err)
		return []slack.MsgOption{slack.MsgOptionText(message, false)}
	}
	return []slack.MsgOption{
		slack.MsgOptionText(message, false),
		slack.MsgOptionBlocks(blocks...),
	}
}

func buildFeedbackBlocks(message, requestID, command, args string) ([]slack.Block, error) {
	msgText := slack.NewTextBlockObject(slack.MarkdownType, message, false, false)
	messageBlock := slack.NewSectionBlock(msgText, nil, nil)

	resolved := strings.TrimSpace(strings.Join([]string{command, args}, " "))
	metaText := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("解釈: `%s`", resolved), false, false)
	metaBlock := slack.NewContextBlock("resolver_feedback_context", metaText)

	correctValue, err := encodeFeedbackActionValue(feedbackActionValue{
		RequestID: requestID,
		Label:     "correct",
		Command:   command,
		Args:      args,
	})
	if err != nil {
		return nil, err
	}
	incorrectValue, err := encodeFeedbackActionValue(feedbackActionValue{
		RequestID: requestID,
		Label:     "incorrect",
		Command:   command,
		Args:      args,
	})
	if err != nil {
		return nil, err
	}

	correctButton := slack.NewButtonBlockElement(
		feedbackCorrectActionID,
		correctValue,
		slack.NewTextBlockObject(slack.PlainTextType, "✅適切", false, false),
	)
	incorrectButton := slack.NewButtonBlockElement(
		feedbackIncorrectActionID,
		incorrectValue,
		slack.NewTextBlockObject(slack.PlainTextType, "❌不適切", false, false),
	)
	actions := slack.NewActionBlock("resolver_feedback_actions", correctButton, incorrectButton)

	return []slack.Block{messageBlock, metaBlock, actions}, nil
}

func encodeFeedbackActionValue(v feedbackActionValue) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal feedback action value: %w", err)
	}
	return string(b), nil
}

func decodeFeedbackActionValue(value string) (feedbackActionValue, error) {
	var v feedbackActionValue
	if err := json.Unmarshal([]byte(value), &v); err != nil {
		return feedbackActionValue{}, fmt.Errorf("failed to decode feedback action value: %w", err)
	}
	return v, nil
}

func newFeedbackBlockActionHandler() socketmode.SocketmodeHandlerFunc {
	return func(event *socketmode.Event, client *socketmode.Client) {
		callback, ok := event.Data.(slack.InteractionCallback)
		if !ok {
			client.Debugf("ignored non-interaction event: %+v", event.Data)
			return
		}
		client.Ack(*event.Request)

		if len(callback.ActionCallback.BlockActions) == 0 {
			return
		}
		action := callback.ActionCallback.BlockActions[0]
		payload, err := decodeFeedbackActionValue(action.Value)
		if err != nil {
			slog.Error("failed to decode feedback action payload", "error", err)
			return
		}

		ctx := resolver.WithRequestID(context.Background(), payload.RequestID)
		ctx = resolver.WithChannel(ctx, "slack_feedback")
		ctx, span := otel.Tracer("slack").Start(ctx, "FeedbackBlockAction")
		defer span.End()
		span.SetAttributes(
			attribute.String("feedback.action_id", action.ActionID),
			attribute.String("feedback.label", payload.Label),
			attribute.String("resolver.request_id", payload.RequestID),
			attribute.String("resolver.resolved_command", payload.Command),
			attribute.String("resolver.resolved_args", payload.Args),
		)

		if action.ActionID == feedbackCorrectActionID {
			recordFeedback(ctx, payload, "")
			if _, err = client.PostEphemeral(
				callback.Channel.ID,
				callback.User.ID,
				slack.MsgOptionText("フィードバックを記録しました。", false),
			); err != nil {
				slog.WarnContext(ctx, "failed to post feedback acknowledgement", "error", err)
			}
			return
		}

		if action.ActionID != feedbackIncorrectActionID {
			return
		}

		modal := buildCorrectionModal(payload)
		if _, err = client.OpenView(callback.TriggerID, modal); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "failed to open correction modal", "error", err)
		}
	}
}

func buildCorrectionModal(payload feedbackActionValue) slack.ModalViewRequest {
	input := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, "例: search and play 宇多田ヒカル", false, false),
		feedbackCorrectionActionID,
	).WithMultiline(false).WithMaxLength(200)

	block := slack.NewInputBlock(
		feedbackCorrectionBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, "本来実行したかったコマンド", false, false),
		nil,
		input,
	)
	block.Optional = true

	privateMetadata, _ := encodeFeedbackActionValue(payload)
	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		CallbackID:      feedbackCorrectionModalID,
		PrivateMetadata: privateMetadata,
		Title:           slack.NewTextBlockObject(slack.PlainTextType, "フィードバック", false, false),
		Submit:          slack.NewTextBlockObject(slack.PlainTextType, "送信", false, false),
		Close:           slack.NewTextBlockObject(slack.PlainTextType, "キャンセル", false, false),
		Blocks:          slack.Blocks{BlockSet: []slack.Block{block}},
	}
}

func newFeedbackViewSubmissionHandler() socketmode.SocketmodeHandlerFunc {
	return func(event *socketmode.Event, client *socketmode.Client) {
		callback, ok := event.Data.(slack.InteractionCallback)
		if !ok {
			return
		}
		client.Ack(*event.Request)

		if callback.Type != slack.InteractionTypeViewSubmission || callback.View.CallbackID != feedbackCorrectionModalID {
			return
		}

		payload, err := decodeFeedbackActionValue(callback.View.PrivateMetadata)
		if err != nil {
			slog.Error("failed to decode feedback private metadata", "error", err)
			return
		}
		correction := extractCorrection(callback.View.State)

		ctx := resolver.WithRequestID(context.Background(), payload.RequestID)
		ctx = resolver.WithChannel(ctx, "slack_feedback")
		ctx, span := otel.Tracer("slack").Start(ctx, "FeedbackViewSubmission")
		defer span.End()
		span.SetAttributes(
			attribute.String("feedback.label", "incorrect"),
			attribute.String("feedback.correction", correction),
			attribute.String("resolver.request_id", payload.RequestID),
			attribute.String("resolver.resolved_command", payload.Command),
			attribute.String("resolver.resolved_args", payload.Args),
		)

		recordFeedback(ctx, feedbackActionValue{
			RequestID: payload.RequestID,
			Label:     "incorrect",
			Command:   payload.Command,
			Args:      payload.Args,
		}, correction)
	}
}

func extractCorrection(state *slack.ViewState) string {
	if state == nil {
		return ""
	}
	blockMap, ok := state.Values[feedbackCorrectionBlockID]
	if !ok {
		return ""
	}
	action, ok := blockMap[feedbackCorrectionActionID]
	if !ok {
		return ""
	}
	return strings.TrimSpace(action.Value)
}

func recordFeedback(ctx context.Context, payload feedbackActionValue, correction string) {
	ctx, span := otel.Tracer("resolver").Start(ctx, "ResolverFeedback.Record")
	defer span.End()
	span.SetAttributes(
		attribute.String("resolver.request_id", payload.RequestID),
		attribute.String("feedback.label", payload.Label),
		attribute.String("feedback.correction", correction),
		attribute.String("resolver.resolved_command", payload.Command),
		attribute.String("resolver.resolved_args", payload.Args),
	)
	resolver.RecordFeedback(ctx, resolver.FeedbackRecord{
		FeedbackLabel:      payload.Label,
		FeedbackCorrection: correction,
		ResolvedCommand:    payload.Command,
		ResolvedArgs:       payload.Args,
	})
}
