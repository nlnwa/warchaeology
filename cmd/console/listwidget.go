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
 */

package console

import (
	"context"
	"fmt"
	"github.com/awesome-gocui/gocui"
	"runtime/debug"
	"strings"
	"time"
)

const (
	reading uint8 = iota
	done
)

type ListWidget struct {
	name                string
	prev                string
	next                string
	selected            int
	selectFunc          func(g *gocui.Gui, widget *ListWidget)
	populateRecordsFunc func(g *gocui.Gui, ctx context.Context, finishedCb func(), widget *ListWidget, data interface{})
	filterFunc          func(interface{}) bool
	records             []interface{}
	filteredRecords     []interface{}
	cancelFunc          context.CancelFunc
	cancelRefreshFunc   context.CancelFunc
	finished            context.Context
	state               uint8
}

func NewListWidget(name string, prev, next string,
	selectFunc func(g *gocui.Gui, widget *ListWidget),
	populateRecordsFunc func(g *gocui.Gui, ctx context.Context, finishedCb func(), widget *ListWidget, data interface{})) *ListWidget {

	return &ListWidget{
		name:                name,
		prev:                prev,
		next:                next,
		selectFunc:          selectFunc,
		populateRecordsFunc: populateRecordsFunc,
		state:               reading,
	}
}

func (w *ListWidget) Init(g *gocui.Gui, data interface{}) {
	if w.cancelFunc != nil {
		w.cancelRefreshFunc()
		w.cancelRefreshFunc = nil
		w.cancelFunc()
		w.cancelFunc = nil
		<-w.finished.Done()
		w.finished = nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.cancelFunc = cancel
	var finishedCb func()
	var refreshCtx context.Context
	refreshCtx, w.cancelRefreshFunc = context.WithCancel(context.Background())
	w.finished, finishedCb = context.WithCancel(context.Background())
	w.records = nil
	w.filteredRecords = nil
	w.selected = -1
	go func() {
		w.populateRecordsFunc(g, ctx, finishedCb, w, data)
	}()
	w.update(g, refreshCtx, w.finished)
}

func (w *ListWidget) Layout(g *gocui.Gui) error {
	if v, err := g.SetView(w.name, 0, 0, 1, 1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorGreen
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = false
		v.Autoscroll = false
		v.Title = w.name
	}
	if err := g.SetKeybinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.KeyHome, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			return w.selectLine(g, view, 0)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.KeyEnd, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			return w.selectLine(g, view, len(w.filteredRecords)-1)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.KeyPgdn, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			_, h := view.Size()
			h--
			return w.selectLine(g, view, w.selected+h)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.KeyPgup, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view != nil {
			_, h := view.Size()
			h--
			return w.selectLine(g, view, w.selected-h)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, 'f', gocui.ModNone, w.search); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.KeyEnter, gocui.ModNone, w.nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.KeyEsc, gocui.ModNone, w.prevView); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.MouseLeft, gocui.ModNone, w.currentView); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.MouseRelease, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.MouseWheelDown, gocui.ModNone, w.cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.name, gocui.MouseWheelUp, gocui.ModNone, w.cursorUp); err != nil {
		return err
	}
	return nil
}

func (w *ListWidget) search(g *gocui.Gui, parent *gocui.View) error {
	maxX, maxY := g.Size()
	var v *gocui.View
	var err error
	if v, err = g.SetView("search", maxX/2-35, maxY/2-1, maxX/2+35, maxY/2+1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorDefault
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = false
		v.Autoscroll = false
		v.Title = "Find in " + parent.Title
		v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝'}
		v.Editable = true
		v.KeybindOnEdit = false
		v.Editor = &SearchEditor{g, w, parent}
		g.Cursor = true
	}
	state.modalView = "search"
	return nil
}

type SearchEditor struct {
	g      *gocui.Gui
	w      *ListWidget
	parent *gocui.View
}

func (s *SearchEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if ch != 0 && mod == 0 {
		v.EditWrite(ch)
		_ = s.w.searchLine(s.g, s.parent, v.ViewBuffer())
		return
	}

	switch key {
	case gocui.KeySpace:
		v.EditWrite(' ')
		_ = s.w.searchLine(s.g, s.parent, v.ViewBuffer())
	case gocui.KeyBackspace, gocui.KeyBackspace2:
		v.EditDelete(true)
		_ = s.w.searchLine(s.g, s.parent, v.ViewBuffer())
	case gocui.KeyDelete:
		v.EditDelete(false)
		_ = s.w.searchLine(s.g, s.parent, v.ViewBuffer())
	case gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0)
	case gocui.KeyArrowRight:
		v.MoveCursor(1, 0)
	case gocui.KeyEsc:
		// If not here the esc key will act like the KeySpace
		fallthrough
	case gocui.KeyEnter:
		s.g.DeleteView(v.Name())
		state.modalView = ""
		s.g.Cursor = false
	default:
	}
}

func (w *ListWidget) prevView(g *gocui.Gui, v *gocui.View) error {
	state.curView = w.prev
	return nil
}

func (w *ListWidget) nextView(g *gocui.Gui, v *gocui.View) error {
	state.curView = w.next
	return nil
}

func (w *ListWidget) currentView(g *gocui.Gui, v *gocui.View) error {
	state.curView = v.Name()
	_, oy := v.Origin()
	_, cy := v.Cursor()
	newSelect := cy + oy
	w.selectLine(g, v, newSelect)
	return nil
}

func (w *ListWidget) cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		w.selectLine(g, v, w.selected+1)
	}
	return nil
}

func (w *ListWidget) cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		w.selectLine(g, v, w.selected-1)
	}
	return nil
}

func (w *ListWidget) searchLine(g *gocui.Gui, v *gocui.View, s string) error {
	if v != nil {
		for i, id := range v.ViewBufferLines() {
			if strings.Contains(id, s) {
				return w.selectLine(g, v, i)
			}
		}
	}
	return nil
}

func (w *ListWidget) selectLine(g *gocui.Gui, v *gocui.View, selected int) error {
	if v != nil {
		ox, oy := v.Origin()
		_, h := v.Size()
		if selected < 0 {
			selected = 0
		}
		if selected >= v.ViewLinesHeight()-1 {
			selected = v.ViewLinesHeight() - 2
		}
		if selected == w.selected {
			return nil
		}
		if w.selected != -1 {
			_ = v.SetLine(w.selected, fmt.Sprintf("%s", w.filteredRecords[w.selected]))
		}
		w.selected = selected
		_ = v.SetHighlight(w.selected, true)

		cy := w.selected - oy
		if cy < 0 {
			oy += cy
			cy = 0
		} else if cy >= h {
			oy += (cy - h) + 1
			cy -= (cy - h) + 1
		}

		if err := v.SetOrigin(ox, oy); err != nil {
			return err
		}
		if err := v.SetCursor(0, cy); err != nil {
			return err
		}

		if l, err := v.Line(w.selected); err != nil || l == "" {
			return nil
		} else {
			v.Subtitle = fmt.Sprintf("%d/%d", w.selected+1, len(w.filteredRecords))
			if len(w.filteredRecords) > 0 {
				w.selectFunc(g, w)
			}
		}
	}
	return nil
}

func (w *ListWidget) update(g *gocui.Gui, ctx context.Context, finished context.Context) {
	time.Sleep(100 * time.Millisecond)
	v, err := g.View(w.name)
	if err != nil {
		panic(err)
	}
	v.Clear()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered in f %s\n%s", r, debug.Stack())
			}
		}()

		l := 0
		for {
			select {
			case <-ctx.Done():
				<-finished.Done()
				return
			case <-finished.Done():
				if l < len(w.records) {
					rec := w.records[l:]
					w.upd(g, ctx, rec)
				}
				return
			default:
				if l < len(w.records) {
					rec := w.records[l:]
					l = l + len(rec)
					w.upd(g, ctx, rec)
				}
			}
		}
	}()
}

func (w *ListWidget) upd(g *gocui.Gui, ctx context.Context, rec []interface{}) {
	g.UpdateAsync(func(gui *gocui.Gui) error {
		v, err := gui.View(w.name)
		if err != nil {
			return err
		}

		for _, r := range rec {
			select {
			case <-ctx.Done():
				return nil
			default:
				if w.filterFunc == nil || w.filterFunc(r) {
					w.filteredRecords = append(w.filteredRecords, r)
					fmt.Fprintf(v, "%s\n", r)
					if len(w.filteredRecords) == 1 {
						w.selectLine(g, v, 0)
						w.selectFunc(g, w)
					}
				}
			}
		}
		v.Subtitle = fmt.Sprintf("%d/%d", w.selected+1, len(w.filteredRecords))
		return nil
	})
}

func (w *ListWidget) refreshFilter(g *gocui.Gui, v *gocui.View) error {
	if w.cancelRefreshFunc != nil {
		w.cancelRefreshFunc()
		w.cancelRefreshFunc = nil
	}
	var ctx context.Context
	ctx, w.cancelRefreshFunc = context.WithCancel(context.Background())
	v.Clear()
	w.filteredRecords = nil
	w.selected = -1

	go func() {
		l := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.finished.Done():
				if l < len(w.records) {
					rec := w.records[l:]
					w.upd(g, ctx, rec)
				}
				return
			default:
				if l < len(w.records) {
					rec := w.records[l:]
					l = l + len(rec)
					w.upd(g, ctx, rec)
				}
			}
		}
	}()
	return nil
}
