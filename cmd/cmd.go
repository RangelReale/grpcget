package grpcget_cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/RangelReale/grpcget"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"gopkg.in/urfave/cli.v1"
)

type Cmd struct {
	App             *cli.App
	GrpcGet         *grpcget.GrpcGet
	GrpcDialOptions []grpc.DialOption
}

func NewCmd() *Cmd {
	ret := &Cmd{
		App: cli.NewApp(),
	}

	ret.App.Flags = []cli.Flag{
		cli.BoolFlag{Name: "plaintext", Usage: "Use plain-text HTTP/2 when connecting to server (no TLS)."},
		cli.BoolFlag{Name: "insecure", Usage: "Skip server certificate and domain verification. (NOT SECURE!). Not valid with -plaintext option."},
		cli.StringFlag{Name: "cacert", Usage: "File containing trusted root certificates for verifying the server. Ignored if -insecure is specified."},
		cli.StringFlag{Name: "cert", Usage: "File containing client certificate (public key), to present to the server. Not valid with -plaintext option. Must also provide -key option."},
		cli.StringFlag{Name: "key", Usage: "File containing client private key, to present to the server. Not valid with -plaintext option. Must also provide -cert option."},
		cli.StringFlag{Name: "connect-timeout", Usage: "The maximum time, in seconds, to wait for connection to be established. Defaults to 10 seconds."},
		cli.StringFlag{Name: "servername", Usage: "Override servername when validating TLS certificate."},
		cli.Float64Flag{Name: "max-time", Usage: "The maximum total time the operation can take. This is useful for preventing batch jobs that use grpcurl from hanging due to slow or bad network links or due to incorrect stream method usage."},
		cli.Float64Flag{Name: "keepalive-time", Usage: "If present, the maximum idle time in seconds, after which a keepalive probe is sent. If the connection remains idle and no keepalive response is received for this same period then the connection is closed and the operation fails."},
	}

	ret.App.Before = func(ctx *cli.Context) error {
		// Do extra validation on arguments and figure out what user asked us to do.
		if ctx.GlobalIsSet("plaintext") && ctx.GlobalIsSet("insecure") {
			return errors.New("The -plaintext and -insecure arguments are mutually exclusive.")
		}
		if ctx.GlobalIsSet("plaintext") && ctx.GlobalString("cert") != "" {
			return errors.New("The -plaintext and -cert arguments are mutually exclusive.")
		}
		if ctx.GlobalIsSet("plaintext") && ctx.GlobalString("key") != "" {
			return errors.New("The -plaintext and -key arguments are mutually exclusive.")
		}
		if (ctx.GlobalString("key") == "") != (ctx.GlobalString("cert") == "") {
			return errors.New("The -cert and -key arguments must be used together and both be present.")
		}
		return nil
	}

	ret.App.Commands = []cli.Command{
		{
			Name:   "list",
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

func (c *Cmd) getGrpcGet(ctx *cli.Context, target string) (*grpcget.GrpcGet, context.Context, context.CancelFunc, error) {
	var gg = c.GrpcGet
	if gg == nil {
		gg = grpcget.NewGrpcGet(grpcget.WithDefaultOutputs(os.Stdout))
	}

	// timeouts
	callctx := context.Background()
	if ctx.GlobalIsSet("max-time") {
		timeout := time.Duration(ctx.GlobalFloat64("max-time") * float64(time.Second))
		callctx, _ = context.WithTimeout(callctx, timeout)
	}
	dialTime := 10 * time.Second
	if ctx.GlobalIsSet("connect-timeout") {
		dialTime = time.Duration(ctx.GlobalFloat64("connect-timeout") * float64(time.Second))
	}
	callctx, cancel := context.WithTimeout(callctx, dialTime)

	// dial options
	var gdopts []grpc.DialOption
	if ctx.GlobalIsSet("keepalive-time") {
		timeout := time.Duration(ctx.GlobalFloat64("keepalive-time") * float64(time.Second))
		gdopts = append(gdopts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    timeout,
			Timeout: timeout,
		}))
	}
	var creds credentials.TransportCredentials
	if !ctx.GlobalIsSet("plaintext") {
		var err error
		creds, err = ClientTransportCredentials(ctx.GlobalIsSet("insecure"), ctx.GlobalString("cacert"), ctx.GlobalString("cert"), ctx.GlobalString("key"))
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Failed to configure transport credentials: %v", err)
		}
		if ctx.GlobalIsSet("servername") {
			if err := creds.OverrideServerName(ctx.GlobalString("servername")); err != nil {
				return nil, nil, nil, fmt.Errorf("Failed to override server name as %q: %v", ctx.GlobalString("serverName"), err)
			}
		}
	} else {
		gdopts = append(gdopts, grpc.WithInsecure())
	}

	// add extra dial options
	gdopts = append(gdopts, c.GrpcDialOptions...)

	// set grpcget options
	gg.SetOpts(grpcget.WithDefaultConnection(target, gdopts...))

	return gg, callctx, cancel, nil
}

// LIST
func (c *Cmd) CmdList(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return errors.New("First argument must be hostname:port")
	}

	service := ""
	if ctx.NArg() > 1 {
		service = ctx.Args().Get(1)
	}

	gget, callctx, cancel, err := c.getGrpcGet(ctx, ctx.Args().Get(0))
	if err != nil {
		return err
	}
	defer cancel()

	if service != "" {
		err := gget.ListService(callctx, service)
		if err != nil {
			return err
		}
	} else {
		err := gget.ListServices(callctx)
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

	gget, callctx, cancel, err := c.getGrpcGet(ctx, ctx.Args().Get(0))
	if err != nil {
		return err
	}
	defer cancel()

	return gget.Describe(callctx, ctx.Args().Get(1))
}

// INVOKE
func (c *Cmd) CmdInvoke(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return errors.New("First argument must be hostname:port")
	}

	if ctx.NArg() < 2 {
		return errors.New("Second argument must be a method name")
	}

	gget, callctx, cancel, err := c.getGrpcGet(ctx, ctx.Args().Get(0))
	if err != nil {
		return err
	}
	defer cancel()

	var params []string
	for pi := 2; pi < ctx.NArg(); pi++ {
		params = append(params, ctx.Args().Get(pi))
	}

	return gget.Invoke(callctx, ctx.Args().Get(1), grpcget.WithInvokeParams(params...))
}

// ClientTransportCredentials builds transport credentials for a GRPC client using the
// given properties. If cacertFile is blank, only standard trusted certs are used to
// verify the server certs. If clientCertFile is blank, the client will not use a client
// certificate. If clientCertFile is not blank then clientKeyFile must not be blank.
func ClientTransportCredentials(insecureSkipVerify bool, cacertFile, clientCertFile, clientKeyFile string) (credentials.TransportCredentials, error) {
	var tlsConf tls.Config

	if clientCertFile != "" {
		// Load the client certificates from disk
		certificate, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not load client key pair: %v", err)
		}
		tlsConf.Certificates = []tls.Certificate{certificate}
	}

	if insecureSkipVerify {
		tlsConf.InsecureSkipVerify = true
	} else if cacertFile != "" {
		// Create a certificate pool from the certificate authority
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(cacertFile)
		if err != nil {
			return nil, fmt.Errorf("could not read ca certificate: %v", err)
		}

		// Append the certificates from the CA
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to append ca certs")
		}

		tlsConf.RootCAs = certPool
	}

	return credentials.NewTLS(&tlsConf), nil
}
