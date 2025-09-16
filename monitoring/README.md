# Prometheus Monitoring for Go Microservices

This monitoring solution provides comprehensive metrics collection for your Go microservices architecture using Prometheus and Grafana.

## Features

### API Metrics
- **HTTP Requests**: Total count, duration, and status codes for REST endpoints
- **gRPC Requests**: Total count, duration, and status codes for gRPC services
- **GraphQL Requests**: Total count, duration, and status codes for GraphQL operations

### System Metrics
- **Memory Usage**: Current memory consumption per service
- **Goroutines**: Number of active goroutines per service
- **CPU Usage**: CPU utilization (placeholder implementation)
- **Uptime**: Service uptime in seconds

### Database Metrics
- **Query Count**: Total number of database queries by operation type
- **Query Duration**: Database query response times
- **Connection Pool**: Active, idle, and total database connections

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Account       │    │   Catalog       │    │   Order         │
│   Service       │    │   Service       │    │   Service       │
│   :8081/:8082   │    │   :8083/:8084   │    │   :8085/:8086   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   GraphQL       │
                    │   Service       │
                    │   :8087/:8088   │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Prometheus    │
                    │   :9090         │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Grafana       │
                    │   :3000         │
                    └─────────────────┘
```

## Quick Start

1. **Start all services with monitoring**:
   ```bash
   docker-compose up -d
   ```

2. **Access the monitoring stack**:
   - **Prometheus**: http://localhost:9090
   - **Grafana**: http://localhost:3000 (admin/admin)
   - **Service Metrics**: 
     - Account: http://localhost:8082/metrics
     - Catalog: http://localhost:8084/metrics
     - Order: http://localhost:8086/metrics
     - GraphQL: http://localhost:8088/metrics

## Metrics Endpoints

Each service exposes metrics on its health check port:

- **Account Service**: `http://localhost:8082/metrics`
- **Catalog Service**: `http://localhost:8084/metrics`
- **Order Service**: `http://localhost:8086/metrics`
- **GraphQL Service**: `http://localhost:8088/metrics`

## Available Metrics

### HTTP Metrics
- `http_requests_total{service, method, endpoint, status_code}`
- `http_request_duration_seconds{service, method, endpoint}`

### gRPC Metrics
- `grpc_requests_total{service, method, status_code}`
- `grpc_request_duration_seconds{service, method}`

### GraphQL Metrics
- `graphql_requests_total{service, operation, status_code}`
- `graphql_request_duration_seconds{service, operation}`

### System Metrics
- `memory_usage_bytes{service}`
- `goroutines_count{service}`
- `cpu_usage_percent{service}`
- `uptime_seconds{service}`

### Database Metrics
- `db_queries_total{service, operation, table}`
- `db_query_duration_seconds{service, operation, table}`
- `db_connections_total{service}`
- `db_connections_active{service}`
- `db_connections_idle{service}`

## Grafana Dashboards

The system includes a pre-configured dashboard with panels for:
- Request rates and response times
- System resource usage
- Database performance metrics
- Service health status

## Configuration

### Prometheus Configuration
Edit `monitoring/prometheus.yml` to modify scrape intervals, add new targets, or configure alerting rules.

### Grafana Configuration
- **Data Sources**: `monitoring/grafana/datasources/prometheus.yml`
- **Dashboards**: `monitoring/grafana/dashboards/`

## Customization

### Adding New Metrics
1. Add new metric definitions in `monitoring/metrics.go`
2. Update the `MetricsCollector` struct
3. Add recording methods for your metrics
4. Update the Grafana dashboard to visualize new metrics

### Service-Specific Metrics
Each service can add custom metrics by:
1. Creating a service-specific metrics collector
2. Recording metrics in your business logic
3. Exposing them through the `/metrics` endpoint

## Troubleshooting

### Common Issues

1. **Metrics not appearing**: Check that services are running and accessible
2. **Prometheus can't scrape**: Verify network connectivity and port accessibility
3. **Grafana can't connect to Prometheus**: Check Prometheus service status

### Debug Commands
```bash
# Check service health
curl http://localhost:8082/health  # Account
curl http://localhost:8084/health  # Catalog
curl http://localhost:8086/health  # Order
curl http://localhost:8088/health  # GraphQL

# Check metrics
curl http://localhost:8082/metrics  # Account metrics
curl http://localhost:9090/targets  # Prometheus targets
```

## Production Considerations

### Security
- Change default Grafana admin password
- Use authentication for Prometheus and Grafana
- Consider network isolation for monitoring stack

### Performance
- Adjust scrape intervals based on your needs
- Configure retention policies for time series data
- Monitor Prometheus resource usage

### Scaling
- Consider Prometheus federation for large deployments
- Use Grafana clustering for high availability
- Implement proper backup strategies for configuration and data

## ### CHANGE THIS #### - Areas for Enhancement

1. **CPU Metrics**: Implement proper CPU usage calculation
2. **Custom Business Metrics**: Add service-specific business metrics
3. **Alerting Rules**: Configure Prometheus alerting rules
4. **Database Exporters**: Add dedicated database exporters for more detailed DB metrics
5. **Service Discovery**: Implement automatic service discovery
6. **Tracing**: Add distributed tracing with Jaeger or similar

