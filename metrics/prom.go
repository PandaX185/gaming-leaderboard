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

	QueueSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_stream_length",
			Help: "Redis stream length",
		},
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
)
