package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const releaseURL = baseURL + "/doc/devel/release"

var versionRegex = regexp.MustCompile(`^go[0-9]+(.[0-9]+)?((rc|beta|\.)[0-9]+)?$`)

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
		if versionRegex.MatchString(file.Name()) {
			versions = append(versions, file.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	fmt.Printf("%s\n", strings.Join(versions, "\n"))
	return nil
}

func (cmd *Command) OutputRemoteVersions(ctx context.Context) error {
	versions, err := cmd.getGoVersions(ctx, true)
	if err != nil {
		return fmt.Errorf("get remote versions: %w", err)
	}

	var buf strings.Builder
	for _, version := range versions {
		buf.WriteString(version.Version + "\n")
	}

	fmt.Printf("%s", &buf)
	return nil
}

func (cmd *Command) InstallSpecifyVersion(ctx context.Context, version string) error {
	target := "golang.org/dl/" + version + "@latest"
	log.Printf("[INFO] run: go install %s", target)
	if _, err := exec.CommandContext(ctx, "go", "install", target).Output(); err != nil {
		return fmt.Errorf("install %s: %w", version, err)
	}

	log.Printf("[INFO] run: go %s download", version)
	if _, err := exec.CommandContext(ctx, version, "download").Output(); err != nil {
		return fmt.Errorf("download %s: %w", version, err)
	}

	log.Printf("[INFO] install success")
	return nil
}
