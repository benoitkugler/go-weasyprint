package structure

import (
	"github.com/benoitkugler/go-weasyprint/css"
	"github.com/labstack/gommon/log"
	"golang.org/x/net/html"
)

//    Turn an element tree with associated CSS style (computed values)
//    into a "before layout" formatting structure / box tree.
//
//    This includes creating anonymous boxes and processing whitespace
//    as necessary.
//
//    :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
//    :license: BSD, see LICENSE for details.

var ()

type state struct {
	quoteDepth    int
	counterValues map[string][]int
	counterScopes [](map[string]bool)
}

// Maps values of the ``display`` CSS property to box types.
func makeBox(elementTag string, style css.StyleDict, content []AllBox) AllBox {
	switch style.Display {
	case "block":
		return NewBlockBox(elementTag, style, content)
	case "list-item":
		return NewBlockBox(elementTag, style, content)
	case "inline":
		return NewInlineBox(elementTag, style, content)
	case "inline-block":
		return NewInlineBlockBox(elementTag, style, content)
	case "table":
		return NewTableBox(elementTag, style, content)
	case "inline-table":
		return NewInlineTableBox(elementTag, style, content)
	case "table-row":
		return NewTableRowBox(elementTag, style, content)
	case "table-row-group":
		return NewTableRowGroupBox(elementTag, style, content)
	case "table-header-group":
		return NewTableRowGroupBox(elementTag, style, content)
	case "table-footer-group":
		return NewTableRowGroupBox(elementTag, style, content)
	case "table-column":
		return NewTableColumnBox(elementTag, style, content)
	case "table-column-group":
		return NewTableColumnGroupBox(elementTag, style, content)
	case "table-cell":
		return NewTableCellBox(elementTag, style, content)
	case "table-caption":
		return NewTableCaptionBox(elementTag, style, content)
	}
	panic("display property not supported")
}

// Build a formatting structure (box tree) from an element tree.
func buildFormattingStructure(elementTree html.Node, styleFor func(element html.Node, pseudoType string) *css.StyleDict, getImageFromUri TBD, baseUrl string) AllBox {
	boxList := elementToBox(elementTree, styleFor, getImageFromUri, baseUrl, nil)
	var box AllBox
	if len(boxList) > 0 {
		box = boxList[0]
	} else { //  No root element
		rootStyleFor := func(element html.Node, pseudoType string) *css.StyleDict {
			style := styleFor(element, pseudoType)
			if style != nil {
				// TODO: we should check that the element has a parent instead.
				if element.Tag == "html" {
					style.Display = "block"
				} else {
					style.Display = "none"
				}
			}
			return style
		}
		box = elementToBox(elementTree, rootStyleFor, getImageFromUri, baseUrl, nil)[0]
	}
	box.SetIsForRootElement(true)
	// If this is changed, maybe update layout.pages.makeMarginBoxes()
	processWhitespace(box, false)
	box = anonymousTableBoxes(box)
	box = inlineInBlock(box)
	box = blockInInline(box)
	box = setViewportOverflow(box)
	return box
}

// Convert an element and its children into a box with children.
//
//    Return a list of boxes. Most of the time the list will have one item but
//    may have zero or more than one.
//
//    Eg.::
//
//        <p>Some <em>emphasised</em> text.</p>
//
//    gives (not actual syntax)::
//
//        BlockBox[
//            TextBox["Some "],
//            InlineBox[
//                TextBox["emphasised"],
//            ],
//            TextBox[" text."],
//        ]
//
//    ``TextBox``es are anonymous inline boxes:
//    See http://www.w3.org/TR/CSS21/visuren.html#anonymous
func elementToBox(element html.Node, styleFor func(element html.Node, pseudoType string) *css.StyleDict,
	getImageFromUri TBD, baseUrl string, state *state) []AllBox {

	if element.Type != html.TextNode && element.Type != html.ElementNode && element.Type != html.DocumentNode {
		// Here we ignore comments and XML processing instructions.
		return nil
	}

	style := styleFor(element)

	// TODO: should be the used value. When does the used value for `display`
	// differ from the computer value?
	display := style.Display
	if display == "none" {
		return nil
	}

	box = makeBox(element.Data, style, nil)

	if state == nil {
		// use a list to have a shared mutable object
		state = &state{
			// Shared mutable objects:
			quoteDepth:    0,                                    // single integer
			counterValues: map[string][]int{},                   // name -> stacked/scoped values
			counterScopes: []map[string]bool{map[string]bool{}}, //  element tree depths -> counter names
		}
	}

	QuoteDepth, counterValues, counterScopes = state.quoteDepth, state.counterValues, state.counterScopes

	updateCounters(state, style)

	var children []AllBox
	if display == "list-item" {
		children = append(children,
			addBoxMarker(box, counterValues, getImageFromUri)...)
	}

	// If this element’s direct children create new scopes, the counter
	// names will be in this new list
	counterScopes.append(set())

	box.firstLetterStyle = styleFor(element, "first-letter")
	box.firstLineStyle = styleFor(element, "first-line")

	children.extend(beforeAfterToBox(
		element, "before", state, styleFor, getImageFromUri))
	text = element.text
	if text {
		children.append(boxes.TextBox.anonymousFrom(box, text))
	}

	for _, childElement := range element {
		children.extend(elementToBox(
			childElement, styleFor, getImageFromUri, baseUrl, state))
		text = childElement.tail
		if text {
			textBox = boxes.TextBox.anonymousFrom(box, text)
			if children && isinstance(children[-1], boxes.TextBox) {
				children[-1].text += textBox.text
			} else {
				children.append(textBox)
			}
		}
	}
	children.extend(beforeAfterToBox(
		element, "after", state, styleFor, getImageFromUri))

	// Scopes created by this element’s children stop here.
	for _, name := range counterScopes.pop() {
		counterValues[name].pop()
		if !counterValues[name] {
			counterValues.pop(name)
		}
	}
	box.children = children
	setContentLists(element, box, style, counterValues)

	// Specific handling for the element. (eg. replaced element)
	return html.handleElement(element, box, getImageFromUri, baseUrl)

}

func processWhitespace(box AllBox, followingCollapsibleSpace bool) bool {}

// Remove and add boxes according to the table model.
//
//Take and return a ``Box`` object.
//
//See http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes
func anonymousTableBoxes(box AllBox) AllBox {}

// Build the structure of lines inside blocks and return a new box tree.
//
//    Consecutive inline-level boxes in a block container box are wrapped into a
//    line box, itself wrapped into an anonymous block box.
//
//    This line box will be broken into multiple lines later.
//
//    This is the first case in
//    http://www.w3.org/TR/CSS21/visuren.html#anonymous-block-level
//
//    Eg.::
//
//        BlockBox[
//            TextBox["Some "],
//            InlineBox[TextBox["text"]],
//            BlockBox[
//                TextBox["More text"],
//            ]
//        ]
//
//    is turned into::
//
//        BlockBox[
//            AnonymousBlockBox[
//                LineBox[
//                    TextBox["Some "],
//                    InlineBox[TextBox["text"]],
//                ]
//            ]
//            BlockBox[
//                LineBox[
//                    TextBox["More text"],
//                ]
//            ]
//        ]
func inlineInBlock(box AllBox) AllBox {}

// Build the structure of blocks inside lines.
//
//    Inline boxes containing block-level boxes will be broken in two
//    boxes on each side on consecutive block-level boxes, each side wrapped
//    in an anonymous block-level box.
//
//    This is the second case in
//    http://www.w3.org/TR/CSS21/visuren.html#anonymous-block-level
//
//    Eg. if this is given::
//
//        BlockBox[
//            LineBox[
//                InlineBox[
//                    TextBox["Hello."],
//                ],
//                InlineBox[
//                    TextBox["Some "],
//                    InlineBox[
//                        TextBox["text"]
//                        BlockBox[LineBox[TextBox["More text"]]],
//                        BlockBox[LineBox[TextBox["More text again"]]],
//                    ],
//                    BlockBox[LineBox[TextBox["And again."]]],
//                ]
//            ]
//        ]
//
//    this is returned::
//
//        BlockBox[
//            AnonymousBlockBox[
//                LineBox[
//                    InlineBox[
//                        TextBox["Hello."],
//                    ],
//                    InlineBox[
//                        TextBox["Some "],
//                        InlineBox[TextBox["text"]],
//                    ]
//                ]
//            ],
//            BlockBox[LineBox[TextBox["More text"]]],
//            BlockBox[LineBox[TextBox["More text again"]]],
//            AnonymousBlockBox[
//                LineBox[
//                    InlineBox[
//                    ]
//                ]
//            ],
//            BlockBox[LineBox[TextBox["And again."]]],
//            AnonymousBlockBox[
//                LineBox[
//                    InlineBox[
//                    ]
//                ]
//            ],
//        ]
func blockInInline(box AllBox) AllBox {}

//  Set a ``viewportOverflow`` attribute on the box for the root element.
//
//    Like backgrounds, ``overflow`` on the root element must be propagated
//    to the viewport.
//
//    See http://www.w3.org/TR/CSS21/visufx.html#overflow
func setViewportOverflow(rootBox AllBox) AllBox {}

// Handle the ``counter-*`` properties.
func updateCounters(state *state, style *css.StyleDict) {
	_, counterValues, counterScopes := state.quoteDepth, state.counterValues, state.counterScopes
	siblingScopes := counterScopes[len(counterScopes)-1]

	for _, nv := range style.CounterReset {
		if siblingScopes[nv.Name] {
			delete(counterValues, nv.Name)
		} else {
			siblingScopes[nv.Name] = true

		}
		counterValues[nv.Name] = append(counterValues[nv.Name], nv.Value)
	}

	// XXX Disabled for now, only exists in Lists3’s editor’s draft.
	//    for name, value in style.counterSet:
	//        values = counterValues.setdefault(name, [])
	//        if not values:
	//            assert name not in siblingScopes
	//            siblingScopes.add(name)
	//            values.append(0)
	//        values[-1] = value

	counterIncrement, cis := style.CounterIncrement, style.CounterIncrement.CI
	if counterIncrement.Auto {
		// "auto" is the initial value but is not valid in stylesheet:
		// there was no counter-increment declaration for this element.
		// (Or the winning value was "initial".)
		// http://dev.w3.org/csswg/css3-lists/#declaring-a-list-item
		if style.Display == "list-item" {
			cis = []css.CounterIncrement{{"list-item", 1}}
		} else {
			cis = nil
		}
	}
	for _, ci := range cis {
		values := counterValues[ci.Name]
		if len(values) == 0 {
			if siblingScopes[ci.Name] {
				log.Fatal("ci.Name shoud'nt be in siblingScopes")
			}
			siblingScopes[ci.Name] = true
			values = append(values, 0)
		}
		values[len(values)-1] += ci.Value
		counterValues[ci.Name] = values
	}
}

// Add a list marker to boxes for elements with ``display: list-item``,
// and yield children to add a the start of the box.

// See http://www.w3.org/TR/CSS21/generate.html#lists
func addBoxMarker(box AllBox, counterValues css.CounterIncrements, getImageFromUri func()) []AllBox {
	style := box.Style()
	imageType, image := style.listStyleImage
	if imageType == "url" {
		// surface may be None here too, in case the image is not available.
		image = getImageFromUri(image)
	}

	if image == nil {
		type_ := style.listStyleType
		if type_ == "none" {
			return nil
		}
		counterValues, has := counterValues["list-item"]
		if !has {
			counterValues = []int{0}
		}
		counterValue = counterValues[len(counterValues)-1]
		markerText = formatListMarker(counterValue, type_)
		markerBox = boxes.TextBox.anonymousFrom(box, markerText)
	} else {
		markerBox = boxes.InlineReplacedBox.anonymousFrom(box, image)
		markerBox.isListMarker = True
	}
	markerBox.elementTag += "::marker"

	position = style.listStylePosition
	if position == "inside" {
		// yield markerBox
		return markerBox
	} else if position == "outside" {
		box.outsideListMarker = markerBox
	}
}
