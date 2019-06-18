package structure

import "github.com/benoitkugler/go-weasyprint/css"

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
	counterValues map[string]int
	counterScopes [](map[string]bool)
}

type Element struct {
	Tag string
}

// Maps values of the ``display`` CSS property to box types.
func makeBox(elementTag TBD, style css.StyleDict, content []AllBox) AllBox {
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
func buildFormattingStructure(elementTree Element, styleFor func(element Element, pseudoType string) *css.StyleDict, getImageFromUri TBD, baseUrl string) AllBox {
	boxList := elementToBox(elementTree, styleFor, getImageFromUri, baseUrl, nil)
	var box AllBox
	if len(boxList) > 0 {
		box = boxList[0]
	} else { //  No root element
		rootStyleFor := func(element Element, pseudoType string) *css.StyleDict {
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
	box.isForRootElement = True
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
func elementToBox(elementTree Element, styleFor func(element Element, pseudoType string) *css.StyleDict,
	getImageFromUri TBD, baseUrl string, state *state) []AllBox {
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
