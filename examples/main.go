package main

import (
	"context"
	"github.com/danesparza/package-assistant/internal/monitor"
	"github.com/sanity-io/litter"
)

func main() {
	repoPath := "/Users/danesparza/work/package-repo"
	files := monitor.FindOldFileVersions(context.Background(), repoPath)
	litter.Dump(files)
}
