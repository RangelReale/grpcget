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
	Params []string
}

func NewParameterInvokeParamSetter(params ...string) *ParameterInvokeParamSetter {
	return &ParameterInvokeParamSetter{
		Params: params,
	}
}

func (i *ParameterInvokeParamSetter) SetInvokeParam(dmh *DynMsgHelper, req *dynamic.Message) error {
	for _, p := range i.Params {
		argname, argvalue, err := i.parseArgumentParam(p)
		if err != nil {
			return err
		}

		err = dmh.SetParamValue(req, argname, argvalue)
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

func WithInvokeParams(params ...string) InvokeOption {
	return func(o *invokeOptions) {
		o.paramSetters = append(o.paramSetters, NewParameterInvokeParamSetter(params...))
	}
}
