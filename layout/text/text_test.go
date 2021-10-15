package text

import (
	"fmt"
	"log"
	"testing"

	"github.com/benoitkugler/go-weasyprint/layout/text/hyphen"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	tu "github.com/benoitkugler/go-weasyprint/utils/testutils"
	"github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango"
	"github.com/benoitkugler/textlayout/pango/fcfonts"
)

var (
	sansFonts = pr.Strings{"DejaVu Sans", "sans"}
	monoFonts = pr.Strings{"DejaVu Sans Mono", "monospace"}
)

const fontmapCache = "test/cache.fc"

var fontmap *fcfonts.FontMap

func init() {
	// this command has to run once
	// fmt.Println("Scanning fonts...")
	// _, err := fontconfig.ScanAndCache(fontmapCache)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fs, err := fontconfig.LoadFontsetFile(fontmapCache)
	if err != nil {
		log.Fatal(err)
	}
	fontmap = fcfonts.NewFontMap(fontconfig.Standard, fs)
}

func assert(t *testing.T, b bool, msg string) {
	if !b {
		t.Fatal(msg)
	}
}

type textContext struct {
	fontmap pango.FontMap
	dict    map[HyphenDictKey]hyphen.Hyphener
}

func (tc textContext) Fontmap() pango.FontMap                            { return tc.fontmap }
func (tc textContext) HyphenCache() map[HyphenDictKey]hyphen.Hyphener    { return tc.dict }
func (tc textContext) StrutLayoutsCache() map[StrutLayoutKey][2]pr.Float { return nil }

// Wrapper for SplitFirstLine() creating a style dict.
func makeText(text string, width pr.MaybeFloat, style pr.Properties) Splitted {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(monoFonts)
	newStyle.UpdateWith(style)
	ct := textContext{fontmap: fontmap, dict: make(map[HyphenDictKey]hyphen.Hyphener)}
	return SplitFirstLine(text, newStyle, ct, width, 0, false)
}

func TestLineContent(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	for _, v := range []struct {
		remaining string
		width     pr.Float
	}{
		{"text for test", 100},
		{"is a text for test", 45},
	} {
		text := "This is a text for test"
		sp := makeText(text, v.width, pr.Properties{"font_family": sansFonts, "font_size": pr.FToV(19)})
		textRunes := []rune(text)
		assert(t, string(textRunes[sp.ResumeAt:]) == v.remaining, "unexpected remaining")
		assert(t, sp.Length+1 == sp.ResumeAt, fmt.Sprintf("%v: expected %d, got %d", v.width, sp.ResumeAt, sp.Length+1)) // +1 for the removed trailing space
	}
}

func TestLineWithAnyWidth(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	sp1 := makeText("some text", nil, nil)
	sp2 := makeText("some text some text", nil, nil)
	assert(t, sp1.Width < sp2.Width, "unexpected width")
}

func TestLineBreaking(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	str := "Thïs is a text for test"
	// These two tests do not really rely on installed fonts
	sp := makeText(str, pr.Float(90), pr.Properties{"font_size": pr.FToV(1)})
	assert(t, sp.ResumeAt == -1, "")

	sp = makeText(str, pr.Float(90), pr.Properties{"font_size": pr.FToV(100)})
	assert(t, string([]rune(str)[sp.ResumeAt:]) == "is a text for test", "")

	sp = makeText(str, pr.Float(100), pr.Properties{"font_family": sansFonts, "font_size": pr.FToV(19)})
	assert(t, string([]rune(str)[sp.ResumeAt:]) == "text for test", "")
}

func TestLineBreakingRTL(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	str := "لوريم ايبسوم دولا"
	// These two tests do not really rely on installed fonts
	sp := makeText(str, pr.Float(90), pr.Properties{"font_size": pr.FToV(1)})
	assert(t, sp.ResumeAt == -1, "")

	sp = makeText(str, pr.Float(90), pr.Properties{"font_size": pr.FToV(100)})
	assert(t, string([]rune(str)[sp.ResumeAt:]) == "ايبسوم دولا", "")
}

func TestTextDimension(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	str := "This is a text for test. This is a test for text.py"
	sp1 := makeText(str, pr.Float(200), pr.Properties{"font_size": pr.FToV(12)})
	sp2 := makeText(str, pr.Float(200), pr.Properties{"font_size": pr.FToV(20)})
	assert(t, sp1.Width*sp1.Height < sp2.Width*sp2.Height, "")
}

func BenchmarkSplitFirstLine(b *testing.B) {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(monoFonts)
	newStyle.UpdateWith(pr.Properties{"font_family": sansFonts, "font_size": pr.FToV(19)})
	ct := textContext{fontmap: fontmap, dict: make(map[HyphenDictKey]hyphen.Hyphener)}

	text := "This is a text for test. This is a test for text.py"
	for i := 0; i < b.N; i++ {
		SplitFirstLine(text, newStyle, ct, pr.Float(200), 0, false)
	}
}
