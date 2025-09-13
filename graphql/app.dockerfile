# Multi-stage build for GraphQL service with security improvements
FROM golang:1.25-alpine AS build

RUN apk --no-cache add gcc g++ make ca-certificates

WORKDIR /go/src/github.com/master-wayne7/go-microservices

COPY go.mod go.sum ./
RUN go mod download

# GraphQL needs all services
COPY graphql graphql
COPY account account
COPY catalog catalog
COPY order order

RUN go build -o /go/bin/app ./graphql

# Production stage with security improvements
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /usr/bin

# Copy binary from build stage
COPY --from=build /go/bin/app .

# Change ownership to non-root user
RUN chown appuser:appgroup app

# Switch to non-root user
USER appuser

# Expose GraphQL service port
EXPOSE 8087

# Expose health check port
EXPOSE 8088

# Health check for container orchestration
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8088/health || exit 1

# Start the application
CMD ["app"]
