package widget

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/awesome-gocui/gocui"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/ui/model"
)

type Entry[T any] struct {
	Item    T
	Display string
	Search  string
}

type ListWidget[T any] struct {
	gui      *gocui.Gui
	viewName string

	items   []Entry[T]
	visible []int

	filter   string
	selected int
	offset   int

	makeText     func(T) string
	makeColor    func(T) string // optional; returns a pre-escaped fg color sequence
	makeSubtitle func(T) string // optional; extra info appended to "N/M" subtitle
	matches      func(item T, search string) bool

	nav Controller

	onSelectionChanged func(item T, ok bool)

	dirty       bool
	flushQueued atomic.Bool
	mu          sync.Mutex
}

func NewListWidget[T any](
	gui *gocui.Gui,
	viewName string,
	makeText func(T) string,
	nav Controller,
	onSelectionChanged func(item T, ok bool),
) *ListWidget[T] {
	w := &ListWidget[T]{
		gui:                gui,
		viewName:           viewName,
		makeText:           makeText,
		nav:                nav,
		onSelectionChanged: onSelectionChanged,
		selected:           -1,
		visible:            make([]int, 0),
		items:              make([]Entry[T], 0),
		dirty:              true,
	}

	return w
}

func (w *ListWidget[T]) Layout(gui *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := gui.SetView(w.viewName, x0, y0, x1, y1, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}
	if errors.Is(err, gocui.ErrUnknownView) {
		v.Title = w.viewName
		v.Highlight = false
		v.Autoscroll = false
		keyBindings := []model.KeyBinding{
			{Key: gocui.KeyArrowDown, Mod: gocui.ModNone, Fn: w.onCursorDown},
			{Key: gocui.KeyArrowUp, Mod: gocui.ModNone, Fn: w.onCursorUp},
			{Key: gocui.KeyHome, Mod: gocui.ModNone, Fn: w.onGotoStart},
			{Key: gocui.KeyEnd, Mod: gocui.ModNone, Fn: w.onGotoEnd},
			{Key: gocui.KeyPgdn, Mod: gocui.ModNone, Fn: w.onPageDown},
			{Key: gocui.KeyPgup, Mod: gocui.ModNone, Fn: w.onPageUp},
			{Key: 'f', Mod: gocui.ModNone, Fn: w.onSearch},
			{Key: '/', Mod: gocui.ModNone, Fn: w.onSearch},
			{Key: gocui.MouseLeft, Mod: gocui.ModNone, Fn: w.onCurrentView},
			{Key: gocui.MouseWheelDown, Mod: gocui.ModNone, Fn: w.onCursorDown},
			{Key: gocui.MouseWheelUp, Mod: gocui.ModNone, Fn: w.onCursorUp},
		}
		for _, kb := range keyBindings {
			if err := gui.SetKeybinding(w.viewName, kb.Key, kb.Mod, kb.Fn); err != nil {
				return err
			}
		}
		w.requestFlush()
	}
	return nil
}

func (w *ListWidget[T]) SetPredicate(matches func(T, string) bool) {
	w.mu.Lock()
	w.matches = matches
	emitClear, emitNew, item := w.rebuildAndEmitLocked()
	w.mu.Unlock()
	w.dispatchEmit(emitClear, emitNew, item)
}

func (w *ListWidget[T]) SetColorizer(makeColor func(T) string) {
	w.mu.Lock()
	w.makeColor = makeColor
	w.mu.Unlock()
}

func (w *ListWidget[T]) SetSubtitleMaker(makeSubtitle func(T) string) {
	w.mu.Lock()
	w.makeSubtitle = makeSubtitle
	w.mu.Unlock()
}

func (w *ListWidget[T]) SetFilter(q string) {
	w.mu.Lock()
	w.filter = normalize(q)
	emitClear, emitNew, item := w.rebuildAndEmitLocked()
	w.mu.Unlock()
	w.dispatchEmit(emitClear, emitNew, item)
}

// rebuildAndEmitLocked rebuilds the visible slice and returns what to emit.
// Must be called with mu held; sets w.dirty.
func (w *ListWidget[T]) rebuildAndEmitLocked() (emitClear bool, emitNew bool, item T) {
	prevHadSelection := w.selected >= 0 && w.selected < len(w.visible)
	prevItemIdx := -1
	if prevHadSelection {
		prevItemIdx = w.visible[w.selected]
	}

	w.rebuildVisibleLocked()
	w.dirty = true

	if len(w.visible) == 0 {
		if prevHadSelection {
			emitClear = true
		}
	} else {
		newItemIdx := w.visible[w.selected]
		if !prevHadSelection || newItemIdx != prevItemIdx {
			emitNew = true
			item = w.items[newItemIdx].Item
		}
	}
	return
}

// dispatchEmit fires onSelectionChanged and requestFlush after the lock is released.
func (w *ListWidget[T]) dispatchEmit(emitClear bool, emitNew bool, item T) {
	if emitClear {
		var zero T
		w.onSelectionChanged(zero, false)
	}
	if emitNew {
		w.onSelectionChanged(item, true)
	}
	w.requestFlush()
}

func (w *ListWidget[T]) SetItems(items []T) {
	entries := make([]Entry[T], 0, len(items))
	for _, item := range items {
		text := w.makeText(item)
		entries = append(entries, Entry[T]{
			Item:    item,
			Display: text,
			Search:  normalize(text),
		})
	}

	var (
		selectedItem T
		ok           bool
	)

	w.mu.Lock()
	w.items = entries
	w.offset = 0
	w.rebuildVisibleLocked()

	if len(w.visible) > 0 {
		w.selected = 0
		ok = true
		selectedItem = w.items[w.visible[0]].Item
	}

	w.dirty = true
	w.mu.Unlock()

	w.onSelectionChanged(selectedItem, ok)

	w.requestFlush()
}

// Clear resets the list to an empty state, clearing selection and scroll position.
func (w *ListWidget[T]) Clear() {
	w.mu.Lock()
	w.items = w.items[:0]
	w.visible = w.visible[:0]
	w.selected = -1
	w.offset = 0
	w.dirty = true
	w.mu.Unlock()

	var zero T
	w.onSelectionChanged(zero, false)
	w.requestFlush()
}

// AddElements appends a batch of items to the list, updating visibility and
// the subtitle incrementally so the user sees progress during large loads.
// onSelectionChanged is only fired when the first visible item is gained.
func (w *ListWidget[T]) AddElements(items []T) {
	// Build entries outside the lock; makeText may be expensive.
	entries := make([]Entry[T], len(items))
	for i, item := range items {
		text := w.makeText(item)
		entries[i] = Entry[T]{
			Item:    item,
			Display: text,
			Search:  normalize(text),
		}
	}

	var (
		emit bool
		sel  T
	)

	w.mu.Lock()

	hadSelection := len(w.visible) > 0 && w.selected >= 0 && w.selected < len(w.visible)

	for _, entry := range entries {
		idx := len(w.items)
		w.items = append(w.items, entry)
		if w.matchesLocked(entry.Item, entry.Search) {
			w.visible = append(w.visible, idx)
		}
	}

	if !hadSelection && len(w.visible) > 0 {
		w.selected = 0
		w.offset = 0
		sel = w.items[w.visible[0]].Item
		emit = true
	}

	w.dirty = true
	w.mu.Unlock()

	if emit {
		w.onSelectionChanged(sel, true)
	}
	w.requestFlush()
}

func (w *ListWidget[T]) flush(g *gocui.Gui) error {
	w.mu.Lock()
	visible := append([]int(nil), w.visible...)
	items := append([]Entry[T](nil), w.items...)
	selected := w.selected
	offset := w.offset
	w.dirty = false
	w.mu.Unlock()

	v, err := g.View(w.viewName)
	if err != nil {
		return err
	}

	_, height := v.Size()
	if height < 1 {
		height = 1
	}

	n := len(visible)
	if n == 0 {
		selected = -1
		offset = 0
	} else {
		if selected < 0 {
			selected = 0
		}
		if selected >= n {
			selected = n - 1
		}
		if selected < offset {
			offset = selected
		}
		if selected >= offset+height {
			offset = selected - height + 1
		}
		if offset < 0 {
			offset = 0
		}
	}

	v.Clear()

	end := min(offset+height, n)

	for row := offset; row < end; row++ {
		entry := items[visible[row]]
		if row == selected {
			fmt.Fprintf(v, "%s> %s%s%s%s\n", model.SGRReset, model.SelectionBg, model.SelectionFg, entry.Display, model.SGRReset)
		} else if w.makeColor != nil {
			fmt.Fprintf(v, "%s%s  %s%s\n", model.SGRReset, w.makeColor(entry.Item), entry.Display, model.ColorReset)
		} else {
			fmt.Fprintf(v, "%s  %s\n", model.SGRReset, entry.Display)
		}
	}

	if n == 0 {
		v.Subtitle = "0/0"
	} else {
		counter := fmt.Sprintf("%d/%d", selected+1, n)
		if w.makeSubtitle != nil {
			if extra := w.makeSubtitle(items[visible[selected]].Item); extra != "" {
				v.Subtitle = extra + "  " + counter
			} else {
				v.Subtitle = counter
			}
		} else {
			v.Subtitle = counter
		}
	}

	w.mu.Lock()
	w.selected = selected
	w.offset = offset
	w.mu.Unlock()

	return nil
}

func (w *ListWidget[T]) requestFlush() {
	if !w.flushQueued.CompareAndSwap(false, true) {
		return
	}

	w.gui.Update(func(g *gocui.Gui) error {
		w.flushQueued.Store(false)
		return w.flush(g)
	})
}

// SelectedItem returns the currently selected item, or the zero value and false
// if nothing is selected. Safe to call from any goroutine.
func (w *ListWidget[T]) SelectedItem() (T, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.selected < 0 || w.selected >= len(w.visible) {
		var zero T
		return zero, false
	}
	return w.items[w.visible[w.selected]].Item, true
}

func (w *ListWidget[T]) setSelected(selected int) {
	w.mu.Lock()

	n := len(w.visible)
	if n == 0 {
		wasSelected := w.selected != -1

		w.selected = -1
		w.dirty = true
		w.mu.Unlock()

		if wasSelected {
			var zero T
			w.onSelectionChanged(zero, false)
		}

		w.requestFlush()
		return
	}

	if selected < 0 {
		selected = 0
	}
	if selected >= n {
		selected = n - 1
	}

	if selected == w.selected {
		w.mu.Unlock()
		return
	}

	w.selected = selected

	itemIdx := w.visible[selected]
	item := w.items[itemIdx].Item

	w.dirty = true
	w.mu.Unlock()

	w.onSelectionChanged(item, true)
	w.requestFlush()
}

func (w *ListWidget[T]) rebuildVisibleLocked() {
	w.visible = w.visible[:0]

	for i, entry := range w.items {
		if w.matchesLocked(entry.Item, entry.Search) {
			w.visible = append(w.visible, i)
		}
	}

	if len(w.visible) == 0 {
		w.selected = -1
		w.offset = 0
		return
	}

	if w.selected >= len(w.visible) {
		w.selected = len(w.visible) - 1
	}
	if w.selected < 0 {
		w.selected = 0
	}
	if w.offset > w.selected {
		w.offset = w.selected
	}
}

func (w *ListWidget[T]) matchesLocked(item T, search string) bool {
	if w.filter != "" && !strings.Contains(search, w.filter) {
		return false
	}
	if w.matches != nil && !w.matches(item, w.filter) {
		return false
	}
	return true
}

func (w *ListWidget[T]) moveSelection(delta int) error {
	w.setSelected(w.selected + delta)
	return nil
}

func (w *ListWidget[T]) pageSelection(v *gocui.View, dir int) error {
	_, h := v.Size()
	if h < 2 {
		h = 2
	}
	w.setSelected(w.selected + dir*(h-1))
	return nil
}

func (w *ListWidget[T]) onCurrentView(_ *gocui.Gui, v *gocui.View) error {
	w.nav.SetCurView(v.Name())

	_, cy := v.Cursor()

	w.mu.Lock()
	offset := w.offset
	w.mu.Unlock()

	w.setSelected(offset + cy)
	return nil
}

func (w *ListWidget[T]) onGotoStart(_ *gocui.Gui, _ *gocui.View) error {
	w.setSelected(0)
	return nil
}

func (w *ListWidget[T]) onGotoEnd(_ *gocui.Gui, _ *gocui.View) error {
	w.mu.Lock()
	last := len(w.visible) - 1
	w.mu.Unlock()
	w.setSelected(last)
	return nil
}

func (w *ListWidget[T]) onPageUp(_ *gocui.Gui, v *gocui.View) error {
	return w.pageSelection(v, -1)
}

func (w *ListWidget[T]) onPageDown(_ *gocui.Gui, v *gocui.View) error {
	return w.pageSelection(v, 1)
}

func (w *ListWidget[T]) onCursorDown(_ *gocui.Gui, _ *gocui.View) error {
	return w.moveSelection(1)
}

func (w *ListWidget[T]) onCursorUp(_ *gocui.Gui, _ *gocui.View) error {
	return w.moveSelection(-1)
}

func (w *ListWidget[T]) onSearch(_ *gocui.Gui, v *gocui.View) error {
	return w.nav.OpenSearchModal(v.Name())
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
