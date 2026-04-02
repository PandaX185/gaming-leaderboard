package queue

import (
	"context"
)

type IQueue interface {
	PublishEvent(ctx context.Context, data Event) error
	GetEvents() chan Event
}
