package main

import (
	"context"
	"log"
	"time"
)

type queryResolver struct {
	server *Server
}

// Accounts implements QueryResolver.
func (q *queryResolver) Accounts(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if id != nil {
		r, err := q.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return []*Account{{
			ID:   r.ID,
			Name: r.Name,
		}}, nil
	}
	skip, take := uint64(0), uint64(0)
	if pagination != nil {
		skip, take = pagination.bounds()
	}
	accountList, err := q.server.accountClient.GetAccounts(ctx, take, skip)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var accounts []*Account
	for _, a := range accountList {
		account := &Account{
			ID:   a.ID,
			Name: a.Name,
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// Products implements QueryResolver.
func (q *queryResolver) Products(ctx context.Context, pagination *PaginationInput, query *string, id []*string) ([]*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if len(id) == 1 && id[0] != nil {
		r, err := q.server.catalogClient.GetProduct(ctx, *id[0])
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return []*Product{{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Price:       r.Price,
		}}, nil
	}
	skip, take := uint64(0), uint64(0)
	if pagination != nil {
		skip, take = pagination.bounds()
	}

	stringIds := make([]string, len(id))
	for i, s := range id {
		if s != nil {
			stringIds[i] = *s
		}
	}
	queryStr := ""
	if query != nil {
		queryStr = *query
	}
	productsList, err := q.server.catalogClient.GetProducts(ctx, skip, take, queryStr, stringIds)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var products []*Product
	for _, p := range productsList {
		product := &Product{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}
		products = append(products, product)
	}
	return products, nil
}

func (p *PaginationInput) bounds() (uint64, uint64) {
	skipValue := uint64(0)
	takeValue := uint64(0)
	if p.Skip != nil {
		skipValue = uint64(*p.Skip)
	}
	if p.Take != nil {
		takeValue = uint64(*p.Take)
	}
	return skipValue, takeValue
}
