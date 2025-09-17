package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kelseyhightower/envconfig"
	"github.com/master-wayne7/go-microservices/monitoring"
)

type AppConfig struct {
	AccountUrl string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogUrl string `envconfig:"CATALOG_SERVICE_URL"`
	OrderUrl   string `envconfig:"ORDER_SERVICE_URL"`
}

// ### CHANGE THIS ####
// Health check handler for container orchestration
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func enforceJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") == "" {
			r.Header.Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	// ### CHANGE THIS #### - Initialize Prometheus metrics
	metrics := monitoring.NewMetricsCollector("graphql-service")
	metrics.SetServiceInfo("1.0.0", "development")

	// Start health check server on separate port
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", healthCheck)
		mux.Handle("/metrics", metrics.PrometheusHandler())
		log.Println("Health check server starting on port 8088...")
		log.Fatal(http.ListenAndServe(":8088", mux))
	}()

	s, err := NewGraphQlServer(
		cfg.AccountUrl,
		cfg.CatalogUrl,
		cfg.OrderUrl,
	)
	if err != nil {
		log.Fatal(err)
	}
	graphqlHandler := handler.NewDefaultServer(s.ToExecutableSchema())

	// ### CHANGE THIS #### - Add GraphQL metrics middleware
	http.Handle("/graphql", monitoring.GraphQLMiddleware(metrics)(enforceJSONContentType(graphqlHandler)))
	http.Handle("/playground", playground.Handler("playground", "/graphql"))

	// ### CHANGE THIS #### - Start system metrics collection
	metrics.StartSystemMetricsCollection(nil)

	// Changed port from 8000 to 8087 to avoid conflicts and maintain consistency
	log.Println("GraphQL server starting on port 8087...")
	log.Fatal(http.ListenAndServe(":8087", nil))
}
