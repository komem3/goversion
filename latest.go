package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/google/subcommands"
)

type latestCmd struct {
	*baseCmd
}

// Execute implements subcommands.Command
func (cmd *latestCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	if err := cmd.OutputLatestVersion(ctx); err != nil {
		log.Printf("[ERR] %v", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

// Name implements subcommands.Command
func (*latestCmd) Name() string {
	return "latest"
}

// SetFlags implements subcommands.Command
func (*latestCmd) SetFlags(*flag.FlagSet) {}

// Synopsis implements subcommands.Command
func (*latestCmd) Synopsis() string {
	return "return latest go version"
}

// Usage implements subcommands.Command
func (*latestCmd) Usage() string {
	return "latest"
}

func NewLatestCmd() subcommands.Command {
	return &latestCmd{&baseCmd{&http.Client{}}}
}

func (cmd *latestCmd) OutputLatestVersion(ctx context.Context) error {
	downloadURL, err := cmd.getDownloadURL(ctx)
	if err != nil {
		return fmt.Errorf("get download URL: %w", err)
	}
	log.Printf("[INFO] latest version: %s", versionRegex.FindString(downloadURL))
	return nil
}
