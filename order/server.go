package order

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/master-wayne7/go-microservices/account"
	"github.com/master-wayne7/go-microservices/catalog"
	"github.com/master-wayne7/go-microservices/monitoring"
	"github.com/master-wayne7/go-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service       Service
	accountClient *account.Client
	catalogClient *catalog.Client
}

func ListenGRPC(s Service, accountUrl, catalogUrl string, port int, metrics *monitoring.MetricsCollector) error {
	accountClient, err := account.NewClient(accountUrl)
	if err != nil {
		return err
	}
	catalogClient, err := catalog.NewClient(catalogUrl)
	if err != nil {
		accountClient.Close()
		return err
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return err
	}

	// ### CHANGE THIS #### - Add gRPC interceptors for metrics
	serv := grpc.NewServer(
		grpc.UnaryInterceptor(monitoring.GRPCUnaryServerInterceptor(metrics)),
		grpc.StreamInterceptor(monitoring.GRPCStreamServerInterceptor(metrics)),
	)
	pb.RegisterOrderServiceServer(serv, &grpcServer{
		service:                         s,
		accountClient:                   accountClient,
		catalogClient:                   catalogClient,
		UnimplementedOrderServiceServer: pb.UnimplementedOrderServiceServer{},
	})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) PostOrder(ctx context.Context, r *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	_, err := s.accountClient.GetAccount(ctx, r.AccountId)
	if err != nil {
		log.Println("Error getting account: ", err)
		return nil, err
	}

	productIds := make([]string, 0, len(r.Products))
	for _, p := range r.Products {
		productIds = append(productIds, p.ProductId)
	}
	orderedProducts, err := s.catalogClient.GetProducts(ctx, 0, 0, "", productIds)
	if err != nil {
		log.Println("error getting products: ", err)
		return nil, err
	}
	products := []OrderedProduct{}
	for _, p := range orderedProducts {
		product := OrderedProduct{
			ID:          p.ID,
			Quantity:    0,
			Price:       p.Price,
			Name:        p.Name,
			Description: p.Description,
		}
		for _, rp := range r.Products {
			if rp.ProductId == p.ID {
				product.Quantity = rp.Quantity
				break
			}
		}
		if product.Quantity != 0 {
			products = append(products, product)
		}
	}
	order, err := s.service.PostOrder(ctx, r.AccountId, products)
	if err != nil {
		log.Printf("Failed to post order: %v", err)
		return nil, fmt.Errorf("could not post order: %w", err)
	}
	orderProto := &pb.Order{
		Id:        order.ID,
		AccountId: order.AccountID,
		Products:  []*pb.Order_OrderProduct{},
	}
	orderProto.CreatedAt, err = order.CreatedAt.MarshalBinary()
	if err != nil {
		log.Printf("Failed to marshal CreatedAt: %v", err)
		return nil, fmt.Errorf("could not marshal order timestamp: %w", err)
	}
	for _, op := range order.Products {
		orderProto.Products = append(orderProto.Products, &pb.Order_OrderProduct{
			Id:          op.ID,
			Name:        op.Name,
			Description: op.Description,
			Price:       op.Price,
			Quantity:    op.Quantity,
		})
	}
	return &pb.PostOrderResponse{
		Order: orderProto,
	}, nil
}

func (s *grpcServer) GetOrdersForAccount(
	ctx context.Context,
	r *pb.GetOrdersForAccountRequest,
) (*pb.GetOrdersForAccountResponse, error) {

	accountsOrder, err := s.service.GetOrdersForAccount(ctx, r.AccountId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	productIdMap := map[string]bool{}
	for _, o := range accountsOrder {
		for _, p := range o.Products {
			productIdMap[p.ID] = true
		}
	}
	productIds := []string{}
	for id := range productIdMap {
		productIds = append(productIds, id)
	}

	products, err := s.catalogClient.GetProducts(ctx, 0, 0, "", productIds)
	if err != nil {
		log.Println("error getting account products: ", err)
		return nil, err
	}
	orders := []*pb.Order{}
	for _, o := range accountsOrder {
		op := &pb.Order{
			AccountId: o.AccountID,
			Id:        o.ID,
			Products:  []*pb.Order_OrderProduct{},
		}
		op.CreatedAt, err = o.CreatedAt.MarshalBinary()
		if err != nil {
			log.Printf("Failed to marshal CreatedAt for order %s: %v", o.ID, err)
			return nil, fmt.Errorf("could not marshal order timestamp: %w", err)
		}
		for _, product := range o.Products {
			for _, p := range products {
				if p.ID == product.ID {
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					break
				}
			}
			op.Products = append(op.Products, &pb.Order_OrderProduct{
				Id:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Quantity:    product.Quantity,
			})
		}
		orders = append(orders, op)
	}
	return &pb.GetOrdersForAccountResponse{
		Order: orders,
	}, nil
}
