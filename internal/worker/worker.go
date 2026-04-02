package worker

import (
	"context"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/metrics"
	"log"
	"os"
	"strconv"
)

type Worker struct {
	qu         queue.IQueue
	maxRetries int
}

func NewWorker(qu queue.IQueue) *Worker {
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

	eventsChan := w.qu.GetEvents()
	for i := 0; i < poolSize; i++ {
		go func() {
			for {
				select {
				case event := <-eventsChan:
					metrics.WorkerInFlight.Inc()
					if err := event.Handler(ctx, event.Payload); err != nil {
						metrics.WorkerErrors.Inc()
						if event.Attempt < w.maxRetries {
							log.Printf("Attempt %d failed, retrying...\n", event.Attempt+1)
							event.Attempt++
							metrics.WorkerRetriesTotal.WithLabelValues(event.Type).Inc()

							go func(e queue.Event) {
								eventsChan <- e
							}(event)
						} else {
							log.Printf("Max retries reached for event %s, discarding", event.Type)
							if event.Ack != nil {
								event.Ack(ctx)
							}
						}
					} else {
						metrics.WorkerProcessed.Inc()
					}
					metrics.WorkerInFlight.Dec()

				case <-ctx.Done():
					log.Println("Worker shutting down...")
					return
				}
			}
		}()
	}
	<-ctx.Done()
}
