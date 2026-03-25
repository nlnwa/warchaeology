package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/nationallibraryofnorway/warchaeology/v5/internal/ui/model"
	widgets "github.com/nationallibraryofnorway/warchaeology/v5/internal/ui/widget"
	"github.com/nlnwa/gowarc/v3"
)

const (
	viewFiles        = "Files"
	viewRecords      = "Records"
	viewRecordHeader = "Header"
	viewRecordErrors = "Errors"
	viewChrome       = "Chrome"

	modalShortcuts = "Shortcuts"
	modalSearch    = "Search"

	// Layout dimensions.
	filesViewHeight  = 9
	recordsListWidth = 51
	recordPanelX0    = recordsListWidth + 1

	// recordBatchSize controls how many records are buffered before flushing
	// to the UI. A prime number avoids phase-locking with the refresh cadence.
	recordBatchSize = 97
)

type App struct {
	// navigation state
	curView    int
	views      []string
	fullscreen bool

	modal string

	// domain state
	dir      string
	files    []string
	file     string
	suffixes []string
	tmpDir   string

	// gui is stored so selection callbacks can schedule view updates.
	gui *gocui.Gui

	// loadCancel cancels the in-progress record-loading goroutine, if any.
	loadCancel context.CancelFunc

	// widgets
	filesWidget   *widgets.ListWidget[FileItem]
	recordsWidget *widgets.ListWidget[RecordItem]

	recordPanel *widgets.RecordPanelWidget

	chromeWidget *widgets.ChromeWidget

	// modals
	shortcutModal *widgets.ShortcutsModal
	searchModal   *widgets.SearchModal
}

func (a *App) SetSearch(viewName string, q string) {
	switch viewName {
	case viewFiles:
		a.filesWidget.SetFilter(q)
	case viewRecords:
		a.recordsWidget.SetFilter(q)
	}
}

func (a *App) ToggleRecordType(recType gowarc.RecordType) error {
	if recType == 0 {
		a.recordsWidget.SetPredicate(nil)
		return nil
	}
	a.recordsWidget.SetPredicate(func(ri RecordItem, _ string) bool {
		return ri.RecordType&recType != 0
	})
	return nil
}

func NewApp(opts *Options) *App {
	dir := opts.Dir
	if dir == "" {
		wd, _ := os.Getwd()
		dir = wd
	}

	return &App{
		dir:      dir,
		files:    opts.Files,
		suffixes: opts.Suffixes,
		tmpDir:   opts.TempDir,
	}
}

func (a *App) Run() error {
	gui, err := gocui.NewGui(gocui.OutputTrue, true)
	if err != nil {
		return err
	}
	defer gui.Close()

	a.gui = gui

	gui.Cursor = false
	gui.Highlight = true
	gui.SupportOverlaps = true
	gui.Mouse = true

	a.filesWidget = widgets.NewListWidget(
		gui,
		viewFiles,
		func(f FileItem) string { return f },
		a,
		a.onFileSelected,
	)

	a.recordsWidget = widgets.NewListWidget(
		gui,
		viewRecords,
		func(i RecordItem) string {
			return i.String()
		},
		a,
		a.onRecordSelected,
	)
	a.recordsWidget.SetColorizer(func(ri RecordItem) string {
		return model.RecordTypeFgColor(ri.RecordType)
	})
	a.recordsWidget.SetSubtitleMaker(func(ri RecordItem) string {
		return fmt.Sprintf("Offset: %d", ri.Offset)
	})
	a.recordPanel = widgets.NewRecordPanelWidget(gui, viewRecordHeader, viewRecordErrors, a)
	a.chromeWidget = widgets.NewChromeWidget(gui, viewChrome, a)

	a.shortcutModal = widgets.NewModalShortcuts(modalShortcuts, a)
	a.searchModal = widgets.NewSearchModal(modalSearch, a)

	a.views = append([]string{viewFiles, viewRecords}, a.recordPanel.ViewNames()...)

	initialized := false
	gui.SetManager(
		a,
		gocui.ManagerFunc(func(_ *gocui.Gui) error {
			if initialized {
				return nil
			}
			initialized = true
			return a.populateFiles(gui)
		}),
	)

	if err := a.registerGlobalKeybindings(gui); err != nil {
		return err
	}

	err = gui.MainLoop()
	if errors.Is(err, gocui.ErrQuit) {
		err = nil
	}
	return err
}

func (a *App) Layout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()

	if err := a.layoutScreen(gui, maxX, maxY); err != nil {
		return err
	}
	if err := a.layoutModal(gui, maxX, maxY); err != nil {
		return err
	}

	// Rebuild view list each frame: errors panel appears/disappears dynamically.
	a.views = append([]string{viewFiles, viewRecords}, a.recordPanel.ViewNames()...)
	if a.curView >= len(a.views) {
		a.curView = len(a.views) - 1
	}

	if a.modal != "" {
		_, _ = gui.SetCurrentView(a.modal)
		_, _ = gui.SetViewOnTop(a.modal)
	} else {
		_, _ = gui.SetCurrentView(a.views[a.curView])
	}
	return nil
}

type Widget = widgets.Widget

func (a *App) layoutScreen(gui *gocui.Gui, maxX, maxY int) error {
	if maxX < 2 || maxY < 2 {
		// Terminal is too small to host valid gocui views.
		return nil
	}

	recordsX1 := min(recordsListWidth, maxX-1)
	recordsY1 := maxY - 2

	chromeY0 := max(maxY-2, 0)

	layout := []struct {
		viewName string
		x0       int
		y0       int
		x1       int
		y1       int
		widget   Widget
	}{
		{
			viewName: viewFiles,
			x0:       0,
			y0:       0,
			x1:       maxX - 1,
			y1:       filesViewHeight,
			widget:   a.filesWidget,
		},
		{
			viewName: viewRecords,
			x0:       0,
			y0:       filesViewHeight + 1,
			x1:       recordsX1,
			y1:       recordsY1,
			widget:   a.recordsWidget,
		},
		{
			viewName: viewChrome,
			x0:       0,
			y0:       chromeY0,
			x1:       maxX - 1,
			y1:       maxY,
			widget:   a.chromeWidget,
		},
	}

	for _, l := range layout {
		if l.x1 <= l.x0 || l.y1 <= l.y0 {
			continue
		}
		if err := l.widget.Layout(gui, l.x0, l.y0, l.x1, l.y1); err != nil {
			return err
		}
	}

	// Frame runes for Files and Records: double-line if focused, single-line otherwise.
	focused := ""
	if a.modal == "" && a.curView < len(a.views) {
		focused = a.views[a.curView]
	}
	for _, l := range layout {
		if l.viewName == viewChrome {
			continue
		}
		v, err := gui.View(l.viewName)
		if err != nil {
			continue
		}
		if l.viewName == focused {
			v.FrameRunes = widgets.DoubleFrame
		} else {
			v.FrameRunes = widgets.SingleFrame
		}
	}

	// Delegate record panel layout (Header, Content, optional Errors).
	panelX0, panelY0 := recordPanelX0, filesViewHeight+1
	panelX1, panelY1 := maxX-1, maxY-2
	if panelX1 > panelX0 && panelY1 > panelY0 {
		if err := a.recordPanel.Layout(gui, panelX0, panelY0, panelX1, panelY1, focused); err != nil {
			return err
		}
	}

	// When fullscreen, resize the focused view to fill the whole panel area and
	// bring it on top so all other views are hidden beneath it.
	if a.fullscreen && a.modal == "" && a.curView < len(a.views) {
		focusedName := a.views[a.curView]
		if _, err := gui.SetView(focusedName, 0, 0, maxX-1, maxY-2, 0); err != nil && !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		_, _ = gui.SetViewOnTop(focusedName)
	}

	return nil
}

type Modal = widgets.Modal

func (a *App) layoutModal(gui *gocui.Gui, maxX, maxY int) error {
	screen := []struct {
		viewName string
		x0       int
		y0       int
		x1       int
		y1       int
		overlaps byte
		modal    Modal
	}{
		{
			viewName: modalShortcuts,
			modal:    a.shortcutModal,
			x0:       maxX/2 - 30,
			y0:       maxY/2 - 15,
			x1:       maxX/2 + 30,
			y1:       maxY/2 + 15,
			overlaps: 0,
		},
		{
			viewName: modalSearch,
			modal:    a.searchModal,
			x0:       maxX/2 - 35,
			y0:       maxY/2 - 1,
			x1:       maxX/2 + 35,
			y1:       maxY/2 + 1,
			overlaps: 0,
		},
	}

	for _, s := range screen {
		if a.modal != s.viewName {
			_ = s.modal.Dispose(gui)
			_ = gui.DeleteView(s.viewName)
			continue
		}
		_, err := gui.SetView(s.viewName, s.x0, s.y0, s.x1, s.y1, s.overlaps)
		if err == nil {
			continue
		}
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		if err := s.modal.Init(gui); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) onShowHelp(_ *gocui.Gui, _ *gocui.View) error {
	a.modal = modalShortcuts
	return nil
}

func (a *App) OpenSearchModal(viewName string) error {
	a.searchModal.SetSourceView(viewName)
	a.modal = modalSearch
	return nil
}

func (a *App) CloseModal(_ *gocui.Gui, _ *gocui.View) error {
	a.modal = ""
	return nil
}

func (a *App) SetCurView(viewName string) {
	for i, v := range a.views {
		if v == viewName {
			a.curView = i
			return
		}
	}
}

func (a *App) onPrevView(gui *gocui.Gui, view *gocui.View) error {
	a.curView = (a.curView - 1 + len(a.views)) % len(a.views)
	return nil
}

func (a *App) onNextView(_ *gocui.Gui, _ *gocui.View) error {
	a.curView = (a.curView + 1) % len(a.views)
	return nil
}

func (a *App) onToggleFullscreen(_ *gocui.Gui, _ *gocui.View) error {
	a.fullscreen = !a.fullscreen
	return nil
}

func (a *App) ToggleLineEndings() error {
	a.recordPanel.ToggleLineEndings()
	item, ok := a.recordsWidget.SelectedItem()
	if !ok {
		return nil
	}
	path := filepath.Join(a.dir, a.file)
	a.loadRecordDetails(a.gui, path, item)
	return nil
}

type RecordItem struct {
	ID         string
	RecordType gowarc.RecordType
	Offset     int64
	Err        error // non-nil for placeholder error records
}

func (r RecordItem) String() string {
	if r.Err != nil {
		return r.ID
	}
	return fmt.Sprintf("%s (%s)", r.ID, (r.RecordType &^ model.ErrorRecordType))
}

type FileItem = string

func (a *App) populateFiles(gui *gocui.Gui) error {
	view, err := gui.View(viewFiles)
	if err != nil {
		return err
	}

	view.Title = a.dir

	var items []FileItem

	if len(a.files) > 0 {
		a.filesWidget.SetItems(a.files)
		return nil
	}

	entries, err := os.ReadDir(a.dir)
	if err != nil {
		return err
	}
	items = make([]FileItem, 0, len(entries))
	for _, entry := range entries {
		for _, suffix := range a.suffixes {
			if strings.HasSuffix(entry.Name(), suffix) {
				items = append(items, entry.Name())
				break
			}
		}
	}
	a.filesWidget.SetItems(items)
	return nil
}

func (a *App) registerGlobalKeybindings(gui *gocui.Gui) error {
	quit := func(*gocui.Gui, *gocui.View) error { return gocui.ErrQuit }
	keyBindings := []model.KeyBinding{
		{Key: gocui.KeyCtrlC, Mod: gocui.ModNone, Fn: quit},
		{Key: 'h', Mod: gocui.ModNone, Fn: a.onShowHelp},
		{Key: 'z', Mod: gocui.ModNone, Fn: a.onToggleFullscreen},
		{Key: gocui.KeyCtrlY, Mod: gocui.ModNone, Fn: a.onToggleMouse},
		{Key: gocui.KeyTab, Mod: gocui.ModNone, Fn: a.onNextView},
		{Key: gocui.KeyBacktab, Mod: gocui.ModNone, Fn: a.onPrevView},
	}

	for _, kb := range keyBindings {
		if err := gui.SetKeybinding("", kb.Key, kb.Mod, kb.Fn); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) onToggleMouse(gui *gocui.Gui, _ *gocui.View) error {
	gui.Mouse = !gui.Mouse
	return nil
}

func (a *App) onFileSelected(item FileItem, ok bool) {
	if a.loadCancel != nil {
		a.loadCancel()
		a.loadCancel = nil
	}

	a.recordsWidget.Clear()

	if !ok {
		return
	}

	a.file = item

	ctx, cancel := context.WithCancel(context.Background())
	a.loadCancel = cancel

	go func() {
		defer cancel()
		path := filepath.Join(a.dir, item)
		_ = a.loadRecords(ctx, path)
	}()
}

func (a *App) onRecordSelected(item RecordItem, ok bool) {
	a.gui.Update(func(g *gocui.Gui) error {
		if !ok {
			a.recordPanel.Clear(g)
			return nil
		}
		path := filepath.Join(a.dir, a.file)
		a.loadRecordDetails(g, path, item)
		return nil
	})
}

// loadRecords reads all records from path and streams them to the records
// widget in batches. It returns when the file is exhausted or ctx is cancelled.
func (a *App) loadRecords(ctx context.Context, path string) error {
	reader, err := gowarc.NewWarcFileReader(path, 0, gowarc.WithBufferTmpDir(a.tmpDir))
	if err != nil {
		return err
	}
	defer reader.Close()

	batch := make([]RecordItem, 0, recordBatchSize)

	for record, err := range reader.Records() {
		if err != nil {
			batch = append(batch, RecordItem{
				ID:         "<read error>",
				Err:        err,
				RecordType: model.ErrorRecordType,
				Offset:     record.Offset,
			})
			if len(batch) >= recordBatchSize {
				a.recordsWidget.AddElements(batch)
				batch = batch[:0]
			}
			continue
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		rt := record.WarcRecord.Type()
		if len(record.Validation) > 0 {
			rt |= model.ErrorRecordType
		}
		item := RecordItem{
			ID:         record.WarcRecord.WarcHeader().Get(gowarc.WarcRecordID),
			RecordType: rt,
			Offset:     record.Offset,
		}
		_ = record.Close()
		batch = append(batch, item)

		if len(batch) >= recordBatchSize {
			a.recordsWidget.AddElements(batch)
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		a.recordsWidget.AddElements(batch)
	}
	return nil
}

// loadRecordDetails opens the record at item.Offset in path and populates the
// header, content, and errors views.
func (a *App) loadRecordDetails(g *gocui.Gui, path string, item RecordItem) {
	a.recordPanel.Clear(g)
	if item.Err != nil {
		a.recordPanel.RenderReadError(g, item.Err)
		return
	}
	reader, err := gowarc.NewWarcFileReader(path, item.Offset, gowarc.WithBufferTmpDir(a.tmpDir))
	if err != nil {
		a.recordPanel.RenderErrors(g, []error{err})
		return
	}
	defer reader.Close()

	rec, err := reader.Next()
	if err != nil {
		a.recordPanel.RenderErrors(g, []error{err})
		return
	}
	defer rec.Close()

	a.recordPanel.PopulateHeader(g, rec)
	a.recordPanel.PopulateContent(g, rec)
	a.recordPanel.RenderErrors(g, rec.Validation)
}
