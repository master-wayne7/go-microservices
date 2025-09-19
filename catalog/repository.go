package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
	"github.com/master-wayne7/go-microservices/monitoring"
)

var (
	ErrNotFound = errors.New("entity not found")
)

type Repository interface {
	Close()
	// Wire metrics into repository
	SetMetrics(mc *monitoring.MetricsCollector)
	PutProduct(ctx context.Context, p Product) error
	GetProductById(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip, take uint64) ([]Product, error)
}

type elasticRepository struct {
	client  *elasticsearch.Client
	metrics *monitoring.MetricsCollector
}

type esSearchResponse struct {
	Hits struct {
		Hits []struct {
			ID     string          `json:"_id"`
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type productDocument struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

var httpTr = &http.Transport{
	MaxIdleConns: 100,

	// Max idle (keep-alive) connections per host (ES node)
	MaxIdleConnsPerHost: 10,

	// How long idle keep-alive connections stay open before being closed
	IdleConnTimeout: 90 * time.Second,

	// Max number of connections per host (0 = unlimited, good if cluster has few nodes)
	MaxConnsPerHost: 100,

	// Disable compression if CPU is more valuable than bandwidth (default true = gzip)
	DisableCompression: false,

	// Time to wait for a TLS handshake
	TLSHandshakeTimeout: 10 * time.Second,

	// Time to wait for a server's first response headers after request is written
	ResponseHeaderTimeout: 30 * time.Second,

	// Max time to wait for "Expect: 100-continue" response
	ExpectContinueTimeout: 1 * time.Second,

	// Whether to reuse TCP connections across different hosts (ES cluster nodes)
	// Default true, usually keep it
	ForceAttemptHTTP2: true,
}

func NewElasticRepository(url string) (Repository, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{url},
		Transport: httpTr,
	})
	if err != nil {
		return nil, err
	}

	return &elasticRepository{client: client}, nil
}

func (r *elasticRepository) Close() {
	httpTr.CloseIdleConnections()
}

// Implement SetMetrics for repository
func (r *elasticRepository) SetMetrics(mc *monitoring.MetricsCollector) {
	r.metrics = mc
}
func (r *elasticRepository) PutProduct(ctx context.Context, p Product) error {
	start := time.Now()
	product := productDocument{
		Name:        p.Name,
		Price:       p.Price,
		Description: p.Description,
	}
	body, err := json.Marshal(product)
	if err != nil {
		return err
	}

	// Index API request
	req := esapi.IndexRequest{
		Index:      "catalog",
		DocumentID: p.ID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("index", "catalog", time.Since(start))
		}
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("index", "catalog", time.Since(start))
		}
		return errors.New("error indexing document")
	}
	if r.metrics != nil {
		r.metrics.RecordDBQuery("index", "catalog", time.Since(start))
	}
	return nil
}
func (r *elasticRepository) GetProductById(ctx context.Context, id string) (*Product, error) {
	start := time.Now()
	req := esapi.GetRequest{
		Index:      "catalog",
		DocumentID: id,
	}

	res, err := req.Do(ctx, r.client.Transport)
	if err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("get", "catalog", time.Since(start))
		}
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("get", "catalog", time.Since(start))
		}
		return nil, fmt.Errorf("error getting document ID=%s: %s", id, res.String())
	}

	// Read raw response body
	data, err := io.ReadAll(res.Body)
	if err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("get", "catalog", time.Since(start))
		}
		return nil, err
	}
	var p productDocument
	err = json.Unmarshal(data, &p)
	if err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("get", "catalog", time.Since(start))
		}
		return nil, err
	}
	if r.metrics != nil {
		r.metrics.RecordDBQuery("get", "catalog", time.Since(start))
	}
	return &Product{
		ID:          id,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}, err
}
func (r *elasticRepository) ListProducts(ctx context.Context, skip, take uint64) ([]Product, error) {
	start := time.Now()
	query := map[string]interface{}{
		"from": skip,
		"size": take,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	// Encode query
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	// Search request
	req := esapi.SearchRequest{
		Index: []string{"catalog"},
		Body:  &buf,
	}

	res, err := req.Do(ctx, r.client.Transport)
	if err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("search", "catalog", time.Since(start))
		}
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("search", "catalog", time.Since(start))
		}
		return nil, fmt.Errorf("error searching products: %s", res.String())
	}

	// Parse response
	var p map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, err
	}

	// Extract hits
	products := make([]Product, 0)
	hitsMap, ok := p["hits"].(map[string]interface{})
	if !ok {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("search", "catalog", time.Since(start))
		}
		return products, nil // Return empty slice if no hits
	}

	hits, ok := hitsMap["hits"].([]interface{})
	if !ok {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("search", "catalog", time.Since(start))
		}
		return products, nil // Return empty slice if hits is not an array
	}

	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})

		// Get the document ID
		docID := hitMap["_id"].(string)

		src := hit.(map[string]interface{})["_source"]
		srcJSON, _ := json.Marshal(src)

		var product productDocument
		if err := json.Unmarshal(srcJSON, &product); err == nil {
			products = append(products, Product{
				ID:          docID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
			})
		}
	}

	if r.metrics != nil {
		r.metrics.RecordDBQuery("search", "catalog", time.Since(start))
	}
	return products, nil

}
func (r *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	start := time.Now()
	// Build query
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": ids,
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("failed to encode query body: %w", err)
	}

	// Perform search
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("catalog"),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("mget", "catalog", time.Since(start))
		}
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("mget", "catalog", time.Since(start))
		}
		return nil, fmt.Errorf("error searching products: %s", res.String())
	}

	// Decode into typed struct
	var esResp esSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	products := make([]Product, 0, len(esResp.Hits.Hits))

	// Convert ES docs into Product
	for _, hit := range esResp.Hits.Hits {
		var doc productDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			continue // Skip invalid docs instead of failing everything
		}

		products = append(products, Product{
			ID:          hit.ID,
			Name:        doc.Name,
			Description: doc.Description,
			Price:       doc.Price,
		})
	}

	if r.metrics != nil {
		r.metrics.RecordDBQuery("mget", "catalog", time.Since(start))
	}
	return products, nil
}

func (r *elasticRepository) SearchProducts(ctx context.Context, query string, skip, take uint64) ([]Product, error) {
	start := time.Now()
	body := map[string]interface{}{
		"from": skip,
		"size": take,
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query":  "*" + query + "*",
				"fields": []string{"name", "description"},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	req := esapi.SearchRequest{
		Index: []string{"catalog"},
		Body:  &buf,
	}

	res, err := req.Do(ctx, r.client.Transport)
	if err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("search", "catalog", time.Since(start))
		}
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("search", "catalog", time.Since(start))
		}
		return nil, fmt.Errorf("error searching products: %s", res.String())
	}

	// Parse response
	var p map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, err
	}

	products := make([]Product, 0)
	hits := p["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})
		docID := hitMap["_id"].(string)

		src := hitMap["_source"]
		srcJSON, _ := json.Marshal(src)

		var product productDocument
		if err := json.Unmarshal(srcJSON, &product); err == nil {
			products = append(products, Product{
				ID:          docID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
			})
		}
	}

	if r.metrics != nil {
		r.metrics.RecordDBQuery("search", "catalog", time.Since(start))
	}
	return products, nil
}
