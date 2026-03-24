package subcommand

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
	switchbotsdk "github.com/nasa9084/go-switchbot/v3"
)

// ActionSpec defines a single action step in a macro.
type ActionSpec struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

// MacroConfig defines a user-configurable macro (sequence of actions).
type MacroConfig struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Shortnames  []string     `json:"shortnames"`
	IgnoreError bool         `json:"ignore_error"`
	Actions     []ActionSpec `json:"actions"`
}

var knownActionTypes = map[string]struct{}{
	"owntone_clear_queue":     {},
	"owntone_play":            {},
	"owntone_pause":           {},
	"owntone_display_outputs": {},
	"yamaha_power_on":         {},
	"yamaha_power_off":        {},
	"yamaha_set_input":        {},
	"yamaha_set_volume":       {},
	"yamaha_set_scene":        {},
	"switchbot_send_command":  {},
	"switchbot_execute_scene": {},
	"wait":                    {},
}

// validateMacros checks all macro action types are known.
func validateMacros(macros []MacroConfig) error {
	var errs []string
	for _, macro := range macros {
		for _, spec := range macro.Actions {
			if _, ok := knownActionTypes[spec.Type]; !ok {
				errs = append(errs, fmt.Sprintf("macro %q: unknown action type %q", macro.Name, spec.Type))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("macro validation failed:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}

func newMacroDefinition(macro MacroConfig) Definition {
	return Definition{
		Name:        macro.Name,
		Description: macro.Description,
		shortnames:  macro.Shortnames,
		Factory: func(def Definition, config Config) Subcommand {
			actions, _ := buildActionsFromSpecs(macro.Actions, config)
			return Subcommand{
				Definition:  def,
				actions:     actions,
				ignoreError: macro.IgnoreError,
			}
		},
	}
}

// resolveParam replaces $-prefixed references with values from config.
func resolveParam(value string, config Config) string {
	switch value {
	case "$LightDeviceID":
		return config.Switchbot.LightDeviceID
	case "$LightSceneID":
		return config.Switchbot.LightSceneID
	case "$AirConditionerID":
		return config.Switchbot.AirConditionerID
	default:
		return value
	}
}

func buildActionsFromSpecs(specs []ActionSpec, config Config) ([]action.Action, error) {
	owntoneClient := owntone.NewClient(config.Owntone)
	yamahaClient := yamaha.NewClient(config.Yamaha)
	switchbotClient := switchbot.NewClient(config.Switchbot)

	var actions []action.Action
	for _, spec := range specs {
		a, err := buildAction(spec, config, owntoneClient, yamahaClient, switchbotClient)
		if err != nil {
			return nil, fmt.Errorf("action type %q: %w", spec.Type, err)
		}
		actions = append(actions, a)
	}
	return actions, nil
}

func buildAction(spec ActionSpec, config Config, owntoneClient *owntone.Client, yamahaClient yamaha.API, switchbotClient *switchbot.CachedClient) (action.Action, error) {
	p := spec.Params
	if p == nil {
		p = map[string]string{}
	}
	resolve := func(key string) string {
		return resolveParam(p[key], config)
	}

	switch spec.Type {
	case "owntone_clear_queue":
		return owntone.NewClearQueueAction(owntoneClient), nil
	case "owntone_play":
		return owntone.NewPlayAction(owntoneClient), nil
	case "owntone_pause":
		return owntone.NewPauseAction(owntoneClient), nil
	case "owntone_display_outputs":
		return owntone.NewDisplayOutputsAction(owntoneClient, p["only_selected"] == "true"), nil
	case "yamaha_power_on":
		return yamaha.NewPowerOnAction(yamahaClient), nil
	case "yamaha_power_off":
		return yamaha.NewPowerOffAction(yamahaClient), nil
	case "yamaha_set_input":
		return yamaha.NewSetInputAction(yamahaClient, p["input"]), nil
	case "yamaha_set_volume":
		vol, err := strconv.Atoi(p["volume"])
		if err != nil {
			return nil, fmt.Errorf("invalid volume %q: %w", p["volume"], err)
		}
		return yamaha.NewSetVolumeAction(yamahaClient, vol), nil
	case "yamaha_set_scene":
		scene, err := strconv.Atoi(p["scene"])
		if err != nil {
			return nil, fmt.Errorf("invalid scene %q: %w", p["scene"], err)
		}
		return yamaha.NewSetSceneAction(yamahaClient, scene), nil
	case "switchbot_send_command":
		deviceID := resolve("device_id")
		cmd, err := parseSwitchBotCommand(p["command"])
		if err != nil {
			return nil, err
		}
		return switchbot.NewSendCommandAction(switchbotClient, deviceID, cmd), nil
	case "switchbot_execute_scene":
		sceneID := resolve("scene_id")
		return switchbot.NewExecuteSceneAction(switchbotClient, sceneID), nil
	case "wait":
		secs, err := strconv.ParseFloat(p["seconds"], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid seconds %q: %w", p["seconds"], err)
		}
		return action.NewNoOpAction(time.Duration(float64(time.Second) * secs)), nil
	default:
		return nil, fmt.Errorf("unknown action type: %q", spec.Type)
	}
}

func parseSwitchBotCommand(cmd string) (switchbotsdk.Command, error) {
	var zero switchbotsdk.Command
	switch cmd {
	case "turn_on":
		return switchbotsdk.TurnOnCommand(), nil
	case "turn_off":
		return switchbotsdk.TurnOffCommand(), nil
	default:
		return zero, fmt.Errorf("unknown switchbot command: %q", cmd)
	}
}
