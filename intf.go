package grpcget

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
)

type ConnectionSupplier interface {
	GetConnection(ctx context.Context) (*grpc.ClientConn, error)
}

/*
type GrpcGetIntf interface {
	IsGrpcGet()
}
*/

type ServiceListOutput interface {
	OutputServiceList(services []string) error
}

type ServiceOutput interface {
	OutputService(service *desc.ServiceDescriptor) error
}

type DescribeOutput interface {
	OutputDescribe(descriptor desc.Descriptor) error
}

type InvokeParamSetter interface {
	SetInvokeParam(req *dynamic.Message) error
}

type InvokeOutput interface {
	OutputInvoke(value proto.Message) error
}
