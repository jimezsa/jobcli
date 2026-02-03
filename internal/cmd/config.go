package cmd

import (
	"fmt"
	"strings"

	"github.com/MrJJimenez/jobcli/internal/config"
)

type ConfigCmd struct {
	Init InitConfigCmd `cmd:"" help:"Write default config and proxies files."`
	Path PathConfigCmd `cmd:"" help:"Print config directory."`
}

type InitConfigCmd struct{}

type PathConfigCmd struct{}

func (c *InitConfigCmd) Run(ctx *Context) error {
	paths, err := config.Init()
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		ctx.UI.Infof("Config already initialized at %s", ctx.ConfigDir)
		return nil
	}
	ctx.UI.Infof("Created: %s", strings.Join(paths, ", "))
	return nil
}

func (c *PathConfigCmd) Run(ctx *Context) error {
	_, err := fmt.Fprintln(ctx.Out, ctx.ConfigDir)
	return err
}
