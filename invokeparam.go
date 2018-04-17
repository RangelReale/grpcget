package grpcget

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/dynamic"
)

//
// ParameterInvokeParamSetter
//
// Params are in the format
// name=value
//
type ParameterInvokeParamSetter struct {
	DynMsgHelper *DynMsgHelper
	Params       []string
}

func NewParameterInvokeParamSetter(params ...string) *ParameterInvokeParamSetter {
	return &ParameterInvokeParamSetter{
		DynMsgHelper: NewDynMsgHelper(),
		Params:       params,
	}
}

func (i *ParameterInvokeParamSetter) SetInvokeParam(req *dynamic.Message) error {
	for _, p := range i.Params {
		argname, argvalue, err := i.parseArgumentParam(p)
		if err != nil {
			return err
		}

		err = i.DynMsgHelper.SetParamValue(req, argname, argvalue)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *ParameterInvokeParamSetter) parseArgumentParam(argument string) (name string, value string, err error) {
	args := strings.Split(argument, "=")
	if len(args) != 2 {
		return "", "", fmt.Errorf("Invoke param must have 2 values, have %d", len(args))
	}

	return args[0], args[1], nil
}

func WithInvokeParams(param ...string) InvokeOption {
	return func(o *invokeOptions) {
		o.paramSetters = append(o.paramSetters, &ParameterInvokeParamSetter{Params: param})
	}
}
