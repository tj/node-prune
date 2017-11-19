// Package prune provides node_modules pruning of unnecessary files.
package prune

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/pkg/errors"
)

// DefaultFiles pruned.
//
// Copied from yarn (mostly).
var DefaultFiles = []string{
	"Makefile",
	"Gulpfile.js",
	"Gruntfile.js",
	".tern-project",
	".gitattributes",
	".editorconfig",
	".*ignore",
	".eslintrc",
	".jshintrc",
	".flowconfig",
	".documentup.json",
	".yarn-metadata.json",
	".travis.yml",
	"LICENSE.txt",
	"LICENSE",
}

// DefaultDirectories pruned.
//
// Copied from yarn (mostly).
var DefaultDirectories = []string{
	"__tests__",
	"test",
	"tests",
	"powered-test",
	"docs",
	"doc",
	"website",
	"images",
	"assets",
	"example",
	"examples",
	"coverage",
	".nyc_output",
}

// DefaultExtensions pruned.
//
// Copied from yarn (mostly).
var DefaultExtensions = []string{
	".md",
	".ts",
}

// Stats for a prune.
type Stats struct {
	FilesTotal   int64
	FilesRemoved int64
	SizeRemoved  int64
}

// Pruner is a module pruner.
type Pruner struct {
	dir   string
	log   log.Interface
	dirs  map[string]struct{}
	exts  map[string]struct{}
	files map[string]struct{}
}

// Option function.
type Option func(*Pruner)

// New with the given options.
func New(options ...Option) *Pruner {
	v := &Pruner{
		dir:   "node_modules",
		log:   log.Log,
		exts:  toMap(DefaultExtensions),
		dirs:  toMap(DefaultDirectories),
		files: toMap(DefaultFiles),
	}

	for _, o := range options {
		o(v)
	}

	return v
}

// WithDir option.
func WithDir(s string) Option {
	return func(v *Pruner) {
		v.dir = s
	}
}

// WithExtensions option.
func WithExtensions(s []string) Option {
	return func(v *Pruner) {
		v.exts = toMap(s)
	}
}

// WithDirectories option.
func WithDirectories(s []string) Option {
	return func(v *Pruner) {
		v.dirs = toMap(s)
	}
}

// WithFiles option.
func WithFiles(s []string) Option {
	return func(v *Pruner) {
		v.files = toMap(s)
	}
}

// Prune performs the pruning.
func (p Pruner) Prune() (*Stats, error) {
	var stats Stats

	err := filepath.Walk(p.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		stats.FilesTotal++

		ctx := p.log.WithField("path", path)

		if !p.prune(path, info) {
			ctx.Debug("keeping")
			return nil
		}

		if info.IsDir() {
			return filepath.SkipDir
		}

		ctx.Debug("prune")
		stats.FilesRemoved++
		stats.SizeRemoved += info.Size()

		if err := os.Remove(path); err != nil {
			return errors.Wrap(err, "removing")
		}

		return nil
	})

	return &stats, err
}

// prune returns true if the file or dir should be pruned.
func (p Pruner) prune(path string, info os.FileInfo) bool {
	// directories
	if info.IsDir() {
		_, ok := p.dirs[info.Name()]
		return ok
	}

	// files
	_, ok := p.files[info.Name()]
	if ok {
		return true
	}

	// extensions
	ext := filepath.Ext(path)
	_, ok = p.exts[ext]
	return ok
}

// toMap returns a map from slice.
func toMap(s []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
