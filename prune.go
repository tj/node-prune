// Package prune provides node_modules pruning of unnecessary files.
package prune

import (
	"os"
	"sync"
	"runtime"
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
	".DS_Store",
	".tern-project",
	".gitattributes",
	".editorconfig",
	".eslintrc",
	".jshintrc",
	".flowconfig",
	".documentup.json",
	".yarn-metadata.json",
	".travis.yml",
	"LICENSE.txt",
	"LICENSE",
	"AUTHORS",
	"CONTRIBUTORS",
	".yarn-integrity",
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
}

// DefaultExtensions pruned.
var DefaultExtensions = []string{
	".md",
	".ts",
	".jst",
	".coffee",
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

type DataObject struct {
	path string
	info os.FileInfo
	wg *sync.WaitGroup
	errorChan chan<- error
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
	var wg sync.WaitGroup
	errorChan := make(chan error)
	done := make(chan bool)
	// data channel carries object details for deletion
	data := make(chan *DataObject)

	// spawn go routines to delete dataObject passed on data channel
	for i:=0 ; i<runtime.NumCPU(); i++ {
		go delete_handler(data)
	}

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
		stats.FilesRemoved++
		stats.SizeRemoved += info.Size()

		// dir stats
		if info.IsDir() {
			s, _ := dirStats(path)
			stats.FilesTotal += s.FilesTotal
			stats.FilesRemoved += s.FilesRemoved
			stats.SizeRemoved += s.SizeRemoved
		}

		// spawn go routine to do the removal
		wg.Add(1)
		// create data object and pass it to data channel for deletion
		data <- &DataObject{path, info, &wg, errorChan}

		// avoid traversing of dir to be removed anyway
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})

	// wait until all removal are finished the signal done
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
		case <-done:
			// all have exited cleanly
			return &stats, err
		case err = <-errorChan:
			// error; discard subsequent errors and return error
			close(errorChan)
			return &stats, err
	}

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

func remove(path string, info os.FileInfo, wg *sync.WaitGroup, errorChan chan<- error) {
	// remove and skip dir
	if info.IsDir() {
		if err := os.RemoveAll(path); err != nil {
			errorChan <- errors.Wrap(err, "removing dir")
			return
		}
		wg.Done()
		return
	}

	// remove file
	if err := os.Remove(path); err != nil {
		errorChan <- errors.Wrap(err, "removing")
		return
	}

	// removal done for given path
	wg.Done()
}

func delete_handler(dataChan <-chan *DataObject)  {
	// consume from dataChan and delete files or dir one after another
	for data := range dataChan {
		remove(data.path, data.info, data.wg, data.errorChan)
	}
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
