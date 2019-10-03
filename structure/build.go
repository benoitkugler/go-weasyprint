package structure

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/benoitkugler/go-weasyprint/structure/counters"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
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

var (
	textContentExtractors = map[string]func(Box) string{
		"text":         boxText,
		"before":       boxTextBefore,
		"after":        boxTextAfter,
		"first-letter": boxTextFirstLetter,
	}
)

type stateShared struct {
	quoteDepth    []int
	counterValues tree.CounterValues
	counterScopes [](map[string]bool)
}

// Maps values of the ``display`` CSS property to box types.
func makeBox(elementTag string, style pr.Properties, content []Box) Box {
	switch style.GetDisplay() {
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
func BuildFormattingStructure(elementTree *utils.HTMLNode, styleFor tree.StyleFor, getImageFromUri gifu, baseUrl string, targetCollector *tree.TargetCollector) Box {
	boxList := elementToBox(elementTree, styleFor, getImageFromUri, baseUrl, nil)
	var box Box
	if len(boxList) > 0 {
		box = boxList[0]
	} else { //  No root element
		rootStyleFor := func(element *utils.HTMLNode, pseudoType string) pr.Properties {
			style := styleFor.Get(element, pseudoType)
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
	box.Box().isForRootElement = true
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
func elementToBox(element *utils.HTMLNode, styleFor tree.StyleFor,
	getImageFromUri gifu, baseUrl string, targetCollector *tree.TargetCollector, state *stateShared) []Box {

	if element.Type != html.TextNode && element.Type != html.ElementNode && element.Type != html.DocumentNode {
		// Here we ignore comments and XML processing instructions.
		return nil
	}

	style := styleFor.Get(element, "")

	// TODO: should be the used value. When does the used value for `display`
	// differ from the computer value?
	display := style.GetDisplay()
	if display == "none" {
		return nil
	}

	box := makeBox(element.Data, style, nil)

	if state == nil {
		// use a list to have a shared mutable object
		state = &stateShared{
			// Shared mutable objects:
			quoteDepth:    []int{0},             // single integer
			counterValues: tree.CounterValues{}, // name -> stacked/scoped values
			counterScopes: []map[string]bool{ //  element tree depths -> counter names
				map[string]bool{},
			},
		}
	}

	counterValues := state.counterValues

	updateCounters(state, style)
	// If this element’s direct children create new scopes, the counter
	// names will be in this new list
	state.counterScopes = append(state.counterScopes, map[string]bool{})

	box.Box().firstLetterStyle = styleFor.Get(element, "first-letter")
	box.Box().firstLineStyle = styleFor.Get(element, "first-line")

	var children, markerBoxes []Box
	if display == "list-item" {
		mb := markerToBox(element, state, style, styleFor, getImageFromUri, targetCollector)
		if mb != nil {
			markerBoxes = []Box{mb}
		}
		children = append(children, markerBoxes...)
	}

	children = append(children, beforeAfterToBox(element, "before", state, styleFor, getImageFromUri, targetCollector)...)
	if element.Type == html.TextNode {
		children = append(children, TextBoxAnonymousFrom(box, element.Data))
	}

	for _, childElement := range utils.NodeChildren(*element) {
		children = append(children, elementToBox(childElement, styleFor, getImageFromUri, baseUrl, state)...)
		// utils.HTMLNode as no notion of tail. Instead, text are converted in text nodes
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
	box.Box().children = children
	setContentLists(element, box, style, counterValues)

	// Specific handling for the element. (eg. replaced element)
	return HandleElement(element, box, getImageFromUri, baseUrl)

}

// Yield the box for ::before or ::after pseudo-element.
func beforeAfterToBox(element *utils.HTMLNode, pseudoType string, state *stateShared, styleFor tree.StyleFor,
	getImageFromUri gifu, targetCollector *tree.TargetCollector) []Box {

	style := styleFor.Get(element, pseudoType)
	if pseudoType != "" && style == nil {
		// Pseudo-elements with no style at all do not get a style dict.
		// Their initial content property computes to "none".
		return nil
	}

	// TODO: should be the computed value. When does the used value for
	// `display` differ from the computer value? It's at least wrong for
	// `content` where "normal" computes as "inhibit" for pseudo elements.
	display := style.GetDisplay()
	content := style.GetContent()
	if display == "none" || content.String == "none" || content.String == "normal" || content.String == "inhibit" {
		return nil
	}

	box := makeBox(fmt.Sprintf("%s::%s", element.Data, pseudoType), style, nil)

	updateCounters(state, style)

	var children []Box
	if display == "list-item" {
		mb := markerToBox(element, state, style, styleFor, getImageFromUri, targetCollector)
		if mb != nil {
			children = append(children, mb)
		}
	}
	children = append(children, contentToBoxes(
		style, box, state.quoteDepth, state.counterValues, getImageFromUri, targetCollector, nil, nil)...)

	box.Box().children = children
	return []Box{box}
}

// Takes the value of a ``content`` property and yield boxes.
func contentToBoxes(style pr.Properties, parentBox Box, quoteDepth []int, counterValues tree.CounterValues,
	getImageFromUri gifu, targetCollector *tree.TargetCollector, context, page interface{}) []Box {

	// Closure to parse the ``parentBoxes`` children all again.
	parseAgain := func(mixinPagebasedCounters tree.CounterValues) {
		// Neither alters the mixed-in nor the cached counter values, no
		// need to deepcopy here
		localCounters := tree.CounterValues{}
		for k, v := range mixinPagebasedCounters {
			localCounters[k] = v
		}
		for k, v := range parentBox.Box().cachedCounterValues {
			localCounters[k] = v
		}

		var localChildren []Box
		localChildren = append(localChildren, contentToBoxes(
			style, parentBox, origQuoteDepth, localCounters,
			getImageFromUri, targetCollector, nil, nil)...)

		// TODO: do we need to add markers here?
		// TODO: redo the formatting structure of the parent instead of hacking
		// the already formatted structure. Find why inlineInBlocks has
		// sometimes already been called, && sometimes not.
		parentChildren := parentBox.Box().children
		if len(parentChildren) == 1 && TypeLineBox.IsInstance(parentChildren[0]) {
			parentChildren[0].Box().children = localChildren
		} else {
			parentBox.Box().children = localChildren
		}
	}

	if style.GetContent().String == "inhibit" {
		return nil
	}

	origQuoteDepth := make([]int, len(quoteDepth))
	for i, v := range quoteDepth {
		origQuoteDepth[i] = v
	}
	cssToken = "content"
	boxList := computeContentList(
		style.GetContent(), parentBox, counterValues, cssToken, parseAgain,
		targetCollector, getImageFromUri, quoteDepth, style.GetQuotes(),
		context, page, nil)
	return boxList
}

var (
	reLineFeeds     = regexp.MustCompile(`\r\n?`)
	reSpaceCollapse = regexp.MustCompile(`[\t ]*\n[\t ]*`)
	reSpace         = regexp.MustCompile(` +`)
)

// First part of "The 'white-space' processing model".
// See http://www.w3.org/TR/CSS21/text.html#white-space-model
// http://dev.w3.org/csswg/css3-text/#white-space-rules
func processWhitespace(_box Box, followingCollapsibleSpace bool) bool {
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
		for _, child := range _box.Box().children {
			switch child.(type) {
			case *TextBox, *InlineBox:
				followingCollapsibleSpace = processWhitespace(
					child, followingCollapsibleSpace)
			default:
				processWhitespace(child, false)
				if child.Box().isInNormalFlow() {
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
func anonymousTableBoxes(box Box) Box {
	if !box.IsParentBox() {
		return box
	}

	// Do recursion.
	boxChildren := box.Box().children
	children := make([]Box, len(boxChildren))
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
func inlineInBlock(box Box) Box {
	if !box.IsParentBox() {
		return box
	}
	baseBox := box.Box()
	boxChildren := baseBox.children

	if len(boxChildren) > 0 && baseBox.leadingCollapsibleSpace == false {
		baseBox.leadingCollapsibleSpace = boxChildren[0].Box().leadingCollapsibleSpace
	}

	var children []Box
	trailingCollapsibleSpace := false
	for _, child := range boxChildren {
		// Keep track of removed collapsing spaces for wrap opportunities, and
		// remove empty text boxes.
		// (They may have been emptied by processWhitespace().)

		if trailingCollapsibleSpace {
			child.Box().leadingCollapsibleSpace = true
		}

		textBox, isTextBox := child.(*TextBox)
		if isTextBox && textBox.text == "" {
			trailingCollapsibleSpace = child.Box().leadingCollapsibleSpace
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

	var newLineChildren, newChildren []Box
	for _, childBox := range children {
		if _, isLineBox := childBox.(*LineBox); isLineBox {
			panic("childBox can't be a LineBox")
		}
		if len(newLineChildren) > 0 && childBox.Box().isAbsolutelyPositioned() {
			newLineChildren = append(newLineChildren, childBox)
		} else if childBox.IsInlineLevelBox() || (len(newLineChildren) > 0 && childBox.Box().isFloated()) {
			// Do not append white space at the start of a line :
			// it would be removed during layout.
			childTextBox, isTextBox := childBox.(*TextBox)
			st := childBox.Box().style.Strings["white_space"]

			// Sequence of white-space was collapsed to a single space by processWhitespace().
			if len(newLineChildren) > 0 || !(isTextBox && childTextBox.text == " " && (st == "normal" || st == "nowrap" || st == "pre-line")) {
				newLineChildren = append(newLineChildren, childBox)
			}
		} else {
			if len(newLineChildren) > 0 {
				// Inlines are consecutive no more: add this line box
				// and create a new one.
				lineBox := LineBoxAnonymousFrom(box, newLineChildren)
				anonymous := BlockBoxAnonymousFrom(box, []Box{lineBox})
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
			anonymous := BlockBoxAnonymousFrom(box, []Box{lineBox})
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
func tableBoxesChildren(box Box, children []Box) Box {
	switch typeBox := box.(type) {
	case *TableColumnBox: // rule 1.1
		// Remove all children.
		children = nil
	case *TableColumnGroupBox: // rule 1.2
		// Remove children other than table-column.
		newChildren := make([]Box, 0, len(children))
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

	newChildren, maxIndex := make([]Box, 0, len(children)), len(children)-1
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
			func(child Box) bool {
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
			func(child Box) bool {
				tf := child.TableFields()
				return tf == nil || !tf.properTableChild
			})
	} else {
		// parentType = type(box)
		children = wrapImproper(box, children, TypeTableBox,
			func(child Box) bool {
				tf := child.TableFields()
				return (tf == nil || !tf.properTableChild ||
					child.IsProperChild(box))
			})
	}
	if box.IsTableBox() {
		return wrapTable(box, children)
	} else {
		box.Box().children = children
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
func wrapTable(box Box, children []Box) Box {
	tableFields := box.TableFields()
	if tableFields == nil {
		panic("wrapTable only takes table boxes")
	}

	// Group table children by type
	var columns, rows, allCaptions []Box
	byType := func(child Box) *[]Box {
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
	var captionTop, captionBottom []Box
	for _, caption := range allCaptions {
		switch caption.Box().style.Strings["caption_side"] {
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
		groupChildren := _group.Box().children
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
		bodyRowGroups []Box
		header        Box
		footer        Box
	)
	for _, _group := range rowGroups {
		group := _group.TableFields()
		display := _group.Box().style.Strings["display"]
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

		groupChildren := group.Box().children
		occupiedCellsByRow := make([]map[int]bool, len(groupChildren))
		for _, row := range groupChildren {
			occupiedCellsInThisRow := occupiedCellsByRow[0]
			occupiedCellsByRow = occupiedCellsByRow[1:]

			// The list is now about rows after this one.
			gridX = 0
			for _, _cell := range row.Box().children {
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
	tf, tableBox := table.TableFields(), table.Box()
	tf.columnGroups = columnGroups
	if tableBox.style.Strings["border_collapse"] == "collapse" {
		tf.collapsedBorderGrid = collapseTableBorders(table, gridWidth, gridHeight)
	}
	var wrapperTypeAF func(Box, []Box) Box
	if TypeInlineTableBox.IsInstance(box) {
		wrapperTypeAF = InlineBlockBoxAnonymousFrom
	} else {
		wrapperTypeAF = BlockBoxAnonymousFrom
	}
	wrapper := wrapperTypeAF(box, append(append(captionTop, table), captionBottom...))
	wrapperBox := wrapper.Box()
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
func collapseTableBorders(table Box, gridWidth, gridHeight int) BorderGrids {
	if gridWidth == 0 || gridHeight == 0 {
		// Don’t bother with empty tables
		return BorderGrids{}
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

	setOneBorder := func(borderGrid [][]border, boxStyle pr.Properties, side css.Side, gridX, gridY int) {
		style := boxStyle.Strings[fmt.Sprintf("border_%s_style", side)]
		width := boxStyle.Values[fmt.Sprintf("border_%s_width", side)]
		color := boxStyle.GetColor(fmt.Sprintf("border_%s_color", side))

		// http://www.w3.org/TR/CSS21/tables.html#border-conflict-resolution
		score := Score{0, width.Value, styleScores[style]}
		if style == "hidden" {
			score[0] = 1
		}

		_style, in := styleMap[style]
		if in {
			style = _style
		}

		previousScore := borderGrid[gridY][gridX].score
		// Strict < so that the earlier call wins in case of a tie.
		if previousScore.lower(score) {
			borderGrid[gridY][gridX] = border{score: score, style: style, width: width.Value, color: color}
		}
	}

	setBorders := func(box Box, x, y, w, h int) {
		style := box.Box().style
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
	for _, rowGroup := range table.Box().children {
		for _, row := range rowGroup.Box().children {
			for _, _cell := range row.Box().children {
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
	for _, rowGroup := range table.Box().children {
		for _, row := range rowGroup.Box().children {
			setBorders(row, 0, gridY, gridWidth, 1)
			gridY += 1
		}
	}

	gridY = 0
	for _, rowGroup := range table.Box().children {
		rowspan := len(rowGroup.Box().children)
		setBorders(rowGroup, 0, gridY, gridWidth, rowspan)
		gridY += rowspan
	}

	for _, columnGroup := range table.TableFields().columnGroups {
		for _, column := range columnGroup.Box().children {
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
	setTransparentBorder := func(box Box, side css.Side, twiceWidth int) {
		st := box.Box().style
		st.Strings[fmt.Sprintf("border_%s_style", side)] = "solid"
		st.Values[fmt.Sprintf("border_%s_width", side)] = css.IntToValue(twiceWidth / 2)
		st.Colors[fmt.Sprintf("border_%s_color", side)] = transparent
	}

	removeBorders := func(box Box) {
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
	for _, rowGroup := range table.Box().children {
		removeBorders(rowGroup)
		for _, row := range rowGroup.Box().children {
			removeBorders(row)
			for _, _cell := range row.Box().children {
				cell := _cell.TableFields()
				setTransparentBorder(_cell, css.Top, maxHorizontalWidth(cell.gridX, gridY, cell.colspan))
				setTransparentBorder(_cell, css.Bottom, maxHorizontalWidth(cell.gridX, gridY+cell.rowspan, cell.colspan))
				setTransparentBorder(_cell, css.Left, maxVerticalWidth(cell.gridX, gridY, cell.rowspan))
				setTransparentBorder(_cell, css.Right, maxVerticalWidth(cell.gridX+cell.colspan, gridY, cell.rowspan))
			}
			gridY += 1
		}
	}

	for _, columnGroup := range table.TableFields().columnGroups {
		removeBorders(columnGroup)
		for _, column := range columnGroup.Box().children {
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

	return BorderGrids{Vertical: verticalBorders, Horizontal: horizontalBorders}
}

//   Wrap consecutive children that do not pass ``test`` in a box of type
// ``wrapperType``.
// ``test`` defaults to children being of the same type as ``wrapperType``.
func wrapImproper(box Box, children []Box, boxType BoxType, test func(Box) bool) []Box {
	var out, improper []Box
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
func blockInInline(box Box) Box {
	if !box.IsParentBox() {
		return box
	}

	var newChildren []Box
	changed := false

	for _, child := range box.Box().children {
		var newChild Box
		if LineBoxIsInstance(child) {
			if len(box.Box().children) != 1 {
				log.Fatalf("Line boxes should have no siblings at this stage, got %r.", box.Box().children)
			}

			var (
				stack          *SkipStack
				newLine, block Box
			)
			for {
				newLine, block, stack = innerBlockInInline(child, stack)
				if block == nil {
					break
				}
				anon := BlockBoxAnonymousFrom(box, []Box{newLine})
				newChildren = append(newChildren, anon)
				newChildren = append(newChildren, blockInInline(block))
				// Loop with the same child && the new stack.
			}

			if len(newChildren) > 0 {
				// Some children were already added, this became a block
				// context.
				newChild = BlockBoxAnonymousFrom(box, []Box{newLine})
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
		box.Box().children = newChildren
	}
	return box
}

type SkipStack struct {
	skip  int
	stack *SkipStack
}

// Find a block-level box in an inline formatting context.
//     If one is found, return ``(newBox, blockLevelBox, resumeAt)``.
//     ``newBox`` contains all of ``box`` content before the block-level box.
//     ``resumeAt`` can be passed as ``skipStack`` in a new call to
//     this function to resume the search just after the block-level box.
//     If no block-level box is found after the position marked by
//     ``skipStack``, return ``(newBox, None, None)``
//
func innerBlockInInline(box Box, skipStack *SkipStack) (Box, Box, *SkipStack) {
	var newChildren []Box
	var blockLevelBox Box
	var resumeAt *SkipStack
	changed := false

	isStart := skipStack == nil
	var skip int
	if isStart {
		skip = 0
	} else {
		skip = skipStack.skip
		skipStack = skipStack.stack
	}

	hasBroken := false
	for index, child := range box.Box().children[skip:] {
		if child.IsBlockLevelBox() && child.Box().isInNormalFlow() {
			if skipStack != nil {
				log.Fatal("Should not skip here")
			}
			blockLevelBox = child
			index += 1 // Resume *after* the block
		} else {
			var newChild Box
			if InlineBoxIsInstance(child) {
				newChild, blockLevelBox, resumeAt = innerBlockInInline(child, skipStack)
				skipStack = nil
			} else {
				if skipStack != nil {
					log.Fatal("Should not skip here")
				}
				newChild = blockInInline(child)
				// blockLevelBox is still None.
			}

			if newChild != child {
				changed = true
			}
			newChildren = append(newChildren, newChild)
		}

		if blockLevelBox != nil {
			resumeAt = &SkipStack{skip: index, stack: resumeAt}
			box = CopyWithChildren(box, newChildren, isStart, false)
			hasBroken = true
			break
		}
	}
	if !hasBroken {
		if changed || skip > 0 {
			box = CopyWithChildren(box, newChildren, isStart, true)
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
func setViewportOverflow(rootBox Box) Box {
	chosenBox := rootBox
	if strings.ToLower(rootBox.Box().elementTag) == "html" &&
		rootBox.Box().style.Strings["overflow"] == "visible" {

		for _, child := range rootBox.Box().children {
			if strings.ToLower(child.Box().elementTag) == "body" {
				chosenBox = child
				break
			}
		}
	}
	rootBox.Box().viewportOverflow = chosenBox.Box().style.Strings["overflow"]
	chosenBox.Box().style.Strings["overflow"] = "visible"
	return rootBox
}

// Compute and return the boxes corresponding to the ``content_list``.
//
// ``parse_again`` is called to compute the ``content_list`` again when
// ``target_collector.lookup_target()`` detected a pending target.
//
// ``build_formatting_structure`` calls
// ``target_collector.check_pending_targets()`` after the first pass to do
// required reparsing.
func computeContentList(contentList pr.SContent, parentBox Box, counterValues tree.CounterValues,
	cssToken string, parseAgain func(tree.CounterValues), targetCollector *tree.TargetCollector,
	getImageFromUri gifu, quoteDepth []int, quoteStyle pr.Quotes, context, page, element interface{}) []Box {

	// TODO: Some computation done here may be done in computed_values
    // instead. We currently miss at least style_for, counters and quotes
    // context in computer. Some work will still need to be done here though,
    // like box creation for URIs.
    var (
		boxlist []Box
		texts []string
	)

    missingCounters = []
    missingTargetCounters = {}
    inPageContext := context != nil && page != nil

    // Collect missing counters during build_formatting_structure.
    // Pointless to collect missing target counters in MarginBoxes.
    needCollectMissing = targetCollector.collecting && ! inPageContext

    // TODO: remove attribute or set a default value in Box class
    if parentBox.Box().cachedCounterValues == nil {
        // Store the counter_values in the parent_box to make them accessible
        // in @page context. Obsoletes the parse_again function's deepcopy.
        // TODO: Is propbably superfluous inPageContext.
		parentBox.Box().cachedCounterValues = counterValues.Copy()
	}

	chunks := make([]string, len(contentList.Contents))
	for i, content := range contentList.Contents {
		switch content.Type {
		case "string":
			chunks[i] = content.String
		case "url":
			if getImageFromUri != nil {
				value := content.Content.(pr.NamedString)
				if value.Name != "external"{
					// Embedding internal references is impossible
                	continue
				}
				image := getImageFromUri(value.String, "")
				if image != nil {
					text :=  strings.Join(texts, "")
					if text != "" {
						boxlist = append(boxlist,	TextBoxAnonymousFrom(parentBox, text))
					}
					texts = nil
					boxlist = append(boxlist,InlineReplacedBoxAnonymousFrom(parentBox, image))
			}
		case "content":
			addedText := textContentExtractors[content.String](box)
			// Simulate the step of white space processing
			// (normally done during the layout)
			addedText = strings.TrimSpace(addedText)
			chunks[i] = addedText
		case "counter":
			cv, has := counterValues[content.String]
			if !has {
				cv = []int{0}
			}
			counterValue := cv[len(cv)-1]
			chunks[i] = format(counterValue, content.CounterStyle)
		case "counters":
			counterName, separator, counterStyle := content.String, content.Separator, content.CounterStyle
			vs, has := counterValues[counterName]
			if !has {
				vs = []int{0}
			}
			cs := make([]string, len(vs))
			for i, counterValue := range vs {
				cs[i] = format(counterValue, counterStyle)
			}
			chunks[i] = strings.Join(cs, separator)
		case "attr":
			chunks[i] = utils.GetAttribute(*element, content.String)
		}
	}
	return strings.Join(chunks, "")
}

// Set the content-lists by strings.
// These content-lists are used in GCPM properties like ``string-set`` and
// ``bookmark-label``.
func setContentLists(element *utils.HTMLNode, box Box, style pr.Properties, counterValues tree.CounterValues) {
	var stringSet []css.NameValue
	if style.StringSet.String != "none" {
		for _, c := range style.StringSet.Contents {
			stringSet = append(stringSet, css.NameValue{
				Name: c.Name, Value: computeContentListString(element, box, counterValues, c.Values),
			})
		}
	}
	box.Box().stringSet = stringSet

	if style.ContentProperties.Name == "none" {
		box.Box().bookmarkLabel = ""
	} else {
		box.Box().bookmarkLabel = computeContentListString(element, box, counterValues, style.ContentProperties.Values)
	}

}

// Handle the ``counter-*`` properties.
func updateCounters(state *stateShared, style pr.Properties) {
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
	if counterIncrement.String == "auto" {
		// "auto" is the initial value but is not valid in stylesheet:
		// there was no counter-increment declaration for this element.
		// (Or the winning value was "initial".)
		// http://dev.w3.org/csswg/css3-lists/#declaring-a-list-item
		if style.Strings["display"] == "list-item" {
			cis = []css.IntString{{Name: "list-item", Value: 1}}
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

// Yield the box for ::marker pseudo-element if there is one.
// https://drafts.csswg.org/css-lists-3/#marker-pseudo
func markerToBox(element *utils.HTMLNode, state *stateShared, parentStyle pr.Properties, styleFor tree.StyleFor,
	getImageFromUri gifu, targetCollector *tree.TargetCollector) Box {
	style := styleFor.Get(element, "marker")

	// TODO: should be the computed value. When does the used value for
	// `display` differ from the computer value? It's at least wrong for
	// `content` where 'normal' computes as 'inhibit' for pseudo elements.

	box := makeBox(element.Data+"::marker", style, nil)
	children := &box.Box().children

	if style.GetDisplay() == "none" {
		return nil
	}

	image := style.GetListStyleImage()

	if content := style.GetContent().String; content != "normal" && content != "inhibit" {
		*children = append(*children, contentToBoxes(style, box, state.quoteDepth, state.counterValues,
			getImageFromUri, targetCollector)...)
	} else {
		if imageUrl, ok := image.(pr.UrlImage); ok {
			// image may be None here too, in case the image is not available.
			image = getImageFromUri(string(imageUrl), "")
			if image != nil {
				markerBox := InlineReplacedBoxAnonymousFrom(box, image)
				*children = append(*children, &markerBox)
			}
		}
		if len(*children) == 0 && style.GetListStyleType() != "none" {
			counterValue_, has := state.counterValues["list-item"]
			if !has {
				counterValue_ = []int{0}
			}
			counterValue := counterValue_[len(counterValue_)-1]
			// TODO: rtl numbered list has the dot on the left
			markerText := counters.FormatListMarker(counterValue, type_)
			markerBox := TextBoxAnonymousFrom(box, markerText)
			markerBox.Box().style.SetWhiteSpace("pre-wrap")
			*children = append(*children, &markerBox)
		}
	}

	if len(*children) == 0 {
		return nil
	}
	var markerBox Box
	if parentStyle.GetListStylePosition() == "outside" {
		markerBox = &BlockBoxAnonymousFrom(box, *children)
		// We can safely edit everything that can't be changed by user style
		// See https://drafts.csswg.org/css-pseudo-4/#marker-pseudo
		markerBox.Box().style.SetPosition(pr.BoolString{String: "absolute"})
		translateX := pr.Dimension{Value: 100, Unit: pr.Percentage}
		if parentStyle.GetDirection() == "ltr" {
			translateX = pr.Dimension{Value: -100, Unit: pr.Percentage}
		}
		translateY := pr.ZeroPixels
		markerBox.Box().style.SetTransform(pr.Transforms{{String: "translate", Dimensions: pr.Dimensions{translateX, translateY}}})
	} else {
		markerBox = &InlineBoxAnonymousFrom(box, *children)
	}
	return markerBox
}

var reHasNonWhitespace = regexp.MustCompile("\\S")

var hasNonWhitespaceDefault = func(text string) bool {
	return reHasNonWhitespace.MatchString(text)
}

// Return true if ``box`` is a TextBox with only whitespace.
func isWhitespace(box Box, hasNonWhitespace func(string) bool) bool {
	if hasNonWhitespace == nil {
		hasNonWhitespace = hasNonWhitespaceDefault
	}
	textBox, is := box.(*TextBox)
	return is && !hasNonWhitespace(textBox.text)
}

func boxText(box Box) string {
	if tBox, is := box.(*TextBox); is {
		return tBox.text
	} else if box.IsParentBox() {
		var chunks []string
		for _, child := range box.descendants() {
			et := child.Box().elementTag
			if !strings.HasSuffix(et, "::before") && !strings.HasSuffix(et, "::after") {
				if tBox, is := child.(*TextBox); is {
					chunks = append(chunks, tBox.text)
				}
			}
		}
		return strings.Join(chunks, "")
	} else {
		return ""
	}
}

func boxTextFirstLetter(box Box) string {
	// TODO: use the same code as := range inlines.firstLetterToBox
	characterFound := false
	firstLetter := ""
	text := []rune(boxText(box))
	tables := []*unicode.RangeTable{unicode.Ps, unicode.Pe, unicode.Pi, unicode.Pf, unicode.Po}
	for len(text) > 0 {
		nextLetter := text[0]
		isPunc := unicode.In(nextLetter, tables...)
		if !isPunc {
			if characterFound {
				break
			}
			characterFound = true
		}
		firstLetter += string(nextLetter)
		text = text[1:]
	}
	return firstLetter
}

func boxTextBefore(box Box) string {
	if box.IsParentBox() {
		var chunks []string
		for _, child := range box.descendants() {
			et := child.Box().elementTag
			if strings.HasSuffix(et, "::before") && !child.IsParentBox() {
				chunks = append(chunks, boxText(child))
			}
		}
		return strings.Join(chunks, "")
	}
	return ""
}

func boxTextAfter(box Box) string {
	if box.IsParentBox() {
		var chunks []string
		for _, child := range box.descendants() {
			et := child.Box().elementTag
			if strings.HasSuffix(et, "::after") && !child.IsParentBox() {
				chunks = append(chunks, boxText(child))
			}
		}
		return strings.Join(chunks, "")
	}
	return ""
}
