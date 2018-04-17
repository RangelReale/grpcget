package grpcget

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

type DMHOption func(*dmhOptions)

type DynMsgHelper struct {
	opts dmhOptions
}

func NewDynMsgHelper(opts ...DMHOption) *DynMsgHelper {
	ret := &DynMsgHelper{}
	for _, o := range opts {
		o(&ret.opts)
	}
	return ret
}

func (h *DynMsgHelper) SetParamValue(msg *dynamic.Message, name, value string) error {
	fields := strings.SplitN(name, ".", 2)
	if len(fields) == 0 {
		return fmt.Errorf("Invoke field name must have at least 1 value, have %d", len(fields))
	}

	fld := msg.FindFieldDescriptorByName(fields[0])
	if fld == nil {
		return fmt.Errorf("Could not find field %s", fields[0])
	}

	if len(fields) == 1 {
		return h.SetFieldParamValue(msg, fld, value)
	} else {
		// Iterate into fields using the rest of the name
		switch fld.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			inner_msg := dynamic.NewMessage(fld.GetMessageType())
			msg.SetField(fld, inner_msg)
			err := h.SetParamValue(inner_msg, fields[1], value)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Cannot interate fields of %s type %s", name, fld.GetType().String())
		}
	}

	return nil
}

func (h *DynMsgHelper) SetFieldParamValue(msg *dynamic.Message, fld *desc.FieldDescriptor, value string) error {
	// set value
	switch fld.GetType() {
	// STRING
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		msg.SetField(fld, value)
		// INT32
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		ivalue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		msg.SetField(fld, int32(ivalue))
		// INT64
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_SINT64:
		ivalue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		msg.SetField(fld, int64(ivalue))
		// UINT32
	case descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_UINT32:
		ivalue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return err
		}
		msg.SetField(fld, uint32(ivalue))
		// UINT64
	case descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_UINT64:
		ivalue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		msg.SetField(fld, uint64(ivalue))
		// FLOAT32
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		ivalue, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return err
		}
		msg.SetField(fld, float32(ivalue))
		// FLOAT64
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		ivalue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		msg.SetField(fld, float64(ivalue))

		// BOOL
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		ivalue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		msg.SetField(fld, ivalue)
	default:
		// try the setters
		for _, setter := range h.opts.fieldSetters {
			ok, err := setter.SetField(msg, fld, value)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}

		return fmt.Errorf("Cannot set value of type %s as string", fld.GetType().String())
	}

	return nil
}

func (h *DynMsgHelper) GetFieldValue(msg *dynamic.Message, fld *desc.FieldDescriptor) (ok bool, value string, err error) {
	for _, fg := range h.opts.fieldValueGetters {
		ok, value, err = fg.GetFieldValue(msg, fld)
		if err != nil {
			return false, "", err
		}
		if ok {
			return true, value, nil
		}
	}
	return false, "", nil
}

// Setter
type DynMsgHelperFieldSetter interface {
	SetField(msg *dynamic.Message, fld *desc.FieldDescriptor, value string) (ok bool, err error)
}

// Getter
type DynMsgHelperFieldValueGetter interface {
	GetFieldValue(msg *dynamic.Message, fld *desc.FieldDescriptor) (ok bool, value string, err error)
}

// DMH options
type dmhOptions struct {
	fieldSetters      []DynMsgHelperFieldSetter
	fieldValueGetters []DynMsgHelperFieldValueGetter
}

func WithDMHFieldSetters(setters ...DynMsgHelperFieldSetter) DMHOption {
	return func(o *dmhOptions) {
		o.fieldSetters = append(o.fieldSetters, setters...)
	}
}

func WithDMHFieldValueGetters(getters ...DynMsgHelperFieldValueGetter) DMHOption {
	return func(o *dmhOptions) {
		o.fieldValueGetters = append(o.fieldValueGetters, getters...)
	}
}
