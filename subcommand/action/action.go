package action

import "context"

type Action interface {
	Run(ctx context.Context, args string) (string, error)
}
