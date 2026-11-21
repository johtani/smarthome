package mcp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/johtani/smarthome/internal/configstore"
	"github.com/johtani/smarthome/subcommand"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestMacroReloadInSameMCPSession(t *testing.T) {
	var mu sync.Mutex
	var actions []string
	device := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		actions = append(actions, r.URL.RawQuery)
		mu.Unlock()
		_, _ = fmt.Fprint(w, `{"response_code":0}`)
	}))
	defer device.Close()
	t.Setenv("SMARTHOME_YAMAHA_URL", device.URL)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"owntone":{"url":"http://localhost"},"switchbot":{"token":"test","secret":"test"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	load := func() (subcommand.Config, error) { return subcommand.LoadConfigFromDir(dir) }
	cfg, err := load()
	if err != nil {
		t.Fatal(err)
	}
	store := configstore.New(cfg)
	s := NewMCPServer()
	unsubscribe := observeTools(s, store)
	defer unsubscribe()
	c, err := client.NewInProcessClient(s)
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	if err := c.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()
	init := mcp.InitializeRequest{}
	init.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	info, err := c.Initialize(ctx, init)
	if err != nil {
		t.Fatal(err)
	}
	if info.Capabilities.Tools == nil || !info.Capabilities.Tools.ListChanged {
		t.Fatal("server did not advertise tools/list_changed")
	}

	const added = `[{"name":"test macro","description":"original","actions":[{"type":"yamaha_power_on"}]}]`
	const changed = `[{"name":"test macro","description":"updated","actions":[{"type":"yamaha_set_input","params":{"input":"hdmi2"}},{"type":"yamaha_power_off"}]}]`
	for _, tt := range []struct {
		name, macros, description string
		wantActions               []string
		reloadFails               bool
	}{
		{"add", added, "original", []string{"power=on"}, false},
		{"change", changed, "updated", []string{"input=hdmi2", "power=standby"}, false},
		{"invalid", "{", "updated", []string{"input=hdmi2", "power=standby"}, true},
		{"delete", "[]", "", nil, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(filepath.Join(dir, "macros.json"), []byte(tt.macros), 0600); err != nil {
				t.Fatal(err)
			}
			if err := store.Reload(load); (err != nil) != tt.reloadFails {
				t.Fatalf("reload returned %v", err)
			}
			list, err := c.ListTools(ctx, mcp.ListToolsRequest{})
			if err != nil {
				t.Fatal(err)
			}
			index := slices.IndexFunc(list.Tools, func(tool mcp.Tool) bool { return tool.Name == "test_macro" })
			if tt.description == "" {
				if index >= 0 {
					t.Fatal("removed macro still listed")
				}
			} else if index < 0 || list.Tools[index].Description != tt.description {
				t.Fatalf("missing macro with description %q", tt.description)
			}
			req := mcp.CallToolRequest{}
			req.Params.Name = "test_macro"
			result, err := c.CallTool(ctx, req)
			if tt.description == "" {
				if err == nil && !result.IsError {
					t.Fatal("removed macro call succeeded")
				}
			} else if err != nil || result.IsError {
				t.Fatalf("call failed: %v, %v", result, err)
			}
			mu.Lock()
			defer mu.Unlock()
			if !slices.Equal(actions, tt.wantActions) {
				t.Errorf("device calls = %v, want %v", actions, tt.wantActions)
			}
			actions = nil
		})
	}
}
