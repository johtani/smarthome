package subcommand

const StartMusicCmd = "start music"

func NewStartMusicCmdDefinition() Definition {
	return Definition{
		Name:        StartMusicCmd,
		Description: "Start Music",
		// 現状、違いがないので再利用している
		Factory: NewFinishMeetingSubcommand,
	}
}
