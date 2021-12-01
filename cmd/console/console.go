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
	"errors"
	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "console <directory>",
		Short: "A shell for working with WARC files",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			var f os.FileInfo
			f, err = os.Lstat(state.dir)
			if err != nil {
				return err
			}
			if !f.IsDir() {
				f := path.Base(state.dir)
				state.dir = path.Dir(state.dir)
				state.files = append(state.files, f)
			}
			if state.suffixes, err = cmd.Flags().GetStringSlice(flag.Suffixes); err != nil {
				return err
			}

			return runE()
		},
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	cmd.Flags().StringSlice(flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)

	return cmd
}

var state = &State{curView: "dir"}

func runE() error {
	os.Setenv("COLORTERM", "truecolor")
	g, err := gocui.NewGui(gocui.OutputTrue, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	state.g = g

	g.Cursor = false
	g.Highlight = true
	g.FgColor = gocui.ColorYellow
	g.BgColor = gocui.ColorDefault
	g.SelFgColor = gocui.ColorCyan
	g.SelBgColor = gocui.ColorDefault
	g.SelFrameColor = gocui.ColorCyan
	g.SupportOverlaps = true
	g.Mouse = true

	nonWidgets := gocui.ManagerFunc(layout)
	fl := gocui.ManagerFunc(flowLayout)

	filesWidget := NewListWidget("dir", "Content_error", "Records", readFile, populateFiles)

	viewRecordWidget := NewRecordWidget("Content", "Records", "dir")

	recordsWidget := NewListWidget("Records", "dir", "Content_header", viewRecordWidget.readRecord, populateRecords)
	state.filter = &recordFilter{}
	recordsWidget.filterFunc = state.filter.filterFunc

	g.SetManager(filesWidget, recordsWidget, viewRecordWidget, nonWidgets, fl)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'e', gocui.ModNone, state.filter.toggleErrorFilter); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("help", gocui.MouseLeft, gocui.ModNone, state.filter.mouseToggleFilter); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'i', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(g, gowarc.Warcinfo)
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(g, gowarc.Request)
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'r', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(g, gowarc.Response)
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'm', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(g, gowarc.Metadata)
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(g, gowarc.Resource)
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'v', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(g, gowarc.Revisit)
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'c', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(g, gowarc.Continuation)
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'n', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return state.filter.toggleRecordTypeFilter(g, gowarc.Conversion)
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'h', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		v := NewShortcutHelpWidget()
		return v.Layout(g)
	}); err != nil {
		log.Panicln(err)
	}

	state.records = recordsWidget

	if state.dir == "" {
		var err error
		state.dir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}
	time.AfterFunc(100*time.Millisecond, func() {
		filesWidget.Init(g, state.dir)
	})

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		panic(err)
	}
	return nil
}

func flowLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	views := g.Views()

	for _, v := range views {
		var x0, y0, x1, y1 int
		switch v.Name() {
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
		_, err := g.SetView(v.Name(), x0, y0, x1, y1, 0)
		if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
	}
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("help", 0, maxY-2, maxX, maxY, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = false
	}
	state.filter.refreshHelp(g)

	v := state.curView
	if state.modalView != "" {
		v = state.modalView
	}
	if _, err := g.SetCurrentView(v); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func readFile(g *gocui.Gui, widget *ListWidget) {
	if len(widget.filteredRecords) > 0 {
		state.file = widget.filteredRecords[widget.selected].(string)
		state.records.Init(g, state.dir+"/"+state.file)
	}
}

type State struct {
	g         *gocui.Gui
	curView   string
	modalView string
	suffixes  []string
	files     []string // Initial files from command line
	dir       string   // Initial dir from command line
	file      string
	records   *ListWidget
	header    string
	content   string
	errors    string
	filter    *recordFilter
}
