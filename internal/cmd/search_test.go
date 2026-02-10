package cmd

import (
	"io"
	"testing"

	"github.com/jimezsa/jobcli/internal/export"
)

func TestResolveFormatWithOutputPathRespectsGlobalFlags(t *testing.T) {
	ctx := &Context{Out: io.Discard, JSONOutput: true}
	got, err := resolveFormat(ctx, SearchOptions{}, "jobs.json")
	if err != nil {
		t.Fatalf("resolveFormat() error = %v", err)
	}
	if got != export.FormatJSON {
		t.Fatalf("resolveFormat() = %q, want %q", got, export.FormatJSON)
	}

	ctx = &Context{Out: io.Discard, PlainText: true}
	got, err = resolveFormat(ctx, SearchOptions{}, "jobs.tsv")
	if err != nil {
		t.Fatalf("resolveFormat() error = %v", err)
	}
	if got != export.FormatTSV {
		t.Fatalf("resolveFormat() = %q, want %q", got, export.FormatTSV)
	}
}
