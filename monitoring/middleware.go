package monitoring

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTPMiddleware creates HTTP middleware for metrics collection
func HTTPMiddleware(metrics *MetricsCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call the next handler
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start)
			statusCode := strconv.Itoa(wrapped.statusCode)

			metrics.RecordHTTPRequest(
				r.Method,
				r.URL.Path,
				statusCode,
				duration,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GRPCUnaryServerInterceptor creates gRPC unary server interceptor for metrics
func GRPCUnaryServerInterceptor(metrics *MetricsCollector) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Record metrics
		duration := time.Since(start)
		statusCode := getGRPCStatusCode(err)

		metrics.RecordGRPCRequest(
			info.FullMethod,
			statusCode,
			duration,
		)

		return resp, err
	}
}

// GRPCStreamServerInterceptor creates gRPC stream server interceptor for metrics
func GRPCStreamServerInterceptor(metrics *MetricsCollector) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Call the handler
		err := handler(srv, ss)

		// Record metrics
		duration := time.Since(start)
		statusCode := getGRPCStatusCode(err)

		metrics.RecordGRPCRequest(
			info.FullMethod,
			statusCode,
			duration,
		)

		return err
	}
}

// getGRPCStatusCode extracts status code from gRPC error
func getGRPCStatusCode(err error) string {
	if err == nil {
		return codes.OK.String()
	}

	if st, ok := status.FromError(err); ok {
		return st.Code().String()
	}

	return codes.Unknown.String()
}

// GraphQLMiddleware creates GraphQL middleware for metrics collection
func GraphQLMiddleware(metrics *MetricsCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call the next handler
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start)
			statusCode := strconv.Itoa(wrapped.statusCode)

			// ### CHANGE THIS #### - Extract GraphQL operation name from request
			// This is a simplified implementation. You might want to parse the GraphQL query
			// to extract the actual operation name
			operation := "unknown"
			if r.Method == "POST" {
				// Try to extract operation from request body or headers
				operation = r.Header.Get("X-GraphQL-Operation")
				if operation == "" {
					operation = "mutation" // Default assumption for POST requests
				}
			} else {
				operation = "query" // Default for GET requests
			}

			metrics.RecordGraphQLRequest(
				operation,
				statusCode,
				duration,
			)
		})
	}
}

// PrometheusHandler returns a handler for Prometheus metrics endpoint
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
