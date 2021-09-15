package counters

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Implement the various counter types and list-style-type values.
//
// These are defined in the same terms as CSS 3 Lists:
// http://dev.w3.org/csswg/css3-lists/#predefined-counters

type CounterStyle map[string]CounterStyleDescriptors

type CounterStyleDescriptors struct {
	Negative        [2]pr.NamedString
	System          CounterStyleSystem
	Prefix, Suffix  pr.NamedString
	Range           []pr.StringRange
	Fallback        string
	Pad             pr.IntNamedString
	Symbols         []pr.NamedString
	AdditiveSymbols []pr.IntNamedString
}

func (desc *CounterStyleDescriptors) Validate() error {
	system := desc.System
	if system == (CounterStyleSystem{}) {
		system = CounterStyleSystem{SecondKeyword: "symbolic"}
	}
	if system.Keyword == "" {
		switch system.SecondKeyword {
		case "cyclic", "fixed", "symbolic":
			if len(desc.Symbols) == 0 {
				return fmt.Errorf("counter style %s needs at least one symbol", system.SecondKeyword)
			}
		case "alphabetic", "numeric":
			if len(desc.Symbols) < 2 {
				return fmt.Errorf("counter style %s needs at least two symbols", system.SecondKeyword)
			}
		case "additive":
			if len(desc.AdditiveSymbols) < 2 {
				return fmt.Errorf("counter style %s needs at least two additive symbols", system.SecondKeyword)
			}
		}
	}
	return nil
}

type CounterStyleSystem struct {
	Keyword, SecondKeyword string
	Number                 int
}

type counterStyleDescriptor struct {
	formatter counterImplementation
	prefix    string
	suffix    string
	fallback  string
	range_    [2]int
}

type counterImplementation = func(value int) (string, bool)

type valueSymbol struct {
	symbol string
	weight int
}

type nonRepeatingSymbols struct {
	symbols    []string
	firstValue int
}

var (
	negative = [2]string{"-", ""}

	// Initial values for counter style descriptors.
	INITIALVALUES = counterStyleDescriptor{
		suffix:   ".",
		range_:   [2]int{int(math.Inf(-1)), int(math.Inf(1))},
		fallback: "decimal",
		// type and symbols ommited here.
	}

	// Maps counter-style names to a dict of descriptors.
	STYLES = map[string]counterStyleDescriptor{
		// Included here for formatListMarker().
		// format() special-cases decimal and does not use this.
		"decimal": INITIALVALUES,
	}

	lowerRoman = []valueSymbol{
		{"m", 1000},
		{"cm", 900},
		{"d", 500},
		{"cd", 400},
		{"c", 100},
		{"xc", 90},
		{"l", 50},
		{"xl", 40},
		{"x", 10},
		{"ix", 9},
		{"v", 5},
		{"iv", 4},
		{"i", 1},
	}

	upperRoman = []valueSymbol{
		{"M", 1000},
		{"CM", 900},
		{"D", 500},
		{"CD", 400},
		{"C", 100},
		{"XC", 90},
		{"L", 50},
		{"XL", 40},
		{"X", 10},
		{"IX", 9},
		{"V", 5},
		{"IV", 4},
		{"I", 1},
	}

	georgian = []valueSymbol{
		{"ჵ", 10000},
		{"ჰ", 9000},
		{"ჯ", 8000},
		{"ჴ", 7000},
		{"ხ", 6000},
		{"ჭ", 5000},
		{"წ", 4000},
		{"ძ", 3000},
		{"ც", 2000},
		{"ჩ", 1000},
		{"შ", 900},
		{"ყ", 800},
		{"ღ", 700},
		{"ქ", 600},
		{"ფ", 500},
		{"ჳ", 400},
		{"ტ", 300},
		{"ს", 200},
		{"რ", 100},
		{"ჟ", 90},
		{"პ", 80},
		{"ო", 70},
		{"ჲ", 60},
		{"ნ", 50},
		{"მ", 40},
		{"ლ", 30},
		{"კ", 20},
		{"ი", 10},
		{"თ", 9},
		{"ჱ", 8},
		{"ზ", 7},
		{"ვ", 6},
		{"ე", 5},
		{"დ", 4},
		{"გ", 3},
		{"ბ", 2},
		{"ა", 1},
	}

	armenian = []valueSymbol{
		{"Ք", 9000},
		{"Փ", 8000},
		{"Ւ", 7000},
		{"Ց", 6000},
		{"Ր", 5000},
		{"Տ", 4000},
		{"Վ", 3000},
		{"Ս", 2000},
		{"Ռ", 1000},
		{"Ջ", 900},
		{"Պ", 800},
		{"Չ", 700},
		{"Ո", 600},
		{"Շ", 500},
		{"Ն", 400},
		{"Յ", 300},
		{"Մ", 200},
		{"Ճ", 100},
		{"Ղ", 90},
		{"Ձ", 80},
		{"Հ", 70},
		{"Կ", 60},
		{"Ծ", 50},
		{"Խ", 40},
		{"Լ", 30},
		{"Ի", 20},
		{"Ժ", 10},
		{"Թ", 9},
		{"Ը", 8},
		{"Է", 7},
		{"Զ", 6},
		{"Ե", 5},
		{"Դ", 4},
		{"Գ", 3},
		{"Բ", 2},
		{"Ա", 1},
	}

	decimalLeadingZero = nonRepeatingSymbols{
		firstValue: -9,
		symbols:    []string{"-09", "-08", "-07", "-06", "-05", "-04", "-03", "-02", "-01", "00", "01", "02", "03", "04", "05", "06", "07", "08", "09"},
	}

	lowerAlpha = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	upperAlpha = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	lowerGreek = []string{"α", "β", "γ", "δ", "ε", "ζ", "η", "θ", "ι", "κ", "λ", "μ", "ν", "ξ", "ο", "π", "ρ", "σ", "τ", "υ", "φ", "χ", "ψ", "ω"}

	disc   = []string{"•"}
	circle = []string{"◦"}
	square = []string{"▪"} // CSS Lists 3 suggests U+25FE BLACK MEDIUM SMALL SQUARE, But I think this one looks better.
)

func init() {
	STYLES["decimal-leading-zero"] = fromInitialValues(counterStyleDescriptor{
		suffix: INITIALVALUES.suffix,
		formatter: func(value int) (s string, b bool) {
			return nonRepeating(decimalLeadingZero, value)
		},
	})
	STYLES["lower-roman"] = fromInitialValues(counterStyleDescriptor{
		suffix: INITIALVALUES.suffix,
		formatter: func(value int) (s string, b bool) {
			return additive(lowerRoman, value)
		},
		range_: [2]int{1, 4999},
	})
	STYLES["upper-roman"] = fromInitialValues(counterStyleDescriptor{
		suffix: INITIALVALUES.suffix,
		formatter: func(value int) (s string, b bool) {
			return additive(upperRoman, value)
		},
		range_: [2]int{1, 4999},
	})
	STYLES["georgian"] = fromInitialValues(counterStyleDescriptor{
		suffix: INITIALVALUES.suffix,
		formatter: func(value int) (s string, b bool) {
			return additive(georgian, value)
		},
		range_: [2]int{1, 19999},
	})
	STYLES["armenian"] = fromInitialValues(counterStyleDescriptor{
		suffix: INITIALVALUES.suffix,
		formatter: func(value int) (s string, b bool) {
			return additive(armenian, value)
		},
		range_: [2]int{1, 9999},
	})
	STYLES["lower-alpha"] = fromInitialValues(counterStyleDescriptor{
		suffix: INITIALVALUES.suffix,
		formatter: func(value int) (s string, b bool) {
			return alphabetic(lowerAlpha, value)
		},
	})
	STYLES["upper-alpha"] = fromInitialValues(counterStyleDescriptor{
		suffix: INITIALVALUES.suffix,
		formatter: func(value int) (s string, b bool) {
			return alphabetic(upperAlpha, value)
		},
	})
	STYLES["lower-greek"] = fromInitialValues(counterStyleDescriptor{
		suffix: INITIALVALUES.suffix,
		formatter: func(value int) (s string, b bool) {
			return alphabetic(lowerGreek, value)
		},
	})
	STYLES["disc"] = fromInitialValues(counterStyleDescriptor{
		suffix: "",
		formatter: func(value int) (s string, b bool) {
			return repeating(disc, value)
		},
	})
	STYLES["circle"] = fromInitialValues(counterStyleDescriptor{
		suffix: "",
		formatter: func(value int) (s string, b bool) {
			return repeating(circle, value)
		},
	})
	STYLES["square"] = fromInitialValues(counterStyleDescriptor{
		suffix: "",
		formatter: func(value int) (s string, b bool) {
			return repeating(square, value)
		},
	})

	// TODO: when @counter-style rules are supported, change override
	//// to bind when a value is generated, not when the @rule is parsed.
	STYLES["lower-latin"] = STYLES["lower-alpha"]
	STYLES["upper-latin"] = STYLES["upper-alpha"]
}

func fromInitialValues(c counterStyleDescriptor) counterStyleDescriptor {
	if c.range_ == [2]int{} {
		c.range_ = INITIALVALUES.range_
	}
	if c.fallback == "" {
		c.fallback = INITIALVALUES.fallback
	}
	return c
}

func reverse(a []string) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}

func abs(v int) int {
	return int(math.Abs(float64(v)))
}

// Implement the algorithm for `type: repeating`.
func repeating(symbols []string, value int) (string, bool) {
	return symbols[(value-1)%len(symbols)], true
}

// Implement the algorithm for `type: alphabetic`.
func alphabetic(symbols []string, value int) (string, bool) {
	if value <= 0 {
		return "", false
	}
	length := len(symbols)
	reversedParts := []string{}
	for value != 0 {
		value -= 1
		reversedParts = append(reversedParts, symbols[value%length])
		value /= length
	}
	reverse(reversedParts)
	return strings.Join(reversedParts, ""), true
}

// Implement the algorithm for `type: non-repeating`.
func nonRepeating(symbols nonRepeatingSymbols, value int) (string, bool) {
	value -= symbols.firstValue
	if 0 <= value && value < len(symbols.symbols) {
		return symbols.symbols[value], true
	}
	return "", false
}

// Implement the algorithm for `type: additive`.
func additive(symbols []valueSymbol, value int) (string, bool) {
	if value == 0 {
		for _, vs := range symbols {
			if vs.weight == 0 {
				return vs.symbol, true
			}
		}
	}
	isNegative := value < 0
	prefix, suffix := negative[0], negative[1]
	var parts []string
	if isNegative {
		value = abs(value)
		parts = []string{prefix}

	}
	for _, vs := range symbols {
		repetitions := value / vs.weight
		parts = append(parts, strings.Repeat(vs.symbol, repetitions))
		value -= vs.weight * repetitions
		if value == 0 {
			if isNegative {
				parts = append(parts, suffix)
			}
			return strings.Join(parts, ""), true

		}
	}
	return "", false //  Failed to find a representation for this value
}

// Return a representation of ``value`` formatted by ``counterStyle``
//or one of its fallback.
//
//The representation includes negative signs, but not the prefix and suffix.
func Format(value int, counterStyle string) string {
	if counterStyle == "none" {
		return ""
	}
	failedStyles := map[string]bool{} // avoid fallback loops
	for {
		if counterStyle == "decimal" || failedStyles[counterStyle] {
			return strconv.Itoa(value)
		}
		style := STYLES[counterStyle]
		low, high := style.range_[0], style.range_[1]
		if low <= value && value <= high {
			representation, ok := style.formatter(value)
			if ok {
				return representation
			}
		}
		failedStyles[counterStyle] = true
		counterStyle = style.fallback
	}
}

// Return a representation of ``value`` formatted for a list marker.
//
//This is the same as :func:`format()`, but includes the counter’s
//prefix and suffix.
func FormatListMarker(value int, counterStyle string) string {
	style := STYLES[counterStyle]
	return style.prefix + Format(value, counterStyle) + style.suffix
}

// ----------------- Unused formatters -------------------------------------------

// // Implement the algorithm for `type: numeric`.
// func numeric(symbols []string, value int) (string, bool) {
// 	if value == 0 {
// 		return symbols[0], true

// 	}
// 	isNegative := value < 0
// 	var reversedParts []string
// 	value = abs(value)
// 	prefix, suffix := negative[0], negative[1]
// 	if isNegative {
// 		reversedParts = []string{suffix}
// 	}
// 	length := len(symbols)
// 	for value != 0 {
// 		reversedParts = append(reversedParts, symbols[value%length])
// 		value /= length
// 	}
// 	if isNegative {
// 		reversedParts = append(reversedParts, prefix)

// 	}
// 	reverse(reversedParts)
// 	return strings.Join(reversedParts, ""), true
// }

//// Implement the algorithm for `type: symbolic`.
//func symbolic(symbols []string, value int) (string, bool) {
//	if value <= 0 {
//		return "", false
//	}
//	length := len(symbols)
//	return strings.Repeat(symbols[value%length], (value-1)/length), true
//}
