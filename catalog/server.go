package catalog

import (
	"context"
	"fmt"
	"net"

	"github.com/master-wayne7/go-microservices/catalog/pb"
	"github.com/master-wayne7/go-microservices/monitoring"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedCatalogServiceServer
	service Service
}

func ListenGRPC(s Service, port int, metrics *monitoring.MetricsCollector) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	// ### CHANGE THIS #### - Add gRPC interceptors for metrics
	serv := grpc.NewServer(
		grpc.UnaryInterceptor(monitoring.GRPCUnaryServerInterceptor(metrics)),
		grpc.StreamInterceptor(monitoring.GRPCStreamServerInterceptor(metrics)),
	)
	pb.RegisterCatalogServiceServer(serv, &grpcServer{service: s, UnimplementedCatalogServiceServer: pb.UnimplementedCatalogServiceServer{}})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) PostProduct(ctx context.Context, r *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	p, err := s.service.PostProduct(ctx, r.Name, r.Description, r.Price)
	if err != nil {
		return nil, err
	}
	return &pb.PostProductResponse{
		Product: &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Price:       p.Price,
			Description: p.Description,
		},
	}, nil
}

func (s *grpcServer) GetProduct(ctx context.Context, r *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	p, err := s.service.GetProduct(ctx, r.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetProductResponse{
		Product: &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Price:       p.Price,
			Description: p.Description,
		},
	}, nil
}
func (s *grpcServer) GetProducts(ctx context.Context, r *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	var res []Product
	var err error

	if r.Query != "" {
		res, err = s.service.SearchProducts(ctx, r.Query, r.Skip, r.Take)
	} else if len(r.Ids) != 0 {
		res, err = s.service.GetProductsByIDs(ctx, r.Ids)
	} else {
		res, err = s.service.GetProducts(ctx, r.Skip, r.Take)
	}
	if err != nil {
		return nil, err
	}

	products := []*pb.Product{}
	for _, p := range res {
		products = append(products, &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Price:       p.Price,
			Description: p.Description,
		})
	}
	return &pb.GetProductsResponse{
		Products: products,
	}, nil
}
