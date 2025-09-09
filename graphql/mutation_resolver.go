package main

import (
	"context"
	"errors"
	"github.com/master-wayne7/go-microservices/order"
	"log"
	"time"
)

var (
	ErrInvalidParameter = errors.New("invalid paramter")
)

type mutationResolver struct {
	server *Server
}

// CreateAccount implements MutationResolver.
func (r *mutationResolver) CreateAccount(ctx context.Context, account *AccountInput) (*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	a, err := r.server.accountClient.PostAccount(ctx, account.Name)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Account{
		ID:   a.ID,
		Name: a.Name,
	}, nil
}

// CreateOrder implements MutationResolver.
func (r *mutationResolver) CreateOrder(ctx context.Context, in *OrderInput) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var products []order.OrderedProduct
	for _, p := range in.Products {
		if *p.Quantity <= 0 {
			return nil, ErrInvalidParameter
		}
		products = append(products, order.OrderedProduct{
			ID:       p.ID,
			Quantity: uint32(*p.Quantity),
		})
	}

	o, err := r.server.orderClient.PostOrder(ctx, in.AccountID, products)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Convert order.OrderedProduct to []*OrderedProducts
	var newProducts []*OrderedProducts
	for _, p := range o.Products {
		newProducts = append(newProducts, &OrderedProducts{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    int(p.Quantity),
		})
	}

	return &Order{
		ID:         o.ID,
		CreatedAt:  o.CreatedAt,
		TotalPrice: o.TotalPrice,
		Products:   newProducts,
	}, nil
}

// CreateProduct implements MutationResolver.
func (r *mutationResolver) CreateProduct(ctx context.Context, product *ProductInput) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	p, err := r.server.catalogClient.PostProduct(ctx, product.Name, product.Description, product.Price)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       product.Price,
	}, nil
}
