package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/master-wayne7/go-microservices/account"
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

	// Start health check server on separate port
	go func() {
		http.HandleFunc("/health", healthCheck)
		log.Println("Health check server starting on port 8082...")
		log.Fatal(http.ListenAndServe(":8082", nil))
	}()

	var r account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = account.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer r.Close()

	// Changed port from 8080 to 8081 to avoid conflicts
	log.Println("Listening on port 8081...")
	s := account.NewService(r)
	log.Fatal(account.ListenGRPC(s, 8081))
}
