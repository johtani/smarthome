package subcommand

import (
	"fmt"
	"github.com/johtani/smarthome/subcommand/action"
	"testing"
)

type okAction struct{}

func (a okAction) Run() (string, error) {
	return "", nil
}

type ngAction struct{}

func (a ngAction) Run() (string, error) {
	return "", fmt.Errorf("something wrong")
}

func TestSubcommand_Exec(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		actions     []action.Action
		ignoreError bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"empty action", fields{"test", "test", []action.Action{}, true}, false},
		{"ok-ng actions expect error", fields{"test", "test", []action.Action{okAction{}, ngAction{}}, false}, true},
		{"ng-ok actions expect error", fields{"test", "test", []action.Action{ngAction{}, okAction{}}, false}, true},
		{"ok-ng actions skip error", fields{"test", "test", []action.Action{okAction{}, ngAction{}}, true}, false},
		{"ng-ok actions skip error", fields{"test", "test", []action.Action{ngAction{}, okAction{}}, true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Definition{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Factory:     nil,
			}
			s := Subcommand{
				Definition:  d,
				actions:     tt.fields.actions,
				ignoreError: tt.fields.ignoreError,
			}
			if _, err := s.Exec(); (err != nil) != tt.wantErr {
				t.Errorf("Exec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
