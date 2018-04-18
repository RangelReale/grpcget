package grpcget

import (
	"fmt"

	"github.com/jhump/protoreflect/dynamic"
)

//
// ParameterInvokeParamSetter
//
// Params are in the format
// name=value
//
type ParameterInvokeParamSetter struct {
	Params []string
}

func NewParameterInvokeParamSetter(params ...string) *ParameterInvokeParamSetter {
	return &ParameterInvokeParamSetter{
		Params: params,
	}
}

func (i *ParameterInvokeParamSetter) SetInvokeParam(dmh *DynMsgHelper, req *dynamic.Message) error {
	for _, p := range i.Params {
		argname, argvalue, err := ParseArgumentParam(p)
		if err != nil {
			return err
		}

		err = dmh.SetParamValue(req, argname, argvalue)
		if err != nil {
			return fmt.Errorf("Error setting param '%s': %v", argname, err)
		}
	}

	return nil
}

func WithInvokeParams(params ...string) InvokeOption {
	return func(o *invokeOptions) {
		o.paramSetters = append(o.paramSetters, NewParameterInvokeParamSetter(params...))
	}
}
