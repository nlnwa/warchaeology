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
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"time"
)

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "console",
		Short: "A shell for working with WARC files",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				state.dir = args[0]
			}
			fmt.Println(state.dir)

			return runE()
		},
	}

	return cmd
}

var state = &State{curView: "dir"}

func runE() error {
	g, err := gocui.NewGui(gocui.OutputTrue, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	state.g = g

	g.Cursor = false
	g.SetManagerFunc(layout)
	g.Highlight = true
	g.FgColor = gocui.ColorYellow
	g.BgColor = gocui.ColorDefault
	g.SelFgColor = gocui.ColorCyan
	g.SelBgColor = gocui.ColorDefault
	g.SelFrameColor = gocui.ColorCyan
	g.SupportOverlaps = true

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	filesWidget := NewListWidget("dir", "records", "records", readFile, populateFiles)
	if err := filesWidget.keybindings(g); err != nil {
		panic(err)
	}

	recordsWidget := NewListWidget("records", "dir", "dir", readRecord, populateRecords)
	state.filter = &recordFilter{}
	recordsWidget.filterFunc = state.filter.filterFunc
	if err := recordsWidget.keybindings(g); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding("", 'e', gocui.ModNone, state.filter.toggleErrorFilter); err != nil {
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
	state.records = recordsWidget

	if state.dir == "" {
		var err error
		state.dir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}
	fmt.Println(state.dir)

	time.AfterFunc(100*time.Millisecond, func() {
		filesWidget.Init(g, state.dir)
	})

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		panic(err)
	}
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("records", 0, 10, 49, maxY-2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorGreen
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = true
		v.Autoscroll = false
		v.Title = "Records"
	}

	if v, err := g.SetView("header", 50, 10, maxX-60, 25, gocui.BOTTOM|gocui.RIGHT); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorDefault
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = false
		v.Title = "WARC header"
	}

	if v, err := g.SetView("content", 50, 25, maxX-60, maxY-2, gocui.TOP|gocui.RIGHT); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorDefault
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = false
		v.Title = "WARC content"
	}

	if v, err := g.SetView("errors", maxX-60, 10, maxX-1, maxY-2, gocui.LEFT); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorRed
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = false
		v.Title = "Errors"
		v.Wrap = true
	}

	if v, err := g.SetView("help", 0, maxY-2, maxX, maxY, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = false
	}
	state.filter.refreshHelp(g)

	if v, err := g.SetView("dir", 0, 0, maxX-1, 9, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = true
	}

	if _, err := g.SetCurrentView(state.curView); err != nil {
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

func readRecord(g *gocui.Gui, widget *ListWidget) {
	r, err := gowarc.NewWarcFileReader(state.dir+"/"+state.file, widget.filteredRecords[widget.selected].(record).offset)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	rec, _, val, err := r.Next()
	if err != nil {
		panic(err)
	}
	defer rec.Close()

	hv, err := g.View("header")
	if err != nil {
		panic(err)
	}
	hv.Clear()
	rec.WarcHeader().Write(hv)

	cv, err := g.View("content")
	if err != nil {
		panic(err)
	}
	cv.Clear()
	rr, err := rec.Block().RawBytes()
	if err != nil {
		panic(err)
	}
	io.Copy(cv, rr)
	rec.ValidateDigest(val)

	ev, err := g.View("errors")
	if err != nil {
		panic(err)
	}
	ev.Clear()
	fmt.Fprintf(ev, "%s\n", val)
}

type State struct {
	g       *gocui.Gui
	curView string
	dir     string
	file    string
	records *ListWidget
	header  string
	content string
	errors  string
	filter  *recordFilter
}
