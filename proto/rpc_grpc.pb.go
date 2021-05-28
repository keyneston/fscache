// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FSCacheClient is the client API for FSCache service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FSCacheClient interface {
	GetFiles(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (FSCache_GetFilesClient, error)
	Shutdown(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type fSCacheClient struct {
	cc grpc.ClientConnInterface
}

func NewFSCacheClient(cc grpc.ClientConnInterface) FSCacheClient {
	return &fSCacheClient{cc}
}

func (c *fSCacheClient) GetFiles(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (FSCache_GetFilesClient, error) {
	stream, err := c.cc.NewStream(ctx, &FSCache_ServiceDesc.Streams[0], "/FSCache/GetFiles", opts...)
	if err != nil {
		return nil, err
	}
	x := &fSCacheGetFilesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type FSCache_GetFilesClient interface {
	Recv() (*Files, error)
	grpc.ClientStream
}

type fSCacheGetFilesClient struct {
	grpc.ClientStream
}

func (x *fSCacheGetFilesClient) Recv() (*Files, error) {
	m := new(Files)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *fSCacheClient) Shutdown(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/FSCache/Shutdown", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FSCacheServer is the server API for FSCache service.
// All implementations must embed UnimplementedFSCacheServer
// for forward compatibility
type FSCacheServer interface {
	GetFiles(*ListRequest, FSCache_GetFilesServer) error
	Shutdown(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedFSCacheServer()
}

// UnimplementedFSCacheServer must be embedded to have forward compatible implementations.
type UnimplementedFSCacheServer struct {
}

func (UnimplementedFSCacheServer) GetFiles(*ListRequest, FSCache_GetFilesServer) error {
	return status.Errorf(codes.Unimplemented, "method GetFiles not implemented")
}
func (UnimplementedFSCacheServer) Shutdown(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Shutdown not implemented")
}
func (UnimplementedFSCacheServer) mustEmbedUnimplementedFSCacheServer() {}

// UnsafeFSCacheServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FSCacheServer will
// result in compilation errors.
type UnsafeFSCacheServer interface {
	mustEmbedUnimplementedFSCacheServer()
}

func RegisterFSCacheServer(s grpc.ServiceRegistrar, srv FSCacheServer) {
	s.RegisterService(&FSCache_ServiceDesc, srv)
}

func _FSCache_GetFiles_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FSCacheServer).GetFiles(m, &fSCacheGetFilesServer{stream})
}

type FSCache_GetFilesServer interface {
	Send(*Files) error
	grpc.ServerStream
}

type fSCacheGetFilesServer struct {
	grpc.ServerStream
}

func (x *fSCacheGetFilesServer) Send(m *Files) error {
	return x.ServerStream.SendMsg(m)
}

func _FSCache_Shutdown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FSCacheServer).Shutdown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/FSCache/Shutdown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FSCacheServer).Shutdown(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// FSCache_ServiceDesc is the grpc.ServiceDesc for FSCache service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FSCache_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "FSCache",
	HandlerType: (*FSCacheServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Shutdown",
			Handler:    _FSCache_Shutdown_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetFiles",
			Handler:       _FSCache_GetFiles_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/rpc.proto",
}
