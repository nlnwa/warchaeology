package widget

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/nationallibraryofnorway/warchaeology/v5/internal/ui/model"
	"github.com/nlnwa/gowarc/v3"
)

type ChromeWidget struct {
	gui      *gocui.Gui
	viewName string

	recType         gowarc.RecordType
	showLineEndings bool

	ctrl Controller
}

func NewChromeWidget(gui *gocui.Gui, viewName string, ctrl Controller) *ChromeWidget {
	return &ChromeWidget{
		gui:      gui,
		viewName: viewName,
		ctrl:     ctrl,
	}
}

func (w *ChromeWidget) Layout(gui *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := gui.SetView(w.viewName, x0, y0, x1, y1, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}
	if errors.Is(err, gocui.ErrUnknownView) {
		v.Frame = false

		if err := gui.SetKeybinding(w.viewName, gocui.MouseLeft, gocui.ModNone, w.onMouseToggle); err != nil {
			return err
		}

		type recordTypeBinding struct {
			key     rune
			recType gowarc.RecordType
		}
		bindings := []recordTypeBinding{
			{key: 'e', recType: model.ErrorRecordType},
			{key: 'i', recType: gowarc.Warcinfo},
			{key: 'q', recType: gowarc.Request},
			{key: 'r', recType: gowarc.Response},
			{key: 'm', recType: gowarc.Metadata},
			{key: 's', recType: gowarc.Resource},
			{key: 'v', recType: gowarc.Revisit},
			{key: 'c', recType: gowarc.Continuation},
			{key: 'n', recType: gowarc.Conversion},
		}
		for _, b := range bindings {
			if err := gui.SetKeybinding("", b.key, gocui.ModNone, func(_ *gocui.Gui, _ *gocui.View) error {
				return w.onToggleRecordType(b.recType)
			}); err != nil {
				return err
			}
		}

		if err := gui.SetKeybinding("", 'l', gocui.ModNone, func(_ *gocui.Gui, _ *gocui.View) error {
			return w.onToggleLineEndings()
		}); err != nil {
			return err
		}
	}

	return w.redraw()
}

func (w *ChromeWidget) onMouseToggle(gui *gocui.Gui, view *gocui.View) error {
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

	lword := strings.ToLower(word)
	if lword == "eol" {
		return w.onToggleLineEndings()
	}
	if recType, ok := stringToRecordType(word); ok {
		return w.onToggleRecordType(recType)
	}
	return nil
}

func (w *ChromeWidget) onToggleLineEndings() error {
	w.showLineEndings = !w.showLineEndings
	if err := w.redraw(); err != nil {
		return err
	}
	return w.ctrl.ToggleLineEndings()
}

func (w *ChromeWidget) onToggleRecordType(recType gowarc.RecordType) error {
	w.recType = w.recType ^ recType

	if err := w.redraw(); err != nil {
		return err
	}

	return w.ctrl.ToggleRecordType(w.recType)
}

func (w *ChromeWidget) redraw() error {
	v, err := w.gui.View(w.viewName)
	if err != nil {
		return err
	}

	v.Clear()
	width, _ := v.Size()
	if width <= 0 {
		return nil
	}

	// Three degradation phases, evaluated by width:
	//   1. full labels + eoL + help  (right-aligned)
	//   2. full labels + eoL  (help hidden)
	//   3. full labels only, clipped at view boundary
	const help = "h: help"
	showHelp := width >= leftWidth+1+len(help)
	showEoL := width >= leftWidth

	tokColored, tokPlain := w.renderTokens()

	var leftColored, leftPlain string
	if showEoL {
		leftColored = tokColored + "  " + colorize("eoL", model.LineEndsColor, w.showLineEndings)
		leftPlain = tokPlain + "  eoL"
	} else {
		leftColored = tokColored
		leftPlain = tokPlain
	}

	if showHelp {
		gap := max(width-len(leftPlain)-len(help), 1)
		_, err = fmt.Fprintf(v, "%s%s%s%s", leftColored, model.SGRReset, strings.Repeat(" ", gap), help)
	} else {
		_, err = fmt.Fprint(v, leftColored+model.SGRReset)
	}
	return err
}

// tokenDef holds the static (label, color, record-type-bit) data for one filter chip.
// The enabled state is dynamic and computed per-draw from ChromeWidget.recType.
type tokenDef struct {
	full    string
	color   gocui.Attribute
	recType gowarc.RecordType
}

// tokenDefs is the ordered list of filter chips. Order determines display
// order. Declared as a package-level var so its width can be pre-computed.
var tokenDefs = []tokenDef{
	{full: "Error", color: model.ErrorColor, recType: model.ErrorRecordType},
	{full: "warcInfo", color: model.WarcInfoColor, recType: gowarc.Warcinfo},
	{full: "reQuest", color: model.RequestColor, recType: gowarc.Request},
	{full: "Response", color: model.ResponseColor, recType: gowarc.Response},
	{full: "Metadata", color: model.MetadataColor, recType: gowarc.Metadata},
	{full: "reVisit", color: model.RevisitColor, recType: gowarc.Revisit},
	{full: "reSource", color: model.ResourceColor, recType: gowarc.Resource},
	{full: "Continuation", color: model.ContinuationColor, recType: gowarc.Continuation},
	{full: "coNversion", color: model.ConversionColor, recType: gowarc.Conversion},
}

// leftWidth is the visible column width of the full left segment ("tokens  eoL"),
// pre-computed once at package init.
var leftWidth = func() int {
	w := 0
	for i, td := range tokenDefs {
		if i > 0 {
			w++ // "|" separator
		}
		w += len(td.full)
	}
	const eolSuffix = 2 + 3 // "  eoL"
	return w + eolSuffix
}()

// renderTokens returns the colorized and plain-text strings for all record-type
// chips joined by "|". Does not include the eoL suffix.
func (w *ChromeWidget) renderTokens() (string, string) {
	colored := make([]string, len(tokenDefs))
	plain := make([]string, len(tokenDefs))
	for i, td := range tokenDefs {
		colored[i] = colorize(td.full, td.color, w.recType&td.recType != 0)
		plain[i] = td.full
	}
	return strings.Join(colored, "|"), strings.Join(plain, "|")
}

var (
	fgBlack = model.EscapeFgColor(gocui.NewRGBColor(0, 0, 0))
	bgBlack = model.EscapeBgColor(gocui.NewRGBColor(0, 0, 0))
)

func colorize(str string, color gocui.Attribute, enabled bool) string {
	if enabled {
		return fmt.Sprintf("%s%s%s%s", model.EscapeBgColor(color), fgBlack, str, model.ColorReset)
	}
	return fmt.Sprintf("%s%s%s%s", model.EscapeFgColor(color), bgBlack, str, model.ColorReset)
}

func indexFunc(r rune) bool {
	// gocui pads view lines with null bytes; treat them as word boundaries.
	return r == ' ' || r == '|' || r == 0
}

func stringToRecordType(s string) (gowarc.RecordType, bool) {
	switch strings.ToLower(s) {
	case "error", "e":
		return model.ErrorRecordType, true
	case "warcinfo", "i":
		return gowarc.Warcinfo, true
	case "request", "q":
		return gowarc.Request, true
	case "response", "r":
		return gowarc.Response, true
	case "metadata", "m":
		return gowarc.Metadata, true
	case "revisit", "v":
		return gowarc.Revisit, true
	case "resource", "s":
		return gowarc.Resource, true
	case "continuation", "c":
		return gowarc.Continuation, true
	case "conversion", "n":
		return gowarc.Conversion, true
	default:
		return 0, false
	}
}
