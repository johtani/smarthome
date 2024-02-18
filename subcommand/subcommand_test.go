package subcommand

import (
	"fmt"
	"github.com/johtani/smarthome/subcommand/action"
	"testing"
)

type okAction struct{}

func (a okAction) Run(_ string) (string, error) {
	return "", nil
}

type ngAction struct{}

func (a ngAction) Run(_ string) (string, error) {
	return "", fmt.Errorf("something wrong")
}

func NewDummySubcommand(definition Definition, _ Config) Subcommand {
	return Subcommand{
		Definition: definition,
		actions:    []action.Action{},
	}
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
			if _, err := s.Exec(""); (err != nil) != tt.wantErr {
				t.Errorf("Exec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefinition_Distance(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		shortnames  []string
		WithArgs    bool
		Factory     func(Definition, Config) Subcommand
		Match       func(message string) (bool, string)
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
		// TODO: Add test cases.
		{name: "only entity name",
			fields: fields{Name: "test", Description: "description", Factory: NewDummySubcommand, shortnames: []string{}},
			args:   args{name: "tess"}, distance: 1, cmd: "test"},
		{name: "hit shortname",
			fields: fields{Name: "tesssss", Description: "description", Factory: NewDummySubcommand, shortnames: []string{"hoge", "test"}},
			args:   args{name: "tess"}, distance: 1, cmd: "test"},
		{name: "hit entity name without hyphen",
			fields: fields{Name: "test-cmd", Description: "description", Factory: NewDummySubcommand, shortnames: []string{}},
			args:   args{name: "tess cmd", withoutHyphen: true}, distance: 1, cmd: "test cmd"},
		{name: "hit shortname without hyphen",
			fields: fields{Name: "tesssss-cmd", Description: "description", Factory: NewDummySubcommand, shortnames: []string{"hoge", "test-cmd"}},
			args:   args{name: "tess cmd", withoutHyphen: true}, distance: 1, cmd: "test cmd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Definition{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				shortnames:  tt.fields.shortnames,
				WithArgs:    tt.fields.WithArgs,
				Factory:     tt.fields.Factory,
			}
			got, got1 := d.Distance(tt.args.name, tt.args.withoutHyphen)
			if got != tt.distance {
				t.Errorf("Distance() distance = %v, want %v", got, tt.distance)
			}
			if got1 != tt.cmd {
				t.Errorf("Distance() cmd = %v, want %v", got1, tt.cmd)
			}
		})
	}
}

func TestCommands_Help(t *testing.T) {
	type fields struct {
		definitions []Definition
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "a and b command help",
			fields: fields{
				definitions: []Definition{
					{Name: "a", Description: "description", Factory: NewDummySubcommand},
					{Name: "b", Description: "description", Factory: NewDummySubcommand, shortnames: []string{"c"}},
				},
			},
			want: "  a : description\n  b [c]: description\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commands{
				definitions: tt.fields.definitions,
			}
			if got := c.Help(); got != tt.want {
				t.Errorf("Help() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommands_Find(t *testing.T) {
	type fields struct {
		definitions []Definition
	}
	type args struct {
		name          string
		withoutHyphen bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		def      Definition
		wantArgs string
		dymMsg   string
		wantErr  bool
	}{
		{
			name: "Exact match",
			fields: fields{
				definitions: []Definition{
					{Name: "abc", Description: "description", Factory: NewDummySubcommand, WithArgs: false},
					{Name: "de-f", Description: "description", Factory: NewDummySubcommand, shortnames: []string{"def"}, WithArgs: false},
				},
			},
			args:     args{name: "abc", withoutHyphen: false},
			def:      Definition{Name: "abc", Description: "description", Factory: NewDummySubcommand, WithArgs: false},
			wantArgs: "",
			dymMsg:   "",
			wantErr:  false,
		},
		{
			name: "Exact Match with Args",
			fields: fields{
				definitions: []Definition{
					{Name: "abc", Description: "description", Factory: NewDummySubcommand, WithArgs: true},
					{Name: "de-f", Description: "description", Factory: NewDummySubcommand, shortnames: []string{"def"}, WithArgs: false},
				},
			},
			args:     args{name: "abc d", withoutHyphen: false},
			def:      Definition{Name: "abc", Description: "description", Factory: NewDummySubcommand, WithArgs: true},
			wantArgs: "d",
			dymMsg:   "",
			wantErr:  false,
		},
		{
			name: "Match Did you mean",
			fields: fields{
				definitions: []Definition{
					{Name: "abc", Description: "description", Factory: NewDummySubcommand, WithArgs: false},
					{Name: "de-f", Description: "description", Factory: NewDummySubcommand, shortnames: []string{"def"}, WithArgs: false},
				},
			},
			args:     args{name: "acb", withoutHyphen: false},
			def:      Definition{Name: "abc", Description: "description", Factory: NewDummySubcommand, WithArgs: false},
			wantArgs: "",
			dymMsg:   "Did you mean \"abc\"?",
			wantErr:  false,
		},
		{
			name: "Exact Match with Args",
			fields: fields{
				definitions: []Definition{
					{Name: "abc", Description: "description", Factory: NewDummySubcommand, WithArgs: true},
					{Name: "de-f", Description: "description", Factory: NewDummySubcommand, shortnames: []string{"def"}, WithArgs: false},
				},
			},
			args:    args{name: "abc", withoutHyphen: false},
			wantErr: true,
		},
		// Error with args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commands{
				definitions: tt.fields.definitions,
			}
			got, got1, got2, err := c.Find(tt.args.name, tt.args.withoutHyphen)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.def.Name {
				t.Errorf("Find() def = %v, want %v", got, tt.def)
			}
			if got1 != tt.wantArgs {
				t.Errorf("Find() args = %v, want %v", got1, tt.wantArgs)
			}
			if got2 != tt.dymMsg {
				t.Errorf("Find() dymMsg = %v, want %v", got2, tt.dymMsg)
			}
		})
	}
}
