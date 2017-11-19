package main

import (
	"flag"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/dustin/go-humanize"

	"github.com/tj/node-prune"
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

	stats, err := prune.Prune(dir)
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
