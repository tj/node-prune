// Package prune provides node_modules pruning of unnecessary files.
package prune

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/apex/log"
)

// DefaultFiles pruned.
//
// Copied from yarn (mostly).
var DefaultFiles = []string{
	"Makefile",
	"Gulpfile.js",
	"Gruntfile.js",
	"gulpfile.js",
	".DS_Store",
	".tern-project",
	".gitattributes",
	".editorconfig",
	".eslintrc",
	"eslint",
	".eslintrc.js",
	".eslintrc.json",
	".eslintignore",
	".stylelintrc",
	"stylelint.config.js",
	".stylelintrc.json",
	".stylelintrc.yaml",
	".stylelintrc.yml",
	".stylelintrc.js",
	".htmllintrc",
	"htmllint.js",
	".lint",
	".npmignore",
	".jshintrc",
	".flowconfig",
	".documentup.json",
	".yarn-metadata.json",
	".travis.yml",
	"appveyor.yml",
	".gitlab-ci.yml",
	"circle.yml",
	".coveralls.yml",
	"CHANGES",
	"LICENSE.txt",
	"LICENSE",
	"license",
	"AUTHORS",
	"CONTRIBUTORS",
	".yarn-integrity",
	".yarnclean",
	"_config.yml",
	".babelrc",
	".yo-rc.json",
	"jest.config.js",
	"karma.conf.js",
	".appveyor.yml",
	"tsconfig.json",
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
	".idea",
	".vscode",
	"website",
	"images",
	"assets",
	"example",
	"examples",
	"coverage",
	".nyc_output",
	".circleci",
	".github",
}

// DefaultExtensions pruned.
var DefaultExtensions = []string{
	".markdown",
	".md",
	".ts",
	".jst",
	".coffee",
	".tgz",
	".swp",
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
	ch    chan func()
	wg    sync.WaitGroup
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
		ch:    make(chan func()),
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
func (p *Pruner) Prune() (*Stats, error) {
	var stats Stats

	p.startN(runtime.NumCPU())
	defer p.stop()

	err := filepath.Walk(p.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		stats.FilesTotal++

		ctx := p.log.WithFields(log.Fields{
			"path": path,
			"size": info.Size(),
			"dir":  info.IsDir(),
		})

		// keep
		if !p.prune(path, info) {
			ctx.Debug("keep")
			return nil
		}

		// prune
		ctx.Info("prune")
		atomic.AddInt64(&stats.FilesRemoved, 1)
		atomic.AddInt64(&stats.SizeRemoved, info.Size())

		// remove and skip dir
		if info.IsDir() {
			p.ch <- func() {
				s, _ := dirStats(path)

				atomic.AddInt64(&stats.FilesTotal, s.FilesTotal)
				atomic.AddInt64(&stats.FilesRemoved, s.FilesRemoved)
				atomic.AddInt64(&stats.SizeRemoved, s.SizeRemoved)

				if err := os.RemoveAll(path); err != nil {
					ctx.WithError(err).Error("removing directory")
				}
			}
			return filepath.SkipDir
		}

		// remove file
		p.ch <- func() {
			if err := os.Remove(path); err != nil {
				ctx.WithError(err).Error("removing file")
			}
		}

		return nil
	})

	return &stats, err
}

// prune returns true if the file or dir should be pruned.
func (p *Pruner) prune(path string, info os.FileInfo) bool {
	// directories
	if info.IsDir() {
		_, ok := p.dirs[info.Name()]
		return ok
	}

	// files
	if _, ok := p.files[info.Name()]; ok {
		return true
	}

	// files exact match
	if _, ok := p.files[path]; ok {
		return true
	}

	// extensions
	ext := filepath.Ext(path)
	_, ok := p.exts[ext]
	return ok
}

// startN starts n loops.
func (p *Pruner) startN(n int) {
	for i := 0; i < n; i++ {
		p.wg.Add(1)
		go p.start()
	}
}

// start loop.
func (p *Pruner) start() {
	defer p.wg.Done()
	for fn := range p.ch {
		fn()
	}
}

// stop loop.
func (p *Pruner) stop() {
	close(p.ch)
	p.wg.Wait()
}

// dirStats returns stats for files in dir.
func dirStats(dir string) (*Stats, error) {
	var stats Stats

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		stats.FilesTotal++
		stats.FilesRemoved++
		stats.SizeRemoved += info.Size()
		return err
	})

	return &stats, err
}

// toMap returns a map from slice.
func toMap(s []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
