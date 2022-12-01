package warcwriterconfig

import (
	"github.com/nlnwa/gowarc"
	"path"
	"strings"
	"time"
)

type IdentityNamer struct {
	gowarc.PatternNameGenerator
}

func NewIdentityNamer(fromFileName string) *IdentityNamer {
	fromFileName = path.Base(fromFileName)
	fromFileName = strings.TrimSuffix(fromFileName, ".gz")
	fromFileName = strings.TrimSuffix(fromFileName, ".arc")
	fromFileName = strings.TrimSuffix(fromFileName, ".warc")

	i := &IdentityNamer{}
	i.Pattern = "%{prefix}s" + fromFileName + ".%{ext}s"
	return i
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
