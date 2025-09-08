package main

import (
	"context"
	"log"
	"time"
)

type accountResolver struct {
	server *Server
}

// Orders implements AccountResolver.
func (a *accountResolver) Orders(ctx context.Context, obj *Account) ([]*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	orderList, err := a.server.orderClient.GetOrdersForAccount(ctx, obj.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var orders []*Order
	for _, o := range orderList {
		var products []*OrderedProducts
		for _, p := range o.Products {
			products = append(products, &OrderedProducts{
				ID:          p.ID,
				Description: p.Description,
				Name:        p.Name,
				Price:       p.Price,
				Quantity:    int(p.Quantity),
			})
		}
		orders = append(orders, &Order{
			Products:   products,
			ID:         o.ID,
			CreatedAt:  o.CreatedAt,
			TotalPrice: o.TotalPrice,
		})
	}
	return orders, nil
}
