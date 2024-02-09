package console

import (
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc"
	"strings"
)

type recordFilter struct {
	error   bool
	recType gowarc.RecordType
}

func (r *recordFilter) filterFunc(rec interface{}) bool {
	if rec == nil {
		return false
	}
	if !r.error && r.recType == 0 {
		return true
	}
	if r.error && rec.(record).hasError {
		return true
	}
	if r.recType&rec.(record).recordType != 0 {
		return true
	}
	return false
}

func (r *recordFilter) mouseToggleFilter(g *gocui.Gui, v *gocui.View) error {
	x, y := v.Cursor()
	str, _ := v.Line(y)

	nl := strings.LastIndexFunc(str[:x], indexFunc)
	if nl == -1 {
		nl = 0
	} else {
		nl = nl + 1
	}
	nr := strings.IndexFunc(str[x:], indexFunc)
	if nr == -1 {
		nr = len(str)
	} else {
		nr = nr + x
	}
	word := str[nl:nr]

	switch strings.ToLower(word) {
	case "error":
		_ = r.toggleErrorFilter(g, v)
	case "warcinfo":
		_ = r.toggleRecordTypeFilter(g, gowarc.Warcinfo)
	case "request":
		_ = r.toggleRecordTypeFilter(g, gowarc.Request)
	case "response":
		_ = r.toggleRecordTypeFilter(g, gowarc.Response)
	case "metadata":
		_ = r.toggleRecordTypeFilter(g, gowarc.Metadata)
	case "revisit":
		_ = r.toggleRecordTypeFilter(g, gowarc.Revisit)
	case "resource":
		_ = r.toggleRecordTypeFilter(g, gowarc.Resource)
	case "continuation":
		_ = r.toggleRecordTypeFilter(g, gowarc.Continuation)
	case "conversion":
		_ = r.toggleRecordTypeFilter(g, gowarc.Conversion)
	}

	return nil
}

func indexFunc(r rune) bool {
	return r == ' ' || r == 0 || r == '|'
}

func (r *recordFilter) toggleErrorFilter(g *gocui.Gui, v *gocui.View) error {
	r.error = !r.error
	v2, err := g.View("Records")
	if err != nil {
		return err
	}
	r.refreshHelp(g)
	return state.records.refreshFilter(g, v2)
}

func (r *recordFilter) toggleRecordTypeFilter(g *gocui.Gui, recType gowarc.RecordType) error {
	r.recType = r.recType ^ recType
	v2, err := g.View("Records")
	if err != nil {
		return err
	}
	r.refreshHelp(g)
	return state.records.refreshFilter(g, v2)
}

func (r *recordFilter) refreshHelp(g *gocui.Gui) {
	sb := strings.Builder{}
	sb.WriteString("|")
	sb.WriteString(filterString("Error", ErrorColor, r.error))
	sb.WriteString(filterString("warcInfo", WarcInfoColor, r.recType&gowarc.Warcinfo != 0))
	sb.WriteString(filterString("reQuest", RequestColor, r.recType&gowarc.Request != 0))
	sb.WriteString(filterString("Response", ResponseColor, r.recType&gowarc.Response != 0))
	sb.WriteString(filterString("Metadata", MetadataColor, r.recType&gowarc.Metadata != 0))
	sb.WriteString(filterString("reVisit", RevisitColor, r.recType&gowarc.Revisit != 0))
	sb.WriteString(filterString("reSource", ResourceColor, r.recType&gowarc.Resource != 0))
	sb.WriteString(filterString("Continuation", ContinuationColor, r.recType&gowarc.Continuation != 0))
	sb.WriteString(filterString("coNversion", ConversionColor, r.recType&gowarc.Conversion != 0))
	if v, err := g.View("help"); err == nil {
		v.Clear()
		txt := "h: help"
		width, _ := v.Size()
		space := width - 85
		fmt.Fprintf(v, "%[1]s%[2]*[3]s", sb.String(), space, txt)
	}
}

func filterString(s string, color gocui.Attribute, on bool) string {
	fg := escapeFgColor(gocui.NewRGBColor(0, 0, 0))
	bg := escapeBgColor(gocui.NewRGBColor(0, 0, 0))
	if on {
		return fmt.Sprintf("%s%s%s%s|", escapeBgColor(color), fg, s, escapeFgColor(gocui.ColorDefault))
	} else {
		return fmt.Sprintf("%s%s%s%s|", escapeFgColor(color), bg, s, escapeFgColor(gocui.ColorDefault))
	}
}
