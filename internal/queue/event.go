package queue

import "context"

type Event struct {
	Type    string
	Payload any
	Handler func(ctx context.Context, payload any) error
	Attempt int
}
