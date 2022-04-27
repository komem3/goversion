package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var versionRegex = regexp.MustCompile(`go[1-9]\.+[0-9]{1,2}(\.+[0-9]{1,2})?`)

type baseCmd struct {
	client *http.Client
}

func (cmd *baseCmd) getDownloadURL(ctx context.Context) (string, error) {
	const downlaodURL = baseURL + "/dl"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downlaodURL, nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	res, err := cmd.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request %s: %w", downlaodURL, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response status is %d", res.StatusCode)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("load the HTML document: %w", err)
	}

	href, exists := doc.Find(".downloadtable .download").FilterFunction(func(_ int, s *goquery.Selection) bool {
		text := s.Text()
		return filepath.Ext(text) == ".gz" && strings.Contains(text, targetOS) && strings.Contains(text, targetArch)
	}).Attr("href")
	if !exists {
		return "", fmt.Errorf("html element does not have href")
	}

	return baseURL + href, nil
}

func (*baseCmd) goEnv(ctx context.Context, env string) string {
	cmd := exec.CommandContext(ctx, "go", "env", env)

	out, err := cmd.Output()
	if err != nil {
		log.Printf("[ERR] go env %s: %v", env, err)
		return ""
	}
	return string(out)
}
