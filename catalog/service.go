package catalog

import (
	"context"

	"github.com/segmentio/ksuid"
)

type Service interface {
	PostProduct(ctx context.Context, name, description string, price float64) (*Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	GetProducts(ctx context.Context, skip, take uint64) ([]Product, error)
	GetProductsByIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip, take uint64) ([]Product, error)
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

type catalogService struct {
	repository Repository
}

// GetProduct implements Service.
func (c *catalogService) GetProduct(ctx context.Context, id string) (*Product, error) {
	return c.repository.GetProductById(ctx, id)
}

// GetProducts implements Service.
func (c *catalogService) GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return c.repository.ListProducts(ctx, skip, take)
}

// GetProductsByIDs implements Service.
func (c *catalogService) GetProductsByIDs(ctx context.Context, ids []string) ([]Product, error) {
	return c.repository.ListProductsWithIDs(ctx, ids)
}

// PostProduct implements Service.
func (c *catalogService) PostProduct(ctx context.Context, name string, description string, price float64) (*Product, error) {
	p := Product{
		ID:          ksuid.New().String(),
		Name:        name,
		Description: description,
		Price:       price,
	}
	err := c.repository.PutProduct(ctx, p)
	return &p, err
}

// SearchProducts implements Service.
func (c *catalogService) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return c.repository.SearchProducts(ctx, query, skip, take)
}

func NewService(r Repository) Service {
	return &catalogService{repository: r}
}
