// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: publisher.proto

package protocol

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

// PublisherNodeClient is the client API for PublisherNode service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PublisherNodeClient interface {
	Notify(ctx context.Context, in *NotifyRequest, opts ...grpc.CallOption) (*NotifyResponse, error)
}

type publisherNodeClient struct {
	cc grpc.ClientConnInterface
}

func NewPublisherNodeClient(cc grpc.ClientConnInterface) PublisherNodeClient {
	return &publisherNodeClient{cc}
}

func (c *publisherNodeClient) Notify(ctx context.Context, in *NotifyRequest, opts ...grpc.CallOption) (*NotifyResponse, error) {
	out := new(NotifyResponse)
	err := c.cc.Invoke(ctx, "/network.zktoro.PublisherNode/Notify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PublisherNodeServer is the server API for PublisherNode service.
// All implementations must embed UnimplementedPublisherNodeServer
// for forward compatibility
type PublisherNodeServer interface {
	Notify(context.Context, *NotifyRequest) (*NotifyResponse, error)
	mustEmbedUnimplementedPublisherNodeServer()
}

// UnimplementedPublisherNodeServer must be embedded to have forward compatible implementations.
type UnimplementedPublisherNodeServer struct {
}

func (UnimplementedPublisherNodeServer) Notify(context.Context, *NotifyRequest) (*NotifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Notify not implemented")
}
func (UnimplementedPublisherNodeServer) mustEmbedUnimplementedPublisherNodeServer() {}

// UnsafePublisherNodeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PublisherNodeServer will
// result in compilation errors.
type UnsafePublisherNodeServer interface {
	mustEmbedUnimplementedPublisherNodeServer()
}

func RegisterPublisherNodeServer(s grpc.ServiceRegistrar, srv PublisherNodeServer) {
	s.RegisterService(&PublisherNode_ServiceDesc, srv)
}

func _PublisherNode_Notify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NotifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PublisherNodeServer).Notify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/network.zktoro.PublisherNode/Notify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PublisherNodeServer).Notify(ctx, req.(*NotifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PublisherNode_ServiceDesc is the grpc.ServiceDesc for PublisherNode service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PublisherNode_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "network.zktoro.PublisherNode",
	HandlerType: (*PublisherNodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Notify",
			Handler:    _PublisherNode_Notify_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "publisher.proto",
}
