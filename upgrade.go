package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

const (
	baseURL       = "https://go.dev"
	defaultGoRoot = "/usr/local/go"
)

const (
	rootUserID = 0
	targetOS   = "linux"
	targetArch = "amd64"
)

const userQuestion = "Current user is not root user.\n" +
	recommendMessage +
	"Do you continue upgrade process?"

const recommendMessage = `Recommend that you run it again with the following command.

	sudo $(go env GOPATH)/bin/goversion upgrade

`

func (cmd *Command) Upgrade(ctx context.Context) error {
	if os.Getuid() != rootUserID &&
		!yesno(userQuestion) {
		log.Printf("[INFO] cancel")
		return nil
	}

	url, err := cmd.getDownloadURL(ctx)
	if err != nil {
		return fmt.Errorf("get download URL: %w", err)
	}

	if !yesno("Do you upgrade to %s?", versionRegex.FindString(url)) {
		log.Printf("[INFO] cancel")
		return nil
	}

	log.Printf("[INFO] target URL is %s", url)

	archiveFile, err := cmd.downloadGo(ctx, url)
	if err != nil {
		return fmt.Errorf("download go: %w", err)
	}

	extractDir, err := cmd.extract(archiveFile)
	if err != nil {
		return fmt.Errorf("extract %s: %w", archiveFile, err)
	}

	if !yesno("Do you really overwrite %s?", defaultGoRoot) {
		log.Printf("[INFO] cancel")
		return nil
	}

	goroot := cmd.goEnv(ctx, "GOROOT")
	if goroot == "" {
		goroot = defaultGoRoot
	}

	if err := os.RemoveAll(goroot); err != nil {
		if os.IsPermission(err) {
			log.Printf("[ERR] %s", recommendMessage)
		}
		return fmt.Errorf("remove %s: %w", goroot, err)
	}
	if err := os.Rename(filepath.Join(extractDir, "go"), goroot); err != nil {
		return fmt.Errorf("rename from %s to %s: %w", extractDir, goroot, err)
	}

	log.Printf("[INFO] upgrade success")

	return nil
}

func (cmd *Command) downloadGo(ctx context.Context, url string) (string, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "*-"+path.Base(url))
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tmpFile.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := cmd.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response status is %d", resp.StatusCode)
	}

	log.Printf("[INFO] download to %s", tmpFile.Name())
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"[INFO] downloading",
	)
	if _, err := io.Copy(io.MultiWriter(tmpFile, bar), resp.Body); err != nil {
		return "", fmt.Errorf("response write to %s: %w", tmpFile.Name(), err)
	}

	return tmpFile.Name(), nil
}

func (*Command) extract(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("new gzip reader: %w", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gr); err != nil {
		return "", fmt.Errorf("copy to buffer: %w", err)
	}

	dir := filename[:strings.LastIndex(filename, ".tar.gz")]
	if err := os.Mkdir(dir, 0o755); err != nil {
		return "", fmt.Errorf("make %s directory: %w", dir, err)
	}

	log.Printf("[INFO] extract to %s", dir)

	tr := tar.NewReader(&buf)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if hdr.Typeflag == tar.TypeDir {
			if err := os.Mkdir(filepath.Join(dir, hdr.Name), hdr.FileInfo().Mode()); err != nil {
				return "", fmt.Errorf("create %s: %w", hdr.Name, err)
			}
			continue
		}

		file, err := os.OpenFile(filepath.Join(dir, hdr.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
		if err != nil {
			return "", fmt.Errorf("create %s: %w", hdr.Name, err)
		}
		if _, err := io.Copy(file, tr); err != nil {
			file.Close()
			return "", fmt.Errorf("copy to %s: %w", file.Name(), err)
		}
		file.Close()
	}

	return dir, nil
}
