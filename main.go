package main

import (
	"context"
	"flag"
	"log"
)

var args = struct {
	latest   bool
	upgrade  bool
	ls       bool
	lsRemote bool
	install  string
}{}

func init() {
	flag.BoolVar(&args.latest, "latest", false, "output latest version")
	flag.BoolVar(&args.upgrade, "upgrade", false, "upgrade go version")
	flag.BoolVar(&args.ls, "ls", false, "output local minor versions")
	flag.BoolVar(&args.lsRemote, "ls-remote", false, "output remote minor versions")
	flag.StringVar(&args.install, "install", "", "install minor version")
}

func main() {
	flag.Parse()

	var (
		ctx = context.Background()
		cmd = NewCommand()
		err error
	)
	switch {
	case args.latest:
		err = cmd.OutputLatestVersion(ctx)
	case args.upgrade:
		err = cmd.Upgrade(ctx)
	case args.ls:
		err = cmd.OutputLocalVersions(ctx)
	case args.lsRemote:
		err = cmd.OutputRemoteVersions(ctx)
	case args.install != "":
		err = cmd.InstallSpecifyVersion(ctx, args.install)
	default:
		flag.PrintDefaults()
	}

	if err != nil {
		log.Fatalf("[ERR] %v", err)
	}
}
