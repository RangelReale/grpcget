package grpcget_dmh_uuid

import (
	"fmt"

	"github.com/RangelReale/fproto-wrap-std/gowrap/gwproto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

//
// Field parser and getter for "github.com/RangelReale/fproto-wrap-std" UUID type
//
// Usage:
// gget := grpcget.NewGrpcGet_Default(grpcget.WithDMHFieldValueParsers(grpcget_dmh_uuid.NewDMHUuid()),
//		grpcget.WithDMHFieldValueGetters(grpcget_dmh_uuid.NewDMHUuid())))

//
// UUID
//
type DMHUuid struct {
}

func NewDMHUuid() *DMHUuid {
	return &DMHUuid{}
}

func (h *DMHUuid) ParseFieldValue(fld *desc.FieldDescriptor, value string) (ok bool, retval interface{}, err error) {
	if fld.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		if fld.AsFieldDescriptorProto().TypeName != nil && *fld.AsFieldDescriptorProto().TypeName == ".fproto_wrap.UUID" {
			return true, &gwproto.UUID{Value: value}, nil
		}
	}
	return false, nil, nil
}

func (h *DMHUuid) GetFieldValue(msg *dynamic.Message, fld *desc.FieldDescriptor) (ok bool, value string, err error) {
	if fld.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		if fld.AsFieldDescriptorProto().TypeName != nil && *fld.AsFieldDescriptorProto().TypeName == ".fproto_wrap.UUID" {
			fvalue := msg.GetField(fld)
			if fvalue != nil {
				switch xvalue := fvalue.(type) {
				case *dynamic.Message:
					p_uuid := &gwproto.UUID{}
					err := xvalue.ConvertTo(p_uuid)
					if err != nil {
						return false, "", err
					}
					return true, p_uuid.Value, nil
				default:
					return false, "", fmt.Errorf("Unknown type for UUID field")
				}
			}
		}
	}
	return false, "", nil
}

//
// NullUuid
//
type DMHNullUuid struct {
}

func NewDMHNullUuid() *DMHNullUuid {
	return &DMHNullUuid{}
}

func (h *DMHNullUuid) ParseFieldValue(fld *desc.FieldDescriptor, value string) (ok bool, retval interface{}, err error) {
	if fld.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		if fld.AsFieldDescriptorProto().TypeName != nil && *fld.AsFieldDescriptorProto().TypeName == ".fproto_wrap.NullUUID" {
			if value == "" {
				return true, &gwproto.NullUUID{Valid: false}, nil
			} else {
				return true, &gwproto.NullUUID{Value: value, Valid: true}, nil
			}
		}
	}
	return false, nil, nil
}

func (h *DMHNullUuid) GetFieldValue(msg *dynamic.Message, fld *desc.FieldDescriptor) (ok bool, value string, err error) {
	if fld.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		if fld.AsFieldDescriptorProto().TypeName != nil && *fld.AsFieldDescriptorProto().TypeName == ".fproto_wrap.NullUUID" {
			fvalue := msg.GetField(fld)
			if fvalue != nil {
				switch xvalue := fvalue.(type) {
				case *dynamic.Message:
					p_NullUuid := &gwproto.NullUUID{}
					err := xvalue.ConvertTo(p_NullUuid)
					if err != nil {
						return false, "", err
					}
					if p_NullUuid.Valid {
						return true, p_NullUuid.Value, nil
					}

					return true, "<null>", nil
				default:
					return false, "", fmt.Errorf("Unknown type for null uuid field")
				}
			}
		}
	}
	return false, "", nil
}
