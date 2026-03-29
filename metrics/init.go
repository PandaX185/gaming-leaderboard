package metrics

import "github.com/prometheus/client_golang/prometheus"

func Init() {
	prometheus.MustRegister(
		RequestsTotal,
		RequestDuration,
		QueueSize,
		WorkerProcessed,
		WorkerErrors,
	)
}
