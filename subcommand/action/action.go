/*
Package action defines the interface for smart home actions.
Actions are the smallest units of work, such as calling an API or controlling a device.
*/
package action

import "context"

// Action is an interface for executing a single smart home action.
type Action interface {
	Run(ctx context.Context, args string) (string, error)
}
