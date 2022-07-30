package main

import (
	"context"
	"fmt"
)

func (cmd *Command) OutputLatestVersion(ctx context.Context) error {
	versions, err := cmd.getGoVersions(ctx, false)
	if err != nil {
		return fmt.Errorf("get download URL: %w", err)
	}
	fmt.Printf("latest version: %s\n", versions[0].Version)
	return nil
}
