package console

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

var (
	ErrorColor        = gocui.NewRGBColor(255, 100, 100)
	WarcInfoColor     = gocui.NewRGBColor(100, 255, 255)
	RequestColor      = gocui.NewRGBColor(100, 255, 100)
	MetadataColor     = gocui.NewRGBColor(255, 0, 255)
	ResponseColor     = gocui.NewRGBColor(100, 150, 150)
	ResourceColor     = gocui.NewRGBColor(120, 100, 50)
	RevisitColor      = gocui.NewRGBColor(255, 255, 50)
	ContinuationColor = gocui.NewRGBColor(120, 100, 50)
	ConversionColor   = gocui.NewRGBColor(120, 100, 50)
)

func escapeFgColor(color gocui.Attribute) string {
	if color == gocui.ColorDefault {
		return "\x1b[0m"
	}
	red, green, blue := color.RGB()
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", red, green, blue)
}

func escapeBgColor(color gocui.Attribute) string {
	if color == gocui.ColorDefault {
		return "\x1b[0m"
	}
	red, green, blue := color.RGB()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", red, green, blue)
}
