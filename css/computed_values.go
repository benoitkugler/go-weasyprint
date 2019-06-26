package css

import "golang.org/x/net/html"

var ZeroPixels = Dimension{Unit: "px", Value: 0}

// Return a dict of computed values.

// :param element: The HTML element these style apply to
// :param pseudo_type: The type of pseudo-element, eg 'before', None
// :param specified: a dict of specified values. Should contain
// 			  values for all properties.
// :param computed: a dict of already known computed values.
// 			 Only contains some properties (or none).
// :param parent_style: a dict of computed values of the parent
// 				 element (should contain values for all properties),
// 				 or ``None`` if ``element`` is the root element.
// :param base_url: The base URL used to resolve relative URLs.
func compute(element html.Node, pseudoType string,
	specified, computed, parent_style,
	root_style StyleDict, base_url string) StyleDict {
	// def computer():
	// """Dummy object that holds attributes."""
	// return 0

	// computer.is_root_element = parent_style is None
	// if parent_style is None:
	// parent_style = INITIAL_VALUES

	// computer.element = element
	// computer.pseudo_type = pseudo_type
	// computer.specified = specified
	// computer.computed = computed
	// computer.parent_style = parent_style
	// computer.root_style = root_style
	// computer.base_url = base_url

	// getter = COMPUTER_FUNCTIONS.get

	// for name in COMPUTING_ORDER:
	// if name in computed:
	// 	# Already computed
	// 	continue

	// value = specified[name]
	// function = getter(name)
	// if function is not None:
	// 	value = function(computer, name, value)
	// # else: same as specified

	// computed[name] = value

	// computed['_weasy_specified_display'] = specified['display']
	return computed
}

type TBD struct{}

type Value struct {
	Dimension
	None    bool
	Auto    bool
	Content bool
}

type GradientValue struct {
	StopPositions []Value
	Center
	SizeType string
	Size     int
}

type Gradient struct {
	Type  string
	Value GradientValue
}

// backgroundImage computes lenghts in gradient background-image.
func backgroundImage(computer TBD, name string, values []Gradient) []Gradient {
	for _, gradient := range values {
		value := gradient.Value
		if gradient.Type == "linear-gradient" || gradient.Type == "radial-gradient" {
			for index, pos := range value.StopPositions {
				if pos.None {
					value.StopPositions[index] = Value{None: true}
				} else {
					value.StopPositions[index] = length(computer, name, pos)
				}
			}
		}
		if gradient.Type == "radial-gradient" {
			value.Center = backgroundPosition(computer, name, []int{value.Center})[0]
			if value.SizeType == "explicit" {
				value.Size = lengthOrPercentageTuple(computer, name, value.Size)
			}
		}
	}
	return values
}

// Compute a length ``value``.
// passing a negative fontSize means null
func length(computer TBD, name string, value Value, fontSize int) Dimension {
	if value.Auto || value.Content {
		return value.Dimension
	}
	if value.Value == 0 {
		return ZEROPIXELS
	}

	unit := value.Unit
	_, in := LENGTHSTOPIXELS[unit]
	var result int
	switch unit {
	case "px":
		return value.Dimension
	case in:
		// Convert absolute lengths to pixels
		result = value.Value * LENGTHSTOPIXELS[unit]
	case "em", "ex", "ch", "rem":
		if fontSize < 0 {
			fontSize = computer.computed["fontSize"]
		}
		switch unit {
		case "ex":
			// TODO: cache
			result = value.value * fontSize * exRatio(computer.computed)
		case "ch":
			// TODO: cache
			// TODO: use context to use @font-face fonts
			layout := text.Layout(nil, fontSize, computer.computed)
			layout.setText("0")
			line := layout.iterLines()[0]
			logicalWidth, _ = text.getSize(line, computer.computed)
			result = value.value * logicalWidth
		case "em":
			result = value.value * fontSize
		case "rem":
			result = value.value * computer.rootStyle["fontSize"]
		default:
			// A percentage or "auto": no conversion needed.
			return value.Dimension
		}
		return Dimension{Value: result, Unit: "px"}
	}
}
