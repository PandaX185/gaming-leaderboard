package worker

import (
	"context"
	"gaming-leaderboard/internal/queue"
	"log"
	"os"
	"strconv"
)

type Worker struct {
	qu         queue.IQueue
	maxRetries int
}

func NewPlayerWorker(qu queue.IQueue) *Worker {
	return &Worker{
		qu: qu,
	}
}

func (w *Worker) SetMaxRetries(retries int) *Worker {
	w.maxRetries = retries
	return w
}

func (w *Worker) Start(ctx context.Context) {
	poolSize, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
	if err != nil {
		poolSize = 100
	}
	for i := 0; i < poolSize; i++ {
		go func() {
			for {
				select {
				case event := <-w.qu.GetEvents():
					if err := event.Handler(ctx, event.Payload); err != nil {
						if event.Attempt < w.maxRetries {
							log.Printf("Attempt %d failed, retrying...\n", event.Attempt+1)
							event.Attempt++
							w.qu.GetEvents() <- event
						} else {
							log.Printf("Max retries reached for event %s, discarding", event.Type)
						}
					}

				case <-ctx.Done():
					log.Println("Worker shutting down...")
					return
				}
			}
		}()
	}
	<-ctx.Done()
}
