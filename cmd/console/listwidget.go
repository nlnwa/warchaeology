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
	"time"
)

const (
	reading uint8 = iota
	done
)

type ListWidget struct {
	this                string
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

func NewListWidget(this, prev, next string,
	selectFunc func(g *gocui.Gui, widget *ListWidget),
	populateRecordsFunc func(g *gocui.Gui, ctx context.Context, finishedCb func(), widget *ListWidget, data interface{})) *ListWidget {

	return &ListWidget{
		this:                this,
		prev:                prev,
		next:                next,
		selectFunc:          selectFunc,
		populateRecordsFunc: populateRecordsFunc,
		state:               reading,
	}
}

func (w *ListWidget) Init(g *gocui.Gui, data interface{}) {
	if w.cancelFunc != nil {
		//if w.cancelRefreshFunc != nil {
		w.cancelRefreshFunc()
		w.cancelRefreshFunc = nil
		//}
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
	//w.finished, finishedCb = context.WithCancel(refreshCtx)
	w.finished, finishedCb = context.WithCancel(context.Background())
	w.records = nil
	w.filteredRecords = nil
	w.selected = -1
	go func() {
		w.populateRecordsFunc(g, ctx, finishedCb, w, data)
	}()
	w.update(g, refreshCtx, w.finished)
}

func (w *ListWidget) keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding(w.this, gocui.KeyArrowDown, gocui.ModNone, w.cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.this, gocui.KeyArrowUp, gocui.ModNone, w.cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.this, gocui.KeyEnter, gocui.ModNone, w.nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(w.this, gocui.KeyEsc, gocui.ModNone, w.prevView); err != nil {
		return err
	}
	return nil
}

func (w *ListWidget) prevView(g *gocui.Gui, v *gocui.View) error {
	state.curView = w.prev
	return nil
}

func (w *ListWidget) nextView(g *gocui.Gui, v *gocui.View) error {
	state.curView = w.next
	return nil
}

func (w *ListWidget) cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		w.scrollView(g, v, 1)
	}
	return nil
}

func (w *ListWidget) cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		w.scrollView(g, v, -1)
	}
	return nil
}

func (w *ListWidget) scrollView(g *gocui.Gui, v *gocui.View, dy int) error {
	if v != nil {
		v.Autoscroll = false
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		_, h := v.Size()
		sy := cy + oy + dy
		if sy < 0 || sy >= v.ViewLinesHeight()-1 {
			return nil
		}
		w.selected = sy

		cy += dy
		if cy < 0 {
			oy += dy
			cy = 0
		} else if cy >= h {
			oy += (cy - h) + 1
			cy -= (cy - h) + 1
		}

		if err := v.SetOrigin(ox, oy); err != nil {
			return err
		}
		if err := v.SetCursor(cx, cy); err != nil {
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
	v, err := g.View(w.this)
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
		v, err := gui.View(w.this)
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
						w.selected = 0
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
