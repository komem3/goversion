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
	"runtime"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

var maxWorkers = runtime.NumCPU() * 4

const defaultGoRoot = "/usr/local/go"

const (
	rootUserID = 0
	targetOS   = "linux"
	targetArch = "amd64"
)

const userQuestion = "Current user is not root user.\n" +
	recommendMessage +
	"Do you continue upgrade process?"

const recommendMessage = "It is recommended to run with administrator privileges.\n"

func (cmd *Command) Upgrade(ctx context.Context) error {
	if os.Getuid() != rootUserID &&
		!yesno(userQuestion) {
		log.Printf("[INFO] cancel")
		return nil
	}

	versions, err := cmd.getGoVersions(ctx, false)
	if err != nil {
		return fmt.Errorf("get download URL: %w", err)
	}

	url := versions[0].getDownloadURL()
	if url == "" {
		return fmt.Errorf("missing install target for %s", versions[0].Version)
	}
	if !yesno("Do you upgrade to %s?", versions[0].Version) {
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

	if err := os.RemoveAll(defaultGoRoot); err != nil {
		if os.IsPermission(err) {
			log.Printf("[ERR] %s", recommendMessage)
		}
		return fmt.Errorf("remove %s: %w", defaultGoRoot, err)
	}
	if err := os.Rename(filepath.Join(extractDir, "go"), defaultGoRoot); err != nil {
		return fmt.Errorf("rename from %s to %s: %w", extractDir, defaultGoRoot, err)
	}

	log.Printf("[INFO] upgrade success")

	return nil
}

func (cmd *Command) getContentLength(ctx context.Context, url string) (int64, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, false, fmt.Errorf("create request: %w", err)
	}

	resp, err := cmd.client.Do(req)
	if err != nil {
		return 0, false, fmt.Errorf("request %s: %w", url, err)
	}
	return resp.ContentLength, resp.Header.Get("Accept-Ranges") == "bytes", nil
}

func (cmd *Command) partialDownload(ctx context.Context, url string, size int64) (*bytes.Buffer, error) {
	chunk := int(size / int64(maxWorkers))
	if size%int64(maxWorkers) != 0 {
		chunk++
	}
	var (
		wg   sync.WaitGroup
		buf  = make([]bytes.Buffer, maxWorkers)
		errs error
		bar  = progressbar.DefaultBytes(
			size,
			"[INFO] downloading",
		)
	)
	for i := 0; i < maxWorkers; i++ {
		i := i
		wg.Add(1)

		go func() {
			defer wg.Done()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("create request: %w", err))
				return
			}

			if i+1 == maxWorkers {
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-", i*chunk))
			} else {
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", i*chunk, (i+1)*chunk-1))
			}

			resp, err := cmd.client.Do(req)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("request %s: %w", url, err))
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusPartialContent {
				errs = errors.Join(errs, fmt.Errorf("response status is %d", resp.StatusCode))
				return
			}

			if _, err := io.Copy(io.MultiWriter(&buf[i], bar), resp.Body); err != nil {
				errs = errors.Join(errs, fmt.Errorf("write response: %w", err))
				return
			}
		}()
	}

	wg.Wait()

	if errs != nil {
		return nil, errs
	}

	var file bytes.Buffer
	file.Grow(int(size))
	for _, b := range buf {
		if _, err := file.Write(b.Bytes()); err != nil {
			return nil, fmt.Errorf("write bytes: %w", err)
		}
	}
	return &file, nil
}

func (cmd *Command) allDownload(ctx context.Context, url string) (*bytes.Buffer, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := cmd.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status is %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"[INFO] downloading",
	)
	if _, err := io.Copy(io.MultiWriter(&buf, bar), resp.Body); err != nil {
		return nil, fmt.Errorf("response write to buffer: %w", err)
	}

	return &buf, nil
}

func (cmd *Command) downloadGo(ctx context.Context, url string) (string, error) {
	size, partialable, err := cmd.getContentLength(ctx, url)
	if err != nil {
		return "", fmt.Errorf("get content length: %w", err)
	}

	var resp *bytes.Buffer
	if partialable {
		resp, err = cmd.partialDownload(ctx, url, size)
		if err != nil {
			return "", fmt.Errorf("partial download: %w", err)
		}
	} else {
		resp, err = cmd.allDownload(ctx, url)
		if err != nil {
			return "", fmt.Errorf("all download: %w", err)
		}
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "*-"+path.Base(url))
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp); err != nil {
		return "", fmt.Errorf("write to %s: %w", tmpFile.Name(), err)
	}
	log.Printf("[INFO] save to %s", tmpFile.Name())

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
