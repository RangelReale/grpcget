package grpcget

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
)

// Interface to supply a connection to GrpcGet
type ConnectionSupplier interface {
	GetConnection(ctx context.Context) (*grpc.ClientConn, error)
}

// Interface that outputs a service list
type ServiceListOutput interface {
	OutputServiceList(services []string) error
}

// Interface that outputs a single service and its methods
type ServiceOutput interface {
	OutputService(service *desc.ServiceDescriptor) error
}

// Interface that outputs descrption of a symbol
type DescribeOutput interface {
	OutputDescribe(descriptor desc.Descriptor) error
}

// Setter for an invoke parameters
type InvokeParamSetter interface {
	SetInvokeParam(dmh *DynMsgHelper, req *dynamic.Message) error
}

// Interface that outputs the response of an invoke
type InvokeOutput interface {
	OutputInvoke(dmh *DynMsgHelper, value proto.Message) error
}
