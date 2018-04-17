package grpcget

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
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

func NewGrpcGet_Default(supplier ConnectionSupplier, opts ...GetOption) *GrpcGet {
	nopts := []GetOption{
		WithDefaultOutputs(os.Stdout),
		WithConnectionSupplier(supplier),
	}
	for _, o := range opts {
		nopts = append(nopts, o)
	}
	return NewGrpcGet(nopts...)
}

func (g *GrpcGet) SetOpts(opts ...GetOption) {
	for _, opt := range opts {
		opt(&g.opts)
	}
}

func (g *GrpcGet) checkConnection(ctx context.Context) (*grpc.ClientConn, error) {
	if g.opts.connectionSupplier == nil {
		return nil, errors.New("Must configure ConnectionSupplier to run this method")
	}

	return g.opts.connectionSupplier.GetConnection(ctx)
}

func (g *GrpcGet) checkRefClient(ctx context.Context) (*grpcreflect.Client, error) {
	conn, err := g.checkConnection(ctx)
	if err != nil {
		return nil, err
	}

	return grpcreflect.NewClient(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(conn)), nil
}

func (g *GrpcGet) ListServices() error {
	if g.opts.outputServiceList == nil {
		return errors.New("Must configure OutputServiceList to run this method")
	}

	ctx := context.Background()

	refClient, err := g.checkRefClient(ctx)
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

func (g *GrpcGet) ListService(service string) error {
	if g.opts.outputServiceList == nil {
		return errors.New("Must configure OutputService to run this method")
	}

	ctx := context.Background()

	refClient, err := g.checkRefClient(ctx)
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

func (g *GrpcGet) Describe(symbol string) error {
	if g.opts.outputDescribe == nil {
		return errors.New("Must configure OutputDescribe to run this method")
	}

	ctx := context.Background()

	refClient, err := g.checkRefClient(ctx)
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

type getOptions struct {
	connectionSupplier ConnectionSupplier

	outputServiceList ServiceListOutput
	outputService     ServiceOutput
	outputDescribe    DescribeOutput
}

func WithDefaultOutputs(w io.Writer) GetOption {
	return func(o *getOptions) {
		o.outputServiceList = NewDefaultServiceListOutput(w)
		o.outputService = NewDefaultServiceOutput(w)
		o.outputDescribe = NewDefaultDescribeOutput(w)
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
