package images

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/validation"
	"github.com/benoitkugler/go-weasyprint/utils"
)

var (
	// re1 = regexp.MustCompile("(?<!e)-")
	re2 = regexp.MustCompile("[ \n\r\t,]+")
	// re3 = regexp.MustCompile(`(\.[0-9-]+)(?=\.)`)

	UNITS = map[string]pr.Float{
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
	// str = re1.ReplaceAllString(str, " -") // TODO:
	str = re2.ReplaceAllString(str, " ")
	// str = re3.ReplaceAllString(str, `\1 `)  // TODO:
	return strings.TrimSpace(str)
}

type floatOrString struct {
	s string
	f float64
}

func toFloat(s string, offset int) (pr.Float, error) {
	rs := []rune(s)
	s = string(rs[:len(rs)-offset])
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("wrong string for float : %s", s)
	}
	return pr.Float(f), nil
}

// Compute size from string, resolving units and percentages.
func size(str string, fontSize pr.MaybeFloat, reference pr.MaybeFloat) pr.Float {
	if str == "" {
		return 0
	}

	f, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return pr.Float(f)
	}
	// Not a float, try something else

	str = strings.SplitN(normalize(str), " ", 2)[0]
	if strings.HasSuffix(str, "%") {
		v, err := toFloat(str, 1)
		if err != nil {
			log.Println(err)
			return 0
		}
		return v * reference.V() / 100
	} else if strings.HasSuffix(str, "em") {
		v, err := toFloat(str, 2)
		if err != nil {
			log.Println(err)
			return 0
		}
		return fontSize.V() * v
	} else if strings.HasSuffix(str, "ex") {
		// Assume that 1em == 2ex
		v, err := toFloat(str, 2)
		if err != nil {
			log.Println(err)
			return 0
		}
		return fontSize.V() * v / 2
	}

	for suffix, unit := range validation.LENGTHUNITS {
		if coefficient, has := pr.LengthsToPixels[unit]; has && strings.HasSuffix(str, suffix) {
			number, err := toFloat(str, len([]rune(suffix)))
			if err != nil {
				log.Println(err)
				return 0
			}
			return number * coefficient
		}
	}

	// Unknown size
	return 0
}

// Return ``viewbox`` of ``node``.
func getViewBox(node *utils.HTMLNode) []pr.Float {
	viewbox := node.Get("viewBox")
	if viewbox == "" {
		return nil
	}

	var out []pr.Float
	for _, position := range strings.Split(normalize(viewbox), " ") {
		f, err := strconv.ParseFloat(position, 64)
		if err != nil {
			log.Printf("wrong string for float %s in viewbox", position)
			return nil
		}
		out = append(out, pr.Float(f))
	}
	return out
}
