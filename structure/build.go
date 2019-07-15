package structure

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/utils"
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

type styleForType = func(element html.Node, pseudoType string) css.StyleDict

type stateShared struct {
	quoteDepth    int
	counterValues map[string][]int
	counterScopes [](map[string]bool)
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Maps values of the ``display`` CSS property to box types.
func makeBox(elementTag string, style css.StyleDict, content []AllBox) AllBox {
	switch style.Strings["display"] {
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
func buildFormattingStructure(elementTree html.Node, styleFor styleForType, getImageFromUri gifu, baseUrl string) AllBox {
	boxList := elementToBox(elementTree, styleFor, getImageFromUri, baseUrl, nil)
	var box AllBox
	if len(boxList) > 0 {
		box = boxList[0]
	} else { //  No root element
		rootStyleFor := func(element html.Node, pseudoType string) css.StyleDict {
			style := styleFor(element, pseudoType)
			if !style.IsZero() {
				// TODO: we should check that the element has a parent instead.
				if element.Data == "html" {
					style.Strings["display"] = "block"
				} else {
					style.Strings["display"] = "none"
				}
			}
			return style
		}
		box = elementToBox(elementTree, rootStyleFor, getImageFromUri, baseUrl, nil)[0]
	}
	box.BaseBox().isForRootElement = true
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
func elementToBox(element html.Node, styleFor styleForType,
	getImageFromUri gifu, baseUrl string, state *stateShared) []AllBox {

	if element.Type != html.TextNode && element.Type != html.ElementNode && element.Type != html.DocumentNode {
		// Here we ignore comments and XML processing instructions.
		return nil
	}

	style := styleFor(element, "")

	// TODO: should be the used value. When does the used value for `display`
	// differ from the computer value?
	display := style.Strings["display"]
	if display == "none" {
		return nil
	}

	box := makeBox(element.Data, style, nil)

	if state == nil {
		// use a list to have a shared mutable object
		state = &stateShared{
			// Shared mutable objects:
			quoteDepth:    0,                  // single integer
			counterValues: map[string][]int{}, // name -> stacked/scoped values
			counterScopes: []map[string]bool{ //  element tree depths -> counter names
				map[string]bool{},
			},
		}
	}

	QuoteDepth, counterValues := state.quoteDepth, state.counterValues

	updateCounters(state, &style)

	var children []AllBox
	if display == "list-item" {
		children = append(children, addBoxMarker(box, counterValues, getImageFromUri)...)
	}

	// If this element’s direct children create new scopes, the counter
	// names will be in this new list
	state.counterScopes = append(state.counterScopes, map[string]bool{})

	box.BaseBox().firstLetterStyle = styleFor(element, "first-letter")
	box.BaseBox().firstLineStyle = styleFor(element, "first-line")

	children = append(children, beforeAfterToBox(element, "before", state, styleFor, getImageFromUri)...)
	if element.Type == html.TextNode {
		children = append(children, TextBoxAnonymousFrom(box, element.Data))
	}

	for _, childElement := range utils.NodeChildren(element) {
		children = append(children, elementToBox(childElement, styleFor, getImageFromUri, baseUrl, state)...)
		// html.Node as no notion of tail. Instead, text are converted in text nodes
	}
	children = append(children, beforeAfterToBox(element, "after", state, styleFor, getImageFromUri)...)

	// Scopes created by this element’s children stop here.
	cs := state.counterScopes[len(state.counterScopes)-1]
	state.counterScopes = state.counterScopes[:len(state.counterScopes)-2]
	for name := range cs {
		counterValues[name] = counterValues[name][:len(state.counterScopes)-2]
		if len(counterValues[name]) == 0 {
			delete(counterValues, name)
		}
	}
	box.BaseBox().children = children
	setContentLists(element, box, style, counterValues)

	// Specific handling for the element. (eg. replaced element)
	return HandleElement(element, box, getImageFromUri, baseUrl)

}

var (
	reLineFeeds     = regexp.MustCompile(`\r\n?`)
	reSpaceCollapse = regexp.MustCompile(`[\t ]*\n[\t ]*`)
	reSpace         = regexp.MustCompile(` +`)
)

// First part of "The 'white-space' processing model".
// See http://www.w3.org/TR/CSS21/text.html#white-space-model
// http://dev.w3.org/csswg/css3-text/#white-space-rules
func processWhitespace(_box AllBox, followingCollapsibleSpace bool) bool {
	box, isTextBox := _box.(*TextBox)
	if isTextBox {
		text := box.text
		if text == "" {
			return followingCollapsibleSpace
		}

		// Normalize line feeds
		text = reLineFeeds.ReplaceAllString(text, "\n")

		styleWhiteSpace := box.style.Strings["white_space"]
		newLineCollapse := styleWhiteSpace == "normal" || styleWhiteSpace == "nowrap"
		spaceCollapse := styleWhiteSpace == "normal" || styleWhiteSpace == "nowrap" || styleWhiteSpace == "pre-line"

		if spaceCollapse {
			// \r characters were removed/converted earlier
			text = reSpaceCollapse.ReplaceAllString(text, "\n")
		}

		if newLineCollapse {
			// TODO: this should be language-specific
			// Could also replace with a zero width space character (U+200B),
			// or no character
			// CSS3: http://www.w3.org/TR/css3-text/#line-break-transform
			text = strings.ReplaceAll(text, "\n", " ")
		}
		if spaceCollapse {
			text = strings.ReplaceAll(text, "\t", " ")
			text = reSpace.ReplaceAllString(text, " ")
			previousText := text
			if followingCollapsibleSpace && strings.HasPrefix(text, " ") {
				text = text[1:]
				box.leadingCollapsibleSpace = true
			}
			followingCollapsibleSpace = strings.HasSuffix(previousText, " ")
		} else {
			followingCollapsibleSpace = false
		}
		box.text = text
		return followingCollapsibleSpace
	}
	if _box.IsParentBox() {
		for _, child := range _box.BaseBox().children {
			switch child.(type) {
			case *TextBox, *InlineBox:
				followingCollapsibleSpace = processWhitespace(
					child, followingCollapsibleSpace)
			default:
				processWhitespace(child, false)
				if child.BaseBox().isInNormalFlow() {
					followingCollapsibleSpace = false
				}
			}
		}
	}
	return followingCollapsibleSpace
}

// Remove and add boxes according to the table model.
//
//Take and return a ``Box`` object.
//
//See http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes
func anonymousTableBoxes(box AllBox) AllBox {
	if !box.IsParentBox() {
		return box
	}

	// Do recursion.
	boxChildren := box.BaseBox().children
	children := make([]AllBox, len(boxChildren))
	for index, child := range boxChildren {
		children[index] = anonymousTableBoxes(child)
	}
	return tableBoxesChildren(box, children)
}

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
func inlineInBlock(box AllBox) AllBox {
	if !box.IsParentBox() {
		return box
	}
	baseBox := box.BaseBox()
	boxChildren := baseBox.children

	if len(boxChildren) > 0 && baseBox.leadingCollapsibleSpace == false {
		baseBox.leadingCollapsibleSpace = boxChildren[0].BaseBox().leadingCollapsibleSpace
	}

	var children []AllBox
	trailingCollapsibleSpace := false
	for _, child := range boxChildren {
		// Keep track of removed collapsing spaces for wrap opportunities, and
		// remove empty text boxes.
		// (They may have been emptied by processWhitespace().)

		if trailingCollapsibleSpace {
			child.BaseBox().leadingCollapsibleSpace = true
		}

		textBox, isTextBox := child.(*TextBox)
		if isTextBox && textBox.text == "" {
			trailingCollapsibleSpace = child.BaseBox().leadingCollapsibleSpace
		} else {
			trailingCollapsibleSpace = false
			children = append(children, inlineInBlock(child))
		}
	}
	if baseBox.trailingCollapsibleSpace == false {
		baseBox.trailingCollapsibleSpace = trailingCollapsibleSpace
	}

	if !box.IsBlockContainerBox() {
		baseBox.children = children
		return box
	}

	var newLineChildren, newChildren []AllBox
	for _, childBox := range children {
		if _, isLineBox := childBox.(*LineBox); isLineBox {
			panic("childBox can't be a LineBox")
		}
		if len(newLineChildren) > 0 && childBox.BaseBox().isAbsolutelyPositioned() {
			newLineChildren = append(newLineChildren, childBox)
		} else if childBox.IsInlineLevelBox() || (len(newLineChildren) > 0 && childBox.BaseBox().isFloated()) {
			// Do not append white space at the start of a line :
			// it would be removed during layout.
			childTextBox, isTextBox := childBox.(*TextBox)
			st := childBox.BaseBox().style.Strings["white_space"]

			// Sequence of white-space was collapsed to a single space by processWhitespace().
			if len(newLineChildren) > 0 || !(isTextBox && childTextBox.text == " " && (st == "normal" || st == "nowrap" || st == "pre-line")) {
				newLineChildren = append(newLineChildren, childBox)
			}
		} else {
			if len(newLineChildren) > 0 {
				// Inlines are consecutive no more: add this line box
				// and create a new one.
				lineBox := LineBoxAnonymousFrom(box, newLineChildren)
				anonymous := BlockBoxAnonymousFrom(box, []AllBox{lineBox})
				newChildren = append(newChildren, anonymous)
				newLineChildren = nil
			}
			newChildren = append(newChildren, childBox)
		}
	}
	if len(newLineChildren) > 0 {
		// There were inlines at the end
		lineBox := LineBoxAnonymousFrom(box, newLineChildren)
		if len(newChildren) > 0 {
			anonymous := BlockBoxAnonymousFrom(box, []AllBox{lineBox})
			newChildren = append(newChildren, anonymous)
		} else {
			// Only inline-level children: one line box
			newChildren = append(newChildren, lineBox)
		}
	}

	baseBox.children = newChildren
	return box
}

// Internal implementation of anonymousTableBoxes().
func tableBoxesChildren(box AllBox, children []AllBox) AllBox {
	switch typeBox := box.(type) {
	case *TableColumnBox: // rule 1.1
		// Remove all children.
		children = nil
	case *TableColumnGroupBox: // rule 1.2
		// Remove children other than table-column.
		newChildren := make([]AllBox, 0, len(children))
		for _, child := range children {
			if _, is := child.(*TableColumnBox); is {
				newChildren = append(newChildren, child)
			}
		}
		children = newChildren

		// Rule XXX (not in the spec): column groups have at least
		// one column child.
		if len(children) == 0 {
			for i := 0; i < typeBox.tableFields.span; i++ {
				children = append(children, TypeTableColumnBox.AnonymousFrom(box, nil))
			}
		}

	}
	tf := box.TableFields()
	// rule 1.3
	if tf != nil && tf.tabularContainer && len(children) >= 2 {
		// TODO: Maybe only remove text if internal is also
		//       a proper table descendant of box.
		// This is what the spec says, but maybe not what browsers do:
		// http://lists.w3.org/Archives/Public/www-style/2011Oct/0567

		// Last child
		internal, text := children[len(children)-2], children[len(children)-1]
		internalTableFields := internal.TableFields()
		if internalTableFields != nil && internalTableFields.internalTableOrCaption && isWhitespace(text, nil) {
			children = children[:len(children)-2]
		}
		// First child
		if len(children) >= 2 {
			text, internal = children[0], children[1]
			internalTableFields := internal.TableFields()
			if internalTableFields.internalTableOrCaption && isWhitespace(text, nil) {
				children = children[1:]
			}
		}
		// Children other than first and last that would be removed by
		// rule 1.3 are also removed by rule 1.4 below.
	}

	newChildren, maxIndex := make([]AllBox, 0, len(children)), len(children)-1
	for index, child := range children {
		// Ignore some whitespace: rule 1.4
		if !(index != 0 && index != maxIndex) {
			prevChild, nextChild := children[index-1], children[index+1]
			prevChildTable, nextChildTable := prevChild.TableFields(), nextChild.TableFields()
			if prevChildTable != nil && prevChildTable.internalTableOrCaption && nextChildTable != nil && nextChildTable.internalTableOrCaption && isWhitespace(child, nil) {
				newChildren = append(newChildren, child)
			}
		}
	}
	children = newChildren

	if box.IsTableBox() {
		// Rule 2.1
		children = wrapImproper(box, children, TypeTableRowBox,
			func(child AllBox) bool {
				tf := child.TableFields()
				return tf != nil && tf.properTableChild
			})
	} else if TypeTableRowGroupBox.IsInstance(box) {
		// Rule 2.2
		children = wrapImproper(box, children, TypeTableRowBox, nil)
	}

	if TypeTableRowBox.IsInstance(box) {
		// Rule 2.3
		children = wrapImproper(box, children, TypeTableCellBox, nil)
	} else {
		// Rule 3.1
		children = wrapImproper(box, children, TypeTableCellBox, TypeTableCellBox.IsInstance)
	}
	// Rule 3.2
	if InlineBoxIsInstance(box) {
		children = wrapImproper(box, children, TypeInlineTableBox,
			func(child AllBox) bool {
				tf := child.TableFields()
				return tf == nil || !tf.properTableChild
			})
	} else {
		// parentType = type(box)
		children = wrapImproper(box, children, TypeTableBox,
			func(child AllBox) bool {
				tf := child.TableFields()
				return (tf == nil || !tf.properTableChild ||
					child.IsProperChild(box))
			})
	}
	if box.IsTableBox() {
		return wrapTable(box, children)
	} else {
		box.BaseBox().children = children
		return box
	}
}

// Take a table box and return it in its table wrapper box.
// Also re-order children and assign grid positions to each column and cell.
// Because of colspan/rowspan works, gridY is implicitly the index of a row,
// but gridX is an explicit attribute on cells, columns and column group.
// http://www.w3.org/TR/CSS21/tables.html#model
// http://www.w3.org/TR/CSS21/tables.html#table-layout
//
// wrapTable will panic if box's children are not table boxes
func wrapTable(box AllBox, children []AllBox) AllBox {
	tableFields := box.TableFields()
	if tableFields == nil {
		panic("wrapTable only takes table boxes")
	}

	// Group table children by type
	var columns, rows, allCaptions []AllBox
	byType := func(child AllBox) *[]AllBox {
		switch {
		case TypeTableColumnBox.IsInstance(child), TypeTableColumnGroupBox.IsInstance(child):
			return &columns
		case TypeTableRowBox.IsInstance(child), TypeTableRowGroupBox.IsInstance(child):
			return &rows
		case TableCaptionBoxIsInstance(child):
			return &allCaptions
		default:
			return nil
		}
	}

	for _, child := range children {
		*byType(child) = append(*byType(child), child)
	}
	// Split top and bottom captions
	var captionTop, captionBottom []AllBox
	for _, caption := range allCaptions {
		switch caption.BaseBox().style.Strings["caption_side"] {
		case "top":
			captionTop = append(captionTop, caption)
		case "bottom":
			captionBottom = append(captionBottom, caption)
		}
	}
	// Assign X positions on the grid to column boxes
	columnGroups := wrapImproper(box, columns, TypeTableColumnGroupBox, nil)
	gridX := 0
	for _, _group := range columnGroups {
		group := _group.TableFields()
		group.gridX = gridX
		groupChildren := _group.BaseBox().children
		if len(groupChildren) > 0 {
			for _, column := range groupChildren {
				// There"s no need to take care of group"s span, as "span=x"
				// already generates x TableColumnBox children
				column.TableFields().gridX = gridX
				gridX += 1
			}
			group.span = len(groupChildren)
		} else {
			gridX += group.span
		}
	}
	gridWidth := gridX

	rowGroups := wrapImproper(box, rows, TypeTableRowGroupBox, nil)
	// Extract the optional header and footer groups.
	var (
		bodyRowGroups []AllBox
		header        AllBox
		footer        AllBox
	)
	for _, _group := range rowGroups {
		group := _group.TableFields()
		display := _group.BaseBox().style.Strings["display"]
		if display == "table-header-group" && header == nil {
			group.isHeader = true
			header = _group
		} else if display == "table-footer-group" && footer == nil {
			group.isFooter = true
			footer = _group
		} else {
			bodyRowGroups = append(bodyRowGroups, _group)
		}
	}

	rowGroups = nil
	if header != nil {
		rowGroups = append(rowGroups, header)
	}
	rowGroups = append(rowGroups, bodyRowGroups...)
	if footer != nil {
		rowGroups = append(rowGroups, footer)
	}

	// Assign a (x,y) position in the grid to each cell.
	// rowspan can not extend beyond a row group, so each row group
	// is independent.
	// http://www.w3.org/TR/CSS21/tables.html#table-layout
	// Column 0 is on the left if direction is ltr, right if rtl.
	// This algorithm does not change.
	gridHeight := 0
	for _, group := range rowGroups {
		// Indexes: row number in the group.
		// Values: set of cells already occupied by row-spanning cells.

		groupChildren := group.BaseBox().children
		occupiedCellsByRow := make([]map[int]bool, len(groupChildren))
		for _, row := range groupChildren {
			occupiedCellsInThisRow := occupiedCellsByRow[0]
			occupiedCellsByRow = occupiedCellsByRow[1:]

			// The list is now about rows after this one.
			gridX = 0
			for _, _cell := range row.BaseBox().children {
				cell := _cell.TableFields()
				// Make sure that the first grid cell is free.
				for occupiedCellsInThisRow[gridX] {
					gridX += 1
				}
				cell.gridX = gridX
				newGridX := gridX + cell.colspan
				// http://www.w3.org/TR/html401/struct/tables.html#adef-rowspan
				if cell.rowspan != 1 {
					maxRowspan := len(occupiedCellsByRow) + 1
					var spannedRows []map[int]bool
					if cell.rowspan == 0 {
						// All rows until the end of the group
						spannedRows = occupiedCellsByRow
						cell.rowspan = maxRowspan
					} else {
						cell.rowspan = min(cell.rowspan, maxRowspan)
						spannedRows = occupiedCellsByRow[:cell.rowspan-1]
					}
					for _, occupiedCells := range spannedRows {
						for i := gridX; i < newGridX; i++ {
							occupiedCells[i] = true
						}
					}
				}
				gridX = newGridX
				gridWidth = max(gridWidth, gridX)
			}
			gridHeight += len(groupChildren)
		}
	}
	table := CopyWithChildren(box, rowGroups, true, true)
	tf, tableBox := table.TableFields(), table.BaseBox()
	tf.columnGroups = columnGroups
	if tableBox.style.Strings["border_collapse"] == "collapse" {
		tf.collapsedBorderGrid = collapseTableBorders(table, gridWidth, gridHeight)
	}
	var wrapperTypeAF func(AllBox, []AllBox) AllBox
	if TypeInlineTableBox.IsInstance(box) {
		wrapperTypeAF = InlineBlockBoxAnonymousFrom
	} else {
		wrapperTypeAF = BlockBoxAnonymousFrom
	}
	wrapper := wrapperTypeAF(box, append(append(captionTop, table), captionBottom...))
	wrapperBox := wrapper.BaseBox()
	wrapperBox.style = wrapperBox.style.Copy()
	wrapperBox.isTableWrapper = true
	if !tableBox.style.Anonymous {
		// Non-inherited properties of the table element apply to one
		// of the wrapper and the table. The other get the initial value.
		// TODO: put this in a method of the table object
		tableStyleItems := tableBox.style.Items()
		for name := range css.TableWrapperBoxProperties {
			tableStyleItems[name].SetOn(name, &wrapperBox.style)
			css.InitialValuesItems[name].SetOn(name, &tableBox.style)
		}
	}
	// else: non-inherited properties already have their initial values

	return wrapper
}

var (
	styleScores = map[string]int{}
	styleMap    = map[string]string{
		"inset":  "ridge",
		"outset": "groove",
	}

	Transparent = css.Color{}
)

func init() {
	styles := []string{"hidden", "double", "solid", "dashed", "dotted", "ridge",
		"outset", "groove", "inset", "none"}
	N := len(styles) - 1
	for i, v := range styles {
		styleScores[v] = N - i
	}
}

type Score [3]int

func (s Score) lower(other Score) bool {
	return s[0] < other[0] || (s[0] == other[0] && (s[1] < other[1] || (s[1] == other[1] && s[2] < other[2])))
}

type border struct {
	score Score
	style string
	width int
	color css.Color
}

// Resolve border conflicts for a table in the collapsing border model.
//     Take a :class:`TableBox`; set appropriate border widths on the table,
//     column group, column, row group, row, && cell boxes; && return
//     a data structure for the resolved collapsed border grid.
//
func collapseTableBorders(table AllBox, gridWidth, gridHeight int) ([][]border, [][]border) {
	if gridWidth == 0 || gridHeight == 0 {
		// Don’t bother with empty tables
		return nil, nil
	}

	transparent := Transparent
	weakNullBorder := border{score: Score{0, 0, styleScores["none"]}, style: "none", width: 0, color: transparent}

	verticalBorders, horizontalBorders := make([][]border, gridHeight), make([][]border, gridHeight+1)
	for y := 0; y < gridHeight+1; y++ {
		l1, l2 := make([]border, gridWidth+1), make([]border, gridWidth)
		for x := 0; x < gridWidth; x++ {
			l1[x] = weakNullBorder
			l2[x] = weakNullBorder
		}
		l1[gridWidth] = weakNullBorder
		if y < gridHeight {
			verticalBorders[y] = l1
		}
		horizontalBorders[y] = l2
	}

	// verticalBorders = [[weakNullBorder for x in range(gridWidth + 1)]
	//                     for y in range(gridHeight)]
	// horizontalBorders = [[weakNullBorder for x in range(gridWidth)]
	//                       for y in range(gridHeight + 1)]

	setOneBorder := func(borderGrid [][]border, boxStyle css.StyleDict, side css.Side, gridX, gridY int) {
		style := boxStyle.Strings[fmt.Sprintf("border_%s_style", side)]
		width := boxStyle.Values[fmt.Sprintf("border_%s_width", side)]
		color := boxStyle.GetColor(fmt.Sprintf("border_%s_color", side))

		// http://www.w3.org/TR/CSS21/tables.html#border-conflict-resolution
		score := Score{0, width, styleScores[style]}
		if style == hidden {
			score[0] = 1
		}

		_style, in := styleMap[style]
		if in {
			style = _style
		}

		previousScore := borderGrid[gridY][gridX].score
		// Strict < so that the earlier call wins in case of a tie.
		if previousScore.lower(score) {
			borderGrid[gridY][gridX] = border{score: score, style: style, width: width, color: color}
		}
	}

	setBorders := func(box AllBox, x, y, w, h int) {
		style = box.BaseBox().style
		for yy := y; yy < y+h; y++ {
			setOneBorder(verticalBorders, style, css.Left, x, yy)
			setOneBorder(verticalBorders, style, css.Right, x+w, yy)
		}
		for xx := x; xx < x+w; xx++ {
			setOneBorder(horizontalBorders, style, css.Top, xx, y)
			setOneBorder(horizontalBorders, style, css.Bottom, xx, y+h)
		}
	}

	// The order is important here:
	// "A style set on a cell wins over one on a row, which wins over a
	//  row group, column, column group and, lastly, table"
	// See http://www.w3.org/TR/CSS21/tables.html#border-conflict-resolution
	strongNullBorder := border{score: Score{1, 0, styleScores["hidden"]}, style: "hidden", width: 0, color: transparent}

	gridY := 0
	for _, rowGroup := range table.BaseBox().children {
		for _, row := range rowGroup.BaseBox().children {
			for _, _cell := range row.BaseBox().children {
				cell := _cell.TableFields()
				// No border inside of a cell with rowspan || colspan
				for xx := cell.gridX + 1; xx < cell.gridX+cell.colspan; xx++ {
					for yy := gridY; yy < gridY+cell.rowspan; yy++ {
						verticalBorders[yy][xx] = strongNullBorder
					}
				}
				for xx := cell.gridX; xx < cell.gridX+cell.colspan; xx++ {
					for yy := gridY + 1; yy < gridY+cell.rowspan; yy++ {
						horizontalBorders[yy][xx] = strongNullBorder
					}
				}
				// The cell’s own borders
				setBorders(_cell, cell.gridX, gridY, cell.colspan, cell.rowspan)
			}
			gridY += 1
		}
	}

	gridY = 0
	for _, rowGroup := range table.BaseBox().children {
		for _, row := range rowGroup.BaseBox().children {
			setBorders(row, 0, gridY, gridWidth, 1)
			gridY += 1
		}
	}

	gridY = 0
	for _, rowGroup := range table.BaseBox().children {
		rowspan := len(rowGroup.BaseBox().children)
		setBorders(rowGroup, 0, gridY, gridWidth, rowspan)
		gridY += rowspan
	}

	for _, columnGroup := range table.TableFields().columnGroups {
		for _, column := range columnGroup.BaseBox().children {
			setBorders(column, column.TableFields().gridX, 0, 1, gridHeight)
		}
	}

	for _, columnGroup := range table.TableFields().columnGroups {
		tf := columnGroup.TableFields()
		setBorders(columnGroup, tf.gridX, 0, tf.span, gridHeight)
	}

	setBorders(table, 0, 0, gridWidth, gridHeight)

	// Now that all conflicts are resolved, set transparent borders of
	// the correct widths on each box. The actual border grid will be
	// painted separately.
	setTransparentBorder := func(box AllBox, side css.Side, twiceWidth int) {
		st := box.BaseBox().style
		st.Strings[fmt.Sprintf("border_%s_style", side)] = "solid"
		st.Values[fmt.Sprintf("border_%s_width", side)] = css.IntToValue(twiceWidth / 2)
		st.Colors[fmt.Sprintf("border_%s_color", side)] = transparent
	}

	removeBorders := func(box AllBox) {
		setTransparentBorder(box, css.Top, 0)
		setTransparentBorder(box, css.Right, 0)
		setTransparentBorder(box, css.Bottom, 0)
		setTransparentBorder(box, css.Left, 0)
	}

	maxVerticalWidth := func(x, y, h int) int {
		var max int
		for _, gridRow := range verticalBorders[y : y+h] {
			width := gridRow[x].width
			if width > max {
				max = width
			}
		}
		return max
	}

	maxHorizontalWidth := func(x, y, w int) int {
		var max int
		for _, _s := range horizontalBorders[y][x : x+w] {
			width := _s.width
			if width > max {
				max = width
			}
		}
		return max
	}

	gridY = 0
	for _, rowGroup := range table.BaseBox().children {
		removeBorders(rowGroup)
		for _, row := range rowGroup.BaseBox().children {
			removeBorders(row)
			for _, _cell := range row.BaseBox().children {
				cell = _cell.TableFields()
				setTransparentBorder(cell, css.Top, maxHorizontalWidth(cell.gridX, gridY, cell.colspan))
				setTransparentBorder(cell, css.Bottom, maxHorizontalWidth(cell.gridX, gridY+cell.rowspan, cell.colspan))
				setTransparentBorder(cell, css.Left, maxVerticalWidth(cell.gridX, gridY, cell.rowspan))
				setTransparentBorder(cell, css.Right, maxVerticalWidth(cell.gridX+cell.colspan, gridY, cell.rowspan))
			}
			gridY += 1
		}
	}

	for _, columnGroup := range table.TableFields().columnGroups {
		removeBorders(columnGroup)
		for _, column := range columnGroup.BaseBox().children {
			removeBorders(column)
		}
	}

	setTransparentBorder(table, css.Top, maxHorizontalWidth(0, 0, gridWidth))
	setTransparentBorder(table, css.Bottom, maxHorizontalWidth(0, gridHeight, gridWidth))
	// "UAs must compute an initial left && right border width for the table
	// by examining the first && last cells in the first row of the table."
	// http://www.w3.org/TR/CSS21/tables.html#collapsing-borders
	// ... so h=1, not gridHeight :
	setTransparentBorder(table, css.Left, maxVerticalWidth(0, 0, 1))
	setTransparentBorder(table, css.Right, maxVerticalWidth(gridWidth, 0, 1))

	return verticalBorders, horizontalBorders
}

//   Wrap consecutive children that do not pass ``test`` in a box of type
// ``wrapperType``.
// ``test`` defaults to children being of the same type as ``wrapperType``.
func wrapImproper(box AllBox, children []AllBox, boxType BoxType, test func(AllBox) bool) []AllBox {
	var out, improper []AllBox
	if test == nil {
		test = boxType.IsInstance
	}
	for _, child := range children {
		if test(child) {
			if len(improper) > 0 {
				wrapper := boxType.AnonymousFrom(box, nil)
				// Apply the rules again on the new wrapper
				out = append(out, tableBoxesChildren(wrapper, improper))
				improper = nil
			}
			out = append(out, child)
		} else {
			// Whitespace either fail the test or were removed earlier,
			// so there is no need to take special care with the definition
			// of "consecutive".
			improper = append(improper, child)
		}
	}
	if len(improper) > 0 {
		wrapper := boxType.AnonymousFrom(box, nil)
		// Apply the rules again on the new wrapper
		out = append(out, tableBoxesChildren(wrapper, improper))
	}
	return out
}

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
func blockInInline(box AllBox) AllBox {
	if !box.IsParentBox() {
        return box
    }

    var newChildren []AllBox
    changed := false

    for _, child := range box.BaseBox().children {
		var newChild AllBox
        if LineBoxIsInstance(child) {
            if len(box.children) != 1 {
				log.Fatalf("Line boxes should have no siblings at this stage, got %r.", box.children)
			}
                
            var (
				stack []int
				newLine []AllBox
			)
            for {
                newLine, block, stack = innerBlockInInline(child, stack)
                if block == nil {
                    break
				} 
				anon := BlockBoxAnonymousFrom(box, []AllBox{newLine})
                newChildren = append(newChildren, anon)
                newChildren = append(newChildren, blockInInline(block))
                // Loop with the same child && the new stack.
			} 
			
			if len(newChildren) > 0 {
                // Some children were already added, this became a block
                // context.
                newChild = BlockBoxAnonymousFrom(box, []AllBox{newLine})
            } else {
                // Keep the single line box as-is, without anonymous blocks.
                newChild = newLine
            }
        } else {
            // Not in an inline formatting context.
            newChild = blockInInline(child)
        }
    
        if newChild != child {
            changed = true
		} 
		newChildren = append(newChildren, newChild)
		}
    if changed {
        box.BaseBox().children = newChildren
	} 
	return box
	}
	


// Find a block-level box in an inline formatting context.
//     If one is found, return ``(newBox, blockLevelBox, resumeAt)``.
//     ``newBox`` contains all of ``box`` content before the block-level box.
//     ``resumeAt`` can be passed as ``skipStack`` in a new call to
//     this function to resume the search just after the block-level box.
//     If no block-level box is found after the position marked by
//     ``skipStack``, return ``(newBox, None, None)``
//     
func innerBlockInInline(box AllBox, skipStack []int) {
    var newChildren []AllBox
    var blockLevelBox AllBox
    resumeAt = None
    changed := false

	isStart := skipStack == nil 
	var skip int 
    if isStart {
        skip = 0
    } else {
		skip = skipStack[0]
		skipStack = skipStack[1:]
    }

    for index, child := range box.BaseBox().children[skip:] {
        if isinstance(child, boxes.BlockLevelBox) && \
                child.isInNormalFlow() {
                }
            assert skipStack is None  // Should not skip here
            blockLevelBox = child
            index += 1  // Resume *after* the block
        else {
            if isinstance(child, boxes.InlineBox) {
                recursion = innerBlockInInline(child, skipStack)
                skipStack = None
                newChild, blockLevelBox, resumeAt = recursion
            } else {
                assert skipStack is None  // Should not skip here
                newChild = blockInInline(child)
                // blockLevelBox is still None.
            } if newChild is not child {
                changed = true
            } newChildren.append(newChild)
        } if blockLevelBox is not None {
            resumeAt = (index, resumeAt)
            box = box.copyWithChildren(
                newChildren, isStart=isStart, isEnd=false)
            break
        }
    } else {
        if changed || skip {
            box = box.copyWithChildren(
                newChildren, isStart=isStart, isEnd=true)
        }
    }

    return box, blockLevelBox, resumeAt
}

//  Set a ``viewportOverflow`` attribute on the box for the root element.
//
//    Like backgrounds, ``overflow`` on the root element must be propagated
//    to the viewport.
//
//    See http://www.w3.org/TR/CSS21/visufx.html#overflow
func setViewportOverflow(rootBox AllBox) AllBox {}

// Handle the ``counter-*`` properties.
func updateCounters(state *stateShared, style *css.StyleDict) {
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
func addBoxMarker(box AllBox, counterValues map[string][]int, getImageFromUri gifu) []AllBox {
	style := box.BaseBox().style
	image := style.ListStyleImage.Image
	if style.ListStyleImage.Type == "url" {
		// surface may be None here too, in case the image is not available.
		image = getImageFromUri(image.Url(), "")
	}
	var markerBox AllBox
	if image == nil {
		type_ := style.Strings["list_style_type"]
		if type_ == "none" {
			return nil
		}
		counterValues, has := counterValues["list-item"]
		if !has {
			counterValues = []int{0}
		}
		counterValue := counterValues[len(counterValues)-1]
		// TODO: rtl numbered list has the dot on the left
		markerText := formatListMarker(counterValue, type_)
		markerBox = TextBoxAnonymousFrom(box, markerText)
	} else {
		markerBox = InlineReplacedBoxAnonymousFrom(box, image)
		markerBox.BaseBox().isListMarker = true
	}
	markerBox.BaseBox().elementTag += "::marker"

	switch style.Strings["list_style_position"] {
	case "inside":
		return markerBox
	case "outside":
		box.outsideListMarker = markerBox
	}
	return nil
}

var reHasNonWhitespace = regexp.MustCompile("\\S")

var hasNonWhitespaceDefault = func(text string) bool {
	return reHasNonWhitespace.MatchString(text)
}

// Return true if ``box`` is a TextBox with only whitespace.
func isWhitespace(box AllBox, hasNonWhitespace func(string) bool) bool {
	if hasNonWhitespace == nil {
		hasNonWhitespace = hasNonWhitespaceDefault
	}
	textBox, is := box.(*TextBox)
	return is && !hasNonWhitespace(textBox.text)
}
