package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
)

// These metrics are copied from the promhttp documentation for now.
var (
	inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "method"},
	)

	// responseSize is partitioned by the HTTP method and handler.
	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{"code", "method"},
	)
)

func init() {
	prometheus.MustRegister(inFlightGauge, counter, duration, responseSize)
}

func addRequestMetrics(name string, h http.Handler) http.Handler {
	return promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": name}),
			promhttp.InstrumentHandlerCounter(counter,
				promhttp.InstrumentHandlerResponseSize(responseSize, h),
			),
		),
	)
}
