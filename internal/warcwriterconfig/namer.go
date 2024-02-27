package warcwriterconfig

import (
	"path"
	"strings"
	"sync"
	"time"

	"github.com/nlnwa/gowarc"
)

func NewIdentityNamer(fromFileName, filePrefix, dir string) gowarc.WarcFileNameGenerator {
	fromFileName = path.Base(fromFileName)
	fromFileName = strings.TrimSuffix(fromFileName, ".gz")
	fromFileName = strings.TrimSuffix(fromFileName, ".arc")
	fromFileName = strings.TrimSuffix(fromFileName, ".warc")

	return &gowarc.PatternNameGenerator{
		Pattern:   "%{prefix}s" + fromFileName + ".%{ext}s",
		Prefix:    filePrefix,
		Directory: dir,
	}
}

func NewNedlibNamer(fromFileName, filePrefix, dir string) gowarc.WarcFileNameGenerator {
	return &gowarc.PatternNameGenerator{
		Pattern:   "%{prefix}s" + fromFileName + "-%04{serial}d-%{hostOrIp}s.%{ext}s",
		Prefix:    filePrefix,
		Directory: dir,
	}
}

var once sync.Once
var defaultNamer gowarc.WarcFileNameGenerator

func NewDefaultNamer(fromFileName, filePrefix, dir string) gowarc.WarcFileNameGenerator {
	once.Do(func() {
		defaultNamer = &gowarc.PatternNameGenerator{
			Prefix:    filePrefix,
			Directory: dir,
		}
	})
	return defaultNamer
}

func parseSubdirPattern(dirPattern string, recordDate string) (string, error) {
	t, err := time.Parse(time.RFC3339, recordDate)
	if err != nil {
		return "", err
	}
	p := strings.ReplaceAll(dirPattern, "{YYYY}", t.Format("2006"))
	p = strings.ReplaceAll(p, "{YY}", t.Format("06"))
	p = strings.ReplaceAll(p, "{MM}", t.Format("01"))
	p = strings.ReplaceAll(p, "{DD}", t.Format("02"))
	return p, nil
}
