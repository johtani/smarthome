package subcommand

import (
	"fmt"
	"smart_home/subcommand/action"
	"testing"
)

type okAction struct{}

func (a okAction) Run() error {
	return nil
}

type ngAction struct{}

func (a ngAction) Run() error {
	return fmt.Errorf("something wrong")
}

func TestSubcommand_Exec(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		actions     []action.Action
		checkConfig func() error
		ignoreError bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"empty action", fields{"test", "test", []action.Action{}, func() error { return nil }, true}, false},
		{"ok-ng actions expect error", fields{"test", "test", []action.Action{okAction{}, ngAction{}}, func() error { return nil }, false}, true},
		{"ng-ok actions expect error", fields{"test", "test", []action.Action{ngAction{}, okAction{}}, func() error { return nil }, false}, true},
		{"ok-ng actions skip error", fields{"test", "test", []action.Action{okAction{}, ngAction{}}, func() error { return nil }, true}, false},
		{"ng-ok actions skip error", fields{"test", "test", []action.Action{ngAction{}, okAction{}}, func() error { return nil }, true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Subcommand{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				actions:     tt.fields.actions,
				checkConfig: tt.fields.checkConfig,
				ignoreError: tt.fields.ignoreError,
			}
			if err := s.Exec(); (err != nil) != tt.wantErr {
				t.Errorf("Exec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubcommand_CheckConfig(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		actions     []action.Action
		checkConfig func() error
		ignoreError bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"ok checkconfig", fields{"test", "test", []action.Action{}, func() error { return nil }, true}, false},
		{"ng checkconfig", fields{"test", "test", []action.Action{}, func() error { return fmt.Errorf("checkConfig error") }, true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Subcommand{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				actions:     tt.fields.actions,
				checkConfig: tt.fields.checkConfig,
				ignoreError: tt.fields.ignoreError,
			}
			if err := s.CheckConfig(); (err != nil) != tt.wantErr {
				t.Errorf("CheckConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
