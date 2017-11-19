package main

import (
	"flag"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
)

func init() {
	log.SetHandler(text.Default)
}

func main() {
	debug := flag.Bool("verbose", false, "Verbose log output.")
	flag.Parse()
	dir := flag.Arg(0)

	start := time.Now()

	if dir == "" {
		dir = "node_modules"
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	stats, err := prune(dir)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	log.WithFields(log.Fields{
		"files_total":   humanize.Comma(stats.FilesTotal),
		"files_removed": humanize.Comma(stats.FilesRemoved),
		"size_removed":  humanize.Bytes(uint64(stats.SizeRemoved)),
		"duration":      time.Since(start).Round(time.Millisecond).String(),
	}).Info("complete")
}

// Stats of the prune.
type Stats struct {
	FilesTotal   int64
	FilesRemoved int64
	SizeRemoved  int64
}

// prune files in dir.
func prune(dir string) (*Stats, error) {
	var stats Stats

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		atomic.AddInt64(&stats.FilesTotal, 1)

		if !prunable(path, info) {
			return nil
		}

		log.WithField("path", path).Debug("prune")
		atomic.AddInt64(&stats.FilesRemoved, 1)
		atomic.AddInt64(&stats.SizeRemoved, info.Size())
		if err := os.Remove(path); err != nil {
			return errors.Wrap(err, "removing")
		}

		return nil
	})

	return &stats, err
}

// prunable returns true if the file should be pruned.
func prunable(path string, info os.FileInfo) bool {
	ext := filepath.Ext(path)
	return ext == ".ts" || ext == ".md"
}
