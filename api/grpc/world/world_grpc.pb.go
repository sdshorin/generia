// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: world/world.proto

package world

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	WorldService_CreateWorld_FullMethodName         = "/world.WorldService/CreateWorld"
	WorldService_GetWorld_FullMethodName            = "/world.WorldService/GetWorld"
	WorldService_GetWorlds_FullMethodName           = "/world.WorldService/GetWorlds"
	WorldService_JoinWorld_FullMethodName           = "/world.WorldService/JoinWorld"
	WorldService_UpdateWorldImage_FullMethodName    = "/world.WorldService/UpdateWorldImage"
	WorldService_UpdateWorldParams_FullMethodName   = "/world.WorldService/UpdateWorldParams"
	WorldService_GetGenerationStatus_FullMethodName = "/world.WorldService/GetGenerationStatus"
	WorldService_HealthCheck_FullMethodName         = "/world.WorldService/HealthCheck"
)

// WorldServiceClient is the client API for WorldService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WorldServiceClient interface {
	// Create a new world
	CreateWorld(ctx context.Context, in *CreateWorldRequest, opts ...grpc.CallOption) (*WorldResponse, error)
	// Get world by ID
	GetWorld(ctx context.Context, in *GetWorldRequest, opts ...grpc.CallOption) (*WorldResponse, error)
	// Get all worlds available to the user
	GetWorlds(ctx context.Context, in *GetWorldsRequest, opts ...grpc.CallOption) (*WorldsResponse, error)
	// Join a world (add to user's available worlds)
	JoinWorld(ctx context.Context, in *JoinWorldRequest, opts ...grpc.CallOption) (*JoinWorldResponse, error)
	// Update world image
	UpdateWorldImage(ctx context.Context, in *UpdateWorldImageRequest, opts ...grpc.CallOption) (*UpdateWorldImageResponse, error)
	// Update generated world parameters
	UpdateWorldParams(ctx context.Context, in *UpdateWorldParamsRequest, opts ...grpc.CallOption) (*UpdateWorldParamsResponse, error)
	// Get world generation status
	GetGenerationStatus(ctx context.Context, in *GetGenerationStatusRequest, opts ...grpc.CallOption) (*GetGenerationStatusResponse, error)
	// Health check
	HealthCheck(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error)
}

type worldServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWorldServiceClient(cc grpc.ClientConnInterface) WorldServiceClient {
	return &worldServiceClient{cc}
}

func (c *worldServiceClient) CreateWorld(ctx context.Context, in *CreateWorldRequest, opts ...grpc.CallOption) (*WorldResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WorldResponse)
	err := c.cc.Invoke(ctx, WorldService_CreateWorld_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *worldServiceClient) GetWorld(ctx context.Context, in *GetWorldRequest, opts ...grpc.CallOption) (*WorldResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WorldResponse)
	err := c.cc.Invoke(ctx, WorldService_GetWorld_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *worldServiceClient) GetWorlds(ctx context.Context, in *GetWorldsRequest, opts ...grpc.CallOption) (*WorldsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WorldsResponse)
	err := c.cc.Invoke(ctx, WorldService_GetWorlds_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *worldServiceClient) JoinWorld(ctx context.Context, in *JoinWorldRequest, opts ...grpc.CallOption) (*JoinWorldResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(JoinWorldResponse)
	err := c.cc.Invoke(ctx, WorldService_JoinWorld_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *worldServiceClient) UpdateWorldImage(ctx context.Context, in *UpdateWorldImageRequest, opts ...grpc.CallOption) (*UpdateWorldImageResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateWorldImageResponse)
	err := c.cc.Invoke(ctx, WorldService_UpdateWorldImage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *worldServiceClient) UpdateWorldParams(ctx context.Context, in *UpdateWorldParamsRequest, opts ...grpc.CallOption) (*UpdateWorldParamsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateWorldParamsResponse)
	err := c.cc.Invoke(ctx, WorldService_UpdateWorldParams_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *worldServiceClient) GetGenerationStatus(ctx context.Context, in *GetGenerationStatusRequest, opts ...grpc.CallOption) (*GetGenerationStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetGenerationStatusResponse)
	err := c.cc.Invoke(ctx, WorldService_GetGenerationStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *worldServiceClient) HealthCheck(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HealthCheckResponse)
	err := c.cc.Invoke(ctx, WorldService_HealthCheck_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorldServiceServer is the server API for WorldService service.
// All implementations must embed UnimplementedWorldServiceServer
// for forward compatibility.
type WorldServiceServer interface {
	// Create a new world
	CreateWorld(context.Context, *CreateWorldRequest) (*WorldResponse, error)
	// Get world by ID
	GetWorld(context.Context, *GetWorldRequest) (*WorldResponse, error)
	// Get all worlds available to the user
	GetWorlds(context.Context, *GetWorldsRequest) (*WorldsResponse, error)
	// Join a world (add to user's available worlds)
	JoinWorld(context.Context, *JoinWorldRequest) (*JoinWorldResponse, error)
	// Update world image
	UpdateWorldImage(context.Context, *UpdateWorldImageRequest) (*UpdateWorldImageResponse, error)
	// Update generated world parameters
	UpdateWorldParams(context.Context, *UpdateWorldParamsRequest) (*UpdateWorldParamsResponse, error)
	// Get world generation status
	GetGenerationStatus(context.Context, *GetGenerationStatusRequest) (*GetGenerationStatusResponse, error)
	// Health check
	HealthCheck(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error)
	mustEmbedUnimplementedWorldServiceServer()
}

// UnimplementedWorldServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedWorldServiceServer struct{}

func (UnimplementedWorldServiceServer) CreateWorld(context.Context, *CreateWorldRequest) (*WorldResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWorld not implemented")
}
func (UnimplementedWorldServiceServer) GetWorld(context.Context, *GetWorldRequest) (*WorldResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorld not implemented")
}
func (UnimplementedWorldServiceServer) GetWorlds(context.Context, *GetWorldsRequest) (*WorldsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorlds not implemented")
}
func (UnimplementedWorldServiceServer) JoinWorld(context.Context, *JoinWorldRequest) (*JoinWorldResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method JoinWorld not implemented")
}
func (UnimplementedWorldServiceServer) UpdateWorldImage(context.Context, *UpdateWorldImageRequest) (*UpdateWorldImageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateWorldImage not implemented")
}
func (UnimplementedWorldServiceServer) UpdateWorldParams(context.Context, *UpdateWorldParamsRequest) (*UpdateWorldParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateWorldParams not implemented")
}
func (UnimplementedWorldServiceServer) GetGenerationStatus(context.Context, *GetGenerationStatusRequest) (*GetGenerationStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGenerationStatus not implemented")
}
func (UnimplementedWorldServiceServer) HealthCheck(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HealthCheck not implemented")
}
func (UnimplementedWorldServiceServer) mustEmbedUnimplementedWorldServiceServer() {}
func (UnimplementedWorldServiceServer) testEmbeddedByValue()                      {}

// UnsafeWorldServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorldServiceServer will
// result in compilation errors.
type UnsafeWorldServiceServer interface {
	mustEmbedUnimplementedWorldServiceServer()
}

func RegisterWorldServiceServer(s grpc.ServiceRegistrar, srv WorldServiceServer) {
	// If the following call pancis, it indicates UnimplementedWorldServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&WorldService_ServiceDesc, srv)
}

func _WorldService_CreateWorld_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateWorldRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorldServiceServer).CreateWorld(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorldService_CreateWorld_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorldServiceServer).CreateWorld(ctx, req.(*CreateWorldRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorldService_GetWorld_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorldRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorldServiceServer).GetWorld(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorldService_GetWorld_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorldServiceServer).GetWorld(ctx, req.(*GetWorldRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorldService_GetWorlds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorldsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorldServiceServer).GetWorlds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorldService_GetWorlds_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorldServiceServer).GetWorlds(ctx, req.(*GetWorldsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorldService_JoinWorld_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinWorldRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorldServiceServer).JoinWorld(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorldService_JoinWorld_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorldServiceServer).JoinWorld(ctx, req.(*JoinWorldRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorldService_UpdateWorldImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateWorldImageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorldServiceServer).UpdateWorldImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorldService_UpdateWorldImage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorldServiceServer).UpdateWorldImage(ctx, req.(*UpdateWorldImageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorldService_UpdateWorldParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateWorldParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorldServiceServer).UpdateWorldParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorldService_UpdateWorldParams_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorldServiceServer).UpdateWorldParams(ctx, req.(*UpdateWorldParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorldService_GetGenerationStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGenerationStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorldServiceServer).GetGenerationStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorldService_GetGenerationStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorldServiceServer).GetGenerationStatus(ctx, req.(*GetGenerationStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorldService_HealthCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorldServiceServer).HealthCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorldService_HealthCheck_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorldServiceServer).HealthCheck(ctx, req.(*HealthCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WorldService_ServiceDesc is the grpc.ServiceDesc for WorldService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WorldService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "world.WorldService",
	HandlerType: (*WorldServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateWorld",
			Handler:    _WorldService_CreateWorld_Handler,
		},
		{
			MethodName: "GetWorld",
			Handler:    _WorldService_GetWorld_Handler,
		},
		{
			MethodName: "GetWorlds",
			Handler:    _WorldService_GetWorlds_Handler,
		},
		{
			MethodName: "JoinWorld",
			Handler:    _WorldService_JoinWorld_Handler,
		},
		{
			MethodName: "UpdateWorldImage",
			Handler:    _WorldService_UpdateWorldImage_Handler,
		},
		{
			MethodName: "UpdateWorldParams",
			Handler:    _WorldService_UpdateWorldParams_Handler,
		},
		{
			MethodName: "GetGenerationStatus",
			Handler:    _WorldService_GetGenerationStatus_Handler,
		},
		{
			MethodName: "HealthCheck",
			Handler:    _WorldService_HealthCheck_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "world/world.proto",
}
