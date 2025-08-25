package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/kagome"
)

const TokenizeIpaCmd = "tokenize ipa"

func NewTokenizeIpaDefinition() Definition {
	return Definition{
		Name:        TokenizeIpaCmd,
		Description: "Tokenize text by Kagome with IPA dic",
		Factory:     NewTokenizeIpaCommand,
		Args: []Arg{
			{"text", "text / sentence for tokenizer input", true, []string{}},
			{"mode", "search mode", false, []string{"-search", "-extended"}},
		},
	}
}

func NewTokenizeIpaCommand(definition Definition, _ Config) Subcommand {
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			kagome.NewKagomeAction(kagome.IPA),
		},
		ignoreError: false,
	}
}

const TokenizeUniCmd = "tokenize uni"

func NewTokenizeUniDefinition() Definition {
	return Definition{
		Name:        TokenizeUniCmd,
		Description: "Tokenize text by Kagome with Uni dic",
		Factory:     NewTokenizeUniCommand,
		Args: []Arg{
			{"text", "text / sentence for tokenizer input", true, []string{}},
		},
	}
}

func NewTokenizeUniCommand(definition Definition, _ Config) Subcommand {
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			kagome.NewKagomeAction(kagome.UNI),
		},
		ignoreError: false,
	}
}

const TokenizeNeologdCmd = "tokenize neologd"

func NewTokenizeNeologdDefinition() Definition {
	return Definition{
		Name:        TokenizeNeologdCmd,
		Description: "Tokenize text by Kagome with Neologd",
		Factory:     NewTokenizeNeologdCommand,
		Args: []Arg{
			{"text", "text / sentence for tokenizer input", true, []string{}},
		},
	}
}

func NewTokenizeNeologdCommand(definition Definition, _ Config) Subcommand {
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			kagome.NewKagomeAction(kagome.NEOLOGD),
		},
		ignoreError: false,
	}
}
