package model

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc/v3"
)

// Record-type colors — vivid foreground colors for dark terminals.
var (
	ErrorColor        = gocui.NewRGBColor(255, 100, 100) // bright red
	WarcInfoColor     = gocui.NewRGBColor(60, 220, 230)  // bright cyan
	RequestColor      = gocui.NewRGBColor(90, 230, 120)  // bright green
	ResponseColor     = gocui.NewRGBColor(100, 160, 255) // bright blue
	MetadataColor     = gocui.NewRGBColor(220, 100, 220) // bright magenta
	ResourceColor     = gocui.NewRGBColor(255, 170, 60)  // orange-amber
	RevisitColor      = gocui.NewRGBColor(230, 220, 70)  // bright gold
	ContinuationColor = gocui.NewRGBColor(80, 210, 190)  // bright teal
	ConversionColor   = gocui.NewRGBColor(190, 130, 255) // bright violet

	// selectionBgColor is the uniform background applied to every selected row.
	selectionBgColor = gocui.NewRGBColor(60, 60, 80)    // dark blue-gray
	selectionFgColor = gocui.NewRGBColor(240, 240, 240) // near-white

	// LineEndsColor is used for the visible line-endings chrome toggle.
	LineEndsColor = gocui.NewRGBColor(180, 180, 60) // muted yellow
)

// Pre-escaped foreground/background sequences, computed once at startup.
var (
	FgError        = EscapeFgColor(ErrorColor)
	FgWarcInfo     = EscapeFgColor(WarcInfoColor)
	FgRequest      = EscapeFgColor(RequestColor)
	FgResponse     = EscapeFgColor(ResponseColor)
	FgMetadata     = EscapeFgColor(MetadataColor)
	FgResource     = EscapeFgColor(ResourceColor)
	FgRevisit      = EscapeFgColor(RevisitColor)
	FgContinuation = EscapeFgColor(ContinuationColor)
	FgConversion   = EscapeFgColor(ConversionColor)
	FgDefault      = EscapeFgColor(gocui.ColorDefault)
	BgDefault      = EscapeBgColor(gocui.ColorDefault)
	// ColorReset resets both foreground and background to terminal defaults.
	ColorReset = BgDefault + FgDefault

	// SGRReset resets all SGR attributes to terminal defaults. Use at the
	// start of each rendered row to prevent attributes from bleeding across rows.
	SGRReset = "\x1b[0m"

	// SelectionBg/SelectionFg are applied to every selected row.
	SelectionBg = EscapeBgColor(selectionBgColor)
	// SelectionFg is near-white text used on the selection background.
	SelectionFg = EscapeFgColor(selectionFgColor)
)

// ErrorRecordType is a synthetic RecordType bit that signals a record has
// validation errors. It is OR'd into RecordItem.RecordType at load time so the
// chrome filter and colorizer treat errors identically to all other record types.
const ErrorRecordType gowarc.RecordType = 256

// RecordTypeFgColor returns the pre-escaped foreground color sequence for rt.
// The ErrorRecordType bit takes priority: a record with errors is shown in
// error color regardless of its actual record type.
func RecordTypeFgColor(rt gowarc.RecordType) string {
	if rt&ErrorRecordType != 0 {
		return FgError
	}
	switch rt &^ ErrorRecordType {
	case gowarc.Warcinfo:
		return FgWarcInfo
	case gowarc.Request:
		return FgRequest
	case gowarc.Response:
		return FgResponse
	case gowarc.Metadata:
		return FgMetadata
	case gowarc.Resource:
		return FgResource
	case gowarc.Revisit:
		return FgRevisit
	case gowarc.Continuation:
		return FgContinuation
	case gowarc.Conversion:
		return FgConversion
	default:
		return FgDefault
	}
}

func EscapeFgColor(color gocui.Attribute) string {
	if color == gocui.ColorDefault {
		return "\x1b[39m"
	}
	r, g, b := color.RGB()
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func EscapeBgColor(color gocui.Attribute) string {
	if color == gocui.ColorDefault {
		return "\x1b[49m"
	}
	r, g, b := color.RGB()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}
