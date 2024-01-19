package action

type Action interface {
	Run(args string) (string, error)
}
