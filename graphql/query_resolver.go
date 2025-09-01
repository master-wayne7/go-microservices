package main

import "context"

type queryResolver struct {
	server *Server
}

// Accounts implements QueryResolver.
func (q *queryResolver) Accounts(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {
	panic("unimplemented")
}

// Products implements QueryResolver.
func (q *queryResolver) Products(ctx context.Context, pagination *PaginationInput, query *string, id *string) ([]*Product, error) {
	panic("unimplemented")
}
