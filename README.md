# Go Microservices – Accounts, Catalog, Orders, GraphQL Gateway

A production-grade, containerized Go microservices stack with gRPC, GraphQL, Prometheus metrics, and Grafana dashboards.

- Account service (PostgreSQL)
- Catalog service (Elasticsearch)
- Order service (PostgreSQL)
- GraphQL gateway (gqlgen) aggregating all services
- Observability: Prometheus + Grafana (per-service dashboards)

---

## Architecture

- Client calls GraphQL → GraphQL calls gRPC services (account, catalog, order)
- Account/Order persist to PostgreSQL; Catalog persists to Elasticsearch
- Each service exposes health/metrics; Prometheus scrapes; Grafana visualizes

Ports (host:container):
- Account: 8081 gRPC, 8082 health/metrics
- Catalog: 8083 gRPC, 8084 health/metrics
- Order: 8085 gRPC, 8086 health/metrics
- GraphQL: 8087 HTTP API, 8088 health/metrics
- Prometheus: 9090, Grafana: 3000

---

### Technology Stack
- **Language**: Go 1.25.0
- **gRPC**: Inter-service communication
- **GraphQL**: API gateway with gqlgen
- **Database**: PostgreSQL (account & order services), Elasticsearch (catalog service)
- **Containerization**: Docker & Docker Compose
- **Protocol Buffers**: Service contracts

---

## Repository structure

```
.
├── account/              # Account microservice (gRPC + Postgres)
│   ├── cmd/account/main.go
│   ├── account.proto
│   ├── server.go / service.go / repository.go
│   ├── up.sql
│   └── app.dockerfile
├── catalog/              # Catalog microservice (gRPC + Elasticsearch)
│   ├── cmd/catalog/main.go
│   ├── catalog.proto
│   ├── server.go / service.go / repository.go
│   └── app.dockerfile
├── order/                # Order microservice (gRPC + Postgres)
│   ├── cmd/order/main.go
│   ├── order.proto
│   ├── server.go / service.go / repository.go
│   └── app.dockerfile
├── graphql/              # GraphQL gateway (gqlgen)
│   ├── main.go
│   ├── schema.graphql
│   └── resolvers
├── monitoring/
│   ├── prometheus.yml
│   └── grafana/
│       ├── datasources/prometheus.yml
│       └── dashboards/*.json     # one per service
├── docker-compose.yml
└── go.mod / go.sum
```

---

## Prerequisites
- Docker + Docker Compose
- Ports 8081–8088, 9090, 3000 available

Optional (local dev):
- Go 1.25+
- protoc + protoc-gen-go + protoc-gen-go-grpc

---

## Quick start (Docker Compose)

```bash
# Clone
git clone https://github.com/master-wayne7/go-microservices.git
cd go-microservices

# Build & run all services + Prometheus + Grafana
docker-compose up --build -d

# Tail logs (example)
docker-compose logs -f graphql
```

Endpoints:
- GraphQL API: http://localhost:8087/graphql
- GraphQL Playground: http://localhost:8087/playground
- Health: http://localhost:8088/health
- Metrics (GraphQL): http://localhost:8088/metrics
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin by default)

---

## Configuration

docker-compose.yml (excerpt):
- Account Postgres: `postgres://postgres:password@account_db:5432/account?sslmode=disable`  
- Order Postgres: `postgres://postgres:password@order_db:5432/order?sslmode=disable`        
- Catalog Elasticsearch: `http://catalog_db:9200`

GraphQL service env:
- ACCOUNT_SERVICE_URL=account:8081
- CATALOG_SERVICE_URL=catalog:8083
- ORDER_SERVICE_URL=order:8085

---

## Local development (without Docker)

Start dependencies (DB/ES) yourself or via docker-compose, then run services:

```bash
cd account && go run cmd/account/main.go
cd catalog && go run cmd/catalog/main.go
cd order && go run cmd/order/main.go
cd graphql && go run main.go
```

---

## Observability

- Each service exposes Prometheus metrics and health:
  - Account: http://localhost:8082/metrics, /health
  - Catalog: http://localhost:8084/metrics, /health
  - Order:   http://localhost:8086/metrics, /health
  - GraphQL: http://localhost:8088/metrics, /health
- Grafana auto-provisions dashboards per service under folders Account, Catalog, Order, GraphQL
- Example metrics (per-service prefix, e.g., account_service_…):
  - http_requests_total, http_request_duration_seconds_bucket
  - grpc_requests_total, grpc_request_duration_seconds_bucket
  - db_queries_total, db_query_duration_seconds_bucket
  - cpu_usage_percent, memory_usage_bytes, goroutines_count, uptime_seconds

---

## GraphQL API Usage
The GraphQL API provides a unified interface to interact with all the microservices.

Playground: http://localhost:8087/playground

Query Accounts
```graphql
query {
  accounts {
    id
    name
  }
}
```

Create an Account
```graphql
mutation {
  createAccount(account: {name: "New Account"}) {
    id
    name
  }
}
```

Query Products
```graphql
query {
  products {
    id
    name
    price
  }
}
```

Create a Product
```graphql
mutation {
  createProduct(product: {name: "New Product", description: "A new product", price: 19.99}) {
    id
    name
    price
  }
}
```

Create an Order
```graphql
mutation {
  createOrder(order: {accountId: "account_id", products: [{id: "product_id", quantity: 2}]}) {
    id
    totalPrice
    products {
      name
      quantity
    }
  }
}
```

Query Account with Orders
```graphql
query {
  accounts(id: "account_id") {
    name
    orders {
      id
      createdAt
      totalPrice
      products {
        name
        quantity
        price
      }
    }
  }
}
```

### Advanced Queries

Pagination and Filtering
```graphql
query {
  products(pagination: {skip: 0, take: 5}, query: "search_term") {
    id
    name
    description
    price
  }
}
```

Calculate Total Spent by an Account
```graphql
query {
  accounts(id: "account_id") {
    name
    orders {
      totalPrice
    }
  }
}
```

---

## Troubleshooting
- Ports in use → free 8081–8088, 9090, 3000 or change mappings in docker-compose.yml  
- Postgres credentials mismatch → update DATABASE_URL in docker-compose.yml       
- Grafana shows no data → ensure Prometheus is scraping /metrics and time range is correct
- Catalog empty → insert products via GraphQL mutation before creating orders

---

## License
This project is licensed under the [MIT License](./LICENSE). See the LICENSE file for details.
