package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MrJJimenez/jobcli/internal/cmd"
	"github.com/MrJJimenez/jobcli/internal/config"
	"github.com/MrJJimenez/jobcli/internal/ui"
	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	cli := cmd.NewCLI()
	applyEnvDefaults(cli)
	versionString := buildVersion()

	parser, err := kong.New(cli,
		kong.Name("jobcli"),
		kong.Description("Job aggregation CLI."),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{"version": versionString},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	kctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		fallbackUI := ui.New(os.Stdout, os.Stderr, ui.NormalizeColorMode(os.Getenv("JOBCLI_COLOR")), false)
		fallbackUI.Errorf("%v", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	configDir, err := config.ConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	colorMode := ui.NormalizeColorMode(cli.Color)
	disableColor := cli.JSON || cli.Plain
	userInterface := ui.New(os.Stdout, os.Stderr, colorMode, disableColor)

	level := zerolog.InfoLevel
	if cli.Verbose {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	runCtx := &cmd.Context{
		Out:        os.Stdout,
		Err:        os.Stderr,
		UI:         userInterface,
		Config:     cfg,
		ConfigDir:  configDir,
		Logger:     logger,
		Verbose:    cli.Verbose,
		JSONOutput: cli.JSON,
		PlainText:  cli.Plain,
		Version:    versionString,
		ColorMode:  colorMode,
	}

	if err := kctx.Run(runCtx); err != nil {
		userInterface.Errorf("%v", err)
		os.Exit(1)
	}
}

func buildVersion() string {
	if commit == "" && date == "" {
		return version
	}
	if commit == "" {
		return fmt.Sprintf("%s (%s)", version, date)
	}
	if date == "" {
		return fmt.Sprintf("%s (%s)", version, commit)
	}
	return fmt.Sprintf("%s (%s, %s)", version, commit, date)
}

func applyEnvDefaults(cli *cmd.CLI) {
	if envBool("JOBCLI_JSON") {
		cli.JSON = true
	}
	if envBool("JOBCLI_VERBOSE") {
		cli.Verbose = true
	}
	if value := os.Getenv("JOBCLI_COLOR"); value != "" {
		cli.Color = value
	}
}

func envBool(key string) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
