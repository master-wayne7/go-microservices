package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/master-wayne7/go-microservices/catalog"
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

	// ### CHANGE THIS #### - Initialize Prometheus metrics
	metrics := monitoring.NewMetricsCollector("catalog-service")
	metrics.SetServiceInfo("1.0.0", "development")

	// Start health check server on separate port
	go func() {
		http.HandleFunc("/health", healthCheck)
		// ### CHANGE THIS #### - Add Prometheus metrics endpoint
		http.Handle("/metrics", metrics.PrometheusHandler())
		log.Println("Health check server starting on port 8084...")
		log.Fatal(http.ListenAndServe(":8084", nil))
	}()

	var r catalog.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = catalog.NewElasticRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer r.Close()

	// ### CHANGE THIS #### - Start system metrics collection
	metrics.StartSystemMetricsCollection(nil)

	// Changed port from 8080 to 8083 to avoid conflicts
	log.Println("Listening on port 8083...")
	s := catalog.NewService(r)
	log.Fatal(catalog.ListenGRPC(s, 8083, metrics))
}
