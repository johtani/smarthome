package subcommand

import (
	"time"

	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

const SearchAndPlayMusicCmd = "search and play"
const SearchPlayCmd = "search play"

func NewSearchAndPlayMusicCmdDefinition() Definition {
	return Definition{
		Name:        SearchAndPlayMusicCmd,
		Description: "Search Music by keyword And play",
		Factory:     NewSearchAndPlayMusicSubcommand,
		shortnames:  []string{SearchPlayCmd},
		Args: []Arg{
			{"keyword", "search phrase or keyword", true, []string{}},
			{"type", "search result type", false, []string{"artist", "album", "track", "genre"}},
		},
	}
}

func NewSearchAndPlayMusicSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewSearchAndPlayAction(owntoneClient),
			action.NewNoOpAction(3 * time.Second),
			owntone.NewSetVolumeAction(owntoneClient, 33),
			owntone.NewDisplayOutputsAction(owntoneClient, true),
		},
		ignoreError: true,
	}
}
