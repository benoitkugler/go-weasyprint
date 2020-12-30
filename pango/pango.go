package pango

import (
	"fmt"
	"log"

	"github.com/benoitkugler/go-weasyprint/fribidi"
	"golang.org/x/text/width"
)

// enables additional checks, to use only during developpement or testing
const debugMode = false

// assert is only used in debug mode
func assert(b bool) {
	if !b {
		log.Fatal("assertion error")
	}
}

func showDebug(where string, line *LayoutLine, state *ParaBreakState) {
	line_width := line.pango_layout_line_get_width()

	fmt.Printf("rem %d + line %d = %d		%s",
		state.remaining_width,
		line_width,
		state.remaining_width+line_width,
		where)
}

// Alignment describes how to align the lines of a `Layout` within the
// available space. If the `Layout` is set to justify
// using pango_layout_set_justify(), this only has effect for partial lines.
type Alignment uint8

const (
	PANGO_ALIGN_LEFT   Alignment = iota // Put all available space on the right
	PANGO_ALIGN_CENTER                  // Center the line within the available space
	PANGO_ALIGN_RIGHT                   // Put all available space on the left
)

// Rectangle represents a rectangle. It is frequently
// used to represent the logical or ink extents of a single glyph or section
// of text. (See, for instance, pango_font_get_glyph_extents())
type Rectangle struct {
	x      int // X coordinate of the left side of the rectangle.
	y      int // Y coordinate of the the top side of the rectangle.
	width  int // width of the rectangle.
	height int // height of the rectangle.
}

const maxInt = int(^uint32(0) >> 1)

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func minL(a, b fribidi.Level) fribidi.Level {
	if a > b {
		return b
	}
	return a
}

func maxG(a, b GlyphUnit) GlyphUnit {
	if a < b {
		return b
	}
	return a
}

// pango_is_zero_width checks `ch` to see if it is a character that should not be
// normally rendered on the screen.  This includes all Unicode characters
// with "ZERO WIDTH" in their name, as well as bidi formatting characters, and
// a few other ones.
func pango_is_zero_width(ch rune) bool {
	//  00AD  SOFT HYPHEN
	//  034F  COMBINING GRAPHEME JOINER
	//
	//  200B  ZERO WIDTH SPACE
	//  200C  ZERO WIDTH NON-JOINER
	//  200D  ZERO WIDTH JOINER
	//  200E  LEFT-TO-RIGHT MARK
	//  200F  RIGHT-TO-LEFT MARK
	//
	//  2028  LINE SEPARATOR
	//
	//  202A  LEFT-TO-RIGHT EMBEDDING
	//  202B  RIGHT-TO-LEFT EMBEDDING
	//  202C  POP DIRECTIONAL FORMATTING
	//  202D  LEFT-TO-RIGHT OVERRIDE
	//  202E  RIGHT-TO-LEFT OVERRIDE
	//
	//  2060  WORD JOINER
	//  2061  FUNCTION APPLICATION
	//  2062  INVISIBLE TIMES
	//  2063  INVISIBLE SEPARATOR
	//
	//  FEFF  ZERO WIDTH NO-BREAK SPACE
	return (ch & ^0x007F == 0x2000 &&
		((ch >= 0x200B && ch <= 0x200F) || (ch >= 0x202A && ch <= 0x202E) ||
			(ch >= 0x2060 && ch <= 0x2063) || (ch == 0x2028))) ||
		(ch == 0x00AD || ch == 0x034F || ch == 0xFEFF)
}

// return true for east asian wide characters
func isWide(r rune) bool {
	switch width.LookupRune(r).Kind() {
	case width.EastAsianFullwidth, width.EastAsianWide:
		return true
	default:
		return false
	}
}
