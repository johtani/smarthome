package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/kagome"
)

// TokenizeIpaCmd is the command name for tokenizing with IPA dictionary.
const TokenizeIpaCmd = "tokenize ipa"

// NewTokenizeIpaDefinition creates the definition for the tokenize ipa command.
func NewTokenizeIpaDefinition() Definition {
	return Definition{
		Name:        TokenizeIpaCmd,
		Description: "Tokenize text by Kagome with IPA dic",
		Factory:     NewTokenizeIpaCommand,
		Args: []Arg{
			{"text", "text / sentence for tokenizer input", true, []string{}, ""},
			{"mode", "search mode", false, []string{"-search", "-extended"}, ""},
		},
	}
}

// NewTokenizeIpaCommand creates a new Subcommand for the tokenize ipa command.
func NewTokenizeIpaCommand(definition Definition, _ Config) Subcommand {
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			kagome.NewKagomeAction(kagome.IPA),
		},
		ignoreError: false,
	}
}

// TokenizeUniCmd is the command name for tokenizing with Uni dictionary.
const (
	// #nosec G101
	TokenizeUniCmd = "tokenize uni"
)

// NewTokenizeUniDefinition creates the definition for the tokenize uni command.
func NewTokenizeUniDefinition() Definition {
	return Definition{
		Name:        TokenizeUniCmd,
		Description: "Tokenize text by Kagome with Uni dic",
		Factory:     NewTokenizeUniCommand,
		Args: []Arg{
			{"text", "text / sentence for tokenizer input", true, []string{}, ""},
		},
	}
}

// NewTokenizeUniCommand creates a new Subcommand for the tokenize uni command.
func NewTokenizeUniCommand(definition Definition, _ Config) Subcommand {
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			kagome.NewKagomeAction(kagome.UNI),
		},
		ignoreError: false,
	}
}

// TokenizeNeologdCmd is the command name for tokenizing with Neologd dictionary.
const (
	// #nosec G101
	TokenizeNeologdCmd = "tokenize neologd"
)

// NewTokenizeNeologdDefinition creates the definition for the tokenize neologd command.
func NewTokenizeNeologdDefinition() Definition {
	return Definition{
		Name:        TokenizeNeologdCmd,
		Description: "Tokenize text by Kagome with Neologd",
		Factory:     NewTokenizeNeologdCommand,
		Args: []Arg{
			{"text", "text / sentence for tokenizer input", true, []string{}, ""},
		},
	}
}

// NewTokenizeNeologdCommand creates a new Subcommand for the tokenize neologd command.
func NewTokenizeNeologdCommand(definition Definition, _ Config) Subcommand {
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			kagome.NewKagomeAction(kagome.NEOLOGD),
		},
		ignoreError: false,
	}
}
