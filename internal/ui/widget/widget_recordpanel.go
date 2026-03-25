package widget

import (
	"bytes"
	"fmt"
	"io"

	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc/v3"
)

var (
	DoubleFrame = []rune{'═', '║', '╔', '╗', '╚', '╝'}
	SingleFrame = []rune{'─', '│', '┌', '┐', '└', '┘'}
)

// RecordPanelWidget manages the layout of the main record pane (header+block)
// and an optional Errors sub-view.
//
// Active views share the same FrameRunes set so that shared-border corners
// render as consistent box-drawing characters.
//
// The Errors view is only present when there are validation errors to show.
// It will be created on the next Layout call after RenderErrors is called
// with a non-empty error slice, and deleted when Clear is called.
type RecordPanelWidget struct {
	gui         *gocui.Gui
	headerName  string
	errorsName  string
	headerTitle string

	header *RecordWidget
	errors *RecordWidget

	// hasErrors controls whether the errors sub-view is included in the layout.
	hasErrors bool
	// errorsOnly hides Header and Content, showing only the errors view.
	errorsOnly bool
	// errorsTitle overrides the default "Validation Errors" title when set.
	errorsTitle string
	// pendingErrs holds errors written by RenderErrors that have not yet been
	// flushed to the errors view (because it may not exist yet).
	pendingErrs []error

	// showLineEndings controls whether CR/LF are rendered as visible escape sequences.
	showLineEndings bool
}

// ToggleLineEndings flips the visible line endings mode on/off.
func (p *RecordPanelWidget) ToggleLineEndings() {
	p.showLineEndings = !p.showLineEndings
}

func NewRecordPanelWidget(gui *gocui.Gui, headerName, errorsName string, ctrl Controller) *RecordPanelWidget {
	return &RecordPanelWidget{
		gui:         gui,
		headerName:  headerName,
		errorsName:  errorsName,
		headerTitle: "WARC",
		header:      NewRecordWidget(headerName, ctrl),
		errors:      NewRecordWidget(errorsName, ctrl),
	}
}

// ViewNames returns the ordered list of focusable view names in this panel.
// The errors view is only included when errors are present.
func (p *RecordPanelWidget) ViewNames() []string {
	if p.errorsOnly {
		return []string{p.errorsName}
	}
	names := []string{p.headerName}
	if p.hasErrors {
		names = append(names, p.errorsName)
	}
	return names
}

// Layout positions, creates, and initialises all active sub-views. It must be
// called every render frame with the panel's allocated screen rectangle.
// focused is the currently focused view name (used to select frame runes).
func (p *RecordPanelWidget) Layout(gui *gocui.Gui, x0, y0, x1, y1 int, focused string) error {

	type viewSpec struct {
		name   string
		x0, y0 int
		x1, y1 int
		widget *RecordWidget
	}

	var specs []viewSpec
	if p.errorsOnly {
		// Errors-only: keep main view in place so content remains cached, with
		// errors overlaying the full panel on top.
		specs = []viewSpec{
			{p.headerName, x0, y0, x1, y1, p.header},
			{p.errorsName, x0, y0, x1, y1, p.errors},
		}
	} else if p.hasErrors {
		// Errors panel below the main record pane, taking ~1/4 of height.
		errHeight := max(4, (y1-y0)/4)
		mainSplit := y1 - errHeight
		specs = []viewSpec{
			{p.headerName, x0, y0, x1, mainSplit, p.header},
			{p.errorsName, x0, mainSplit + 1, x1, y1, p.errors},
		}
	} else {
		// No errors: main record pane fills the full area.
		_ = gui.DeleteView(p.errorsName)
		specs = []viewSpec{
			{p.headerName, x0, y0, x1, y1, p.header},
		}
	}

	for _, s := range specs {
		if err := s.widget.Layout(gui, s.x0, s.y0, s.x1, s.y1); err != nil {
			return err
		}
		v, err := gui.View(s.name)
		if err != nil {
			return err
		}
		// Each view gets its own frame runes based on whether it is focused.
		if s.name == focused {
			v.FrameRunes = DoubleFrame
		} else {
			v.FrameRunes = SingleFrame
		}

		// Set stable titles for active views.
		switch s.name {
		case p.headerName:
			v.Title = p.headerTitle
		case p.errorsName:
			if p.errorsTitle != "" {
				v.Title = p.errorsTitle
			} else {
				v.Title = "Validation Errors"
			}
		}

		// Flush any pending error text now that the view exists.
		if s.name == p.errorsName && p.pendingErrs != nil {
			v.Clear()
			_ = v.SetOrigin(0, 0)
			for _, e := range p.pendingErrs {
				_, _ = fmt.Fprintf(v, "%v\n", e)
			}
			if p.errorsOnly {
				v.Subtitle = ""
			} else {
				v.Subtitle = fmt.Sprintf("%d errors", len(p.pendingErrs))
			}
			p.pendingErrs = nil
		}
	}

	// Ensure errors view is on top when it overlays header/content.
	if p.errorsOnly {
		_, _ = gui.SetViewOnTop(p.errorsName)
	}

	return nil
}

// PopulateHeader updates the main-pane title. Main-pane content is rendered by
// PopulateContent so header and block can be shown in a single scrollable view.
func (p *RecordPanelWidget) PopulateHeader(gui *gocui.Gui, rec gowarc.Record) {
	p.headerTitle = fmt.Sprint(rec.WarcRecord.Version())
}

// PopulateContent writes header+block into the main record view, rendering
// CR/LF as visible escape sequences when enabled.
func (p *RecordPanelWidget) PopulateContent(gui *gocui.Gui, rec gowarc.Record) {
	v, err := gui.View(p.headerName)
	if err != nil {
		return
	}
	v.Clear()
	_ = v.SetOrigin(0, 0)
	p.headerTitle = fmt.Sprint(rec.WarcRecord.Version())

	warcRecord := rec.WarcRecord
	if _, ok := warcRecord.Block().(gowarc.PayloadBlock); ok {
		_ = warcRecord.Block().Cache()
	}

	blockSize := warcRecord.Block().Size()

	r, err := warcRecord.Block().RawBytes()
	if err != nil {
		_, _ = warcRecord.WarcHeader().Write(v)
		_, _ = fmt.Fprint(v, "\r\n")
		_, _ = fmt.Fprintf(v, "<error reading block: %v>\n", err)
		return
	}

	var src io.Reader
	if p.showLineEndings {
		var buf bytes.Buffer
		_, _ = warcRecord.WarcHeader().Write(&buf)
		_, _ = buf.WriteString("\r\n") // header-block separator
		_, _ = io.Copy(&buf, r)
		_, _ = buf.WriteString("\r\n\r\n") // end-of-record separator
		src = newVisibleFilter(&buf)
	} else {
		src = io.MultiReader(warcHeaderAndSeparatorReader(warcRecord), r)
	}
	if _, err := io.Copy(v, src); err != nil {
		_, _ = fmt.Fprintf(v, "<error rendering block: %v>\n", err)
		return
	}

	subtitle := fmt.Sprintf("Record size: %dB, Block size: %dB", rec.Size, blockSize)
	if pb, ok := warcRecord.Block().(gowarc.PayloadBlock); ok {
		if pr, err := pb.PayloadBytes(); err == nil {
			if ps, err := io.Copy(io.Discard, pr); err == nil {
				subtitle = fmt.Sprintf("%s, Payload size: %dB", subtitle, ps)
			}
		}
	}
	v.Subtitle = subtitle
}

func warcHeaderAndSeparatorReader(warcRecord gowarc.WarcRecord) io.Reader {
	var header bytes.Buffer
	_, _ = warcRecord.WarcHeader().Write(&header)
	_, _ = header.WriteString("\r\n")
	return &header
}

// RenderReadError shows a single read error in a full-panel error view,
// hiding the Header and Content views.
func (p *RecordPanelWidget) RenderReadError(gui *gocui.Gui, err error) {
	p.hasErrors = true
	p.errorsOnly = true
	p.errorsTitle = "Read Error"
	p.pendingErrs = []error{err}
	if v, vErr := gui.View(p.errorsName); vErr == nil {
		v.Clear()
		_ = v.SetOrigin(0, 0)
		_, _ = fmt.Fprintf(v, "%v\n", err)
		v.Subtitle = ""
		p.pendingErrs = nil
	}
}

// RenderErrors stores validation errors for display and marks the errors view
// as needed. The errors view is created (or updated) on the next Layout call.
func (p *RecordPanelWidget) RenderErrors(gui *gocui.Gui, errs []error) {
	p.hasErrors = len(errs) > 0
	if !p.hasErrors {
		p.pendingErrs = nil
		return
	}
	p.pendingErrs = errs
	// If the view already exists (errors were already shown), write immediately.
	if v, err := gui.View(p.errorsName); err == nil {
		v.Clear()
		_ = v.SetOrigin(0, 0)
		for _, e := range errs {
			_, _ = fmt.Fprintf(v, "%v\n", e)
		}
		v.Subtitle = fmt.Sprintf("%d errors", len(errs))
		p.pendingErrs = nil
	}
}

// Clear empties all sub-views and hides the errors panel.
func (p *RecordPanelWidget) Clear(gui *gocui.Gui) {
	p.hasErrors = false
	p.errorsOnly = false
	p.errorsTitle = ""
	p.headerTitle = "WARC"
	p.pendingErrs = nil
	for _, name := range []string{p.headerName} {
		if v, err := gui.View(name); err == nil {
			v.Clear()
			v.Subtitle = ""
			_ = v.SetOrigin(0, 0)
		}
	}
	// Errors view will be deleted on the next Layout call.
}
