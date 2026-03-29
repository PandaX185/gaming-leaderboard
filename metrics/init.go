package metrics

import "github.com/prometheus/client_golang/prometheus"

func Init() {
	prometheus.MustRegister(
		RequestsTotal,
		RequestDuration,
		RequestsInFlight,
		QueueSize,
		QueueSizeByStream,
		QueuePublishedTotal,
		QueueConsumedTotal,
		QueueAckTotal,
		QueueReadErrorsTotal,
		WorkerProcessed,
		WorkerErrors,
		WorkerRetriesTotal,
		WorkerInFlight,
	)
}
