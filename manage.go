package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const releaseURL = baseURL + "/doc/devel/release"

var minorVersionRegex = regexp.MustCompile("^go[0-9]+(.[0-9]+)?$")

func (cmd *Command) OutputLocalVersions(ctx context.Context) error {
	gopath := cmd.goEnv(ctx, "GOPATH")
	if gopath == "" {
		return nil
	}

	files, err := os.ReadDir(filepath.Join(gopath, "bin"))
	if err != nil {
		return fmt.Errorf("read %s: %w", filepath.Join(gopath, "bin"), err)
	}

	var versions []string
	for _, file := range files {
		if minorVersionRegex.MatchString(file.Name()) {
			versions = append(versions, file.Name())
		}
	}

	fmt.Printf("local versions\n%s\n", strings.Join(versions, "\n"))
	return nil
}
