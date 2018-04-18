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

// DynMsgHelper is a helper for setting and getting values from a *dynamic.Message struct
// It supports getter and setter customizations to support custom data types.
type DynMsgHelper struct {
	opts dmhOptions
}

// Creates a new DynMsgHelper
func NewDynMsgHelper(opts ...DMHOption) *DynMsgHelper {
	ret := &DynMsgHelper{}
	for _, o := range opts {
		o(&ret.opts)
	}
	return ret
}

// Sets a parameter value into the message. The name can have "." to set values inside another messages, like
// address.street_name.
func (h *DynMsgHelper) SetParamValue(msg *dynamic.Message, name, value string) error {
	fields := strings.SplitN(name, ".", 2)
	if len(fields) == 0 {
		return fmt.Errorf("Invoke field name must have at least 1 value, have %d", len(fields))
	}
	if len(fields) == 1 {
		fields = append(fields, "")
	}

	fld := msg.FindFieldDescriptorByName(fields[0])
	if fld == nil {
		return fmt.Errorf("Could not find field '%s'", fields[0])
	}

	return h.internalSetParamValue(&setParamSetter_Default{msg: msg, fld: fld}, fld, fields[0], fields[1], value)
}

// Helper for param setter
type setParamSetter interface {
	GetMsg() *dynamic.Message
	GetValue() (ok bool, val interface{})
	SetValue(val interface{}) error
	IsRepeated() bool
}

// Default
type setParamSetter_Default struct {
	msg *dynamic.Message
	fld *desc.FieldDescriptor
}

func (s *setParamSetter_Default) GetMsg() *dynamic.Message {
	return s.msg
}

func (s *setParamSetter_Default) IsRepeated() bool {
	return false
}

func (s *setParamSetter_Default) GetValue() (ok bool, val interface{}) {
	if !s.msg.HasField(s.fld) {
		return false, nil
	}
	return true, s.msg.GetField(s.fld)
}

func (s *setParamSetter_Default) SetValue(val interface{}) error {
	return s.msg.TrySetField(s.fld, val)
}

// Map
type setParamSetter_Map struct {
	msg *dynamic.Message
	fld *desc.FieldDescriptor
	key interface{}
}

func (s *setParamSetter_Map) GetMsg() *dynamic.Message {
	return s.msg
}

func (s *setParamSetter_Map) IsRepeated() bool {
	return false
}

func (s *setParamSetter_Map) GetValue() (ok bool, val interface{}) {
	if !s.msg.HasField(s.fld) {
		return false, nil
	}
	if !s.msg.HasMapField(s.fld, s.key) {
		return false, nil
	}
	return true, s.msg.GetMapField(s.fld, s.key)
}

func (s *setParamSetter_Map) SetValue(val interface{}) error {
	if !s.msg.HasField(s.fld) {
		err := s.msg.TrySetField(s.fld, make(map[interface{}]interface{}))
		if err != nil {
			return err
		}
	}
	return s.msg.TryPutMapField(s.fld, s.key, val)
}

// Repeated
type setParamSetter_Repeated struct {
	msg *dynamic.Message
	fld *desc.FieldDescriptor
	key int
}

func (s *setParamSetter_Repeated) GetMsg() *dynamic.Message {
	return s.msg
}

func (s *setParamSetter_Repeated) IsRepeated() bool {
	return true
}

func (s *setParamSetter_Repeated) GetValue() (ok bool, val interface{}) {
	if !s.msg.HasField(s.fld) {
		return false, nil
	}
	if s.key >= s.msg.FieldLength(s.fld) {
		return false, nil
	}
	return true, s.msg.GetRepeatedField(s.fld, s.key)
}

func (s *setParamSetter_Repeated) SetValue(val interface{}) error {
	// HACK: SetField with a 0-length array don't create the field, the first call must have at least one value
	if !s.msg.HasField(s.fld) {
		if s.key != 0 {
			return fmt.Errorf("The first repeated field key must be 0")
		}
		return s.msg.TrySetField(s.fld, []interface{}{val})
	} else {
		if s.msg.FieldLength(s.fld) == s.key-1 {
			// if same as the last one, set its value
			return s.msg.TrySetRepeatedField(s.fld, s.key, val)
		} else if s.msg.FieldLength(s.fld) == s.key {
			// if one more that last one, add field
			return s.msg.TryAddRepeatedField(s.fld, val)
		} else {
			return fmt.Errorf("Invalid index %d for repeated field, repeated fields must be set in order")
		}
	}
}

func (h *DynMsgHelper) internalSetParamValue(setter setParamSetter, fld *desc.FieldDescriptor, fldname, restname, value string) error {
	if len(restname) == 0 {
		pval, err := h.ParseFieldParamValue(fld, value)
		if err != nil {
			return err
		}
		err = setter.SetValue(pval)
		if err != nil {
			return err
		}
	} else {
		if fld.IsMap() {
			mfields := strings.SplitN(restname, ".", 2)
			if len(mfields) == 0 {
				return fmt.Errorf("Invoke map field name must have at least 1 value, have %d", len(mfields))
			}

			keyvalue, err := h.MustParseScalarFieldValue(fld.GetMapKeyType(), mfields[0])
			if err != nil {
				return err
			}

			return h.internalSetParamValue(&setParamSetter_Map{msg: setter.GetMsg(), fld: fld, key: keyvalue}, fld.GetMapValueType(), fldname, mfields[1], value)
		} else if fld.IsRepeated() && !setter.IsRepeated() {
			rfields := strings.SplitN(restname, ".", 2)
			if len(rfields) == 0 {
				return fmt.Errorf("Invoke repeated field name must have at least 1 value, have %d", len(rfields))
			}

			keyvalue, err := strconv.ParseInt(rfields[0], 10, 32)
			if err != nil {
				return fmt.Errorf("Repeated key must be an integer")
			}

			return h.internalSetParamValue(&setParamSetter_Repeated{msg: setter.GetMsg(), fld: fld, key: int(keyvalue)}, fld, fldname, rfields[1], value)
		} else {
			// Iterate into fields using the rest of the name
			switch fld.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				var inner_msg *dynamic.Message
				// allows setting more values on the same message, by getting the previous value if available
				if has, lastval := setter.GetValue(); has {
					inner_msg = lastval.(*dynamic.Message)
				} else {
					inner_msg = dynamic.NewMessage(fld.GetMessageType())
					err := setter.SetValue(inner_msg)
					if err != nil {
						return err
					}
				}

				err := h.SetParamValue(inner_msg, restname, value)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("Cannot interate fields of %s type %s", fldname, fld.GetType().String())
			}
		}
	}

	return nil
}

func (h *DynMsgHelper) MustParseScalarFieldValue(fld *desc.FieldDescriptor, value string) (retval interface{}, err error) {
	var supported bool
	supported, retval, err = h.ParseScalarFieldValue(fld, value)
	if err != nil {
		return nil, err
	}
	if !supported {
		return nil, fmt.Errorf("Value type must be scalar")
	}
	return retval, nil
}

func (h *DynMsgHelper) ParseScalarFieldValue(fld *desc.FieldDescriptor, value string) (supported bool, retval interface{}, err error) {
	// parse value
	switch fld.GetType() {
	// STRING
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return true, value, nil
		// INT32
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		ivalue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return true, nil, err
		}
		return true, int32(ivalue), nil
		// INT64
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_SINT64:
		ivalue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return true, nil, err
		}
		return true, int64(ivalue), nil
		// UINT32
	case descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_UINT32:
		ivalue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return true, nil, err
		}
		return true, uint32(ivalue), nil
		// UINT64
	case descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_UINT64:
		ivalue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return true, nil, err
		}
		return true, uint64(ivalue), nil
		// FLOAT32
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		ivalue, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return true, nil, err
		}
		return true, float32(ivalue), nil
		// FLOAT64
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		ivalue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return true, nil, err
		}
		return true, float64(ivalue), nil
		// BOOL
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		ivalue, err := strconv.ParseBool(value)
		if err != nil {
			return true, nil, err
		}
		return true, ivalue, nil
	}
	return false, nil, nil
}

// Sets the value of a field on the message.
// It supports DynMsgHelperFieldSetter for types that are not scalar.
func (h *DynMsgHelper) ParseFieldParamValue(fld *desc.FieldDescriptor, value string) (interface{}, error) {
	supported, parseval, err := h.ParseScalarFieldValue(fld, value)
	if err != nil {
		return nil, err
	}
	if !supported {
		// try the setters
		for _, parser := range h.opts.fieldValueParsers {
			ok, retval, err := parser.ParseFieldValue(fld, value)
			if err != nil {
				return nil, err
			}
			if ok {
				return retval, nil
			}
		}

		return nil, fmt.Errorf("Cannot set value of type %s as string", fld.GetType().String())
	}

	return parseval, nil
}

// Gets the value of a field using a DynMsgHelperFieldValueGetter
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
type DynMsgHelperFieldValueParser interface {
	ParseFieldValue(fld *desc.FieldDescriptor, value string) (ok bool, retval interface{}, err error)
}

// Getter
type DynMsgHelperFieldValueGetter interface {
	GetFieldValue(msg *dynamic.Message, fld *desc.FieldDescriptor) (ok bool, value string, err error)
}

// DMH options
type dmhOptions struct {
	fieldValueParsers []DynMsgHelperFieldValueParser
	fieldValueGetters []DynMsgHelperFieldValueGetter
}

func WithDMHFieldValueParsers(setters ...DynMsgHelperFieldValueParser) DMHOption {
	return func(o *dmhOptions) {
		o.fieldValueParsers = append(o.fieldValueParsers, setters...)
	}
}

func WithDMHFieldValueGetters(getters ...DynMsgHelperFieldValueGetter) DMHOption {
	return func(o *dmhOptions) {
		o.fieldValueGetters = append(o.fieldValueGetters, getters...)
	}
}
