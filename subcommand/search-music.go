package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

const SearchMusicCmd = "search music"

func NewSearchMusicCmdDefinition() Definition {
	return Definition{
		Name:        SearchMusicCmd,
		Description: "Search Music by keyword",
		Factory:     NewSearchMusicSubcommand,
	}
}

func NewSearchMusicSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewSearchAndDisplayAction(owntoneClient),
		},
		ignoreError: true,
	}
}
