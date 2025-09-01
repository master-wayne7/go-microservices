package catalog

import (
	"errors"

	"github.com/elastic/go-elasticsearch"
)

var (
	ErrNotFound = errors.New("entity not found")
)

type Repository interface {
	Close()
	PutProduct()
	GetProductById()
	ListProducts()
	ListProductsWithIDs()
	SearchProducts()
}

type elasticRepository struct {
	client *elasticsearch.Client
}
