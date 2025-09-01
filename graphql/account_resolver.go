package main

import "context"

type accountResolver struct {
	server *Server
}

// Orders implements AccountResolver.
func (a *accountResolver) Orders(ctx context.Context, obj *Account) ([]*Order, error) {
	panic("unimplemented")
}
