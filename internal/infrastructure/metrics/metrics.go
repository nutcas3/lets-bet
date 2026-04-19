// Package metrics provides Prometheus RED (Rate/Errors/Duration) instrumentation
// for our HTTP services. Register the middleware once per service and expose
// /metrics via [Handler].
package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Recorder owns the registered Prometheus collectors.
type Recorder struct {
	registry        *prometheus.Registry
	requestTotal    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	inFlight        prometheus.Gauge
}

// New creates a Recorder with a fresh registry scoped to this service.
// The `service` label is baked into every metric.
func New(service string) *Recorder {
	r := prometheus.NewRegistry()
	labels := prometheus.Labels{"service": service}

	reqTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   "betting",
			Subsystem:   "http",
			Name:        "requests_total",
			Help:        "Total HTTP requests served.",
			ConstLabels: labels,
		},
		[]string{"method", "path", "status"},
	)
	reqDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   "betting",
			Subsystem:   "http",
			Name:        "request_duration_seconds",
			Help:        "HTTP request latency in seconds.",
			ConstLabels: labels,
			Buckets:     []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "path"},
	)
	inFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   "betting",
			Subsystem:   "http",
			Name:        "in_flight_requests",
			Help:        "Currently in-flight HTTP requests.",
			ConstLabels: labels,
		},
	)

	r.MustRegister(reqTotal, reqDuration, inFlight)
	r.MustRegister(prometheus.NewGoCollector())
	r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	return &Recorder{
		registry:        r,
		requestTotal:    reqTotal,
		requestDuration: reqDuration,
		inFlight:        inFlight,
	}
}

// Handler exposes /metrics bound to this recorder's registry.
func (r *Recorder) Handler() http.Handler {
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
}

// Registry returns the underlying registry for registering custom collectors.
func (r *Recorder) Registry() *prometheus.Registry { return r.registry }

// RegisterRoutes wires /metrics onto a gorilla/mux router.
func (r *Recorder) RegisterRoutes(router *mux.Router) {
	router.Handle("/metrics", r.Handler()).Methods(http.MethodGet)
}

// Middleware records RED metrics for every request passing through.
func (r *Recorder) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.inFlight.Inc()
		start := time.Now()

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, req)

		route := routeTemplate(req)
		elapsed := time.Since(start).Seconds()
		r.requestDuration.WithLabelValues(req.Method, route).Observe(elapsed)
		r.requestTotal.WithLabelValues(req.Method, route, strconv.Itoa(sw.status)).Inc()
		r.inFlight.Dec()
	})
}

// routeTemplate returns the mux route template (e.g. "/api/users/{id}") so
// cardinality stays bounded. Falls back to the raw path.
func routeTemplate(r *http.Request) string {
	if route := mux.CurrentRoute(r); route != nil {
		if tpl, err := route.GetPathTemplate(); err == nil {
			return tpl
		}
	}
	return r.URL.Path
}

type statusWriter struct {
	http.ResponseWriter
	status   int
	once     sync.Once
	hijacked bool
}

func (s *statusWriter) WriteHeader(code int) {
	s.once.Do(func() { s.status = code })
	s.ResponseWriter.WriteHeader(code)
}
