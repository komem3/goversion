package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(NewUpgradeCmd(), "")
	subcommands.Register(NewLatestCmd(), "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
