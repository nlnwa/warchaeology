package console

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/awesome-gocui/gocui"
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
	selectFunc          func(gui *gocui.Gui, widget *ListWidget)
	populateRecordsFunc func(gui *gocui.Gui, ctx context.Context, finishedCb func(), widget *ListWidget, data interface{})
	filterFunc          func(interface{}) bool
	records             []interface{}
	filteredRecords     []interface{}
	cancelFunc          context.CancelFunc
	cancelRefreshFunc   context.CancelFunc
	finished            context.Context
	state               uint8
}

func NewListWidget(name string, prev, next string,
	selectFunc func(gui *gocui.Gui, widget *ListWidget),
	populateRecordsFunc func(gui *gocui.Gui, ctx context.Context, finishedCb func(), widget *ListWidget, data interface{})) *ListWidget {

	return &ListWidget{
		name:                name,
		prev:                prev,
		next:                next,
		selectFunc:          selectFunc,
		populateRecordsFunc: populateRecordsFunc,
		state:               reading,
	}
}

func (widgetList *ListWidget) Init(gui *gocui.Gui, data interface{}) {
	if widgetList.cancelFunc != nil {
		widgetList.cancelRefreshFunc()
		widgetList.cancelRefreshFunc = nil
		widgetList.cancelFunc()
		widgetList.cancelFunc = nil
		<-widgetList.finished.Done()
		widgetList.finished = nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	widgetList.cancelFunc = cancel
	var finishedCb func()
	var refreshCtx context.Context
	refreshCtx, widgetList.cancelRefreshFunc = context.WithCancel(context.Background())
	widgetList.finished, finishedCb = context.WithCancel(context.Background())
	widgetList.records = nil
	widgetList.filteredRecords = nil
	widgetList.selected = -1
	go func() {
		widgetList.populateRecordsFunc(gui, ctx, finishedCb, widgetList, data)
	}()
	widgetList.update(gui, refreshCtx, widgetList.finished)
}

func (widgetList *ListWidget) Layout(gui *gocui.Gui) error {
	if view, err := gui.SetView(widgetList.name, 0, 0, 1, 1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.FgColor = gocui.ColorGreen
		view.BgColor = gocui.ColorDefault
		view.SelBgColor = gocui.ColorWhite
		view.SelFgColor = gocui.ColorBlack
		view.Highlight = false
		view.Autoscroll = false
		view.Title = widgetList.name
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.KeyArrowDown, gocui.ModNone, widgetList.cursorDown); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.KeyArrowUp, gocui.ModNone, widgetList.cursorUp); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.KeyHome, gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		if view != nil {
			return widgetList.selectLine(gui, view, 0)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.KeyEnd, gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		if view != nil {
			return widgetList.selectLine(gui, view, len(widgetList.filteredRecords)-1)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.KeyPgdn, gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		if view != nil {
			_, height := view.Size()
			height--
			return widgetList.selectLine(gui, view, widgetList.selected+height)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.KeyPgup, gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		if view != nil {
			_, height := view.Size()
			height--
			return widgetList.selectLine(gui, view, widgetList.selected-height)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, 'f', gocui.ModNone, widgetList.search); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.KeyEnter, gocui.ModNone, widgetList.nextView); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.KeyEsc, gocui.ModNone, widgetList.prevView); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.MouseLeft, gocui.ModNone, widgetList.currentView); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.MouseRelease, gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		return nil
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.MouseWheelDown, gocui.ModNone, widgetList.cursorDown); err != nil {
		return err
	}
	if err := gui.SetKeybinding(widgetList.name, gocui.MouseWheelUp, gocui.ModNone, widgetList.cursorUp); err != nil {
		return err
	}
	return nil
}

func (widgetList *ListWidget) search(gui *gocui.Gui, parent *gocui.View) error {
	maxX, maxY := gui.Size()
	var view *gocui.View
	var err error
	if view, err = gui.SetView("search", maxX/2-35, maxY/2-1, maxX/2+35, maxY/2+1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.FgColor = gocui.ColorDefault
		view.BgColor = gocui.ColorDefault
		view.SelBgColor = gocui.ColorWhite
		view.SelFgColor = gocui.ColorBlack
		view.Highlight = false
		view.Autoscroll = false
		view.Title = "Find in " + parent.Title
		view.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝'}
		view.Editable = true
		view.KeybindOnEdit = false
		view.Editor = &SearchEditor{gui, widgetList, parent}
		gui.Cursor = true
	}
	state.modalView = "search"
	return nil
}

type SearchEditor struct {
	gui        *gocui.Gui
	widgetList *ListWidget
	parent     *gocui.View
}

func (searchEditor *SearchEditor) Edit(view *gocui.View, key gocui.Key, characterAsInteger rune, modifier gocui.Modifier) {
	if characterAsInteger != 0 && modifier == 0 {
		view.EditWrite(characterAsInteger)
		_ = searchEditor.widgetList.searchLine(searchEditor.gui, searchEditor.parent, view.ViewBuffer())
		return
	}

	switch key {
	case gocui.KeySpace:
		view.EditWrite(' ')
		_ = searchEditor.widgetList.searchLine(searchEditor.gui, searchEditor.parent, view.ViewBuffer())
	case gocui.KeyBackspace, gocui.KeyBackspace2:
		view.EditDelete(true)
		_ = searchEditor.widgetList.searchLine(searchEditor.gui, searchEditor.parent, view.ViewBuffer())
	case gocui.KeyDelete:
		view.EditDelete(false)
		_ = searchEditor.widgetList.searchLine(searchEditor.gui, searchEditor.parent, view.ViewBuffer())
	case gocui.KeyInsert:
		view.Overwrite = !view.Overwrite
	case gocui.KeyArrowLeft:
		view.MoveCursor(-1, 0)
	case gocui.KeyArrowRight:
		view.MoveCursor(1, 0)
	case gocui.KeyEsc:
		// If not here the esc key will act like the KeySpace
		fallthrough
	case gocui.KeyEnter:
		_ = searchEditor.gui.DeleteView(view.Name())
		state.modalView = ""
		searchEditor.gui.Cursor = false
	default:
	}
}

func (widgetList *ListWidget) prevView(gui *gocui.Gui, view *gocui.View) error {
	state.curView = widgetList.prev
	return nil
}

func (widgetList *ListWidget) nextView(gui *gocui.Gui, view *gocui.View) error {
	state.curView = widgetList.next
	return nil
}

func (widgetList *ListWidget) currentView(gui *gocui.Gui, view *gocui.View) error {
	state.curView = view.Name()
	_, oy := view.Origin()
	_, cy := view.Cursor()
	newSelect := cy + oy
	_ = widgetList.selectLine(gui, view, newSelect)
	return nil
}

func (widgetList *ListWidget) cursorDown(gui *gocui.Gui, view *gocui.View) error {
	if view != nil {
		_ = widgetList.selectLine(gui, view, widgetList.selected+1)
	}
	return nil
}

func (widgetList *ListWidget) cursorUp(gui *gocui.Gui, view *gocui.View) error {
	if view != nil {
		_ = widgetList.selectLine(gui, view, widgetList.selected-1)
	}
	return nil
}

func (widgetList *ListWidget) searchLine(gui *gocui.Gui, view *gocui.View, targetString string) error {
	if view != nil {
		for line, id := range view.ViewBufferLines() {
			if strings.Contains(id, targetString) {
				return widgetList.selectLine(gui, view, line)
			}
		}
	}
	return nil
}

func (widgetList *ListWidget) selectLine(gui *gocui.Gui, view *gocui.View, selected int) error {
	if view != nil {
		originX, originY := view.Origin()
		_, height := view.Size()
		if selected < 0 {
			selected = 0
		}
		if selected >= view.ViewLinesHeight()-1 {
			selected = view.ViewLinesHeight() - 2
		}

		// Do nothing if no movement or content is empty
		if selected == widgetList.selected || view.ViewLinesHeight() == 0 {
			return nil
		}

		if widgetList.selected != -1 {
			_ = view.SetLine(widgetList.selected, fmt.Sprintf("%s", widgetList.filteredRecords[widgetList.selected]))
		}
		widgetList.selected = selected
		_ = view.SetHighlight(widgetList.selected, true)

		cursorPositionY := widgetList.selected - originY
		if cursorPositionY < 0 {
			originY += cursorPositionY
			cursorPositionY = 0
		} else if cursorPositionY >= height {
			originY += (cursorPositionY - height) + 1
			cursorPositionY -= (cursorPositionY - height) + 1
		}

		if err := view.SetOrigin(originX, originY); err != nil {
			return err
		}
		if err := view.SetCursor(0, cursorPositionY); err != nil {
			return err
		}

		if line, err := view.Line(widgetList.selected); err != nil || line == "" {
			return nil
		} else {
			view.Subtitle = fmt.Sprintf("%d/%d", widgetList.selected+1, len(widgetList.filteredRecords))
			if len(widgetList.filteredRecords) > 0 {
				widgetList.selectFunc(gui, widgetList)
			}
		}
	}
	return nil
}

func (widgetList *ListWidget) update(gui *gocui.Gui, ctx context.Context, finished context.Context) {
	time.Sleep(100 * time.Millisecond)
	view, err := gui.View(widgetList.name)
	if err != nil {
		panic(err)
	}
	view.Clear()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered in f %s\n%s", r, debug.Stack())
			}
		}()

		length := 0
		for {
			select {
			case <-ctx.Done():
				<-finished.Done()
				return
			case <-finished.Done():
				if length < len(widgetList.records) {
					rec := widgetList.records[length:]
					widgetList.upd(gui, ctx, rec)
				}
				return
			default:
				if length < len(widgetList.records) {
					rec := widgetList.records[length:]
					length = length + len(rec)
					widgetList.upd(gui, ctx, rec)
				}
			}
		}
	}()
}

func (widgetList *ListWidget) upd(gui *gocui.Gui, ctx context.Context, rec []interface{}) {
	gui.UpdateAsync(func(guiInner *gocui.Gui) error {
		view, err := guiInner.View(widgetList.name)
		if err != nil {
			return err
		}

		for _, r := range rec {
			select {
			case <-ctx.Done():
				return nil
			default:
				if widgetList.filterFunc == nil || widgetList.filterFunc(r) {
					widgetList.filteredRecords = append(widgetList.filteredRecords, r)
					fmt.Fprintf(view, "%s\n", r)
					if len(widgetList.filteredRecords) == 1 {
						_ = widgetList.selectLine(gui, view, 0)
						widgetList.selectFunc(gui, widgetList)
					}
				}
			}
		}
		view.Subtitle = fmt.Sprintf("%d/%d", widgetList.selected+1, len(widgetList.filteredRecords))
		return nil
	})
}

func (widgetList *ListWidget) refreshFilter(gui *gocui.Gui, view *gocui.View) error {
	if widgetList.cancelRefreshFunc != nil {
		widgetList.cancelRefreshFunc()
		widgetList.cancelRefreshFunc = nil
	}

	// Skip filtering if records is empty and no read operation is in progress
	if len(widgetList.records) == 0 && widgetList.finished == nil {
		return nil
	}

	var ctx context.Context
	ctx, widgetList.cancelRefreshFunc = context.WithCancel(context.Background())
	view.Clear()
	widgetList.filteredRecords = nil
	widgetList.selected = -1

	go func() {
		length := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-widgetList.finished.Done():
				if length < len(widgetList.records) {
					rec := widgetList.records[length:]
					widgetList.upd(gui, ctx, rec)
				}
				return
			default:
				if length < len(widgetList.records) {
					rec := widgetList.records[length:]
					length = length + len(rec)
					widgetList.upd(gui, ctx, rec)
				}
			}
		}
	}()
	return nil
}
