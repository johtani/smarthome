package action

type Action interface {
	Run() (string, error)
}
