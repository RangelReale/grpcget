package grpcget

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"time"
)

//
// Default implementations for all interfaces
//

//
// ConnectionSupplier - Default
//
type DefaultConnectionSupplier struct {
	Ctx    context.Context
	Target string
	Opts   []grpc.DialOption
}

func NewDefaultConnectionSupplier(ctx context.Context, target string, opts ...grpc.DialOption) *DefaultConnectionSupplier {
	return &DefaultConnectionSupplier{
		Ctx:    ctx,
		Target: target,
		Opts:   opts,
	}
}

func (d *DefaultConnectionSupplier) GetConnection(ctx context.Context) (*grpc.ClientConn, error) {
	curctx := d.Ctx
	if curctx == nil {
		curctx = ctx
	}

	_, has_deadline := curctx.Deadline()
	if has_deadline {
		var cancel context.CancelFunc
		curctx, cancel = context.WithCancel(curctx)
		defer cancel()
	}

	return grpc.DialContext(curctx, d.Target, d.Opts...)
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
		_, err := fmt.Fprintf(d.Out, "\t%s(%s) returns (%s)\n", mt.GetName(), mt.GetInputType().GetFullyQualifiedName(), mt.GetOutputType().GetFullyQualifiedName())
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
		err = d.DumpMethod(1, sd, true)
		if err != nil {
			return err
		}
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
		err := d.DumpMethod(level+1, mt, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DefaultDescribeOutput) DumpMethod(level int, mtd *desc.MethodDescriptor, complete bool) error {
	levelStr := strings.Repeat("\t", level)

	_, err := fmt.Fprintf(d.Out, "%s%s(%s) returns (%s)\n", levelStr, mtd.GetName(), mtd.GetInputType().GetName(), mtd.GetOutputType().GetName())
	if err != nil {
		return err
	}

	if complete {
		_, err = fmt.Fprintf(d.Out, "%s\tRequest: %s\n", levelStr, mtd.GetInputType().GetFullyQualifiedName())
		if err != nil {
			return err
		}

		err = d.DumpMessage(level+2, mtd.GetInputType())
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(d.Out, "%s\tResponse: %s\n", levelStr, mtd.GetInputType().GetFullyQualifiedName())
		if err != nil {
			return err
		}

		err = d.DumpMessage(level+2, mtd.GetOutputType())
		if err != nil {
			return err
		}
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

//
// InvokeOutput
//
type DefaultInvokeOutput struct {
	Out io.Writer
}

func NewDefaultInvokeOutput(out io.Writer) *DefaultInvokeOutput {
	return &DefaultInvokeOutput{
		Out: out,
	}
}

func (d *DefaultInvokeOutput) OutputInvoke(dmh *DynMsgHelper, value proto.Message) error {
	return d.DumpMessageCheck(dmh, 0, value)
}

func (d *DefaultInvokeOutput) DumpMessageCheck(dmh *DynMsgHelper, level int, msg interface{}) error {
	levelStr := strings.Repeat("\t", level)

	switch xvalue := msg.(type) {
	case *dynamic.Message:
		return d.DumpMessage(dmh, level, xvalue)
	case *empty.Empty:
		_, err := fmt.Fprintf(d.Out, "%s<empty>\n", levelStr)
		return err
	case *timestamp.Timestamp:
		t, err := ptypes.Timestamp(xvalue)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(d.Out, "%s%s\n", levelStr, t.Format(time.RFC3339))
		return err
	}

	return fmt.Errorf("Unknown message type, cannot output")
}

func (d *DefaultInvokeOutput) DumpMessage(dmh *DynMsgHelper, level int, msg *dynamic.Message) error {
	if msg == nil {
		return nil
	}

	levelStr := strings.Repeat("\t", level)

	for _, fld := range msg.GetKnownFields() {
		var value string

		// check if has getter plugin
		has_getter, getter_value, err := dmh.GetFieldValue(msg, fld)
		if err != nil {
			return err
		}
		if has_getter {
			value = getter_value
		} else {
			switch fld.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				value = msg.GetField(fld).(string)
				// INT32
			case descriptor.FieldDescriptorProto_TYPE_SFIXED32,
				descriptor.FieldDescriptorProto_TYPE_INT32,
				descriptor.FieldDescriptorProto_TYPE_SINT32,
				descriptor.FieldDescriptorProto_TYPE_ENUM:
				value = fmt.Sprintf("%d", msg.GetField(fld).(int32))
				// INT64
			case descriptor.FieldDescriptorProto_TYPE_SFIXED64,
				descriptor.FieldDescriptorProto_TYPE_INT64,
				descriptor.FieldDescriptorProto_TYPE_SINT64:
				value = fmt.Sprintf("%d", msg.GetField(fld).(int64))
				// UINT32
			case descriptor.FieldDescriptorProto_TYPE_FIXED32,
				descriptor.FieldDescriptorProto_TYPE_UINT32:
				value = fmt.Sprintf("%d", msg.GetField(fld).(uint32))
				// UINT64
			case descriptor.FieldDescriptorProto_TYPE_FIXED64,
				descriptor.FieldDescriptorProto_TYPE_UINT64:
				value = fmt.Sprintf("%d", msg.GetField(fld).(uint64))
				// FLOAT32
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				value = fmt.Sprintf("%f", msg.GetField(fld).(float32))
				// FLOAT64
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				value = fmt.Sprintf("%f", msg.GetField(fld).(float64))
				// BOOL
			case descriptor.FieldDescriptorProto_TYPE_BOOL:
				value = fmt.Sprintf("%v", msg.GetField(fld).(bool))
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				value = ""
			default:
				value = "Unknown"
			}
		}

		is_print := true
		if fld.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE && !msg.HasField(fld) {
			is_print = false
		}

		if is_print {
			var opt string
			if fld.IsMap() {
				opt = "[map]"
			} else if fld.IsRepeated() {
				opt = "[]"
			}

			fmt.Fprintf(d.Out, "%s%s%s: %s\n", levelStr, fld.GetName(), opt, value)

			if !has_getter {
				// Dump sub messages
				switch fld.GetType() {
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					if fld.IsMap() {
						// map fields have value of map[interface{}]interface{}
						f_map := msg.GetField(fld).(map[interface{}]interface{})

						for ridx, ritem := range f_map {
							fmt.Fprintf(d.Out, "%s\t- %v\n", levelStr, ridx)
							err := d.DumpMessageCheck(dmh, level+1, ritem)
							if err != nil {
								return err
							}
						}
					} else if fld.IsRepeated() {
						for ridx := 0; ridx < msg.FieldLength(fld); ridx++ {
							fmt.Fprintf(d.Out, "%s\t-\n", levelStr)
							err := d.DumpMessageCheck(dmh, level+1, msg.GetRepeatedField(fld, ridx))
							if err != nil {
								return err
							}
						}
					} else {
						err := d.DumpMessageCheck(dmh, level+1, msg.GetField(fld))
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
