package subcommand

const StartMusicCmd = "start-music"

func NewStartMusicCmdDefinition() Definition {
	return Definition{
		StartMusicCmd,
		"Start Music",
		// 現状、違いがないので再利用している
		NewFinishMeetingSubcommand,
	}
}
