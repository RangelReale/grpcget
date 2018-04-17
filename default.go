package grpcget

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/grpc"
)

//
// ConnectionSupplier - Default
//
type DefaultConnectionSupplier struct {
	Target string
	Opts   []grpc.DialOption
}

func NewDefaultConnectionSupplier(target string, opts ...grpc.DialOption) *DefaultConnectionSupplier {
	return &DefaultConnectionSupplier{
		Target: target,
		Opts:   opts,
	}
}

func (d *DefaultConnectionSupplier) GetConnection(ctx context.Context) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, d.Target, d.Opts...)
}

//
// ConnectionSupplier - Connection
//
type ConnectionConnectionSupplier struct {
	Conn *grpc.ClientConn
}

func NewConnectionConnectionSupplier(conn *grpc.ClientConn) *ConnectionConnectionSupplier {
	return &ConnectionConnectionSupplier{
		Conn: conn,
	}
}

func (d *ConnectionConnectionSupplier) GetConnection(ctx context.Context) (*grpc.ClientConn, error) {
	return d.Conn, nil
}

//
// ServiceListOutput
//
type DefaultServiceListOutput struct {
	Out io.Writer
}

func NewDefaultServiceListOutput(out io.Writer) *DefaultServiceListOutput {
	return &DefaultServiceListOutput{
		Out: out,
	}
}

func (d *DefaultServiceListOutput) OutputServiceList(services []string) error {
	for _, s := range services {
		_, err := fmt.Fprintln(d.Out, s)
		if err != nil {
			return err
		}
	}
	return nil
}

//
// ServiceOutput
//
type DefaultServiceOutput struct {
	Out io.Writer
}

func NewDefaultServiceOutput(out io.Writer) *DefaultServiceOutput {
	return &DefaultServiceOutput{
		Out: out,
	}
}

func (d *DefaultServiceOutput) OutputService(service *desc.ServiceDescriptor) error {
	for _, mt := range service.GetMethods() {
		_, err := fmt.Fprintf(d.Out, "%s.%s(%s) returns (%s)\n", service.GetFullyQualifiedName(), mt.GetName(), mt.GetInputType().GetFullyQualifiedName(), mt.GetOutputType().GetFullyQualifiedName())
		if err != nil {
			return err
		}
	}

	return nil
}

//
// DescribeOutput
//
type DefaultDescribeOutput struct {
	Out io.Writer
}

func NewDefaultDescribeOutput(out io.Writer) *DefaultDescribeOutput {
	return &DefaultDescribeOutput{
		Out: out,
	}
}

func (d *DefaultDescribeOutput) OutputDescribe(descriptor desc.Descriptor) error {
	var err error

	switch sd := descriptor.(type) {
	case *desc.ServiceDescriptor:
		fmt.Fprintf(d.Out, "Service: %s\n", sd.GetFullyQualifiedName())
		err = d.DumpService(1, sd)
		if err != nil {
			return err
		}
	case *desc.MethodDescriptor:
		fmt.Fprintf(d.Out, "Service RPC: %s\n", sd.GetFullyQualifiedName())
	case *desc.EnumDescriptor:
		fmt.Fprintf(d.Out, "Enum: %s\n", sd.GetFullyQualifiedName())
		err = d.DumpEnum(1, sd)
		if err != nil {
			return err
		}
	case *desc.MessageDescriptor:
		fmt.Fprintf(d.Out, "Message: %s\n", sd.GetFullyQualifiedName())
		err = d.DumpMessage(1, sd)
		if err != nil {
			return err
		}
	case *desc.FieldDescriptor:
		fmt.Fprintf(d.Out, "Field: %s\n", sd.GetFullyQualifiedName())
		err = d.DumpField(1, sd)
		if err != nil {
			return err
		}
	default:
		fmt.Fprintf(d.Out, "Unknown: %s\n", sd.GetFullyQualifiedName())
	}

	return nil
}

func (d *DefaultDescribeOutput) DumpEnum(level int, enum *desc.EnumDescriptor) error {
	levelStr := strings.Repeat("\t", level)

	for _, ev := range enum.GetValues() {
		fmt.Fprintf(d.Out, "%s%s = %d\n", levelStr, ev.GetName(), ev.GetNumber())
	}

	return nil
}

func (d *DefaultDescribeOutput) DumpMessage(level int, msg *desc.MessageDescriptor) error {
	for _, fld := range msg.GetFields() {
		err := d.DumpField(level, fld)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DefaultDescribeOutput) DumpService(level int, svc *desc.ServiceDescriptor) error {
	for _, mt := range svc.GetMethods() {
		err := d.DumpMethod(level+1, mt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DefaultDescribeOutput) DumpMethod(level int, mtd *desc.MethodDescriptor) error {
	levelStr := strings.Repeat("\t", level)

	_, err := fmt.Fprintf(d.Out, "%s%s(%s) returns (%s)\n", levelStr, mtd.GetName(), mtd.GetInputType().GetFullyQualifiedName(), mtd.GetOutputType().GetFullyQualifiedName())
	if err != nil {
		return err
	}

	return nil
}

func (d *DefaultDescribeOutput) DumpField(level int, fld *desc.FieldDescriptor) error {
	levelStr := strings.Repeat("\t", level)
	tn := ""
	if fld.AsFieldDescriptorProto().TypeName != nil {
		tn = fmt.Sprintf(" [%s]", *fld.AsFieldDescriptorProto().TypeName)
	}
	opt := ""
	if fld.IsRepeated() && !fld.IsMap() {
		opt += "[]"
	}
	if fld.IsRequired() {
		opt += "*"
	}

	tp := fld.GetType().String()
	if fld.IsMap() {
		tp = fmt.Sprintf("map[%s]%s", fld.GetMapKeyType().GetType().String(), fld.GetType().String())
	}

	fn := fld.GetName()

	if fld.GetOneOf() != nil {
		fn = fmt.Sprintf("(oneof %s).%s", fld.GetOneOf().GetName(), fn)
	}

	fmt.Fprintf(d.Out, "%s%s%s: %s%s\n", levelStr, fn, opt, tp, tn)

	switch fld.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		d.DumpMessage(level+1, fld.GetMessageType())
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		d.DumpEnum(level+1, fld.GetEnumType())
	}

	return nil
}