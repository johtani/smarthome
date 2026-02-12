package yamaha

import (
	"context"
	"strings"
	"testing"
)

type mockYamahaAPI struct {
	setSceneFunc  func(ctx context.Context, scene int) error
	setVolumeFunc func(ctx context.Context, volume int) error
	powerOnFunc   func(ctx context.Context) error
	powerOffFunc  func(ctx context.Context) error
	setInputFunc  func(ctx context.Context, input string) error
}

func (m *mockYamahaAPI) SetScene(ctx context.Context, scene int) error {
	return m.setSceneFunc(ctx, scene)
}
func (m *mockYamahaAPI) SetVolume(ctx context.Context, volume int) error {
	return m.setVolumeFunc(ctx, volume)
}
func (m *mockYamahaAPI) PowerOn(ctx context.Context) error {
	return m.powerOnFunc(ctx)
}
func (m *mockYamahaAPI) PowerOff(ctx context.Context) error {
	return m.powerOffFunc(ctx)
}
func (m *mockYamahaAPI) SetInput(ctx context.Context, input string) error {
	return m.setInputFunc(ctx, input)
}

func TestActions(t *testing.T) {
	mock := &mockYamahaAPI{
		setSceneFunc: func(_ context.Context, _ int) error {
			return nil
		},
		setVolumeFunc: func(_ context.Context, _ int) error {
			return nil
		},
		powerOnFunc: func(_ context.Context) error {
			return nil
		},
		powerOffFunc: func(_ context.Context) error {
			return nil
		},
		setInputFunc: func(_ context.Context, _ string) error {
			return nil
		},
	}

	t.Run("PowerOnAction", func(t *testing.T) {
		action := NewPowerOnAction(mock)
		got, err := action.Run(context.Background(), "")
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if !strings.Contains(got, "Power on") {
			t.Errorf("Unexpected result: %s", got)
		}
	})

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
		action := NewSetVolumeAction(mock, 70)
		got, err := action.Run(context.Background(), "")
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if !strings.Contains(got, "volume to 70") {
			t.Errorf("Unexpected result: %s", got)
		}
	})

	t.Run("SetInputAction", func(t *testing.T) {
		action := NewSetInputAction(mock, "airplay")
		got, err := action.Run(context.Background(), "")
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if !strings.Contains(got, "input to airplay") {
			t.Errorf("Unexpected result: %s", got)
		}
	})
}
