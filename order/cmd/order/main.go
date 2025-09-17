package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/master-wayne7/go-microservices/monitoring"
	"github.com/master-wayne7/go-microservices/order"
	"github.com/tinrab/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	AccountURL  string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL  string `envconfig:"CATALOG_SERVICE_URL"`
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

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	if cfg.AccountURL == "" {
		log.Fatal("ACCOUNT_SERVICE_URL environment variable is required")
	}
	if cfg.CatalogURL == "" {
		log.Fatal("CATALOG_SERVICE_URL environment variable is required")
	}

	// ### CHANGE THIS #### - Initialize Prometheus metrics
	metrics := monitoring.NewMetricsCollector("order-service")
	metrics.SetServiceInfo("1.0.0", "development")

	// Start health check server on separate port
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", healthCheck)
		mux.Handle("/metrics", metrics.PrometheusHandler())
		log.Println("Health check server starting on port 8086...")
		log.Fatal(http.ListenAndServe(":8086", mux))
	}()

	var r order.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = order.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer r.Close()

	// ### CHANGE THIS #### - Start system metrics collection
	metrics.StartSystemMetricsCollection(r.(*order.PostgresRepository).DB())

	// Changed port from 8080 to 8085 to avoid conflicts
	log.Println("Listening on port 8085...")
	s := order.NewService(r)
	log.Fatal(order.ListenGRPC(s, cfg.AccountURL, cfg.CatalogURL, 8085, metrics))
}
