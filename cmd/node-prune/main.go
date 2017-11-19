package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/dustin/go-humanize"

	"github.com/tj/node-prune"
)

func init() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.WarnLevel)
}

func main() {
	debug := flag.Bool("verbose", false, "Verbose log output.")
	flag.Parse()
	dir := flag.Arg(0)

	start := time.Now()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	var options []prune.Option

	if dir != "" {
		options = append(options, prune.WithDir(dir))
	}

	p := prune.New(options...)

	stats, err := p.Prune()
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	println()
	defer println()

	output("files total", humanize.Comma(stats.FilesTotal))
	output("files removed", humanize.Comma(stats.FilesRemoved))
	output("size removed", humanize.Bytes(uint64(stats.SizeRemoved)))
	output("duration", time.Since(start).Round(time.Millisecond).String())
}

func output(name, val string) {
	fmt.Printf("\x1b[1m%20s\x1b[0m %s\n", name, val)
}
