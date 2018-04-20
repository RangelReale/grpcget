package grpcget_dmh_google

import (
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

//
// Field parser and getter for "google/protobuf/wrappers.proto" types
//

//
// Wrappers
//
type DMHWrappers struct {
}

func NewDMHWrappers() *DMHWrappers {
	return &DMHWrappers{}
}

func (h *DMHWrappers) ParseFieldValue(fld *desc.FieldDescriptor, value string) (ok bool, retval interface{}, err error) {
	if fld.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		if fld.AsFieldDescriptorProto().TypeName != nil && fld.GetMessageType().GetFile().GetName() == "google/protobuf/wrappers.proto" {
			switch *fld.AsFieldDescriptorProto().TypeName {
			case ".google.protobuf.DoubleValue":
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return false, nil, err
				}
				return true, &wrappers.DoubleValue{Value: v}, nil
			case ".google.protobuf.FloatValue":
				v, err := strconv.ParseFloat(value, 32)
				if err != nil {
					return false, nil, err
				}
				return true, &wrappers.FloatValue{Value: float32(v)}, nil
			case ".google.protobuf.Int32Value":
				v, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return false, nil, err
				}
				return true, &wrappers.Int32Value{Value: int32(v)}, nil
			case ".google.protobuf.Int64Value":
				v, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return false, nil, err
				}
				return true, &wrappers.Int64Value{Value: v}, nil
			case ".google.protobuf.UInt32Value":
				v, err := strconv.ParseUint(value, 10, 32)
				if err != nil {
					return false, nil, err
				}
				return true, &wrappers.UInt32Value{Value: uint32(v)}, nil
			case ".google.protobuf.UInt64Value":
				v, err := strconv.ParseUint(value, 10, 64)
				if err != nil {
					return false, nil, err
				}
				return true, &wrappers.UInt64Value{Value: v}, nil
			case ".google.protobuf.StringValue":
				return true, &wrappers.StringValue{Value: value}, nil
			case ".google.protobuf.BoolValue":
				v, err := strconv.ParseBool(value)
				if err != nil {
					return false, nil, err
				}
				return true, &wrappers.BoolValue{Value: v}, nil
			case ".google.protobuf.BytesValue":
				return true, &wrappers.BytesValue{Value: []byte(value)}, nil
			}
		}
	}
	return false, nil, nil
}

func (h *DMHWrappers) GetFieldValue(msg *dynamic.Message, fld *desc.FieldDescriptor) (ok bool, value string, err error) {
	if fld.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		if fld.AsFieldDescriptorProto().TypeName != nil && fld.GetMessageType().GetFile().GetName() == "google/protobuf/wrappers.proto" {
			fvalue := msg.GetField(fld)
			if fvalue != nil {
				var p_type proto.Message

				switch xvalue := fvalue.(type) {
				case *dynamic.Message:
					if xvalue != nil { // can be a pointer to nil
						switch *fld.AsFieldDescriptorProto().TypeName {
						case ".google.protobuf.DoubleValue":
							p_type = &wrappers.DoubleValue{}
						case ".google.protobuf.FloatValue":
							p_type = &wrappers.FloatValue{}
						case ".google.protobuf.Int32Value":
							p_type = &wrappers.Int32Value{}
						case ".google.protobuf.Int64Value":
							p_type = &wrappers.Int64Value{}
						case ".google.protobuf.UInt32Value":
							p_type = &wrappers.UInt32Value{}
						case ".google.protobuf.UInt64Value":
							p_type = &wrappers.UInt64Value{}
						case ".google.protobuf.StringValue":
							p_type = &wrappers.StringValue{}
						case ".google.protobuf.BoolValue":
							p_type = &wrappers.BoolValue{}
						case ".google.protobuf.BytesValue":
							p_type = &wrappers.BytesValue{}
						default:
							return false, "", fmt.Errorf("Unknown wrapper type: %s", *fld.AsFieldDescriptorProto().TypeName)
						}

						if p_type != nil {
							err := xvalue.ConvertTo(p_type)
							if err != nil {
								return false, "", err
							}
							return h.getProtoValue(p_type)
						}
					} else {
						return true, "", nil
					}
				case *wrappers.DoubleValue, *wrappers.FloatValue, *wrappers.BoolValue, *wrappers.BytesValue, *wrappers.Int32Value, *wrappers.Int64Value,
					*wrappers.UInt32Value, *wrappers.UInt64Value, *wrappers.StringValue:
					p_type = xvalue.(proto.Message)
				default:
					return false, "", fmt.Errorf("Unknown type for wrapper field")
				}

				if p_type != nil {
					return h.getProtoValue(p_type)
				} else {
					return false, "", fmt.Errorf("Unknown type for wrapper field")
				}
			}
		}
	}
	return false, "", nil
}

func (h *DMHWrappers) getProtoValue(i interface{}) (bool, string, error) {
	switch x := i.(type) {
	case *wrappers.DoubleValue:
		return true, fmt.Sprintf("%f", x.Value), nil
	case *wrappers.FloatValue:
		return true, fmt.Sprintf("%f", x.Value), nil
	case *wrappers.Int32Value:
		return true, fmt.Sprintf("%d", x.Value), nil
	case *wrappers.Int64Value:
		return true, fmt.Sprintf("%d", x.Value), nil
	case *wrappers.UInt32Value:
		return true, fmt.Sprintf("%d", x.Value), nil
	case *wrappers.UInt64Value:
		return true, fmt.Sprintf("%d", x.Value), nil
	case *wrappers.StringValue:
		return true, fmt.Sprintf("%s", x.Value), nil
	case *wrappers.BoolValue:
		return true, fmt.Sprintf("%t", x.Value), nil
	case *wrappers.BytesValue:
		return true, fmt.Sprintf("%s", x.Value), nil
	default:
		return false, "", fmt.Errorf("Unknown type for wrapper field")
	}
}
