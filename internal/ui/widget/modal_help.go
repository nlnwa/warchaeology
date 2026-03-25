package widget

import (
	"github.com/awesome-gocui/gocui"
)

const shortcutHelpTitle = "Help"

const shortcutHelpText = `Navigation:
	Tab / Shift+Tab: next / previous view
	Esc: close modal
	Ctrl+C: quit

Search:
	f or /: find in current window

Filters:
	e: error
	i: warcinfo
	q: request
	r: response
	m: metadata
	v: revisit
	s: resource
	c: continuation
	n: conversion

View:
	l: toggle line endings
	z: toggle fullscreen
	Ctrl+Y: toggle mouse (off = terminal copy/paste)`

type ShortcutsModal struct {
	viewName string
	ctrl     Controller
}

func NewModalShortcuts(viewName string, ctrl Controller) *ShortcutsModal {
	return &ShortcutsModal{
		viewName: viewName,
		ctrl:     ctrl,
	}
}

func (w *ShortcutsModal) Init(gui *gocui.Gui) error {
	view, err := gui.View(w.viewName)
	if err != nil {
		return err
	}

	view.Title = shortcutHelpTitle
	view.FrameRunes = DoubleFrame
	view.Editable = true
	view.Editor = gocui.EditorFunc(func(_ *gocui.View, _ gocui.Key, _ rune, _ gocui.Modifier) {
		_ = w.ctrl.CloseModal(gui, nil)
	})

	view.WriteString(shortcutHelpText)

	return nil
}

func (w *ShortcutsModal) Dispose(gui *gocui.Gui) error {
	return nil
}
