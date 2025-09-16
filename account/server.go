//go:generate protoc --go_out=./pb --go-grpc_out=./pb --go_opt=module=github.com/master-wayne7/go-microservices account.proto
package account

import (
	"context"
	"fmt"
	"net"

	"github.com/master-wayne7/go-microservices/account/pb"
	"github.com/master-wayne7/go-microservices/monitoring"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
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
	pb.RegisterAccountServiceServer(serv, &grpcServer{service: s, UnimplementedAccountServiceServer: pb.UnimplementedAccountServiceServer{}})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) PostAccount(ctx context.Context, r *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	a, err := s.service.PostAccount(ctx, r.Name)
	if err != nil {
		return nil, err
	}

	return &pb.PostAccountResponse{
		Account: &pb.Account{
			Id:   a.ID,
			Name: a.Name,
		},
	}, nil
}

func (s *grpcServer) GetAccount(ctx context.Context, r *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	a, err := s.service.GetAccount(ctx, r.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetAccountResponse{
		Account: &pb.Account{
			Id:   a.ID,
			Name: a.Name,
		},
	}, nil
}

func (s *grpcServer) GetAccounts(ctx context.Context, r *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	a, err := s.service.GetAccounts(ctx, r.Skip, r.Take)
	if err != nil {
		return nil, err
	}

	accounts := []*pb.Account{}
	for _, p := range a {
		accounts = append(accounts, &pb.Account{
			Id:   p.ID,
			Name: p.Name,
		})
	}

	return &pb.GetAccountsResponse{
		Accounts: accounts,
	}, nil
}
