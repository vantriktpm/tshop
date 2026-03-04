// Package grpc provides gRPC server for internal service-to-service (e.g. product-service calls order-service).
package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OrderServiceServer is the gRPC API for order-service (internal).
type OrderServiceServer struct {
	// UnimplementedOrderServiceServer if using protobuf; for scaffold we use empty struct
}

// GetOrderByID can be called by other services via gRPC (mTLS in production).
func (s *OrderServiceServer) GetOrderByID(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error) {
	if req == nil || req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id required")
	}
	// TODO: call usecase GetOrder, return proto response
	return &GetOrderResponse{OrderId: req.OrderId, Status: "pending"}, nil
}

// GetOrderRequest / GetOrderResponse - in production generate from .proto
type GetOrderRequest struct {
	OrderId string
}

type GetOrderResponse struct {
	OrderId string
	Status  string
}

// RegisterOrderGRPC registers the order gRPC server.
func RegisterOrderGRPC(srv *grpc.Server, orderSvc *OrderServiceServer) {
	// pb.RegisterOrderServiceServer(srv, orderSvc)
	_ = orderSvc
}
