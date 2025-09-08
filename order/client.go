package order

import (
	"context"
	"log"
	"time"

	"github.com/master-wayne7/go-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := pb.NewOrderServiceClient(conn)
	return &Client{conn: conn, service: c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostOrder(ctx context.Context, accountId string, products []OrderedProduct) (*Order, error) {
	protoProducts := []*pb.PostOrderRequest_OrderProduct{}
	for _, p := range products {
		protoProducts = append(protoProducts, &pb.PostOrderRequest_OrderProduct{
			ProductId: p.ID,
			Quantity:  p.Quantity,
		})
	}
	resp, err := c.service.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: accountId,
		Products:  protoProducts,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	newOrder := resp.Order
	newOrderCreateAt := time.Time{}
	newOrderCreateAt.UnmarshalBinary(newOrder.CreatedAt)
	return &Order{
		ID:         newOrder.Id,
		CreatedAt:  newOrderCreateAt,
		TotalPrice: newOrder.TotalPrice,
		AccountID:  newOrder.AccountId,
		Products:   products,
	}, nil

}

func (c *Client) GetOrdersForAccount(ctx context.Context, accountId string) ([]Order, error) {
	r, err := c.service.GetOrdersForAccount(ctx, &pb.GetOrdersForAccountRequest{
		AccountId: accountId,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	orders := []Order{}
	for _, orderProto := range r.Order {
		newOrder := Order{
			ID:         orderProto.Id,
			TotalPrice: orderProto.TotalPrice,
			AccountID:  orderProto.AccountId,
		}
		newOrder.CreatedAt = time.Time{}
		newOrder.CreatedAt.UnmarshalBinary(orderProto.CreatedAt)
		prodcucts := []OrderedProduct{}
		for _, p := range orderProto.Products {
			prodcucts = append(prodcucts, OrderedProduct{
				ID:          p.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
				Quantity:    p.Quantity,
			})
		}
		newOrder.Products = prodcucts
		orders = append(orders, newOrder)
	}
	return orders, nil
}
