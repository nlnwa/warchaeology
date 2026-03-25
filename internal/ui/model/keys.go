package model

import "github.com/awesome-gocui/gocui"

type KeyBinding struct {
	Key any
	Mod gocui.Modifier
	Fn  func(*gocui.Gui, *gocui.View) error
}
