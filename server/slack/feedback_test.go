package slack

import (
	"testing"

	"github.com/slack-go/slack"
)

func TestFeedbackActionValueRoundTrip(t *testing.T) {
	in := feedbackActionValue{
		RequestID: "req-1",
		Label:     "correct",
		Command:   "light on",
		Args:      "",
	}
	encoded, err := encodeFeedbackActionValue(in)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	got, err := decodeFeedbackActionValue(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if got != in {
		t.Fatalf("decoded value mismatch: got %+v, want %+v", got, in)
	}
}

func TestBuildFeedbackBlocks(t *testing.T) {
	blocks, err := buildFeedbackBlocks("ok", "req-1", "light on", "")
	if err != nil {
		t.Fatalf("buildFeedbackBlocks failed: %v", err)
	}
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}

	actionBlock, ok := blocks[2].(*slack.ActionBlock)
	if !ok {
		t.Fatalf("expected action block at index 2, got %T", blocks[2])
	}
	if len(actionBlock.Elements.ElementSet) != 2 {
		t.Fatalf("expected 2 action elements, got %d", len(actionBlock.Elements.ElementSet))
	}

	btnCorrect, ok := actionBlock.Elements.ElementSet[0].(*slack.ButtonBlockElement)
	if !ok {
		t.Fatalf("expected first element button, got %T", actionBlock.Elements.ElementSet[0])
	}
	if btnCorrect.ActionID != feedbackCorrectActionID {
		t.Fatalf("expected action id %q, got %q", feedbackCorrectActionID, btnCorrect.ActionID)
	}

	payload, err := decodeFeedbackActionValue(btnCorrect.Value)
	if err != nil {
		t.Fatalf("decode button payload failed: %v", err)
	}
	if payload.RequestID != "req-1" || payload.Label != "correct" || payload.Command != "light on" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestExtractCorrection(t *testing.T) {
	state := &slack.ViewState{
		Values: map[string]map[string]slack.BlockAction{
			feedbackCorrectionBlockID: {
				feedbackCorrectionActionID: {
					Value: "search and play 宇多田ヒカル",
				},
			},
		},
	}

	got := extractCorrection(state)
	if got != "search and play 宇多田ヒカル" {
		t.Fatalf("unexpected correction value: %q", got)
	}
}

func TestBuildResponseMessageOptions_FeedbackDisabled(t *testing.T) {
	options := buildResponseMessageOptions(false, "ok", "req-1", "light on", "")
	if len(options) != 1 {
		t.Fatalf("expected one msg option, got %d", len(options))
	}
}
