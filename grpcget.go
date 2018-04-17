package grpcget

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type GetOption func(*getOptions)

type GrpcGet struct {
	opts getOptions
}

func NewGrpcGet(opts ...GetOption) *GrpcGet {
	ret := &GrpcGet{}
	ret.SetOpts(opts...)
	return ret
}

func NewGrpcGet_Default(opts ...GetOption) *GrpcGet {
	nopts := []GetOption{
		WithDefaultOutputs(os.Stdout),
	}
	for _, o := range opts {
		nopts = append(nopts, o)
	}
	return NewGrpcGet(nopts...)
}

func (g *GrpcGet) SetOpts(opts ...GetOption) *GrpcGet {
	for _, opt := range opts {
		opt(&g.opts)
	}
	return g
}

func (g *GrpcGet) checkConnection(ctx context.Context) (*grpc.ClientConn, error) {
	if g.opts.connectionSupplier == nil {
		return nil, errors.New("Must configure ConnectionSupplier to run this method")
	}

	return g.opts.connectionSupplier.GetConnection(ctx)
}

func (g *GrpcGet) checkRefClient(ctx context.Context) (*grpcreflect.Client, *grpc.ClientConn, error) {
	conn, err := g.checkConnection(ctx)
	if err != nil {
		return nil, nil, err
	}

	return grpcreflect.NewClient(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(conn)), conn, nil
}

func (g *GrpcGet) ListServices(ctx context.Context) error {
	if g.opts.outputServiceList == nil {
		return errors.New("Must configure OutputServiceList to run this method")
	}

	refClient, _, err := g.checkRefClient(ctx)
	if err != nil {
		return err
	}

	services, err := refClient.ListServices()
	if err != nil {
		return err
	}

	err = g.opts.outputServiceList.OutputServiceList(services)
	if err != nil {
		return err
	}

	return nil
}

func (g *GrpcGet) ListService(ctx context.Context, service string) error {
	if g.opts.outputServiceList == nil {
		return errors.New("Must configure OutputService to run this method")
	}

	refClient, _, err := g.checkRefClient(ctx)
	if err != nil {
		return err
	}

	svc, err := refClient.ResolveService(service)
	if err != nil {
		return err
	}

	err = g.opts.outputService.OutputService(svc)
	if err != nil {
		return err
	}

	return nil
}

func (g *GrpcGet) Describe(ctx context.Context, symbol string) error {
	if g.opts.outputDescribe == nil {
		return errors.New("Must configure OutputDescribe to run this method")
	}

	refClient, _, err := g.checkRefClient(ctx)
	if err != nil {
		return err
	}

	file, err := refClient.FileContainingSymbol(symbol)
	if err != nil {
		return err
	}

	d := file.FindSymbol(symbol)
	if d == nil {
		return fmt.Errorf("Symbol %s not found", symbol)
	}

	err = g.opts.outputDescribe.OutputDescribe(d)
	if err == nil {
		return err
	}

	return nil
}

type InvokeOption func(*invokeOptions)

func (g *GrpcGet) Invoke(ctx context.Context, method string, opts ...InvokeOption) error {
	if g.opts.outputInvoke == nil {
		return errors.New("Must configure OutputInvoke to run this method")
	}

	refClient, conn, err := g.checkRefClient(ctx)
	if err != nil {
		return err
	}

	file, err := refClient.FileContainingSymbol(method)
	if err != nil {
		return err
	}

	d := file.FindSymbol(method)
	if d == nil {
		return fmt.Errorf("Method %s not found", method)
	}

	md, ok := d.(*desc.MethodDescriptor)
	if !ok {
		return fmt.Errorf("Symbol %s is not a method", method)
	}

	var iopts invokeOptions
	for _, o := range opts {
		o(&iopts)
	}

	// create dynamic message
	req := dynamic.NewMessage(md.GetInputType())

	// create dyn msg helper
	dmh := NewDynMsgHelper(g.opts.dmhOpts...)

	// set input parameters
	for _, setter := range iopts.paramSetters {
		err = setter.SetInvokeParam(dmh, req)
		if err != nil {
			return err
		}
	}

	// call
	callctx := context.Background()

	stub := grpcdynamic.NewStub(conn)

	var respHeaders metadata.MD
	var respTrailers metadata.MD

	resp, err := stub.InvokeRpc(callctx, md, req, grpc.Trailer(&respTrailers), grpc.Header(&respHeaders))
	if err != nil {
		return err
	}

	// output
	err = g.opts.outputInvoke.OutputInvoke(dmh, resp)
	if err != nil {
		return err
	}

	return nil
}

// Get options
type getOptions struct {
	connectionSupplier ConnectionSupplier

	outputServiceList ServiceListOutput
	outputService     ServiceOutput
	outputDescribe    DescribeOutput
	outputInvoke      InvokeOutput

	dmhOpts []DMHOption
}

func WithDefaultOutputs(w io.Writer) GetOption {
	return func(o *getOptions) {
		o.outputServiceList = NewDefaultServiceListOutput(w)
		o.outputService = NewDefaultServiceOutput(w)
		o.outputDescribe = NewDefaultDescribeOutput(w)
		o.outputInvoke = NewDefaultInvokeOutput(w)
	}
}

func WithConnectionSupplier(supplier ConnectionSupplier) GetOption {
	return func(o *getOptions) {
		o.connectionSupplier = supplier
	}
}

func WithDefaultConnection(target string, opts ...grpc.DialOption) GetOption {
	return func(o *getOptions) {
		o.connectionSupplier = NewDefaultConnectionSupplier(target, opts...)
	}
}

func WithConnection(conn *grpc.ClientConn) GetOption {
	return func(o *getOptions) {
		o.connectionSupplier = NewConnectionConnectionSupplier(conn)
	}
}

func WithOutputServiceList(output ServiceListOutput) GetOption {
	return func(o *getOptions) {
		o.outputServiceList = output
	}
}

func WithOutputService(output ServiceOutput) GetOption {
	return func(o *getOptions) {
		o.outputService = output
	}
}

func WithOutputDescribe(output DescribeOutput) GetOption {
	return func(o *getOptions) {
		o.outputDescribe = output
	}
}

func WithOutputInvoke(output InvokeOutput) GetOption {
	return func(o *getOptions) {
		o.outputInvoke = output
	}
}

func WithDMHOpts(opts ...DMHOption) GetOption {
	return func(o *getOptions) {
		o.dmhOpts = append(o.dmhOpts, opts...)
	}
}

// Invoke options
type invokeOptions struct {
	paramSetters []InvokeParamSetter
}
