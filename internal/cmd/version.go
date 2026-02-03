package cmd

import "fmt"

type VersionCmd struct{}

func (v *VersionCmd) Run(ctx *Context) error {
	_, err := fmt.Fprintln(ctx.Out, ctx.Version)
	return err
}
