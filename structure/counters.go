package structure

import (
	"math"
	"strconv"
	"strings"
)

//Implement the various counter types and list-style-type values.
//
//These are defined in the same terms as CSS 3 Lists:
//http://dev.w3.org/csswg/css3-lists/#predefined-counters
//
//:copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
//:license: BSD, see LICENSE for details.

type counterStyleDescriptor struct {
	//negative [2]string

	prefix   string
	suffix   string
	range_   [2]int
	fallback string

	//type_ string
	//symbols TBD

	formatter counterImplementation
}

type counterImplementation = func(value int) (string, bool)

type valueSymbol struct {
	weight int
	symbol string
}

type nonRepeatingSymbols struct {
	firstValue int
	symbols    []string
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
		{1000, "m"}, {900, "cm"}, {500, "d"}, {400, "cd"},
		{100, "c"}, {90, "xc"}, {50, "l"}, {40, "xl"},
		{10, "x"}, {9, "ix"}, {5, "v"}, {4, "iv"},
		{1, "i"},
	}

	upperRoman = []valueSymbol{
		{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
		{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
		{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"},
		{1, "I"},
	}

	georgian = []valueSymbol{
		{10000, "ჵ"}, {9000, "ჰ"}, {8000, "ჯ"}, {7000, "ჴ"}, {6000, "ხ"},
		{5000, "ჭ"}, {4000, "წ"}, {3000, "ძ"}, {2000, "ც"}, {1000, "ჩ"},
		{900, "შ"}, {800, "ყ"}, {700, "ღ"}, {600, "ქ"},
		{500, "ფ"}, {400, "ჳ"}, {300, "ტ"}, {200, "ს"}, {100, "რ"},
		{90, "ჟ"}, {80, "პ"}, {70, "ო"}, {60, "ჲ"},
		{50, "ნ"}, {40, "მ"}, {30, "ლ"}, {20, "კ"}, {10, "ი"},
		{9, "თ"}, {8, "ჱ"}, {7, "ზ"}, {6, "ვ"},
		{5, "ე"}, {4, "დ"}, {3, "გ"}, {2, "ბ"}, {1, "ა"},
	}

	armenian = []valueSymbol{
		{9000, "Ք"}, {8000, "Փ"}, {7000, "Ւ"}, {6000, "Ց"},
		{5000, "Ր"}, {4000, "Տ"}, {3000, "Վ"}, {2000, "Ս"}, {1000, "Ռ"},
		{900, "Ջ"}, {800, "Պ"}, {700, "Չ"}, {600, "Ո"},
		{500, "Շ"}, {400, "Ն"}, {300, "Յ"}, {200, "Մ"}, {100, "Ճ"},
		{90, "Ղ"}, {80, "Ձ"}, {70, "Հ"}, {60, "Կ"},
		{50, "Ծ"}, {40, "Խ"}, {30, "Լ"}, {20, "Ի"}, {10, "Ժ"},
		{9, "Թ"}, {8, "Ը"}, {7, "Է"}, {6, "Զ"},
		{5, "Ե"}, {4, "Դ"}, {3, "Գ"}, {2, "Բ"}, {1, "Ա"},
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

//// Implement the algorithm for `type: numeric`.
//func numeric(symbols []string, value int) (string, bool) {
//	if value == 0 {
//		return symbols[0], true
//
//	}
//	isNegative := value < 0
//	var reversedParts []string
//	value = abs(value)
//	prefix, suffix := negative[0], negative[1]
//	if isNegative {
//		reversedParts = []string{suffix}
//	}
//	length := len(symbols)
//	for value != 0 {
//		reversedParts = append(reversedParts, symbols[value%length])
//		value /= length
//	}
//	if isNegative {
//		reversedParts = append(reversedParts, prefix)
//
//	}
//	reverse(reversedParts)
//	return strings.Join(reversedParts, ""), true
//}

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

//// Implement the algorithm for `type: symbolic`.
//func symbolic(symbols []string, value int) (string, bool) {
//	if value <= 0 {
//		return "", false
//	}
//	length := len(symbols)
//	return strings.Repeat(symbols[value%length], (value-1)/length), true
//}

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
func format(value int, counterStyle string) string {
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
	return style.prefix + format(value, counterStyle) + style.suffix
}
