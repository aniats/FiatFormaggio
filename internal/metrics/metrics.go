package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fiat_formaggio_requests_total",
			Help: "Total number of requests handled by the bot",
		},
		[]string{"command"},
	)

	ActiveSessions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "fiat_formaggio_active_sessions",
			Help: "Number of active bot sessions",
		},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fiat_formaggio_request_duration_seconds",
			Help:    "Duration of requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"command"},
	)

	DatabaseConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "fiat_formaggio_database_connections",
			Help: "Number of active database connections",
		},
	)
)

func Init() {
	prometheus.MustRegister(TotalRequests)
	prometheus.MustRegister(ActiveSessions)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(DatabaseConnections)

	ActiveSessions.Set(0)
	DatabaseConnections.Set(1)
}

func RecordRequest(command string, duration time.Duration) {
	TotalRequests.WithLabelValues(command).Inc()
	RequestDuration.WithLabelValues(command).Observe(duration.Seconds())
}

func IncrementActiveSessions() {
	ActiveSessions.Inc()
}

func DecrementActiveSessions() {
	ActiveSessions.Dec()
}