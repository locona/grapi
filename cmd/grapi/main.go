package main

import (
	"fmt"
	"os"

	"github.com/izumin5210/grapi/pkg/grapicmd"
	"github.com/izumin5210/grapi/pkg/grapicmd/cmd"
)

var (
	// Name is application name
	Name string
	// Version is application version
	Version string
	// Revision describes current commit hash generated by `git describe --always`.
	Revision string

	inReader  = os.Stdin
	outWriter = os.Stdout
	errWriter = os.Stderr
)

func main() {
	err := cmd.NewGrapiCommand(grapicmd.NewConfig(
		Name,
		Version,
		Revision,
		inReader,
		outWriter,
		errWriter,
	)).Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
