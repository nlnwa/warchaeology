package widget

import (
	"errors"

	"github.com/awesome-gocui/gocui"
	"github.com/nationallibraryofnorway/warchaeology/v5/internal/ui/model"
)

type RecordWidget struct {
	viewName string
	ctrl     Controller
}

func NewRecordWidget(viewName string, ctrl Controller) *RecordWidget {
	return &RecordWidget{viewName: viewName, ctrl: ctrl}
}

func (w *RecordWidget) Layout(gui *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := gui.SetView(w.viewName, x0, y0, x1, y1, 0)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}
	if errors.Is(err, gocui.ErrUnknownView) {
		v.Wrap = true
		keyBindings := []model.KeyBinding{
			{Key: gocui.KeyArrowDown, Mod: gocui.ModNone, Fn: w.onCursorDown},
			{Key: gocui.KeyArrowUp, Mod: gocui.ModNone, Fn: w.onCursorUp},
			{Key: gocui.KeyHome, Mod: gocui.ModNone, Fn: w.onGotoStart},
			{Key: gocui.KeyEnd, Mod: gocui.ModNone, Fn: w.onGotoEnd},
			{Key: gocui.KeyPgup, Mod: gocui.ModNone, Fn: w.onPageUp},
			{Key: gocui.KeyPgdn, Mod: gocui.ModNone, Fn: w.onPageDown},
			{Key: gocui.MouseLeft, Mod: gocui.ModNone, Fn: w.onFocus},
			{Key: gocui.MouseWheelDown, Mod: gocui.ModNone, Fn: w.onCursorDown},
			{Key: gocui.MouseWheelUp, Mod: gocui.ModNone, Fn: w.onCursorUp},
		}
		for _, kb := range keyBindings {
			if err := gui.SetKeybinding(w.viewName, kb.Key, kb.Mod, kb.Fn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *RecordWidget) onGotoStart(gui *gocui.Gui, view *gocui.View) error {
	return scroll(view, -view.ViewLinesHeight())
}

func (w *RecordWidget) onGotoEnd(gui *gocui.Gui, view *gocui.View) error {
	return scroll(view, view.ViewLinesHeight())
}

func (w *RecordWidget) onPageUp(gui *gocui.Gui, view *gocui.View) error {
	_, y := view.Size()
	return scroll(view, -(y - 1))
}

func (w *RecordWidget) onPageDown(gui *gocui.Gui, view *gocui.View) error {
	_, y := view.Size()
	return scroll(view, y-1)
}

func (w *RecordWidget) onCursorDown(gui *gocui.Gui, view *gocui.View) error {
	return scroll(view, 1)
}

func (w *RecordWidget) onCursorUp(gui *gocui.Gui, view *gocui.View) error {
	return scroll(view, -1)
}

func (w *RecordWidget) onFocus(_ *gocui.Gui, v *gocui.View) error {
	w.ctrl.SetCurView(v.Name())
	return nil
}

func scroll(view *gocui.View, dy int) error {
	_, h := view.Size()
	ch := view.ViewLinesHeight()
	if h >= ch {
		return nil
	}

	ox, oy := view.Origin()
	scrollDestinationY := min(ch-h, min(max(oy+dy, 0), ch-h))
	return view.SetOrigin(ox, scrollDestinationY)
}
