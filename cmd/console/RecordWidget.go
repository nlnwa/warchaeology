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
 *
 */

package console

import (
	"bytes"
	"fmt"
	"io"

	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/flag"
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

func (w *RecordWidget) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	dynamicColumnWidth := maxX - 60
	if dynamicColumnWidth < 51 {
		dynamicColumnWidth = 51
	}

	if v, err := g.SetView(w.headerView, 50, 10, dynamicColumnWidth, 30, gocui.BOTTOM|gocui.RIGHT); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorDefault
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = false
		v.Wrap = true
		v.Title = "WARC header"
	}

	if v, err := g.SetView(w.contentView, 50, 30, dynamicColumnWidth, maxY-2, gocui.TOP|gocui.RIGHT); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorDefault
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = false
		v.Wrap = true
		v.Title = "WARC content"
	}

	if v, err := g.SetView(w.errorView, dynamicColumnWidth, 10, maxX-1, maxY-2, gocui.LEFT); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorRed
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = false
		v.Wrap = true
		v.Title = "Errors"
	}

	_ = w.addKeybindings(g, w.headerView)
	_ = w.addKeybindings(g, w.contentView)
	_ = w.addKeybindings(g, w.errorView)

	return nil
}

func (w *RecordWidget) addKeybindings(g *gocui.Gui, widget string) error {
	if err := g.SetKeybinding(widget, gocui.KeyArrowDown, gocui.ModNone, w.cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.KeyArrowUp, gocui.ModNone, w.cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.KeyHome, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			return w.scroll(view, -view.ViewLinesHeight())
		}
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.KeyEnd, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			return w.scroll(view, view.ViewLinesHeight())
		}
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.KeyPgdn, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			_, h := view.Size()
			h--
			return w.scroll(view, h)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.KeyPgup, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			_, h := view.Size()
			h--
			return w.scroll(view, -h)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.MouseWheelDown, gocui.ModNone, w.cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.MouseWheelUp, gocui.ModNone, w.cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.KeyEnter, gocui.ModNone, w.nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.KeyEsc, gocui.ModNone, w.prevView); err != nil {
		return err
	}
	if err := g.SetKeybinding(widget, gocui.MouseLeft, gocui.ModNone, w.currentView); err != nil {
		return err
	}
	return nil
}

func (w *RecordWidget) readRecord(g *gocui.Gui, widget *ListWidget) {
	r, err := gowarc.NewWarcFileReader(state.dir+"/"+state.file, widget.filteredRecords[widget.selected].(record).offset,
		gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if err != nil {
		panic(err)
	}
	defer func() { _ = r.Close() }()

	rec, offset, val, err := r.Next()
	if err != nil {
		panic(err)
	}
	defer func() { _ = rec.Close() }()

	w.poopulateHeader(g, rec, offset)
	w.poopulateContent(g, rec)

	if err := rec.ValidateDigest(val); err != nil {
		*val = append(*val, err)
	}

	if err := rec.Close(); err != nil {
		*val = append(*val, err)
	}

	ev, err := g.View(w.errorView)
	if err != nil {
		panic(err)
	}
	ev.Clear()
	_, _ = fmt.Fprintf(ev, "%s\n", val)
}

func (w *RecordWidget) poopulateHeader(g *gocui.Gui, rec gowarc.WarcRecord, offset int64) {
	view, err := g.View(w.headerView)
	if err != nil {
		panic(err)
	}
	view.Clear()
	view.Subtitle = fmt.Sprintf("Offset: %d", offset)
	f := &visibleLineEndingFilter{view}
	_, _ = f.Write([]byte(rec.Version().String() + "\r\n"))
	_, _ = rec.WarcHeader().Write(f)
}

func (w *RecordWidget) poopulateContent(g *gocui.Gui, rec gowarc.WarcRecord) {
	view, err := g.View(w.contentView)
	if err != nil {
		panic(err)
	}
	view.Clear()
	if _, ok := rec.Block().(gowarc.PayloadBlock); ok {
		// Cache block when there is a defined payload so that we can get the payload size later.
		_ = rec.Block().Cache()
	}
	f := &visibleLineEndingFilter{view}
	rr, err := rec.Block().RawBytes()
	if err != nil {
		panic(err)
	}
	content, err := io.ReadAll(rr)
	if err != nil {
		panic(err)
	}
	_, _ = f.Write(content)

	subtitle := fmt.Sprintf("Blocksize: %d", len(content))
	if p, ok := rec.Block().(gowarc.PayloadBlock); ok {
		rr, err := p.PayloadBytes()
		if err != nil {
			panic(err)
		}
		n, err := io.Copy(io.Discard, rr)
		if err != nil {
			panic(err)
		}
		subtitle = fmt.Sprintf("%s, PayloadSize: %d", subtitle, n)
	}
	view.Subtitle = subtitle
}

type visibleLineEndingFilter struct {
	w io.Writer
}

func (f *visibleLineEndingFilter) Write(p []byte) (n int, err error) {
	p = colorizeReplaceAll(p, []byte("\r"), []byte("\\r"))
	p = colorizeReplaceAll(p, []byte("\n"), []byte("\\n\n"))
	return f.w.Write(p)
}

func colorizeReplaceAll(source, old, replacement []byte) []byte {
	reset := escapeFgColor(gocui.ColorDefault)
	v := fmt.Sprintf("%s%s%s", escapeFgColor(gocui.ColorGreen), replacement, reset)
	return bytes.ReplaceAll(source, old, []byte(v))
}

func (w *RecordWidget) cursorDown(g *gocui.Gui, v *gocui.View) error {
	return w.scroll(v, 1)
}

func (w *RecordWidget) cursorUp(g *gocui.Gui, v *gocui.View) error {
	return w.scroll(v, -1)
}

func (w *RecordWidget) scroll(v *gocui.View, delta int) error {
	if v != nil {
		_, viewHeight := v.Size()
		contentHeight := v.ViewLinesHeight()
		if viewHeight >= contentHeight {
			return nil
		}

		ox, oy := v.Origin()
		ny := oy + delta
		if ny < 0 {
			ny = 0
		}
		if contentHeight-viewHeight < ny {
			ny = contentHeight - viewHeight
		}
		_ = v.SetOrigin(ox, ny)
	}
	return nil
}

func (w *RecordWidget) prevView(g *gocui.Gui, v *gocui.View) error {
	switch state.curView {
	case w.errorView:
		state.curView = w.contentView
	case w.contentView:
		state.curView = w.headerView
	case w.headerView:
		state.curView = w.prev
	}
	return nil
}

func (w *RecordWidget) nextView(g *gocui.Gui, v *gocui.View) error {
	switch state.curView {
	case w.headerView:
		state.curView = w.contentView
	case w.contentView:
		state.curView = w.errorView
	case w.errorView:
		state.curView = w.next
	}
	return nil
}

func (w *RecordWidget) currentView(g *gocui.Gui, v *gocui.View) error {
	state.curView = v.Name()
	return nil
}
