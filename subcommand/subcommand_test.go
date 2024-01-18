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
		{"empty action", fields{Name: "test", Description: "test", actions: []action.Action{}, ignoreError: true}, false},
		{"ok-ng actions expect error", fields{Name: "test", Description: "test", actions: []action.Action{okAction{}, ngAction{}}}, true},
		{"ng-ok actions expect error", fields{Name: "test", Description: "test", actions: []action.Action{ngAction{}, okAction{}}}, true},
		{"ok-ng actions skip error", fields{Name: "test", Description: "test", actions: []action.Action{okAction{}, ngAction{}}, ignoreError: true}, false},
		{"ng-ok actions skip error", fields{Name: "test", Description: "test", actions: []action.Action{ngAction{}, okAction{}}, ignoreError: true}, false},
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

func TestEntry_Distance(t *testing.T) {
	type fields struct {
		Name       string
		definition Definition
		shortnames []string
	}
	type args struct {
		name          string
		withoutHyphen bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		distance int
		cmd      string
	}{
		{name: "only entity name",
			fields: fields{Name: "test", definition: Definition{Name: "tess", Description: "description", Factory: NewHelpSubcommand}, shortnames: []string{}},
			args:   args{name: "tess"}, distance: 1, cmd: "test"},
		{name: "hit shortname",
			fields: fields{Name: "tesssss", definition: Definition{Name: "tesssss", Description: "description", Factory: NewHelpSubcommand}, shortnames: []string{"hoge", "test"}},
			args:   args{name: "tess"}, distance: 1, cmd: "test"},
		{name: "hit entity name without hyphen",
			fields: fields{Name: "test-cmd", definition: Definition{Name: "test-cmd", Description: "description", Factory: NewHelpSubcommand}, shortnames: []string{}},
			args:   args{name: "tess cmd", withoutHyphen: true}, distance: 1, cmd: "test cmd"},
		{name: "hit shortname without hyphen",
			fields: fields{Name: "tesssss-cmd", definition: Definition{Name: "tesssss-cmd", Description: "description", Factory: NewHelpSubcommand}, shortnames: []string{"hoge", "test-cmd"}},
			args:   args{name: "tess cmd", withoutHyphen: true}, distance: 1, cmd: "test cmd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEntry(
				tt.fields.Name,
				tt.fields.definition,
				tt.fields.shortnames,
			)
			distance, cmd := e.Distance(tt.args.name, tt.args.withoutHyphen)
			if distance != tt.distance {
				t.Errorf("Distance() distance = %v, want %v", distance, tt.distance)
			}
			if cmd != tt.cmd {
				t.Errorf("Distance() cmd = %v, want %v", cmd, tt.cmd)
			}
		})
	}
}
