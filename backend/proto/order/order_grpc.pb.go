// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// OrderClient is the client API for Order service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OrderClient interface {
	//订单
	CreateOrder(ctx context.Context, in *OrderRequest, opts ...grpc.CallOption) (*OrderInfoResponse, error)
}

type orderClient struct {
	cc grpc.ClientConnInterface
}

func NewOrderClient(cc grpc.ClientConnInterface) OrderClient {
	return &orderClient{cc}
}

func (c *orderClient) CreateOrder(ctx context.Context, in *OrderRequest, opts ...grpc.CallOption) (*OrderInfoResponse, error) {
	out := new(OrderInfoResponse)
	err := c.cc.Invoke(ctx, "/Order/CreateOrder", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrderServer is the server API for Order service.
// All implementations must embed UnimplementedOrderServer
// for forward compatibility
type OrderServer interface {
	//订单
	CreateOrder(context.Context, *OrderRequest) (*OrderInfoResponse, error)
	mustEmbedUnimplementedOrderServer()
}

// UnimplementedOrderServer must be embedded to have forward compatible implementations.
type UnimplementedOrderServer struct {
}

func (UnimplementedOrderServer) CreateOrder(context.Context, *OrderRequest) (*OrderInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOrder not implemented")
}
func (UnimplementedOrderServer) mustEmbedUnimplementedOrderServer() {}

// UnsafeOrderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OrderServer will
// result in compilation errors.
type UnsafeOrderServer interface {
	mustEmbedUnimplementedOrderServer()
}

func RegisterOrderServer(s grpc.ServiceRegistrar, srv OrderServer) {
	s.RegisterService(&Order_ServiceDesc, srv)
}

func _Order_CreateOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServer).CreateOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Order/CreateOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServer).CreateOrder(ctx, req.(*OrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Order_ServiceDesc is the grpc.ServiceDesc for Order service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Order_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Order",
	HandlerType: (*OrderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateOrder",
			Handler:    _Order_CreateOrder_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "order/order.proto",
}
