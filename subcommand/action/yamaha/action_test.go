package yamaha

import (
	"context"
	"strings"
	"testing"
)

type mockYamahaAPI struct {
	setSceneFunc  func(ctx context.Context, scene int) error
	setVolumeFunc func(ctx context.Context, volume int) error
	powerOffFunc  func(ctx context.Context) error
}

func (m *mockYamahaAPI) SetScene(ctx context.Context, scene int) error {
	return m.setSceneFunc(ctx, scene)
}
func (m *mockYamahaAPI) SetVolume(ctx context.Context, volume int) error {
	return m.setVolumeFunc(ctx, volume)
}
func (m *mockYamahaAPI) PowerOff(ctx context.Context) error {
	return m.powerOffFunc(ctx)
}

func TestActions(t *testing.T) {
	mock := &mockYamahaAPI{
		setSceneFunc: func(ctx context.Context, scene int) error {
			return nil
		},
		setVolumeFunc: func(ctx context.Context, volume int) error {
			return nil
		},
		powerOffFunc: func(ctx context.Context) error {
			return nil
		},
	}

	t.Run("PowerOffAction", func(t *testing.T) {
		action := NewPowerOffAction(mock)
		got, err := action.Run(context.Background(), "")
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if !strings.Contains(got, "Power off") {
			t.Errorf("Unexpected result: %s", got)
		}
	})

	t.Run("SetSceneAction", func(t *testing.T) {
		action := NewSetSceneAction(mock, 1)
		got, err := action.Run(context.Background(), "")
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if !strings.Contains(got, "scene to No.1") {
			t.Errorf("Unexpected result: %s", got)
		}
	})

	t.Run("SetVolumeAction", func(t *testing.T) {
		action := NewSetVolumeAction(mock)
		got, err := action.Run(context.Background(), "")
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if !strings.Contains(got, "volume to 70") {
			t.Errorf("Unexpected result: %s", got)
		}
	})
}
