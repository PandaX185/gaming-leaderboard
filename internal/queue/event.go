package queue

import "context"

type Event struct {
	Type    string                                       `json:"type"`
	Payload any                                          `json:"payload"`
	Handler func(ctx context.Context, payload any) error `json:"-"`
	Ack     func(ctx context.Context) error              `json:"-"`
	Attempt int                                          `json:"attempt"`
}
