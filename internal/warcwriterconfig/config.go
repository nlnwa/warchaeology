/*
 * Copyright 2021 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package warcwriterconfig

import (
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/utils"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const DefaultDateFormat = "2006-1-2"

type WarcWriterConfig struct {
	Compress              bool
	CompressionLevel      int
	ConcurrentWriters     int
	MaxFileSize           int64
	FilePrefix            string
	DefaultTime           time.Time
	OutDir                string
	Flush                 bool
	WarcVersion           *gowarc.WarcVersion
	WarcFileNameGenerator string
	SubDirPattern         string
	writers               map[string]*gowarc.WarcFileWriter
	WarcInfoFunc          func(recordBuilder gowarc.WarcRecordBuilder) error
	writersGuard          sync.Mutex
	OneToOneWriter        bool
}

func NewFromViper() (*WarcWriterConfig, error) {
	var err error
	var outDir string
	if outDir, err = filepath.Abs(viper.GetString(flag.WarcDir)); err != nil {
		return nil, err
	}
	if outDir, err = filepath.EvalSymlinks(outDir); err != nil {
		return nil, err
	}
	if f, err := os.Lstat(outDir); err != nil {
		return nil, fmt.Errorf("could not write to output directory '%s': %w", outDir, err.(*os.PathError).Err)
	} else if !f.IsDir() {
		return nil, fmt.Errorf("could not write to output directory: '%s' is not a directory", outDir)
	}

	var version *gowarc.WarcVersion
	switch viper.GetString(flag.WarcVersion) {
	case "1.0":
		version = gowarc.V1_0
	case "1.1":
		version = gowarc.V1_1
	default:
		return nil, fmt.Errorf("unknown WARC version: %s", viper.GetString(flag.WarcVersion))
	}

	var defaultDate time.Time
	if t, err := time.Parse(DefaultDateFormat, viper.GetString(flag.DefaultDate)); err != nil {
		return nil, err
	} else {
		defaultDate = t.Add(12 * time.Hour)
	}

	return &WarcWriterConfig{
		Compress:              viper.GetBool(flag.Compress),
		CompressionLevel:      viper.GetInt(flag.CompressionLevel),
		ConcurrentWriters:     viper.GetInt(flag.ConcurrentWriters),
		MaxFileSize:           utils.ParseSizeInBytes(viper.GetString(flag.FileSize)),
		DefaultTime:           defaultDate,
		OutDir:                outDir,
		FilePrefix:            viper.GetString(flag.FilePrefix),
		SubDirPattern:         viper.GetString(flag.SubdirPattern),
		WarcFileNameGenerator: viper.GetString(flag.NameGenerator),
		Flush:                 viper.GetBool(flag.Flush),
		WarcVersion:           version,
		writers:               map[string]*gowarc.WarcFileWriter{},
	}, nil
}

func (w *WarcWriterConfig) GetWarcWriter(fromFileName, warcDate string) *gowarc.WarcFileWriter {
	var namer gowarc.WarcFileNameGenerator
	var lookupKey string
	var dir string

	if w.OneToOneWriter {
		// Only one writer with unrestricted size to allow for one to one mapping
		w.ConcurrentWriters = 1
		w.MaxFileSize = 0
	}

	s, err := parseSubdirPattern(w.SubDirPattern, warcDate)
	if err != nil {
		panic(err)
	}
	if s != "" {
		dir = w.OutDir + "/" + s
	} else {
		dir = w.OutDir
	}
	lookupKey = s

	switch w.WarcFileNameGenerator {
	case "identity":
		namer = NewIdentityNamer(fromFileName, w.FilePrefix, dir)
	default:
		namer = NewDefaultNamer(fromFileName, w.FilePrefix, dir)
	}

	if w.OneToOneWriter {
		if err := os.MkdirAll(dir, 0777); err != nil {
			panic(err)
		}

		return gowarc.NewWarcFileWriter(
			gowarc.WithMaxConcurrentWriters(w.ConcurrentWriters),
			gowarc.WithCompression(w.Compress),
			gowarc.WithCompressionLevel(w.CompressionLevel),
			gowarc.WithMaxFileSize(w.MaxFileSize),
			gowarc.WithRecordOptions(gowarc.WithVersion(w.WarcVersion)),
			gowarc.WithFileNameGenerator(namer),
			gowarc.WithFlush(w.Flush),
			gowarc.WithWarcInfoFunc(w.WarcInfoFunc),
			gowarc.WithRecordOptions(gowarc.WithVersion(w.WarcVersion), gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir))),
		)
	} else {
		w.writersGuard.Lock()
		defer w.writersGuard.Unlock()

		if ww, ok := w.writers[lookupKey]; ok {
			return ww
		} else {
			ww = gowarc.NewWarcFileWriter(
				gowarc.WithMaxConcurrentWriters(w.ConcurrentWriters),
				gowarc.WithCompression(w.Compress),
				gowarc.WithCompressionLevel(w.CompressionLevel),
				gowarc.WithMaxFileSize(w.MaxFileSize),
				gowarc.WithRecordOptions(gowarc.WithVersion(w.WarcVersion)),
				gowarc.WithFileNameGenerator(namer),
				gowarc.WithFlush(w.Flush),
				gowarc.WithWarcInfoFunc(w.WarcInfoFunc),
				gowarc.WithRecordOptions(gowarc.WithVersion(w.WarcVersion), gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir))),
			)
			w.writers[lookupKey] = ww
			if err := os.MkdirAll(dir, 0777); err != nil {
				panic(err)
			}
			return ww
		}
	}
}

func (w *WarcWriterConfig) Close() {
	for _, writer := range w.writers {
		err := writer.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error closing WARC writer: %v\n", err)
		}
	}
}
