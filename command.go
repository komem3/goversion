package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
)

type GoVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []struct {
		Filename string `json:"filename"`
		OS       string `json:"os"`
		Arch     string `json:"arch"`
		Version  string `json:"version"`
		Sha256   string `json:"sha_256"`
		Size     int    `json:"size"`
		Kind     string `json:"kind"`
	} `json:"files"`
}

type GoVersions []*GoVersion

const (
	baseURL     = "https://go.dev/dl/?mode=json"
	downloadURL = "https://storage.googleapis.com/golang/"
)

var versionRegex = regexp.MustCompile(`go[1-9]\.+[0-9]{1,2}(\.+[0-9]{1,2})?`)

type Command struct {
	client *http.Client
}

func NewCommand() *Command {
	return &Command{&http.Client{}}
}

func (cmd *Command) getGoVersions(ctx context.Context, all bool) (GoVersions, error) {
	u := baseURL
	if all {
		u += "&include=all"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	res, err := cmd.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", u, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status is %d", res.StatusCode)
	}
	defer res.Body.Close()

	var versions GoVersions
	if err := json.NewDecoder(res.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}
	return versions, nil
}

func (g *GoVersion) getDownloadURL() string {
	for _, file := range g.Files {
		if file.Arch == runtime.GOARCH && file.OS == runtime.GOOS {
			return downloadURL + file.Filename
		}
	}
	return ""
}

func (*Command) goEnv(ctx context.Context, env string) string {
	cmd := exec.CommandContext(ctx, "go", "env", env)

	out, err := cmd.Output()
	if err != nil {
		log.Printf("[ERR] go env %s: %v", env, err)
		return ""
	}

	output := string(out)
	if output[len(output)-1] == '\n' {
		output = output[:len(output)-1]
	}
	return output
}
