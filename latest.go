package main

import (
	"context"
	"fmt"
)

func (cmd *Command) OutputLatestVersion(ctx context.Context) error {
	downloadURL, err := cmd.getDownloadURL(ctx)
	if err != nil {
		return fmt.Errorf("get download URL: %w", err)
	}
	fmt.Printf("latest version: %s\n", versionRegex.FindString(downloadURL))
	return nil
}
