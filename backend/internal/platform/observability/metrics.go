package observability

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type Metrics struct {
	Registry                *prometheus.Registry
	HTTPRequests            *prometheus.CounterVec
	HTTPRequestDuration     *prometheus.HistogramVec
	ProviderRequests        *prometheus.CounterVec
	ProviderRequestDuration *prometheus.HistogramVec
	ProviderCache           *prometheus.CounterVec
	ProviderCircuit         *prometheus.GaugeVec
	RedisCommands           *prometheus.CounterVec
	AuthEvents              *prometheus.CounterVec
	PlaceQueries            *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()
	metrics := &Metrics{
		Registry: registry,
		HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "alongtu_http_requests_total",
			Help: "Total API requests grouped by stable route template and status.",
		}, []string{"method", "route", "status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "alongtu_http_request_duration_seconds",
			Help:    "API request latency grouped by stable route template.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		ProviderRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "alongtu_provider_requests_total",
			Help: "Total map provider calls by operation and bounded outcome.",
		}, []string{"operation", "outcome"}),
		ProviderRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "alongtu_provider_request_duration_seconds",
			Help:    "Map provider call latency by operation and bounded outcome.",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 1.5, 2.2, 3, 5},
		}, []string{"operation", "outcome"}),
		ProviderCache: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "alongtu_provider_cache_total",
			Help: "Provider cache decisions by operation and bounded state.",
		}, []string{"operation", "state"}),
		ProviderCircuit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "alongtu_provider_circuit_state",
			Help: "Current provider circuit state (1 for active state, 0 otherwise).",
		}, []string{"state"}),
		RedisCommands: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "alongtu_redis_commands_total",
			Help: "Redis commands grouped by command type and bounded outcome.",
		}, []string{"command", "outcome"}),
		AuthEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "alongtu_auth_events_total",
			Help: "Authentication events grouped by operation and bounded outcome.",
		}, []string{"operation", "outcome"}),
		PlaceQueries: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "alongtu_place_queries_total",
			Help: "Place queries grouped by operation and bounded outcome.",
		}, []string{"operation", "outcome"}),
	}
	registry.MustRegister(
		metrics.HTTPRequests,
		metrics.HTTPRequestDuration,
		metrics.ProviderRequests,
		metrics.ProviderRequestDuration,
		metrics.ProviderCache,
		metrics.ProviderCircuit,
		metrics.RedisCommands,
		metrics.AuthEvents,
		metrics.PlaceQueries,
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)
	metrics.SetProviderCircuit("closed")
	return metrics
}

func (m *Metrics) RegisterDatabasePool(pool *pgxpool.Pool) error {
	return m.Registry.Register(newDatabasePoolCollector(pool))
}

func (m *Metrics) RedisHook() redis.Hook { return redisMetricsHook{metrics: m} }

func (m *Metrics) ObserveAuth(operation, outcome string) {
	m.AuthEvents.WithLabelValues(operation, outcome).Inc()
}

func (m *Metrics) ObservePlaceQuery(operation, outcome string) {
	m.PlaceQueries.WithLabelValues(operation, outcome).Inc()
}

func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(writer, r)
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		m.HTTPRequests.WithLabelValues(r.Method, route, strconv.Itoa(writer.status)).Inc()
		m.HTTPRequestDuration.WithLabelValues(r.Method, route).Observe(time.Since(started).Seconds())
	})
}

func (m *Metrics) ObserveProvider(operation, outcome string, duration time.Duration) {
	m.ProviderRequests.WithLabelValues(operation, outcome).Inc()
	m.ProviderRequestDuration.WithLabelValues(operation, outcome).Observe(duration.Seconds())
}

func (m *Metrics) ObserveProviderCache(operation, state string) {
	m.ProviderCache.WithLabelValues(operation, state).Inc()
}

func (m *Metrics) SetProviderCircuit(state string) {
	for _, candidate := range []string{"closed", "open", "half_open"} {
		value := 0.0
		if candidate == state {
			value = 1
		}
		m.ProviderCircuit.WithLabelValues(candidate).Set(value)
	}
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

type databasePoolCollector struct {
	pool        *pgxpool.Pool
	connections *prometheus.Desc
	waits       *prometheus.Desc
}

func newDatabasePoolCollector(pool *pgxpool.Pool) *databasePoolCollector {
	return &databasePoolCollector{
		pool:        pool,
		connections: prometheus.NewDesc("alongtu_database_pool_connections", "PostgreSQL pool connections by state.", []string{"state"}, nil),
		waits:       prometheus.NewDesc("alongtu_database_pool_acquire_count_total", "Total PostgreSQL pool acquire attempts.", nil, nil),
	}
}

func (c *databasePoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.connections
	ch <- c.waits
}

func (c *databasePoolCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()
	states := map[string]int32{
		"total":    stat.TotalConns(),
		"acquired": stat.AcquiredConns(),
		"idle":     stat.IdleConns(),
		"max":      stat.MaxConns(),
	}
	for state, value := range states {
		ch <- prometheus.MustNewConstMetric(c.connections, prometheus.GaugeValue, float64(value), state)
	}
	ch <- prometheus.MustNewConstMetric(c.waits, prometheus.CounterValue, float64(stat.AcquireCount()))
}

type redisMetricsHook struct{ metrics *Metrics }

func (h redisMetricsHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := next(ctx, network, addr)
		h.metrics.RedisCommands.WithLabelValues("dial", redisOutcome(err)).Inc()
		return conn, err
	}
}

func (h redisMetricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		err := next(ctx, cmd)
		h.metrics.RedisCommands.WithLabelValues(cmd.Name(), redisOutcome(err)).Inc()
		return err
	}
}

func (h redisMetricsHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		err := next(ctx, cmds)
		h.metrics.RedisCommands.WithLabelValues("pipeline", redisOutcome(err)).Inc()
		return err
	}
}

func redisOutcome(err error) string {
	if err == nil {
		return "success"
	}
	if errors.Is(err, redis.Nil) {
		return "miss"
	}
	return "error"
}

func (w *statusWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}
