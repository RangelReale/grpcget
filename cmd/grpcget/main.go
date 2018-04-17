package main

import (
	"fmt"
	"os"

	"github.com/RangelReale/grpcget/cmd"
)

func main() {
	cmd := grpcget_cmd.NewCmd()

	// you can set a cmd.GrpcGet if you want to customize the getter

	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
