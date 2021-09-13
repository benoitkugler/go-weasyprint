package images

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/benoitkugler/go-weasyprint/utils"
)

var (
	re1 = regexp.MustCompile("(?<!e)-")
	re2 = regexp.MustCompile("[ \n\r\t,]+")
	re3 = regexp.MustCompile(`(\.[0-9-]+)(?=\.)`)

	UNITS = map[string]float64{
		"mm": 1 / 25.4,
		"cm": 1 / 2.54,
		"in": 1,
		"pt": 1 / 72.,
		"pc": 1 / 6.,
		"px": 1,
	}
)

// Normalize a string corresponding to an array of various values.
func normalize(str string) string {
	str = strings.ReplaceAll(str, "E", "e")
	str = re1.ReplaceAllString(str, " -")
	str = re2.ReplaceAllString(str, " ")
	str = re3.ReplaceAllString(str, `\1 `)
	return strings.TrimSpace(str)
}

type floatOrString struct {
	s string
	f float64
}

func toFloat(s string, offset int) float64 {
	rs := []rune(s)
	s = string(rs[:len(rs)-offset])
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatalf("wrong string for float : %s", s)
	}
	return f
}

// Replace a ``str`` with units by a float value.
// If ``reference`` is a float, it is used as reference for percentages. If it
// is ``"x"``, we use the viewport width as reference. If it is ``"y"``, we
// use the viewport height as reference. If it is ``"xy"``, we use
// ``(viewportWidth ** 2 + viewportHeight ** 2) ** .5 / 2 ** .5`` as
// reference.
// reference="xy"
func size(surface FakeSurface, str string, reference floatOrString) float64 {
	if str == "" {
		return 0
	}

	f, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return f
	}
	// Not a float, try something else

	// No surface (for parsing only)
	if (surface == FakeSurface{}) {
		return 0
	}

	str = strings.SplitN(normalize(str), " ", 1)[0]
	if strings.HasSuffix(str, "%") {
		if reference.s == "x" {
			reference.f = surface.contextWidth
		} else if reference.s == "y" {
			reference.f = surface.contextHeight
		} else if reference.s == "xy" {
			reference.f = math.Sqrt(math.Pow(surface.contextWidth, 2)+math.Pow(surface.contextHeight, 2)) / math.Sqrt(2)
		}
		return toFloat(str, 1) * reference.f / 100
	} else if strings.HasSuffix(str, "em") {
		return surface.fontSize * toFloat(str, 2)
	} else if strings.HasSuffix(str, "ex") {
		// Assume that 1em == 2ex
		return surface.fontSize * toFloat(str, 2) / 2
	}

	for unit, coefficient := range UNITS {
		if strings.HasSuffix(str, unit) {
			number := toFloat(str, len([]rune(unit)))
			return number * surface.dpi * coefficient
		}
	}

	// Unknown size
	return 0
}

// Return ``(width, height, viewbox)`` of ``node``.
// If ``reference`` is ``true``, we can rely on surface size to resolve
// percentages.
// reference=true
func nodeFormat(surface FakeSurface, node *utils.HTMLNode, reference bool) (float64, float64, []float64) {
	var refW, refH floatOrString // init at 0
	if reference {
		refW.s = "x"
		refH.s = "y"
	}
	widthS := node.Get("width")
	if widthS == "" {
		widthS = "100%"
	}
	heightS := node.Get("height")
	if heightS == "" {
		heightS = "100%"
	}
	width := size(surface, widthS, refW)
	height := size(surface, heightS, refH)
	viewbox := node.Get("viewBox")
	var viewboxF []float64
	if viewbox != "" {
		viewbox = re2.ReplaceAllString(viewbox, " ")
		for _, position := range strings.Split(viewbox, " ") {
			f, err := strconv.ParseFloat(position, 64)
			if err != nil {
				log.Fatalf("wring string for float %s", position)
			}
			viewboxF = append(viewboxF, f)
		}
		if width == 0 {
			width = viewboxF[2]
		}
		if height == 0 {
			height = viewboxF[3]
		}
	}
	return width, height, viewboxF
}
