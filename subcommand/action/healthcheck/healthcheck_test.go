package healthcheck

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockYamahaAPI struct {
	getDeviceInfoFunc func(ctx context.Context) error
}

func (m *mockYamahaAPI) SetScene(ctx context.Context, scene int) error    { return nil }
func (m *mockYamahaAPI) SetVolume(ctx context.Context, volume int) error  { return nil }
func (m *mockYamahaAPI) PowerOn(ctx context.Context) error                { return nil }
func (m *mockYamahaAPI) PowerOff(ctx context.Context) error               { return nil }
func (m *mockYamahaAPI) SetInput(ctx context.Context, input string) error { return nil }
func (m *mockYamahaAPI) GetDeviceInfo(ctx context.Context) error {
	return m.getDeviceInfoFunc(ctx)
}

func TestYamahaHealthCheckAction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mockYamahaAPI{
			getDeviceInfoFunc: func(ctx context.Context) error {
				return nil
			},
		}
		action := NewYamahaHealthCheckAction(mock)
		got, err := action.Run(context.Background(), "")
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if got != "Yamaha: OK" {
			t.Errorf("got %q, want %q", got, "Yamaha: OK")
		}
	})

	t.Run("Failure", func(t *testing.T) {
		mock := &mockYamahaAPI{
			getDeviceInfoFunc: func(ctx context.Context) error {
				return errors.New("connection error")
			},
		}
		action := NewYamahaHealthCheckAction(mock)
		got, err := action.Run(context.Background(), "")
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if !strings.Contains(got, "Yamaha: Error") {
			t.Errorf("expected error message, got %q", got)
		}
	})
}
