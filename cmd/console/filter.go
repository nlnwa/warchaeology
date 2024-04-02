package console

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc"
)

type recordFilter struct {
	error   bool
	recType gowarc.RecordType
}

func (recordfilter *recordFilter) filterFunc(rec interface{}) bool {
	if rec == nil {
		return false
	}
	if !recordfilter.error && recordfilter.recType == 0 {
		return true
	}
	if recordfilter.error && rec.(record).hasError {
		return true
	}
	if recordfilter.recType&rec.(record).recordType != 0 {
		return true
	}
	return false
}

func (recordfilter *recordFilter) mouseToggleFilter(gui *gocui.Gui, view *gocui.View) error {
	x, y := view.Cursor()
	str, _ := view.Line(y)

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
		_ = recordfilter.toggleErrorFilter(gui, view)
	case "warcinfo":
		_ = recordfilter.toggleRecordTypeFilter(gui, gowarc.Warcinfo)
	case "request":
		_ = recordfilter.toggleRecordTypeFilter(gui, gowarc.Request)
	case "response":
		_ = recordfilter.toggleRecordTypeFilter(gui, gowarc.Response)
	case "metadata":
		_ = recordfilter.toggleRecordTypeFilter(gui, gowarc.Metadata)
	case "revisit":
		_ = recordfilter.toggleRecordTypeFilter(gui, gowarc.Revisit)
	case "resource":
		_ = recordfilter.toggleRecordTypeFilter(gui, gowarc.Resource)
	case "continuation":
		_ = recordfilter.toggleRecordTypeFilter(gui, gowarc.Continuation)
	case "conversion":
		_ = recordfilter.toggleRecordTypeFilter(gui, gowarc.Conversion)
	}

	return nil
}

func indexFunc(recordfilter rune) bool {
	return recordfilter == ' ' || recordfilter == 0 || recordfilter == '|'
}

func (recordfilter *recordFilter) toggleErrorFilter(gui *gocui.Gui, view *gocui.View) error {
	recordfilter.error = !recordfilter.error
	recordView, err := gui.View("Records")
	if err != nil {
		return err
	}
	recordfilter.refreshHelp(gui)
	return state.records.refreshFilter(gui, recordView)
}

func (recordfilter *recordFilter) toggleRecordTypeFilter(gui *gocui.Gui, recType gowarc.RecordType) error {
	recordfilter.recType = recordfilter.recType ^ recType
	recordView, err := gui.View("Records")
	if err != nil {
		return err
	}
	recordfilter.refreshHelp(gui)
	return state.records.refreshFilter(gui, recordView)
}

func (recordfilter *recordFilter) refreshHelp(gui *gocui.Gui) {
	toolbarString := strings.Builder{}
	toolbarString.WriteString("|")
	toolbarString.WriteString(filterString("Error", ErrorColor, recordfilter.error))
	toolbarString.WriteString(filterString("warcInfo", WarcInfoColor, recordfilter.recType&gowarc.Warcinfo != 0))
	toolbarString.WriteString(filterString("reQuest", RequestColor, recordfilter.recType&gowarc.Request != 0))
	toolbarString.WriteString(filterString("Response", ResponseColor, recordfilter.recType&gowarc.Response != 0))
	toolbarString.WriteString(filterString("Metadata", MetadataColor, recordfilter.recType&gowarc.Metadata != 0))
	toolbarString.WriteString(filterString("reVisit", RevisitColor, recordfilter.recType&gowarc.Revisit != 0))
	toolbarString.WriteString(filterString("reSource", ResourceColor, recordfilter.recType&gowarc.Resource != 0))
	toolbarString.WriteString(filterString("Continuation", ContinuationColor, recordfilter.recType&gowarc.Continuation != 0))
	toolbarString.WriteString(filterString("coNversion", ConversionColor, recordfilter.recType&gowarc.Conversion != 0))
	if view, err := gui.View("help"); err == nil {
		view.Clear()
		helpText := "h: help"
		width, _ := view.Size()
		space := width - 85
		fmt.Fprintf(view, "%[1]s%[2]*[3]s", toolbarString.String(), space, helpText)
	}
}

func filterString(s string, color gocui.Attribute, enabled bool) string {
	foreground := escapeFgColor(gocui.NewRGBColor(0, 0, 0))
	background := escapeBgColor(gocui.NewRGBColor(0, 0, 0))
	if enabled {
		return fmt.Sprintf("%s%s%s%s|", escapeBgColor(color), foreground, s, escapeFgColor(gocui.ColorDefault))
	} else {
		return fmt.Sprintf("%s%s%s%s|", escapeFgColor(color), background, s, escapeFgColor(gocui.ColorDefault))
	}
}
