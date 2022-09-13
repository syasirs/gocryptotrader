// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: btrpc.proto

package btrpc

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

// BacktesterServiceClient is the client API for BacktesterService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BacktesterServiceClient interface {
	ExecuteStrategyFromFile(ctx context.Context, in *ExecuteStrategyFromFileRequest, opts ...grpc.CallOption) (*ExecuteStrategyResponse, error)
	ExecuteStrategyFromConfig(ctx context.Context, in *ExecuteStrategyFromConfigRequest, opts ...grpc.CallOption) (*ExecuteStrategyResponse, error)
	ListAllRuns(ctx context.Context, in *ListAllRunsRequest, opts ...grpc.CallOption) (*ListAllRunsResponse, error)
	StartRunByID(ctx context.Context, in *StartRunByIDRequest, opts ...grpc.CallOption) (*ExecuteStrategyResponse, error)
	StopRunByID(ctx context.Context, in *StopRunByIDRequest, opts ...grpc.CallOption) (*StopRunByIDResponse, error)
	StopAllRuns(ctx context.Context, in *StopAllRunsRequest, opts ...grpc.CallOption) (*StopAllRunsResponse, error)
}

type backtesterServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBacktesterServiceClient(cc grpc.ClientConnInterface) BacktesterServiceClient {
	return &backtesterServiceClient{cc}
}

func (c *backtesterServiceClient) ExecuteStrategyFromFile(ctx context.Context, in *ExecuteStrategyFromFileRequest, opts ...grpc.CallOption) (*ExecuteStrategyResponse, error) {
	out := new(ExecuteStrategyResponse)
	err := c.cc.Invoke(ctx, "/btrpc.BacktesterService/ExecuteStrategyFromFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *backtesterServiceClient) ExecuteStrategyFromConfig(ctx context.Context, in *ExecuteStrategyFromConfigRequest, opts ...grpc.CallOption) (*ExecuteStrategyResponse, error) {
	out := new(ExecuteStrategyResponse)
	err := c.cc.Invoke(ctx, "/btrpc.BacktesterService/ExecuteStrategyFromConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *backtesterServiceClient) ListAllRuns(ctx context.Context, in *ListAllRunsRequest, opts ...grpc.CallOption) (*ListAllRunsResponse, error) {
	out := new(ListAllRunsResponse)
	err := c.cc.Invoke(ctx, "/btrpc.BacktesterService/ListAllRuns", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *backtesterServiceClient) StartRunByID(ctx context.Context, in *StartRunByIDRequest, opts ...grpc.CallOption) (*ExecuteStrategyResponse, error) {
	out := new(ExecuteStrategyResponse)
	err := c.cc.Invoke(ctx, "/btrpc.BacktesterService/StartRunByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *backtesterServiceClient) StopRunByID(ctx context.Context, in *StopRunByIDRequest, opts ...grpc.CallOption) (*StopRunByIDResponse, error) {
	out := new(StopRunByIDResponse)
	err := c.cc.Invoke(ctx, "/btrpc.BacktesterService/StopRunByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *backtesterServiceClient) StopAllRuns(ctx context.Context, in *StopAllRunsRequest, opts ...grpc.CallOption) (*StopAllRunsResponse, error) {
	out := new(StopAllRunsResponse)
	err := c.cc.Invoke(ctx, "/btrpc.BacktesterService/StopAllRuns", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BacktesterServiceServer is the server API for BacktesterService service.
// All implementations must embed UnimplementedBacktesterServiceServer
// for forward compatibility
type BacktesterServiceServer interface {
	ExecuteStrategyFromFile(context.Context, *ExecuteStrategyFromFileRequest) (*ExecuteStrategyResponse, error)
	ExecuteStrategyFromConfig(context.Context, *ExecuteStrategyFromConfigRequest) (*ExecuteStrategyResponse, error)
	ListAllRuns(context.Context, *ListAllRunsRequest) (*ListAllRunsResponse, error)
	StartRunByID(context.Context, *StartRunByIDRequest) (*ExecuteStrategyResponse, error)
	StopRunByID(context.Context, *StopRunByIDRequest) (*StopRunByIDResponse, error)
	StopAllRuns(context.Context, *StopAllRunsRequest) (*StopAllRunsResponse, error)
	mustEmbedUnimplementedBacktesterServiceServer()
}

// UnimplementedBacktesterServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBacktesterServiceServer struct {
}

func (UnimplementedBacktesterServiceServer) ExecuteStrategyFromFile(context.Context, *ExecuteStrategyFromFileRequest) (*ExecuteStrategyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteStrategyFromFile not implemented")
}
func (UnimplementedBacktesterServiceServer) ExecuteStrategyFromConfig(context.Context, *ExecuteStrategyFromConfigRequest) (*ExecuteStrategyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteStrategyFromConfig not implemented")
}
func (UnimplementedBacktesterServiceServer) ListAllRuns(context.Context, *ListAllRunsRequest) (*ListAllRunsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAllRuns not implemented")
}
func (UnimplementedBacktesterServiceServer) StartRunByID(context.Context, *StartRunByIDRequest) (*ExecuteStrategyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartRunByID not implemented")
}
func (UnimplementedBacktesterServiceServer) StopRunByID(context.Context, *StopRunByIDRequest) (*StopRunByIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopRunByID not implemented")
}
func (UnimplementedBacktesterServiceServer) StopAllRuns(context.Context, *StopAllRunsRequest) (*StopAllRunsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopAllRuns not implemented")
}
func (UnimplementedBacktesterServiceServer) mustEmbedUnimplementedBacktesterServiceServer() {}

// UnsafeBacktesterServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BacktesterServiceServer will
// result in compilation errors.
type UnsafeBacktesterServiceServer interface {
	mustEmbedUnimplementedBacktesterServiceServer()
}

func RegisterBacktesterServiceServer(s grpc.ServiceRegistrar, srv BacktesterServiceServer) {
	s.RegisterService(&BacktesterService_ServiceDesc, srv)
}

func _BacktesterService_ExecuteStrategyFromFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteStrategyFromFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BacktesterServiceServer).ExecuteStrategyFromFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/btrpc.BacktesterService/ExecuteStrategyFromFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BacktesterServiceServer).ExecuteStrategyFromFile(ctx, req.(*ExecuteStrategyFromFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BacktesterService_ExecuteStrategyFromConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteStrategyFromConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BacktesterServiceServer).ExecuteStrategyFromConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/btrpc.BacktesterService/ExecuteStrategyFromConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BacktesterServiceServer).ExecuteStrategyFromConfig(ctx, req.(*ExecuteStrategyFromConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BacktesterService_ListAllRuns_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAllRunsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BacktesterServiceServer).ListAllRuns(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/btrpc.BacktesterService/ListAllRuns",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BacktesterServiceServer).ListAllRuns(ctx, req.(*ListAllRunsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BacktesterService_StartRunByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartRunByIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BacktesterServiceServer).StartRunByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/btrpc.BacktesterService/StartRunByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BacktesterServiceServer).StartRunByID(ctx, req.(*StartRunByIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BacktesterService_StopRunByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopRunByIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BacktesterServiceServer).StopRunByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/btrpc.BacktesterService/StopRunByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BacktesterServiceServer).StopRunByID(ctx, req.(*StopRunByIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BacktesterService_StopAllRuns_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopAllRunsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BacktesterServiceServer).StopAllRuns(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/btrpc.BacktesterService/StopAllRuns",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BacktesterServiceServer).StopAllRuns(ctx, req.(*StopAllRunsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BacktesterService_ServiceDesc is the grpc.ServiceDesc for BacktesterService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BacktesterService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "btrpc.BacktesterService",
	HandlerType: (*BacktesterServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ExecuteStrategyFromFile",
			Handler:    _BacktesterService_ExecuteStrategyFromFile_Handler,
		},
		{
			MethodName: "ExecuteStrategyFromConfig",
			Handler:    _BacktesterService_ExecuteStrategyFromConfig_Handler,
		},
		{
			MethodName: "ListAllRuns",
			Handler:    _BacktesterService_ListAllRuns_Handler,
		},
		{
			MethodName: "StartRunByID",
			Handler:    _BacktesterService_StartRunByID_Handler,
		},
		{
			MethodName: "StopRunByID",
			Handler:    _BacktesterService_StopRunByID_Handler,
		},
		{
			MethodName: "StopAllRuns",
			Handler:    _BacktesterService_StopAllRuns_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "btrpc.proto",
}
