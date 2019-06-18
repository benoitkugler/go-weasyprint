package css

import "golang.org/x/net/html"

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
