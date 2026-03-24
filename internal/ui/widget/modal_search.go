package widget

import (
	"strings"

	"github.com/awesome-gocui/gocui"
)

type SearchModal struct {
	viewName   string
	sourceView string // the list view that triggered the search
	ctrl       Controller
}

func NewSearchModal(viewName string, ctrl Controller) *SearchModal {
	return &SearchModal{viewName: viewName, ctrl: ctrl}
}

// SetSourceView records which list view opened the modal. Called by the
// Navigator implementation before the modal is made visible.
func (w *SearchModal) SetSourceView(viewName string) {
	w.sourceView = viewName
}

func (w *SearchModal) Init(gui *gocui.Gui) error {
	view, err := gui.View(w.viewName)
	if err != nil {
		return err
	}

	view.Title = "Find in " + w.sourceView
	view.FrameRunes = DoubleFrame
	view.Editable = true
	view.Editor = gocui.DefaultEditor

	gui.Cursor = true

	if err := gui.SetKeybinding(w.viewName, gocui.KeyEnter, gocui.ModNone, w.onSearch); err != nil {
		return err
	}
	return gui.SetKeybinding(w.viewName, gocui.KeyEsc, gocui.ModNone, w.ctrl.CloseModal)
}

func (w *SearchModal) Dispose(gui *gocui.Gui) error {
	gui.Cursor = false
	return nil
}

func (w *SearchModal) onSearch(gui *gocui.Gui, view *gocui.View) error {
	q := strings.TrimSpace(view.Buffer())
	w.ctrl.SetSearch(w.sourceView, q)
	return w.ctrl.CloseModal(gui, view)
}
