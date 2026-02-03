package cmd

import (
	"io"

	"github.com/MrJJimenez/jobcli/internal/config"
	"github.com/MrJJimenez/jobcli/internal/ui"
	"github.com/rs/zerolog"
)

type Context struct {
	Out        io.Writer
	Err        io.Writer
	UI         *ui.UI
	Config     config.Config
	ConfigDir  string
	Logger     zerolog.Logger
	Verbose    bool
	JSONOutput bool
	PlainText  bool
	Version    string
	ColorMode  ui.ColorMode
}
