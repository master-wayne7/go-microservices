package main

import "context"

type mutationResolver struct {
	server *Server
}

// CreateAccount implements MutationResolver.
func (r *mutationResolver) CreateAccount(ctx context.Context, account *AccountInput) (*Account, error) {
	panic("unimplemented")
}

// CreateOrder implements MutationResolver.
func (r *mutationResolver) CreateOrder(ctx context.Context, order *OrderInput) (*Order, error) {
	panic("unimplemented")
}

// CreateProduct implements MutationResolver.
func (r *mutationResolver) CreateProduct(ctx context.Context, product *ProductInput) (*Product, error) {
	panic("unimplemented")
}
