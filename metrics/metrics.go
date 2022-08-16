package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Latency = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "method_duration",
		Help: "Latency of the service",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	}, []string{"method"})

	Subscriptions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_subscriptions",
		Help: "Number of total active subscriptions",
	})

	SuccessfullCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "method_count",
		Help: "Number of successful calls to the service",
	}, []string{"method"})

	FailedCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "method_failure_count",
		Help: "Number of failed calls to the service",
	}, []string{"method"})
)

func init() {

	prometheus.Register(Latency)

	prometheus.Register(Subscriptions)

	prometheus.Register(SuccessfullCalls)

	prometheus.Register(FailedCalls)
}
