package catalog

import (
	"context"

	"github.com/master-wayne7/go-microservices/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.CatalogServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := pb.NewCatalogServiceClient(conn)
	return &Client{conn: conn, service: c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostProduct(ctx context.Context, name, description string, price float64) (*Product, error) {
	r, err := c.service.PostProduct(
		ctx,
		&pb.PostProductRequest{
			Name:        name,
			Description: description,
			Price:       price,
		},
	)
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          r.Product.Id,
		Name:        r.Product.Name,
		Price:       r.Product.Price,
		Description: r.Product.Description,
	}, nil
}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	r, err := c.service.GetProduct(
		ctx,
		&pb.GetProductRequest{
			Id: id,
		},
	)
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          r.Product.Id,
		Name:        r.Product.Name,
		Price:       r.Product.Price,
		Description: r.Product.Description,
	}, nil
}

func (c *Client) GetProducts(ctx context.Context, skip, take uint64, query string, ids []string) ([]Product, error) {
	r, err := c.service.GetProducts(
		ctx,
		&pb.GetProductsRequest{
			Take:  take,
			Skip:  skip,
			Query: query,
			Ids:   ids,
		},
	)
	if err != nil {
		return nil, err
	}

	products := make([]Product, 0)
	for _, p := range r.Products {
		products = append(products, Product{
			ID:          p.Id,
			Name:        p.Name,
			Price:       p.Price,
			Description: p.Description,
		})
	}
	return products, nil
}
