package grpcget_cmd

import (
	"errors"
	"os"

	"github.com/RangelReale/grpcget"
	"google.golang.org/grpc"
	"gopkg.in/urfave/cli.v1"
)

type Cmd struct {
	App     *cli.App
	GrpcGet *grpcget.GrpcGet
}

func NewCmd() *Cmd {
	ret := &Cmd{
		App: cli.NewApp(),
	}

	ret.App.Commands = []cli.Command{
		{
			Name: "list",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "service"},
			},
			Action: ret.CmdList,
		},
		{
			Name:   "desc",
			Action: ret.CmdDescribe,
		},
		{
			Name:   "invoke",
			Action: ret.CmdInvoke,
		},
	}

	return ret
}

func (c *Cmd) Run() error {
	return c.App.Run(os.Args)
}

func (c *Cmd) getGrpcGet() *grpcget.GrpcGet {
	if c.GrpcGet != nil {
		return c.GrpcGet
	}

	return grpcget.NewGrpcGet(grpcget.WithDefaultOutputs(os.Stdout))
}

// LIST
func (c *Cmd) CmdList(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return errors.New("First argument must be hostname:port")
	}

	gget := c.getGrpcGet().SetOpts(grpcget.WithDefaultConnection(ctx.Args().Get(0), grpc.WithInsecure()))

	if ctx.String("service") != "" {
		err := gget.ListService(ctx.String("service"))
		if err != nil {
			return err
		}
	} else {
		err := gget.ListServices()
		if err != nil {
			return err
		}
	}

	return nil
}

// DESCRIBE
func (c *Cmd) CmdDescribe(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return errors.New("First argument must be hostname:port")
	}

	if ctx.NArg() < 2 {
		return errors.New("Second argument must be type name")
	}

	gget := c.getGrpcGet().SetOpts(grpcget.WithDefaultConnection(ctx.Args().Get(0), grpc.WithInsecure()))

	err := gget.Describe(ctx.Args().Get(1))
	return err
}

func (c *Cmd) CmdInvoke(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return errors.New("First argument must be hostname:port")
	}

	if ctx.NArg() < 2 {
		return errors.New("Second argument must be a method name")
	}

	gget := c.getGrpcGet().SetOpts(grpcget.WithDefaultConnection(ctx.Args().Get(0), grpc.WithInsecure()))

	var params []string
	for pi := 2; pi < ctx.NArg(); pi++ {
		params = append(params, ctx.Args().Get(pi))
	}

	err := gget.Invoke(ctx.Args().Get(1), grpcget.WithInvokeParams(params...))
	if err != nil {
		return err
	}

	return nil
}
