package order

import (
	"context"
	"time"
)

type Service interface {
	PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error)
	GetOrdersForAccount(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error)
}

type Order struct {
	ID         string           `json:"id"`
	CreatedAt  time.Time        `json:"created_at"`
	TotalPrice float64          `json:"total_price"`
	AccountID  string           `json:"account_id"`
	Products   []OrderedProduct `json:"products"`
}

type OrderedProduct struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Descriptiobn string  `json:"description"`
	Price        float64 `json:"price"`
	Quantity     uint32  `json:"quantity"`
}

type orderService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &orderService{repository: r}
}

func (s *orderService) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {

}
func (s *orderService) GetOrdersForAccount(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {

}
