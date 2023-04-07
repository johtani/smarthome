package owntone

type SetVolumeAction struct {
	name   string
	volume int
	c      Client
}

func (a SetVolumeAction) Run() error {
	err := a.c.SetVolume(a.volume)
	if err != nil {
		return err
	}
	println("owntone set volume action succeeded.")
	return nil
}

func NewSetVolumeAction() SetVolumeAction {
	return SetVolumeAction{
		"Set Owntone Volume",
		33,
		NewOwntoneClient(),
	}
}
