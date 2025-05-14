package warcwriterconfig

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nlnwa/gowarc/v2"
)

func NewIdentityNamer(path, filePrefix, dir string) gowarc.WarcFileNameGenerator {
	basename := filepath.Base(path)
	basename = strings.TrimSuffix(basename, ".gz")
	basename = strings.TrimSuffix(basename, ".arc")
	basename = strings.TrimSuffix(basename, ".warc")

	return &gowarc.PatternNameGenerator{
		Pattern:   "%{prefix}s" + basename + ".%{ext}s",
		Prefix:    filePrefix,
		Directory: dir,
	}
}

func NewNedlibNamer(path, filePrefix, dir string) gowarc.WarcFileNameGenerator {
	filename := filepath.Base(path)
	return &gowarc.PatternNameGenerator{
		Pattern:   "%{prefix}s" + filename + "-%04{serial}d-%{hostOrIp}s.%{ext}s",
		Prefix:    filePrefix,
		Directory: dir,
	}
}

var once sync.Once
var defaultNamer gowarc.WarcFileNameGenerator

func NewDefaultNamer(filePrefix, dir string) gowarc.WarcFileNameGenerator {
	once.Do(func() {
		defaultNamer = &gowarc.PatternNameGenerator{
			Prefix:    filePrefix,
			Directory: dir,
		}
	})
	return defaultNamer
}

func parseSubdirPattern(dirPattern string, t time.Time) string {
	p := strings.ReplaceAll(dirPattern, "{YYYY}", t.Format("2006"))
	p = strings.ReplaceAll(p, "{YY}", t.Format("06"))
	p = strings.ReplaceAll(p, "{MM}", t.Format("01"))
	p = strings.ReplaceAll(p, "{DD}", t.Format("02"))
	return p
}
