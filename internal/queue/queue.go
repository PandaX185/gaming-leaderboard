package queue

import (
	"context"
)

type IQueue interface {
	PublishEvent(ctx context.Context, data any) error
	GetEvents() chan Event
}
