package subcommand

import (
	"time"

	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

// PlayRandomPlaylistCmd is the command name for playing random playlist music.
const PlayRandomPlaylistCmd = "play random playlist"

// PlayRandomArtistCmd is the command name for playing random artist music.
const PlayRandomArtistCmd = "play random artist"

// PlayRandomGenreCmd is the command name for playing random genre music.
const PlayRandomGenreCmd = "play random genre"

// NewPlayRandomPlaylistCmdDefinition creates the definition for the random playlist command.
func NewPlayRandomPlaylistCmdDefinition() Definition {
	return Definition{
		Name:        PlayRandomPlaylistCmd,
		Description: "Play music from a random playlist",
		Factory:     NewPlayRandomPlaylistSubcommand,
	}
}

// NewPlayRandomArtistCmdDefinition creates the definition for the random artist command.
func NewPlayRandomArtistCmdDefinition() Definition {
	return Definition{
		Name:        PlayRandomArtistCmd,
		Description: "Play music from a random artist",
		Factory:     NewPlayRandomArtistSubcommand,
	}
}

// NewPlayRandomGenreCmdDefinition creates the definition for the random genre command.
func NewPlayRandomGenreCmdDefinition() Definition {
	return Definition{
		Name:        PlayRandomGenreCmd,
		Description: "Play music from a random genre",
		Factory:     NewPlayRandomGenreSubcommand,
	}
}

func newPlayRandomMusicSubcommand(definition Definition, config Config, mode string) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewClearQueueAction(owntoneClient),
			yamaha.NewPowerOnAction(yamahaClient),
			yamaha.NewSetInputAction(yamahaClient, "airplay"),
			action.NewFixedArgsAction(owntone.NewPlayAction(owntoneClient), mode),
			action.NewNoOpAction(3 * time.Second),
			yamaha.NewSetVolumeAction(yamahaClient, 39),
			owntone.NewDisplayOutputsAction(owntoneClient, true),
		},
		ignoreError: true,
	}
}

// NewPlayRandomPlaylistSubcommand creates a new Subcommand for random playlist playback.
func NewPlayRandomPlaylistSubcommand(definition Definition, config Config) Subcommand {
	return newPlayRandomMusicSubcommand(definition, config, "")
}

// NewPlayRandomArtistSubcommand creates a new Subcommand for random artist playback.
func NewPlayRandomArtistSubcommand(definition Definition, config Config) Subcommand {
	return newPlayRandomMusicSubcommand(definition, config, "artist")
}

// NewPlayRandomGenreSubcommand creates a new Subcommand for random genre playback.
func NewPlayRandomGenreSubcommand(definition Definition, config Config) Subcommand {
	return newPlayRandomMusicSubcommand(definition, config, "genre")
}
