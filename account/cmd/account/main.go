package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/master-wayne7/go-microservices/account"
	"github.com/master-wayne7/go-microservices/monitoring"
	"github.com/tinrab/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

// Health check handler for container orchestration
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	// ✅ Initialize metrics only once
	metrics := monitoring.NewMetricsCollector("account-service")
	metrics.SetServiceInfo("1.0.0", "development")

	// start health + metrics server
	go func() {
		mux := http.NewServeMux()
		// Wrap health and metrics with HTTP metrics middleware
		mux.Handle("/health", monitoring.HTTPMiddleware(metrics)(http.HandlerFunc(healthCheck)))
		mux.Handle("/metrics", monitoring.HTTPMiddleware(metrics)(metrics.PrometheusHandler()))
		log.Println("Health + metrics server starting on port 8082...")
		log.Fatal(http.ListenAndServe(":8082", mux))
	}()

	// ✅ Connect to DB with retry
	var r account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = account.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println("DB connection failed, retrying...", err)
		}
		return
	})
	defer r.Close()

	// Wire metrics into repository for DB query metrics
	if pr, ok := r.(*account.PostgresRepository); ok {
		pr.SetMetrics(metrics)
	}

	// ✅ Start system + DB metrics collection
	// Pass DB handle for DB metrics
	metrics.StartSystemMetricsCollection(r.(*account.PostgresRepository).DB())

	// ✅ Start gRPC server with metrics interceptors
	log.Println("Account service gRPC listening on port 8081...")
	s := account.NewService(r)
	log.Fatal(account.ListenGRPC(s, 8081, metrics))
}
