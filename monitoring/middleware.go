package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
			// Capture and restore body for parsing operation
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body.Close()
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start)
			statusCode := strconv.Itoa(wrapped.statusCode)

			// Extract GraphQL operation name from request body
			operation := "unknown"
			if r.Method == "POST" && len(bodyBytes) > 0 {
				var payload struct {
					OperationName string `json:"operationName"`
					Query         string `json:"query"`
				}
				if err := json.Unmarshal(bodyBytes, &payload); err == nil {
					if payload.OperationName != "" {
						operation = payload.OperationName
					} else if len(payload.Query) > 0 {
						// Heuristic: determine if query starts with mutation or query
						q := bytes.TrimSpace([]byte(payload.Query))
						if bytes.HasPrefix(bytes.ToLower(q), []byte("mutation")) {
							operation = "mutation"
						} else if bytes.HasPrefix(bytes.ToLower(q), []byte("query")) {
							operation = "query"
						}
					}
				}
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
