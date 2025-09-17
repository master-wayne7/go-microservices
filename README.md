# Go Microservices Project

A microservices-based e-commerce application built with Go, featuring gRPC services, GraphQL API, and containerized deployment.

## 🆕 **Latest Updates**

### ✅ **All 4 Services Completed**
- **Account Service** - User account management with gRPC API ✅ **COMPLETED**
- **Catalog Service** - Product catalog management with Elasticsearch backend ✅ **COMPLETED**
- **Order Service** - Order processing with PostgreSQL backend ✅ **COMPLETED**
- **GraphQL Gateway** - Unified API gateway using GraphQL ✅ **COMPLETED**

### 🔧 **Current Work in Progress**
- **Metrics and Monitoring** - Implementing Prometheus metrics and Grafana dashboards 🔄 **IN PROGRESS** 


---

## 🏗️ Architecture

This project implements a microservices architecture with the following components:

### Services
- **Account Service** (`/account`) - User account management with gRPC API ✅ **COMPLETED**
- **Catalog Service** (`/catalog`) - Product catalog management with Elasticsearch backend ✅ **COMPLETED**
- **Order Service** (`/order`) - Order processing with PostgreSQL backend ✅ **COMPLETED**
- **GraphQL Gateway** (`/graphql`) - Unified API gateway using GraphQL ✅ **COMPLETED**

### Technology Stack
- **Language**: Go 1.25.0
- **gRPC**: Inter-service communication
- **GraphQL**: API gateway with gqlgen
- **Database**: PostgreSQL (account & order services), Elasticsearch (catalog service)
- **Containerization**: Docker & Docker Compose ✅ **PRODUCTION READY**
- **Protocol Buffers**: Service contracts

## 📁 Project Structure

```
go-microservices/
├── account/           # Account microservice ✅ PRODUCTION READY
│   ├── account.proto  # gRPC service definition
│   ├── server.go      # gRPC server implementation
│   ├── service.go     # Business logic
│   ├── repository.go  # Data access layer
│   ├── app.dockerfile # Production Docker image ✅ ENHANCED
│   └── pb/           # Generated protobuf files
├── catalog/          # Catalog microservice ✅ PRODUCTION READY
│   ├── catalog.proto # gRPC service definition
│   ├── server.go     # gRPC server implementation
│   ├── service.go    # Business logic
│   ├── repository.go # Elasticsearch data layer
│   ├── client.go     # gRPC client library
│   ├── app.dockerfile # Production Docker image ✅ ENHANCED
│   └── pb/          # Generated protobuf files
├── order/            # Order microservice ✅ COMPLETED
│   ├── order.proto   # gRPC service definition ✅ COMPLETED
│   ├── server.go     # gRPC server implementation ✅ COMPLETED
│   ├── service.go    # Business logic ✅ COMPLETED
│   ├── repository.go # PostgreSQL data layer ✅ COMPLETED
│   ├── client.go     # gRPC client library ✅ COMPLETED
│   ├── app.dockerfile # Production Docker image ✅ COMPLETED
│   ├── pb/          # Generated protobuf files ✅ COMPLETED
│   └── up.sql       # Database schema ✅ COMPLETED
├── graphql/          # GraphQL API gateway ✅ ENHANCED
│   ├── schema.graphql # GraphQL schema
│   ├── main.go       # GraphQL server ✅ ENHANCED
│   ├── app.dockerfile # Production Docker image ✅ NEW
│   └── generated.go  # Auto-generated resolvers
├── docker-compose.yml # Container orchestration ✅ PRODUCTION READY
└── go.mod            # Go module definition
```

## 🚀 Getting Started

### Prerequisites
- Go 1.25.0 or later
- Protocol Buffers compiler (`protoc`)
- Docker and Docker Compose ✅ **ENHANCED**
- PostgreSQL (for account & order services)
- Elasticsearch (for catalog service)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-microservices
   ```

2. **Install Go dependencies**
   ```bash
   go mod tidy
   ```

3. **Install Protocol Buffers tools**
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

4. **Generate protobuf files**
   ```bash
   # For account service
   cd account
   go generate
   
   # For catalog service
   cd ../catalog
   go generate
   ```

### Running the Services

#### **Production-Ready Docker Deployment** ✅ **NEW**
```bash
# Start all services with health checks and proper orchestration
docker-compose up --build

# View service status and health
docker-compose ps
docker-compose logs -f [service-name]

# Scale services (when implemented)
docker-compose up --scale order=3
```

#### Development Mode
```bash
# Start all services with Docker Compose
docker-compose up

# Or run individual services
cd account && go run cmd/account/main.go
cd catalog && go run cmd/catalog/main.go
cd order && go run cmd/order/main.go
cd graphql && go run main.go
```

## 🔧 Development

### **New Development Workflow** ✅ **ENHANCED**
1. **Service Development**: Each service has its own Docker image with health checks
2. **Database Integration**: Proper schema management with `up.sql` files
3. **Health Monitoring**: Built-in health endpoints for container orchestration
4. **Security**: Non-root execution and minimal attack surface

### Adding New Services
1. Create a new service directory
2. Define your `.proto` file for gRPC contracts
3. Generate Go code: `go generate`
4. Implement service, repository, and server layers
5. Add to `docker-compose.yml` with health checks
6. Create production-ready Docker image

### GraphQL Schema Updates
1. Modify `graphql/schema.graphql`
2. Regenerate resolvers: `go run github.com/99designs/gqlgen generate`

### Protocol Buffer Updates
1. Modify `.proto` files
2. Regenerate Go code: `go generate`

## 📊 API Endpoints

### **Updated Service Ports** ✅ **NEW**
- **Account Service**: `localhost:8081` (gRPC), `localhost:8082` (Health)
- **Catalog Service**: `localhost:8083` (gRPC), `localhost:8084` (Health)
- **Order Service**: `localhost:8085` (gRPC), `localhost:8086` (Health)
- **GraphQL Gateway**: `localhost:8087` (API), `localhost:8088` (Health)

### GraphQL Gateway
- **URL**: `http://localhost:8087/graphql` ✅ **UPDATED**
- **Playground**: `http://localhost:8087/playground` ✅ **UPDATED**

### gRPC Services
- **Account Service**: `localhost:8081` ✅ **UPDATED**
- **Catalog Service**: `localhost:8083` ✅ **UPDATED**
- **Order Service**: `localhost:8085` ✅ **NEW**

## 🗄️ Database Schema

Each microservice has its own database with proper health monitoring:
- `account_db` - User accounts (PostgreSQL) ✅ **HEALTH MONITORED**
- `catalog_db` - Product catalog (Elasticsearch) ✅ **HEALTH MONITORED**
- `order_db` - Orders and transactions (PostgreSQL) ✅ **NEW & HEALTH MONITORED**

## 🔄 Service Communication

- **Internal**: gRPC for inter-service communication ✅ **ENHANCED**
- **External**: GraphQL API gateway for client applications ✅ **ENHANCED**
- **Health Monitoring**: Built-in health checks for orchestration ✅ **NEW**
- **Data Consistency**: Transaction-based order processing ✅ **NEW**

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run specific service tests
go test ./account/...
go test ./catalog/...
go test ./order/... ✅ **NEW**
go test ./graphql/...

# Test health endpoints
curl http://localhost:8082/health  # Account
curl http://localhost:8084/health  # Catalog
curl http://localhost:8086/health  # Order
curl http://localhost:8088/health  # GraphQL
```

## 📦 Deployment

### **Production-Ready Docker** ✅ **ENHANCED**
```bash
# Build optimized production images
docker-compose build

# Run with health monitoring
docker-compose up -d

# Monitor service health
docker-compose ps
docker-compose logs -f [service-name]

# Scale services (when implemented)
docker-compose up --scale order=3
```

### **Health Check Monitoring** ✅ **NEW**
- **Automatic health monitoring** every 30 seconds
- **Service dependency management** - services wait for healthy dependencies
- **Container restart policies** based on health status
- **Production-ready orchestration** with proper startup sequences

### Kubernetes (planned)
- Helm charts for each service
- Service mesh integration
- Horizontal pod autoscaling
- **Health check integration** ✅ **READY**

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. **Add health checks** for new services ✅ **REQUIRED**
5. **Use non-root users** in Docker images ✅ **REQUIRED**
6. **Test with Docker Compose** before submitting ✅ **REQUIRED**
7. Submit a pull request

## 📝 License

This project is under development and subject to change.

## 🔮 Roadmap

- [x] Complete Account Service implementation ✅
- [x] Complete Catalog Service implementation ✅
- [x] Complete Order Service implementation ✅
- [x] Complete GraphQL Gateway implementation ✅
- [x] Fix Docker port conflicts ✅
- [x] Implement health checks ✅
- [x] Security hardening ✅
- [x] Production-ready Docker images ✅
- [ ] Add monitoring and logging 🔄 **IN PROGRESS**
- [ ] Kubernetes deployment
- [ ] Service mesh integration

## 🎯 **Today's Major Achievements**

### **Infrastructure & DevOps**
- ✅ **Eliminated port conflicts** - All services can now run simultaneously
- ✅ **Production-ready Docker images** - Security hardened with non-root users
- ✅ **Health check system** - Container orchestration ready
- ✅ **Service dependency management** - Proper startup sequences

### **Order Service Development**
- ✅ **Complete service architecture** - Service, repository, and server layers
- ✅ **PostgreSQL integration** - Transaction-based order processing
- ✅ **Product relationship management** - Normalized order-product structure
- ✅ **Docker containerization** - Production-ready deployment

### **Security & Best Practices**
- ✅ **Non-root execution** - All services run as unprivileged users
- ✅ **Minimal attack surface** - Multi-stage builds with essential dependencies only
- ✅ **Health monitoring** - Built-in monitoring for production environments
- ✅ **Proper service isolation** - Each service has dedicated ports and health endpoints

---

**Note**: All 4 microservices (Account, Catalog, Order, and GraphQL Gateway) are now fully implemented and production-ready. Currently working on metrics and monitoring implementation with Prometheus and Grafana.
