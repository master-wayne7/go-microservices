# Multi-stage build for Account service with security improvements
FROM golang:1.23-alpine AS build

# Install build dependencies
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory
WORKDIR /go/src/github.com/master-wayne7/go-microservices

# Copy dependency files
COPY go.mod go.sum ./
COPY vendor vendor

# Copy Account service code
COPY account account

# Build the application
RUN GO111MODULE=on go build -mod vendor -o /go/bin/app ./account/cmd/account

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

# Expose main service port (changed from 8080 to 8081)
EXPOSE 8081

# Expose health check port
EXPOSE 8082

# Health check for container orchestration
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Start the application
CMD ["app"]