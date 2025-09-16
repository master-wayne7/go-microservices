package monitoring

import (
	"database/sql"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector holds all Prometheus metrics and a private registry
type MetricsCollector struct {
	registry *prometheus.Registry

	// Service name for labeling
	serviceName string

	// API Metrics
	httpRequestsTotal      *prometheus.CounterVec
	httpRequestDuration    *prometheus.HistogramVec
	grpcRequestsTotal      *prometheus.CounterVec
	grpcRequestDuration    *prometheus.HistogramVec
	graphqlRequestsTotal   *prometheus.CounterVec
	graphqlRequestDuration *prometheus.HistogramVec

	// System Metrics
	cpuUsageGauge    prometheus.Gauge
	memoryUsageGauge prometheus.Gauge
	goroutinesGauge  prometheus.Gauge
	uptimeGauge      prometheus.Gauge

	// Database Metrics
	dbConnectionsGauge  prometheus.Gauge
	dbQueriesTotal      *prometheus.CounterVec
	dbQueryDuration     *prometheus.HistogramVec
	dbConnectionsActive prometheus.Gauge
	dbConnectionsIdle   prometheus.Gauge

	// Service Info
	serviceInfo *prometheus.GaugeVec

	startTime time.Time
}

// NewMetricsCollector creates a new metrics collector for a service with its own registry
func NewMetricsCollector(serviceName string) *MetricsCollector {
	// Create a new registry instead of using the global DefaultRegisterer
	reg := prometheus.NewRegistry()

	metricPrefix := strings.ReplaceAll(serviceName, "-", "_")
	labels := prometheus.Labels{"service": serviceName}

	// instantiate metrics (use prometheus.New* not promauto)
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        metricPrefix + "_http_requests_total",
			Help:        "Total number of HTTP requests",
			ConstLabels: labels,
		},
		[]string{"method", "endpoint", "status_code"},
	)
	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        metricPrefix + "_http_request_duration_seconds",
			Help:        "HTTP request duration in seconds",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	grpcRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        metricPrefix + "_grpc_requests_total",
			Help:        "Total number of gRPC requests",
			ConstLabels: labels,
		},
		[]string{"method", "status_code"},
	)
	grpcRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        metricPrefix + "_grpc_request_duration_seconds",
			Help:        "gRPC request duration in seconds",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	graphqlRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        metricPrefix + "_graphql_requests_total",
			Help:        "Total number of GraphQL requests",
			ConstLabels: labels,
		},
		[]string{"operation", "status_code"},
	)
	graphqlRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        metricPrefix + "_graphql_request_duration_seconds",
			Help:        "GraphQL request duration in seconds",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	cpuUsageGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metricPrefix + "_cpu_usage_percent",
		Help:        "CPU usage percentage",
		ConstLabels: labels,
	})
	memoryUsageGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metricPrefix + "_memory_usage_bytes",
		Help:        "Memory usage in bytes",
		ConstLabels: labels,
	})
	goroutinesGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metricPrefix + "_goroutines_count",
		Help:        "Number of goroutines",
		ConstLabels: labels,
	})
	uptimeGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metricPrefix + "_uptime_seconds",
		Help:        "Service uptime in seconds",
		ConstLabels: labels,
	})

	dbConnectionsGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metricPrefix + "_db_connections_total",
		Help:        "Total number of database connections",
		ConstLabels: labels,
	})
	dbQueriesTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        metricPrefix + "_db_queries_total",
			Help:        "Total number of database queries",
			ConstLabels: labels,
		},
		[]string{"operation", "table"},
	)
	dbQueryDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        metricPrefix + "_db_query_duration_seconds",
			Help:        "Database query duration in seconds",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)
	dbConnectionsActive := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metricPrefix + "_db_connections_active",
		Help:        "Number of active database connections",
		ConstLabels: labels,
	})
	dbConnectionsIdle := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metricPrefix + "_db_connections_idle",
		Help:        "Number of idle database connections",
		ConstLabels: labels,
	})

	serviceInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        metricPrefix + "_service_info",
		Help:        "Service information",
		ConstLabels: labels,
	}, []string{"version", "environment"})

	// Register all collectors to this private registry
	reg.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		grpcRequestsTotal,
		grpcRequestDuration,
		graphqlRequestsTotal,
		graphqlRequestDuration,
		cpuUsageGauge,
		memoryUsageGauge,
		goroutinesGauge,
		uptimeGauge,
		dbConnectionsGauge,
		dbQueriesTotal,
		dbQueryDuration,
		dbConnectionsActive,
		dbConnectionsIdle,
		serviceInfo,
	)

	return &MetricsCollector{
		registry:               reg,
		serviceName:            serviceName,
		httpRequestsTotal:      httpRequestsTotal,
		httpRequestDuration:    httpRequestDuration,
		grpcRequestsTotal:      grpcRequestsTotal,
		grpcRequestDuration:    grpcRequestDuration,
		graphqlRequestsTotal:   graphqlRequestsTotal,
		graphqlRequestDuration: graphqlRequestDuration,
		cpuUsageGauge:          cpuUsageGauge,
		memoryUsageGauge:       memoryUsageGauge,
		goroutinesGauge:        goroutinesGauge,
		uptimeGauge:            uptimeGauge,
		dbConnectionsGauge:     dbConnectionsGauge,
		dbQueriesTotal:         dbQueriesTotal,
		dbQueryDuration:        dbQueryDuration,
		dbConnectionsActive:    dbConnectionsActive,
		dbConnectionsIdle:      dbConnectionsIdle,
		serviceInfo:            serviceInfo,
		startTime:              time.Now(),
	}
}

// PrometheusHandler returns an HTTP handler that serves this collector's registry
func (mc *MetricsCollector) PrometheusHandler() http.Handler {
	return promhttp.HandlerFor(mc.registry, promhttp.HandlerOpts{})
}

// RecordHTTPRequest records HTTP request metrics
func (mc *MetricsCollector) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	mc.httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	mc.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordGRPCRequest records gRPC request metrics
func (mc *MetricsCollector) RecordGRPCRequest(method, statusCode string, duration time.Duration) {
	mc.grpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
	mc.grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordGraphQLRequest records GraphQL request metrics
func (mc *MetricsCollector) RecordGraphQLRequest(operation, statusCode string, duration time.Duration) {
	mc.graphqlRequestsTotal.WithLabelValues(operation, statusCode).Inc()
	mc.graphqlRequestDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordDBQuery records database query metrics
func (mc *MetricsCollector) RecordDBQuery(operation, table string, duration time.Duration) {
	mc.dbQueriesTotal.WithLabelValues(operation, table).Inc()
	mc.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// UpdateSystemMetrics updates system metrics
func (mc *MetricsCollector) UpdateSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mc.memoryUsageGauge.Set(float64(m.Alloc))
	mc.goroutinesGauge.Set(float64(runtime.NumGoroutine()))
	mc.uptimeGauge.Set(time.Since(mc.startTime).Seconds())

	// Placeholder for CPU - replace with a real measurement in production
	mc.cpuUsageGauge.Set(float64(runtime.NumGoroutine()) * 0.1)
}

// UpdateDBMetrics updates database connection metrics
func (mc *MetricsCollector) UpdateDBMetrics(db *sql.DB) {
	if db == nil {
		return
	}
	stats := db.Stats()
	mc.dbConnectionsGauge.Set(float64(stats.OpenConnections))
	mc.dbConnectionsActive.Set(float64(stats.InUse))
	mc.dbConnectionsIdle.Set(float64(stats.Idle))
}

// SetServiceInfo sets service information
func (mc *MetricsCollector) SetServiceInfo(version, environment string) {
	mc.serviceInfo.WithLabelValues(version, environment).Set(1)
}

// StartSystemMetricsCollection starts a goroutine to collect system metrics periodically
func (mc *MetricsCollector) StartSystemMetricsCollection(db *sql.DB) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mc.UpdateSystemMetrics()
			if db != nil {
				mc.UpdateDBMetrics(db)
			}
		}
	}()
}
