package console

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/awesome-gocui/gocui"
	"github.com/nationallibraryofnorway/warchaeology/v4/cmd/internal/flag"
	"github.com/nlnwa/gowarc/v3"
	"github.com/spf13/viper"
)

type RecordWidget struct {
	name        string
	prev        string
	next        string
	headerView  string
	contentView string
	errorView   string
}

func NewRecordWidget(name, prev, next string) *RecordWidget {
	return &RecordWidget{
		name:        name,
		prev:        prev,
		next:        next,
		headerView:  name + "_header",
		contentView: name + "_content",
		errorView:   name + "_error",
	}
}

func (recordWidget *RecordWidget) Layout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	if state.fullView {
		if _, err := gui.SetView(recordWidget.contentView, 0, 0, maxX-1, maxY-2, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
		if view, err := gui.View(recordWidget.contentView); err == nil {
			view.FgColor = gocui.ColorDefault
			view.BgColor = gocui.ColorDefault
			view.SelBgColor = gocui.ColorWhite
			view.SelFgColor = gocui.ColorBlack
			view.Highlight = false
			view.Wrap = true
			view.Title = "WARC content [FULL]"
		}
		_ = gui.DeleteView(recordWidget.headerView)
		_ = gui.DeleteView(recordWidget.errorView)

		_ = recordWidget.addKeybindings(gui, recordWidget.contentView)
		return nil
	}

	dynamicColumnWidth := max(maxX-60, 51)

	if view, err := gui.SetView(recordWidget.headerView, 50, 10, dynamicColumnWidth, 30, gocui.BOTTOM|gocui.RIGHT); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.FgColor = gocui.ColorDefault
		view.BgColor = gocui.ColorDefault
		view.SelBgColor = gocui.ColorWhite
		view.SelFgColor = gocui.ColorBlack
		view.Highlight = false
		view.Wrap = true
		view.Title = "WARC header"
	}

	if view, err := gui.SetView(recordWidget.contentView, 50, 30, dynamicColumnWidth, maxY-2, gocui.TOP|gocui.RIGHT); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.FgColor = gocui.ColorDefault
		view.BgColor = gocui.ColorDefault
		view.SelBgColor = gocui.ColorWhite
		view.SelFgColor = gocui.ColorBlack
		view.Highlight = false
		view.Wrap = true
		view.Title = "WARC content [z]"
	}

	if view, err := gui.SetView(recordWidget.errorView, dynamicColumnWidth, 10, maxX-1, maxY-2, gocui.LEFT); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.FgColor = gocui.ColorRed
		view.BgColor = gocui.ColorDefault
		view.SelBgColor = gocui.ColorWhite
		view.SelFgColor = gocui.ColorBlack
		view.Highlight = false
		view.Wrap = true
		view.Title = "Errors"
	}

	_ = recordWidget.addKeybindings(gui, recordWidget.headerView)
	_ = recordWidget.addKeybindings(gui, recordWidget.contentView)
	_ = recordWidget.addKeybindings(gui, recordWidget.errorView)

	return nil
}

func (recordWidget *RecordWidget) addKeybindings(gui *gocui.Gui, widget string) error {
	if err := gui.SetKeybinding(widget, gocui.KeyArrowDown, gocui.ModNone, recordWidget.cursorDown); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.KeyArrowUp, gocui.ModNone, recordWidget.cursorUp); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.KeyHome, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			return recordWidget.scroll(view, -view.ViewLinesHeight())
		}
		return nil
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.KeyEnd, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			return recordWidget.scroll(view, view.ViewLinesHeight())
		}
		return nil
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.KeyPgdn, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			_, h := view.Size()
			h--
			return recordWidget.scroll(view, h)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.KeyPgup, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			_, h := view.Size()
			h--
			return recordWidget.scroll(view, -h)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.MouseWheelDown, gocui.ModNone, recordWidget.cursorDown); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.MouseWheelUp, gocui.ModNone, recordWidget.cursorUp); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.KeyEnter, gocui.ModNone, recordWidget.nextView); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.KeyEsc, gocui.ModNone, recordWidget.prevView); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widget, gocui.MouseLeft, gocui.ModNone, recordWidget.currentView); err != nil {
		return err
	}
	return nil
}

func (recordWidget *RecordWidget) readRecord(gui *gocui.Gui, widget *ListWidget) {
	selected := widget.filteredRecords[widget.selected].(record)
	if selected.errMsg != "" {
		recordWidget.renderErrors(gui, []error{fmt.Errorf("%s", selected.errMsg)})
		recordWidget.clearViews(gui)
		return
	}

	warcFileReader, err := gowarc.NewWarcFileReader(filepath.Join(state.dir, state.file), selected.offset,
		gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if err != nil {
		recordWidget.renderErrors(gui, []error{err})
		recordWidget.clearViews(gui)
		return
	}
	defer func() { _ = warcFileReader.Close() }()

	record, err := warcFileReader.Next()
	if err != nil {
		recordWidget.renderErrors(gui, []error{err})
		recordWidget.clearViews(gui)
		return
	}
	defer record.Close()

	warcRecord := record.WarcRecord
	offset := record.Offset
	validation := append([]error{}, record.Validation...)

	recordWidget.populateHeader(gui, warcRecord, offset, record.Size)
	recordWidget.populateContent(gui, warcRecord)

	digestValidation, err := warcRecord.ValidateDigest()
	validation = append(validation, digestValidation...)
	if err != nil {
		validation = append(validation, err)
	}
	recordWidget.renderErrors(gui, validation)
}

func (recordWidget *RecordWidget) populateHeader(gui *gocui.Gui, warcRecord gowarc.WarcRecord, offset int64, size int64) {
	view, err := gui.View(recordWidget.headerView)
	if err != nil {
		return
	}
	view.Clear()
	if size > 0 {
		view.Subtitle = fmt.Sprintf("Offset: %d, Size: %dB", offset, size)
	} else {
		view.Subtitle = fmt.Sprintf("Offset: %d", offset)
	}
	visibleLineEndingFilter := &visibleLineEndingFilter{view}
	_, _ = visibleLineEndingFilter.Write([]byte(warcRecord.Version().String() + "\r\n"))
	_, _ = warcRecord.WarcHeader().Write(visibleLineEndingFilter)
}

func (recordWidget *RecordWidget) populateContent(gui *gocui.Gui, warcRecord gowarc.WarcRecord) {
	view, err := gui.View(recordWidget.contentView)
	if err != nil {
		return
	}
	view.Clear()
	if _, ok := warcRecord.Block().(gowarc.PayloadBlock); ok {
		// Cache block when there is a defined payload so that we can get the payload size later.
		_ = warcRecord.Block().Cache()
	}
	visibleLineEndingFilter := &visibleLineEndingFilter{view}
	ioReader, err := warcRecord.Block().RawBytes()
	if err != nil {
		_, _ = fmt.Fprintf(view, "<error reading block: %v>\n", err)
		return
	}
	bytesWritten, err := io.Copy(visibleLineEndingFilter, ioReader)
	if err != nil {
		_, _ = fmt.Fprintf(view, "<error rendering block: %v>\n", err)
		return
	}

	subtitle := fmt.Sprintf("Blocksize: %d", bytesWritten)
	if payloadblock, ok := warcRecord.Block().(gowarc.PayloadBlock); ok {
		ioReader, err := payloadblock.PayloadBytes()
		if err != nil {
			view.Subtitle = subtitle
			return
		}
		bytesWritten, err := io.Copy(io.Discard, ioReader)
		if err != nil {
			view.Subtitle = subtitle
			return
		}
		subtitle = fmt.Sprintf("%s, PayloadSize: %d", subtitle, bytesWritten)
	}
	view.Subtitle = subtitle
}

func (recordWidget *RecordWidget) clearViews(gui *gocui.Gui) {
	for _, name := range []string{recordWidget.headerView, recordWidget.contentView} {
		if view, err := gui.View(name); err == nil {
			view.Clear()
		}
	}
}

func (recordWidget *RecordWidget) renderErrors(gui *gocui.Gui, errs []error) {
	view, err := gui.View(recordWidget.errorView)
	if err != nil {
		return
	}
	view.Clear()
	if len(errs) == 0 {
		_, _ = fmt.Fprintln(view, "[]")
		return
	}
	for _, validationErr := range errs {
		if validationErr != nil {
			_, _ = fmt.Fprintf(view, "%v\n", validationErr)
		}
	}
}

type visibleLineEndingFilter struct {
	ioWriter io.Writer
}

func (visiblelineEndingFilter *visibleLineEndingFilter) Write(content []byte) (bytesWritten int, err error) {
	content = colorizeReplaceAll(content, []byte("\r"), []byte("\\r"))
	content = colorizeReplaceAll(content, []byte("\n"), []byte("\\n\n"))
	return visiblelineEndingFilter.ioWriter.Write(content)
}

func colorizeReplaceAll(source, old, replacement []byte) []byte {
	reset := escapeFgColor(gocui.ColorDefault)
	coloredString := fmt.Sprintf("%s%s%s", escapeFgColor(gocui.ColorGreen), replacement, reset)
	return bytes.ReplaceAll(source, old, []byte(coloredString))
}

func (recordWidget *RecordWidget) cursorDown(gui *gocui.Gui, view *gocui.View) error {
	return recordWidget.scroll(view, 1)
}

func (recordWidget *RecordWidget) cursorUp(gui *gocui.Gui, view *gocui.View) error {
	return recordWidget.scroll(view, -1)
}

func (recordWidget *RecordWidget) scroll(view *gocui.View, ScrollDelta int) error {
	if view != nil {
		_, viewHeight := view.Size()
		contentHeight := view.ViewLinesHeight()
		if viewHeight >= contentHeight {
			return nil
		}

		originX, originY := view.Origin()
		scrollDestinationY := max(originY+ScrollDelta, 0)
		scrollDestinationY = min(contentHeight-viewHeight, min(scrollDestinationY, contentHeight-viewHeight))
		_ = view.SetOrigin(originX, scrollDestinationY)
	}
	return nil
}

func (recordWidget *RecordWidget) prevView(gui *gocui.Gui, view *gocui.View) error {
	if state.fullView && state.curView == recordWidget.contentView {
		state.curView = recordWidget.prev
		return nil
	}

	switch state.curView {
	case recordWidget.errorView:
		state.curView = recordWidget.contentView
	case recordWidget.contentView:
		state.curView = recordWidget.headerView
	case recordWidget.headerView:
		state.curView = recordWidget.prev
	}
	return nil
}

func (recordWidget *RecordWidget) nextView(gui *gocui.Gui, view *gocui.View) error {
	if state.fullView && state.curView == recordWidget.contentView {
		state.curView = recordWidget.next
		return nil
	}

	switch state.curView {
	case recordWidget.headerView:
		state.curView = recordWidget.contentView
	case recordWidget.contentView:
		state.curView = recordWidget.errorView
	case recordWidget.errorView:
		state.curView = recordWidget.next
	}
	return nil
}

func (recordWidget *RecordWidget) currentView(gui *gocui.Gui, view *gocui.View) error {
	state.curView = view.Name()
	return nil
}

func (recordWidget *RecordWidget) toggleContentFullscreen(gui *gocui.Gui, view *gocui.View) error {
	state.fullView = !state.fullView
	if state.fullView {
		state.curView = recordWidget.contentView
	}
	return nil
}
