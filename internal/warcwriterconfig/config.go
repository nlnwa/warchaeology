package warcwriterconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v4/internal/hooks"
	"github.com/nlnwa/warchaeology/v4/internal/util"
)

const DefaultDateFormat = "2006-1-2"

type WarcWriterConfig struct {
	FilePrefix            string
	DefaultTime           time.Time
	OutDir                string
	TmpDir                string
	Flush                 bool
	WarcVersion           *gowarc.WarcVersion
	WarcFileNameGenerator string
	SubDirPattern         string
	writers               map[string]*gowarc.WarcFileWriter
	WarcInfoFunc          func(recordBuilder gowarc.WarcRecordBuilder) error
	writersGuard          sync.Mutex
	OneToOneWriter        bool
	openOutputFileHook    hooks.OpenOutputFileHook
	closeOutputFileHook   hooks.CloseOutputFileHook
	WarcFileWriterOptions []gowarc.WarcFileWriterOption
}

type WarcWriterOptions struct {
	WarcVersion           string
	OutDir                string
	DefaultTime           string
	OpenOutputFileHook    string
	CloseOutputFileHook   string
	Compress              bool
	CompressionLevel      int
	ConcurrentWriters     int
	MaxFileSize           string
	FilePrefix            string
	SubDirPattern         string
	WarcFileNameGenerator string
	Flush                 bool
	OneToOneWriter        bool
	WarcInfoFunc          func(recordBuilder gowarc.WarcRecordBuilder) error
	TmpDir                string
}

func defaultWarcWriterOptions() *WarcWriterOptions {
	return &WarcWriterOptions{
		WarcVersion:       "1.1",
		DefaultTime:       time.Now().Format(DefaultDateFormat),
		Compress:          true,
		CompressionLevel:  1,
		ConcurrentWriters: 1,
	}
}

func WithDefaultTime(time string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.DefaultTime = time
	}
}

func WithWarcVersion(version string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.WarcVersion = version
	}
}

func WithOutDir(outDir string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.OutDir = outDir
	}
}

func WithOpenOutputFileHook(hook string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.OpenOutputFileHook = hook
	}
}

func WithCloseOutputFileHook(hook string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.CloseOutputFileHook = hook
	}
}

func WithCompress(compress bool) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.Compress = compress
	}
}

func WithCompressionLevel(level int) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.CompressionLevel = level
	}
}

func WithConcurrentWriters(writers int) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.ConcurrentWriters = writers
	}
}

func WithMaxFileSize(size string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.MaxFileSize = size
	}
}

func WithFilePrefix(prefix string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.FilePrefix = prefix
	}
}

func WithSubDirPattern(pattern string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.SubDirPattern = pattern
	}
}

func WithWarcFileNameGenerator(generator string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.WarcFileNameGenerator = generator
	}
}

func WithFlush(flush bool) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.Flush = flush
	}
}

func WithOneToOneWriter(oneToOne bool) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.OneToOneWriter = oneToOne
	}
}

func WithWarcInfoFunc(f func(recordBuilder gowarc.WarcRecordBuilder) error) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.WarcInfoFunc = f
	}
}

func WithBufferTmpDir(tmpDir string) func(*WarcWriterOptions) {
	return func(w *WarcWriterOptions) {
		w.TmpDir = tmpDir
	}
}

func New(cmd string, options ...func(*WarcWriterOptions)) (*WarcWriterConfig, error) {
	o := defaultWarcWriterOptions()
	for _, option := range options {
		option(o)
	}

	var err error
	var outDir string
	if outDir, err = filepath.Abs(o.OutDir); err != nil {
		return nil, err
	}
	if outDir, err = filepath.EvalSymlinks(outDir); err != nil {
		return nil, err
	}
	if f, err := os.Lstat(outDir); err != nil {
		return nil, fmt.Errorf("failed to stat output directory '%s': %w", outDir, err.(*os.PathError).Err)
	} else if !f.IsDir() {
		return nil, fmt.Errorf("specified output directory is not a directory: %s", outDir)
	}

	var defaultTime time.Time
	if t, err := time.Parse(DefaultDateFormat, o.DefaultTime); err != nil {
		return nil, err
	} else {
		defaultTime = t.Add(12 * time.Hour)
	}

	var version *gowarc.WarcVersion
	switch o.WarcVersion {
	case "1.0":
		version = gowarc.V1_0
	case "1.1":
		version = gowarc.V1_1
	default:
		return nil, fmt.Errorf("unknown WARC version: %s", o.WarcVersion)
	}

	openOutputFileHook, err := hooks.NewOpenOutputFileHook(cmd, o.OpenOutputFileHook)
	if err != nil {
		return nil, err
	}

	closeOutputFileHook, err := hooks.NewCloseOutputFileHook(cmd, o.CloseOutputFileHook)
	if err != nil {
		return nil, err
	}

	if o.OneToOneWriter {
		// Only one writer with unrestricted size to allow for one to one mapping
		o.ConcurrentWriters = 1
		o.MaxFileSize = "0"
		o.WarcFileNameGenerator = "identity"
	}

	warcFileWriterOptions := []gowarc.WarcFileWriterOption{
		gowarc.WithMaxConcurrentWriters(o.ConcurrentWriters),
		gowarc.WithCompression(o.Compress),
		gowarc.WithCompressionLevel(o.CompressionLevel),
		gowarc.WithMaxFileSize(util.ParseSizeInBytes(o.MaxFileSize)),
		gowarc.WithFlush(o.Flush),
		gowarc.WithWarcInfoFunc(o.WarcInfoFunc),
		gowarc.WithRecordOptions(gowarc.WithVersion(version), gowarc.WithBufferTmpDir(o.TmpDir)),
	}

	return &WarcWriterConfig{
		DefaultTime:           defaultTime,
		OutDir:                outDir,
		FilePrefix:            o.FilePrefix,
		SubDirPattern:         o.SubDirPattern,
		WarcFileNameGenerator: o.WarcFileNameGenerator,
		openOutputFileHook:    openOutputFileHook,
		closeOutputFileHook:   closeOutputFileHook,
		writers:               make(map[string]*gowarc.WarcFileWriter),
		WarcFileWriterOptions: warcFileWriterOptions,
		WarcVersion:           version,
		OneToOneWriter:        o.OneToOneWriter,
	}, nil
}

func (w *WarcWriterConfig) GetWarcWriter(path, warcDate string) (*gowarc.WarcFileWriter, error) {
	var namer gowarc.WarcFileNameGenerator
	var dir string

	subDir, err := parseSubdirPattern(w.SubDirPattern, warcDate)
	if err != nil {
		return nil, err
	}

	if subDir != "" {
		dir = filepath.Join(w.OutDir, subDir)
	} else {
		dir = w.OutDir
	}
	// create output directory if it does not exist
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}

	switch w.WarcFileNameGenerator {
	case "identity":
		namer = NewIdentityNamer(path, w.FilePrefix, dir)
	case "nedlib":
		namer = NewNedlibNamer(path, w.FilePrefix, dir)
	default:
		namer = NewDefaultNamer(w.FilePrefix, dir)
	}

	var ww *gowarc.WarcFileWriter

	if w.OneToOneWriter {
		ww = gowarc.NewWarcFileWriter(
			append(w.WarcFileWriterOptions,
				gowarc.WithFileNameGenerator(namer),
				gowarc.WithBeforeFileCreationHook(w.openOutputFileHook.WithSrcFileName(path).Run),
				gowarc.WithAfterFileCreationHook(w.closeOutputFileHook.WithSrcFileName(path).Run),
			)...,
		)
		return ww, nil
	}

	w.writersGuard.Lock()
	defer w.writersGuard.Unlock()

	ww, ok := w.writers[subDir]
	if ok {
		return ww, nil
	}

	ww = gowarc.NewWarcFileWriter(
		append(w.WarcFileWriterOptions,
			gowarc.WithFileNameGenerator(namer),
		)...,
	)
	w.writers[subDir] = ww

	return ww, nil
}

func (w *WarcWriterConfig) Close() error {
	var lastErr error
	for _, writer := range w.writers {
		err := writer.Close()
		if err != nil {
			lastErr = fmt.Errorf("error closing WARC writer: %w", err)
		}
	}
	return lastErr
}
