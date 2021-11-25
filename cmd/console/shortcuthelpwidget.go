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
 *
 */

package console

import "github.com/awesome-gocui/gocui"

const shortcutHelpWidgetName = "Keyboard shortcuts"

type ShortcutHelpWidget struct {
}

func NewShortcutHelpWidget() *ShortcutHelpWidget {
	return &ShortcutHelpWidget{}
}

func (w *ShortcutHelpWidget) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(shortcutHelpWidgetName, maxX/2-30, maxY/2-20, maxX/2+30, maxY/2+20, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorGreen
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = false
		v.Autoscroll = false
		v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝'}
		v.Editable = true
		v.KeybindOnEdit = false
		v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
			state.modalView = ""
			g.Cursor = false
			g.DeleteView(shortcutHelpWidgetName)
		})
		v.Title = shortcutHelpWidgetName
		v.WriteString("f: find in current window\n\nFilters:\ne: error\ni: warcinfo\nq: request\nr: response\nm: metadata\nv: revisit\ns: resource\nc: continuation\nn: conversion")
	}
	state.modalView = shortcutHelpWidgetName
	return nil
}
