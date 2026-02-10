package cmd

import (
	"fmt"

	"github.com/jimezsa/jobcli/internal/seen"
)

type SeenCmd struct {
	Diff   SeenDiffCmd   `cmd:"" help:"Write unseen jobs (A-B) to JSON."`
	Update SeenUpdateCmd `cmd:"" help:"Merge new jobs into seen history JSON."`
}

type SeenDiffCmd struct {
	New   string `name:"new" required:"" help:"Path to new jobs JSON file (A)."`
	Seen  string `name:"seen" required:"" help:"Path to seen jobs JSON file (B). Missing file is treated as empty."`
	Out   string `name:"out" required:"" help:"Output path for unseen jobs JSON file (C)."`
	Stats bool   `name:"stats" help:"Print comparison stats."`
}

type SeenUpdateCmd struct {
	Seen  string `name:"seen" required:"" help:"Path to seen jobs JSON file (B). Missing file is treated as empty."`
	Input string `name:"input" required:"" help:"Path to input jobs JSON file to merge into seen history."`
	Out   string `name:"out" required:"" help:"Output path for updated seen jobs JSON."`
	Stats bool   `name:"stats" help:"Print merge stats."`
}

func (c *SeenDiffCmd) Run(ctx *Context) error {
	newJobs, err := seen.ReadJobs(c.New)
	if err != nil {
		return fmt.Errorf("read --new: %w", err)
	}
	seenJobs, err := seen.ReadJobsAllowMissing(c.Seen)
	if err != nil {
		return fmt.Errorf("read --seen: %w", err)
	}

	unseenJobs, stats := seen.Diff(newJobs, seenJobs)
	if err := seen.WriteJobs(c.Out, unseenJobs); err != nil {
		return fmt.Errorf("write --out: %w", err)
	}

	if c.Stats {
		_, err := fmt.Fprintf(
			ctx.Out,
			"total_new=%d total_seen=%d invalid_skipped=%d unseen_emitted=%d\n",
			stats.TotalNew,
			stats.TotalSeen,
			stats.InvalidSkipped(),
			stats.Unseen,
		)
		return err
	}

	return nil
}

func (c *SeenUpdateCmd) Run(ctx *Context) error {
	seenJobs, err := seen.ReadJobsAllowMissing(c.Seen)
	if err != nil {
		return fmt.Errorf("read --seen: %w", err)
	}
	inputJobs, err := seen.ReadJobs(c.Input)
	if err != nil {
		return fmt.Errorf("read --input: %w", err)
	}

	mergedJobs, stats := seen.Merge(seenJobs, inputJobs)
	if err := seen.WriteJobs(c.Out, mergedJobs); err != nil {
		return fmt.Errorf("write --out: %w", err)
	}

	if c.Stats {
		_, err := fmt.Fprintf(
			ctx.Out,
			"total_seen=%d total_input=%d invalid_skipped=%d added=%d total_out=%d\n",
			stats.TotalSeen,
			stats.TotalInput,
			stats.InvalidSkipped(),
			stats.Added,
			stats.TotalOut,
		)
		return err
	}

	return nil
}
