package widget

import (
	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc/v3"
)

// Controller is the single interface that the application implements and all
// widgets depend on for navigation, search, record filtering, and display toggling.
type Controller interface {
	// Navigation / modal management
	SetCurView(name string)

	OpenSearchModal(viewName string) error
	CloseModal(*gocui.Gui, *gocui.View) error

	// Search
	SetSearch(viewName string, query string)

	// Record type filter
	ToggleRecordType(recType gowarc.RecordType) error

	// Visible line endings toggle
	ToggleLineEndings() error
}

// Widget manages one or more gocui views. Layout is called every frame with
// the allocated screen rectangle. The widget is responsible for creating its
// own view via gui.SetView, performing one-time setup when the view is first
// created (detected via gocui.ErrUnknownView), and redrawing its content
// whenever bounds or internal state change.
type Widget interface {
	Layout(gui *gocui.Gui, x0, y0, x1, y1 int) error
}

// Modal is implemented by overlay dialogs that need setup/teardown when
// shown/hidden.
type Modal interface {
	Init(gui *gocui.Gui) error
	Dispose(gui *gocui.Gui) error
}
