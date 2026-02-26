package console

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/nationallibraryofnorway/warchaeology/v4/cmd/internal/flag"
	"github.com/nlnwa/gowarc/v3"
	"github.com/spf13/cobra"
)

const (
	viewDir     = "dir"
	viewRecords = "Records"
	viewHelp    = "help"
)

var state = newState()

func newState() *State {
	return &State{curView: viewDir}
}

func NewCmdConsole() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "console DIR",
		Short: "A shell for working with WARC files",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := state.Complete(cmd, args); err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return Run()
		},
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	cmd.Flags().StringSlice(flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)

	return cmd
}

func (state *State) Complete(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("missing input directory")
	}
	if len(args) > 0 {
		state.dir = args[0]
	}
	var err error
	if state.dir, err = filepath.Abs(state.dir); err != nil {
		return err
	}
	if state.dir, err = filepath.EvalSymlinks(state.dir); err != nil {
		return err
	}
	var fileInfo os.FileInfo
	fileInfo, err = os.Lstat(state.dir)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		selectedFile := filepath.Base(state.dir)
		state.dir = filepath.Dir(state.dir)
		state.files = append(state.files, selectedFile)
	}
	if state.suffixes, err = cmd.Flags().GetStringSlice(flag.Suffixes); err != nil {
		return err
	}

	return nil

}

func Run() error {
	os.Setenv("COLORTERM", "truecolor")
	gui, err := gocui.NewGui(gocui.OutputTrue, true)
	if err != nil {
		return err
	}
	defer gui.Close()

	state.gui = gui

	gui.Cursor = false
	gui.Highlight = true
	gui.FgColor = gocui.ColorYellow
	gui.BgColor = gocui.ColorDefault
	gui.SelFgColor = gocui.ColorCyan
	gui.SelBgColor = gocui.ColorDefault
	gui.SelFrameColor = gocui.ColorCyan
	gui.SupportOverlaps = true
	gui.Mouse = true

	nonWidgets := gocui.ManagerFunc(state.layout)
	flowLayout := gocui.ManagerFunc(flowLayout)

	filesWidget := NewListWidget(viewDir, "Content_error", viewRecords, state.readFile, populateFiles)

	viewRecordWidget := NewRecordWidget("Content", "Records", "dir")

	recordsWidget := NewListWidget(viewRecords, viewDir, "Content_header", viewRecordWidget.readRecord, populateRecords)
	state.filter = &recordFilter{}
	recordsWidget.filterFunc = state.filter.filterFunc

	gui.SetManager(filesWidget, recordsWidget, viewRecordWidget, nonWidgets, flowLayout)

	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := registerFilterBindings(gui); err != nil {
		return err
	}
	if err := gui.SetKeybinding(viewHelp, gocui.MouseLeft, gocui.ModNone, state.filter.mouseToggleFilter); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'h', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		shortcutHelpWidget := NewShortcutHelpWidget()
		return shortcutHelpWidget.Layout(gui)
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'z', gocui.ModNone, viewRecordWidget.toggleContentFullscreen); err != nil {
		return err
	}

	state.records = recordsWidget

	if state.dir == "" {
		var err error
		state.dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	time.AfterFunc(100*time.Millisecond, func() {
		filesWidget.Init(gui, state.dir)
	})

	if err := gui.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		return err
	}
	return nil
}

func flowLayout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	views := gui.Views()

	for _, view := range views {
		var x0, y0, x1, y1 int
		switch view.Name() {
		case viewRecords:
			x0 = 0
			y0 = 10
			x1 = 49
			y1 = maxY - 2
		case viewDir:
			x0 = 0
			y0 = 0
			x1 = maxX - 1
			y1 = 9
		default:
			continue
		}
		_, err := gui.SetView(view.Name(), x0, y0, x1, y1, 0)
		if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
	}
	return nil
}

func (state *State) layout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()

	if view, err := gui.SetView(viewHelp, 0, maxY-2, maxX, maxY, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.Frame = false
		view.Editable = false
	}
	state.filter.refreshHelp(gui)

	newView := state.curView
	if state.modalView != "" {
		newView = state.modalView
	}
	if _, err := gui.SetCurrentView(newView); err != nil {
		if _, fallbackErr := gui.SetCurrentView(viewDir); fallbackErr != nil {
			return fallbackErr
		}
		state.curView = viewDir
	}

	return nil
}

func quit(gui *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (state *State) readFile(gui *gocui.Gui, widget *ListWidget) {
	if len(widget.filteredRecords) > 0 {
		state.file = widget.filteredRecords[widget.selected].(string)
		state.records.Init(gui, filepath.Join(state.dir, state.file))
	}
}

func registerFilterBindings(gui *gocui.Gui) error {
	if err := gui.SetKeybinding("", 'e', gocui.ModNone, state.filter.toggleErrorFilter); err != nil {
		return err
	}

	type recordTypeBinding struct {
		key     rune
		recType gowarc.RecordType
	}

	bindings := []recordTypeBinding{
		{key: 'i', recType: gowarc.Warcinfo},
		{key: 'q', recType: gowarc.Request},
		{key: 'r', recType: gowarc.Response},
		{key: 'm', recType: gowarc.Metadata},
		{key: 's', recType: gowarc.Resource},
		{key: 'v', recType: gowarc.Revisit},
		{key: 'c', recType: gowarc.Continuation},
		{key: 'n', recType: gowarc.Conversion},
	}

	for _, binding := range bindings {
		recType := binding.recType
		if err := gui.SetKeybinding("", binding.key, gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
			return state.filter.toggleRecordTypeFilter(gui, recType)
		}); err != nil {
			return err
		}
	}

	return nil
}

type State struct {
	gui       *gocui.Gui
	curView   string
	modalView string
	suffixes  []string
	files     []string // Initial files from command line
	dir       string   // Initial dir from command line
	file      string
	records   *ListWidget
	filter    *recordFilter
	fullView  bool
}
