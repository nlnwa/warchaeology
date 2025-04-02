package console

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v4/cmd/internal/flag"
	"github.com/spf13/cobra"
)

var state = &State{curView: "dir"}

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
		state.dir = filepath.Dir(state.dir)
		state.files = append(state.files, filepath.Base(state.dir))
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

	filesWidget := NewListWidget("dir", "Content_error", "Records", state.readFile, populateFiles)

	viewRecordWidget := NewRecordWidget("Content", "Records", "dir")

	recordsWidget := NewListWidget("Records", "dir", "Content_header", viewRecordWidget.readRecord, populateRecords)
	state.filter = &recordFilter{}
	recordsWidget.filterFunc = state.filter.filterFunc

	gui.SetManager(filesWidget, recordsWidget, viewRecordWidget, nonWidgets, flowLayout)

	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", 'e', gocui.ModNone, state.filter.toggleErrorFilter); err != nil {
		return err
	}
	if err := gui.SetKeybinding("help", gocui.MouseLeft, gocui.ModNone, state.filter.mouseToggleFilter); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'i', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(gui, gowarc.Warcinfo)
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'q', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(gui, gowarc.Request)
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'r', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(gui, gowarc.Response)
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'm', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(gui, gowarc.Metadata)
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 's', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(gui, gowarc.Resource)
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'v', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(gui, gowarc.Revisit)
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'c', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(gui, gowarc.Continuation)
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'n', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(gui, gowarc.Conversion)
	}); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", 'h', gocui.ModNone, func(guiInner *gocui.Gui, view *gocui.View) error {
		shortcutHelpWidget := NewShortcutHelpWidget()
		return shortcutHelpWidget.Layout(gui)
	}); err != nil {
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
		case "Records":
			x0 = 0
			y0 = 10
			x1 = 49
			y1 = maxY - 2
		case "dir":
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

	if view, err := gui.SetView("help", 0, maxY-2, maxX, maxY, 0); err != nil {
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
		return err
	}

	return nil
}

func quit(gui *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (state *State) readFile(gui *gocui.Gui, widget *ListWidget) {
	if len(widget.filteredRecords) > 0 {
		state.file = widget.filteredRecords[widget.selected].(string)
		state.records.Init(gui, state.dir+"/"+state.file)
	}
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
}
