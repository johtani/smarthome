package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

// SearchMusicCmd is the command name for searching music.
const SearchMusicCmd = "search music"

// NewSearchMusicCmdDefinition creates the definition for the search music command.
func NewSearchMusicCmdDefinition() Definition {
	return Definition{
		Name:        SearchMusicCmd,
		Description: "Search Music by keyword",
		Factory:     NewSearchMusicSubcommand,
		Args: []Arg{
			{"keyword", "search phrase or keyword", true, []string{}, ""},
			{"type", "search result type", false, []string{"artist", "album", "track", "genre"}, "type:"},
		},
	}
}

// NewSearchMusicSubcommand creates a new Subcommand for the search music command.
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
