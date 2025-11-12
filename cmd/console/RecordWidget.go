package console

import (
	"bytes"
	"fmt"
	"io"

	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v4/cmd/internal/flag"
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
		view.Title = "WARC content"
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
	warcFileReader, err := gowarc.NewWarcFileReader(state.dir+"/"+state.file, widget.filteredRecords[widget.selected].(record).offset,
		gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if err != nil {
		panic(err)
	}
	defer func() { _ = warcFileReader.Close() }()

	warcRecord, offset, validation, err := warcFileReader.Next()
	if err != nil {
		panic(err)
	}
	defer func() { _ = warcRecord.Close() }()

	recordWidget.poopulateHeader(gui, warcRecord, offset)
	recordWidget.poopulateContent(gui, warcRecord)

	if err := warcRecord.ValidateDigest(validation); err != nil {
		*validation = append(*validation, err)
	}

	if err := warcRecord.Close(); err != nil {
		*validation = append(*validation, err)
	}

	view, err := gui.View(recordWidget.errorView)
	if err != nil {
		panic(err)
	}
	view.Clear()
	_, _ = fmt.Fprintf(view, "%s\n", validation)
}

func (recordWidget *RecordWidget) poopulateHeader(gui *gocui.Gui, warcRecord gowarc.WarcRecord, offset int64) {
	view, err := gui.View(recordWidget.headerView)
	if err != nil {
		panic(err)
	}
	view.Clear()
	view.Subtitle = fmt.Sprintf("Offset: %d", offset)
	visibleLineEndingFilter := &visibleLineEndingFilter{view}
	_, _ = visibleLineEndingFilter.Write([]byte(warcRecord.Version().String() + "\r\n"))
	_, _ = warcRecord.WarcHeader().Write(visibleLineEndingFilter)
}

func (recordWidget *RecordWidget) poopulateContent(gui *gocui.Gui, warcRecord gowarc.WarcRecord) {
	view, err := gui.View(recordWidget.contentView)
	if err != nil {
		panic(err)
	}
	view.Clear()
	if _, ok := warcRecord.Block().(gowarc.PayloadBlock); ok {
		// Cache block when there is a defined payload so that we can get the payload size later.
		_ = warcRecord.Block().Cache()
	}
	visibleLineEndingFilter := &visibleLineEndingFilter{view}
	ioReader, err := warcRecord.Block().RawBytes()
	if err != nil {
		panic(err)
	}
	content, err := io.ReadAll(ioReader)
	if err != nil {
		panic(err)
	}
	_, _ = visibleLineEndingFilter.Write(content)

	subtitle := fmt.Sprintf("Blocksize: %d", len(content))
	if payloadblock, ok := warcRecord.Block().(gowarc.PayloadBlock); ok {
		ioReader, err := payloadblock.PayloadBytes()
		if err != nil {
			panic(err)
		}
		bytesWritten, err := io.Copy(io.Discard, ioReader)
		if err != nil {
			panic(err)
		}
		subtitle = fmt.Sprintf("%s, PayloadSize: %d", subtitle, bytesWritten)
	}
	view.Subtitle = subtitle
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
