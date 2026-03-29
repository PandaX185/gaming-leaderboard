package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	RequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current in-flight HTTP requests",
		},
		[]string{"method", "endpoint"},
	)

	QueueSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_stream_length",
			Help: "Redis stream length",
		},
	)

	QueueSizeByStream = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_length",
			Help: "Queue length by backend stream or queue name",
		},
		[]string{"queue"},
	)

	QueuePublishedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_published_total",
			Help: "Total queue publish attempts",
		},
		[]string{"event_type", "result"},
	)

	QueueConsumedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_consumed_total",
			Help: "Total consumed queue events",
		},
		[]string{"event_type", "source"},
	)

	QueueAckTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_ack_total",
			Help: "Total queue ACK attempts",
		},
		[]string{"event_type", "result"},
	)

	QueueReadErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_read_errors_total",
			Help: "Total queue read errors",
		},
		[]string{"operation"},
	)

	WorkerProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_processed_total",
			Help: "Total processed events",
		},
	)

	WorkerErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_errors_total",
			Help: "Total worker errors",
		},
	)

	WorkerRetriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_retries_total",
			Help: "Total worker retries",
		},
		[]string{"event_type"},
	)

	WorkerInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_events_in_flight",
			Help: "Current number of worker events being processed",
		},
	)
)
