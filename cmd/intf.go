package grpcget_cmd

type CmdOverride interface {
	OverrideTargetAddress(target string) string
	OverrideServiceName(service string) string
	OverrideDescribeSymbolName(symbol string) string
	OverrideInvokeMethodName(method string) string
}
