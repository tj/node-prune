package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/dustin/go-humanize"

	"github.com/tj/node-prune/internal/prune"
)

func init() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.WarnLevel)
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Globs of files that should not be pruned
var exclusionGlobs arrayFlags

// Globs of files that should always be pruned in addition to the defaults
var inclusionGlobs arrayFlags

func main() {
	debug := flag.Bool("verbose", false, "Verbose log output.")
	flag.Var(&exclusionGlobs, "exclude", "Glob of files that should not be pruned. Can be specified multiple times.")
	flag.Var(&inclusionGlobs, "include", "Globs of files that should always be pruned in addition to the defaults. Can be specified multiple times.")
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

	if len(exclusionGlobs) > 0 {
		options = append(options, prune.WithExceptions(exclusionGlobs))
	}

	if len(inclusionGlobs) > 0 {
		options = append(options, prune.WithGlobs(inclusionGlobs))
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
