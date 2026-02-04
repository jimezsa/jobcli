package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/muesli/termenv"
)

type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

const LinkColor = "#87CEEB"

type UI struct {
	Out          io.Writer
	Err          io.Writer
	Output       *termenv.Output
	ErrOutput    *termenv.Output
	ColorEnabled bool
}

func New(out io.Writer, err io.Writer, mode ColorMode, disableColor bool) *UI {
	output := termenv.NewOutput(out)
	errOutput := termenv.NewOutput(err)

	colorEnabled := shouldEnableColor(output, mode, disableColor)
	return &UI{
		Out:          out,
		Err:          err,
		Output:       output,
		ErrOutput:    errOutput,
		ColorEnabled: colorEnabled,
	}
}

func shouldEnableColor(output *termenv.Output, mode ColorMode, disableColor bool) bool {
	if disableColor {
		return false
	}

	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}

	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default:
		return output.ColorProfile() != termenv.Ascii
	}
}

func (u *UI) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	msg = strings.TrimRight(msg, "\n")
	if u.ColorEnabled {
		msg = u.ErrOutput.String(msg).Foreground(u.ErrOutput.Color("1")).String()
	}
	fmt.Fprintln(u.Err, msg)
}

func (u *UI) Warnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	msg = strings.TrimRight(msg, "\n")
	if u.ColorEnabled {
		msg = u.ErrOutput.String(msg).Foreground(u.ErrOutput.Color("3")).String()
	}
	fmt.Fprintln(u.Err, msg)
}

func (u *UI) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	msg = strings.TrimRight(msg, "\n")
	if u.ColorEnabled {
		msg = u.Output.String(msg).Foreground(u.Output.Color("4")).String()
	}
	fmt.Fprintln(u.Out, msg)
}

func (u *UI) Successf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	msg = strings.TrimRight(msg, "\n")
	if u.ColorEnabled {
		msg = u.Output.String(msg).Foreground(u.Output.Color("2")).String()
	}
	fmt.Fprintln(u.Out, msg)
}

func ColorizeLink(output *termenv.Output, enabled bool, text string) string {
	if !enabled || output == nil {
		return text
	}
	return output.String(text).Foreground(output.Color(LinkColor)).String()
}

func (u *UI) LinkText(text string) string {
	return ColorizeLink(u.Output, u.ColorEnabled, text)
}

func NormalizeColorMode(value string) ColorMode {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case string(ColorAlways):
		return ColorAlways
	case string(ColorNever):
		return ColorNever
	default:
		return ColorAuto
	}
}
