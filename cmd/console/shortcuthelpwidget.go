package console

import "github.com/awesome-gocui/gocui"

const shortcutHelpWidgetName = "Keyboard shortcuts"

type ShortcutHelpWidget struct {
}

func NewShortcutHelpWidget() *ShortcutHelpWidget {
	return &ShortcutHelpWidget{}
}

func (shortcutHelpWidget *ShortcutHelpWidget) Layout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	if view, err := gui.SetView(shortcutHelpWidgetName, maxX/2-30, maxY/2-20, maxX/2+30, maxY/2+20, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.FgColor = gocui.ColorGreen
		view.BgColor = gocui.ColorDefault
		view.SelBgColor = gocui.ColorWhite
		view.SelFgColor = gocui.ColorBlack
		view.Highlight = false
		view.Autoscroll = false
		view.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝'}
		view.Editable = true
		view.KeybindOnEdit = false
		view.Editor = gocui.EditorFunc(func(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
			state.modalView = ""
			gui.Cursor = false
			_ = gui.DeleteView(shortcutHelpWidgetName)
		})
		view.Title = shortcutHelpWidgetName
		view.WriteString("f: find in current window\n\nFilters:\ne: error\ni: warcinfo\nq: request\nr: response\nm: metadata\nv: revisit\ns: resource\nc: continuation\nn: conversion")
	}
	state.modalView = shortcutHelpWidgetName
	return nil
}
