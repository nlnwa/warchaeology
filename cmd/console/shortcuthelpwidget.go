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
			_ = g.DeleteView(shortcutHelpWidgetName)
		})
		v.Title = shortcutHelpWidgetName
		v.WriteString("f: find in current window\n\nFilters:\ne: error\ni: warcinfo\nq: request\nr: response\nm: metadata\nv: revisit\ns: resource\nc: continuation\nn: conversion")
	}
	state.modalView = shortcutHelpWidgetName
	return nil
}
