package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Layout for tables and internal table boxes.

func reverseBoxes(in []Box) []Box {
	N := len(in)
	out := make([]Box, N)
	for i, v := range in {
		out[N-1-i] = v
	}
	return out
}

// func tableLayout(context, table, maxPositionY, skipStack, containingBlock,
//                  pageIsEmpty, absoluteBoxes, fixedBoxes) {
//                  }
//     """Layout for a table box."""
//     // Avoid a circular import
//     from .blocks import (
//         blockContainerLayout, blockLevelPageBreak,
//         findEarlierPageBreak)

//     columnWidths = table.columnWidths

//     if table.Style.GetBorderCollapse() == "separate" {
//         borderSpacingX, borderSpacingY = table.Style["borderSpacing"]
//     } else {
//         borderSpacingX = 0
//         borderSpacingY = 0
//     }

//     // TODO: reverse this for direction: rtl
//     columnPositions = table.columnPositions = []
//     positionX = table.contentBoxX()
//     rowsX = positionX + borderSpacingX
//     for width := range columnWidths {
//         positionX += borderSpacingX
//         columnPositions.append(positionX)
//         positionX += width
//     } rowsWidth = positionX - rowsX

//     if table.Style.GetBorderCollapse() == "collapse" {
//         if skipStack {
//             skippedGroups, groupSkipStack = skipStack
//             if groupSkipStack {
//                 skippedRows, _ = groupSkipStack
//             } else {
//                 skippedRows = 0
//             } for group := range table.children[:skippedGroups] {
//                 skippedRows += len(group.children)
//             }
//         } else {
//             skippedRows = 0
//         } _, horizontalBorders = table.collapsedBorderGrid
//         if horizontalBorders {
//             table.BorderTopWidth = max(
//                 width for _, (_, width, )
//                 := range horizontalBorders[skippedRows]) / 2
//         }
//     }

//     // Make this a sub-function so that many local variables like rowsX
//     // don"t need to be passed as parameters.
//     def groupLayout(group, positionY, maxPositionY,
//                      pageIsEmpty, skipStack) {
//                      }
//         resumeAt = None
//         nextPage = {"break": "any", "page": None}
//         originalPageIsEmpty = pageIsEmpty
//         resolvePercentages(group, containingBlock=table)
//         group.positionX = rowsX
//         group.positionY = positionY
//         group.width = rowsWidth
//         newGroupChildren = []
//         // For each rows, cells for which this is the last row (with rowspan)
//         endingCellsByRow = [[] for row := range group.children]

//         isGroupStart = skipStack  == nil
//         if isGroupStart {
//             skip = 0
//         } else {
//             skip, skipStack = skipStack
//             assert ! skipStack  // No breaks inside rows for now
//         } for i, row := range enumerate(group.children[skip:]) {
//             indexRow = i + skip
//             row.index = indexRow
//         }

//             if newGroupChildren {
//                 pageBreak = blockLevelPageBreak(
//                     newGroupChildren[-1], row)
//                 if pageBreak := range ("page", "recto", "verso", "left", "right") {
//                     nextPage["break"] = pageBreak
//                     resumeAt = (indexRow, None)
//                     break
//                 }
//             }

//             resolvePercentages(row, containingBlock=table)
//             row.positionX = rowsX
//             row.positionY = positionY
//             row.width = rowsWidth
//             // Place cells at the top of the row && layout their content
//             newRowChildren = []
//             for cell := range row.children {
//                 spannedWidths = columnWidths[cell.gridX:][:cell.colspan]
//                 // In the fixed layout the grid width is set by cells in
//                 // the first row && column elements.
//                 // This may be less than the previous value of cell.colspan
//                 // if that would bring the cell beyond the grid width.
//                 cell.colspan = len(spannedWidths)
//                 if cell.colspan == 0 {
//                     // The cell is entierly beyond the grid width, remove it
//                     // entierly. Subsequent cells := range the same row have greater
//                     // gridX, so they are beyond too.
//                     cellIndex = row.children.index(cell)
//                     ignoredCells = row.children[cellIndex:]
//                     LOGGER.warning("This table row has more columns than "
//                                    "the table, ignored %i cells: %r",
//                                    len(ignoredCells), ignoredCells)
//                     break
//                 } resolvePercentages(cell, containingBlock=table)
//                 cell.positionX = columnPositions[cell.gridX]
//                 cell.positionY = row.positionY
//                 cell.marginTop = 0
//                 cell.marginLeft = 0
//                 cell.width = 0
//                 bordersPlusPadding = cell.borderWidth()  // with width==0
//                 // TODO: we should remove the number of columns with no
//                 // originating cells to cell.colspan, see
//                 // testLayoutTableAuto49
//                 cell.width = (
//                     sum(spannedWidths) +
//                     borderSpacingX * (cell.colspan - 1) -
//                     bordersPlusPadding)
//                 // The computed height is a minimum
//                 cell.computedHeight = cell.height
//                 cell.height = "auto"
//                 cell, _, _, _, _ = blockContainerLayout(
//                     context, cell,
//                     maxPositionY=float("inf"),
//                     skipStack=None,
//                     pageIsEmpty=true,
//                     absoluteBoxes=absoluteBoxes,
//                     fixedBoxes=fixedBoxes)
//                 cell.empty = ! any(
//                     child.isFloated() || child.isInNormalFlow()
//                     for child := range cell.children)
//                 cell.contentHeight = cell.height
//                 if cell.computedHeight != "auto" {
//                     cell.height = max(cell.height, cell.computedHeight)
//                 } newRowChildren.append(cell)
//             }

//             row = row.copyWithChildren(newRowChildren)

//             // Table height algorithm
//             // http://www.w3.org/TR/CSS21/tables.html#height-layout

//             // cells with vertical-align: baseline
//             baselineCells = []
//             for cell := range row.children {
//                 verticalAlign = cell.style["verticalAlign"]
//                 if verticalAlign := range ("top", "middle", "bottom") {
//                     cell.verticalAlign = verticalAlign
//                 } else {
//                     // Assume "baseline" for any other value
//                     cell.verticalAlign = "baseline"
//                     cell.baseline = cellBaseline(cell)
//                     baselineCells.append(cell)
//                 }
//             } if baselineCells {
//                 row.baseline = max(cell.baseline for cell := range baselineCells)
//                 for cell := range baselineCells {
//                     extra = row.baseline - cell.baseline
//                     if cell.baseline != row.baseline && extra {
//                         addTopPadding(cell, extra)
//                     }
//                 }
//             }

//             // row height
//             for cell := range row.children {
//                 endingCellsByRow[cell.rowspan - 1].append(cell)
//             } endingCells = endingCellsByRow.pop(0)
//             if endingCells:  // := range this row
//                 if row.height == "auto" {
//                     rowBottomY = max(
//                         cell.positionY + cell.borderHeight()
//                         for cell := range endingCells)
//                     row.height = max(rowBottomY - row.positionY, 0)
//                 } else {
//                     row.height = max(row.height, max(
//                         rowCell.height for rowCell := range endingCells))
//                     rowBottomY = cell.positionY + row.height
//                 }
//             else {
//                 rowBottomY = row.positionY
//                 row.height = 0
//             }

//             if ! baselineCells {
//                 row.baseline = rowBottomY
//             }

//             // Add extra padding to make the cells the same height as the row
//             // && honor vertical-align
//             for cell := range endingCells {
//                 cellBottomY = cell.positionY + cell.borderHeight()
//                 extra = rowBottomY - cellBottomY
//                 if extra {
//                     if cell.verticalAlign == "bottom" {
//                         addTopPadding(cell, extra)
//                     } else if cell.verticalAlign == "middle" {
//                         extra /= 2.
//                         addTopPadding(cell, extra)
//                         cell.paddingBottom += extra
//                     } else {
//                         cell.paddingBottom += extra
//                     }
//                 } if cell.computedHeight != "auto" {
//                     verticalAlignShift = 0
//                     if cell.verticalAlign == "middle" {
//                         verticalAlignShift = (
//                             cell.computedHeight - cell.contentHeight) / 2
//                     } else if cell.verticalAlign == "bottom" {
//                         verticalAlignShift = (
//                             cell.computedHeight - cell.contentHeight)
//                     } if verticalAlignShift > 0 {
//                         for child := range cell.children {
//                             child.translate(dy=verticalAlignShift)
//                         }
//                     }
//                 }
//             }

//             nextPositionY = row.positionY + row.height + borderSpacingY
//             // Break if this row overflows the page, unless there is no
//             // other content on the page.
//             if nextPositionY > maxPositionY && ! pageIsEmpty {
//                 if newGroupChildren {
//                     previousRow = newGroupChildren[-1]
//                     pageBreak = blockLevelPageBreak(previousRow, row)
//                     if pageBreak == "avoid" {
//                         earlierPageBreak = findEarlierPageBreak(
//                             newGroupChildren, absoluteBoxes, fixedBoxes)
//                         if earlierPageBreak {
//                             newGroupChildren, resumeAt = earlierPageBreak
//                             break
//                         }
//                     } else {
//                         resumeAt = (indexRow, None)
//                         break
//                     }
//                 } if originalPageIsEmpty {
//                     resumeAt = (indexRow, None)
//                 } else {
//                     return None, None, nextPage
//                 } break
//             }

//             positionY = nextPositionY
//             newGroupChildren.append(row)
//             pageIsEmpty = false

//         // Do ! keep the row group if we made a page break
//         // before any of its rows || with "avoid"
//         if resumeAt && ! originalPageIsEmpty && (
//                 group.style["breakInside"] := range ("avoid", "avoid-page") or
//                 ! newGroupChildren) {
//                 }
//             return None, None, nextPage

//         group = group.copyWithChildren(
//             newGroupChildren,
//             isStart=isGroupStart, isEnd=resumeAt  == nil )

//         // Set missing baselines := range a second loop because of rowspan
//         for row := range group.children {
//             if row.baseline  == nil  {
//                 if row.children {
//                     // lowest bottom content edge
//                     row.baseline = max(
//                         cell.contentBoxY() + cell.height
//                         for cell := range row.children) - row.positionY
//                 } else {
//                     row.baseline = 0
//                 }
//             }
//         } group.height = positionY - group.positionY
//         if group.children {
//             // The last border spacing is outside of the group.
//             group.height -= borderSpacingY
//         }

//         return group, resumeAt, nextPage

//     def bodyGroupsLayout(skipStack, positionY, maxPositionY,
//                            pageIsEmpty) {
//                            }
//         if skipStack  == nil  {
//             skip = 0
//         } else {
//             skip, skipStack = skipStack
//         } newTableChildren = []
//         resumeAt = None
//         nextPage = {"break": "any", "page": None}

//         for i, group := range enumerate(table.children[skip:]) {
//             indexGroup = i + skip
//             group.index = indexGroup
//         }

//             if group.isHeader || group.isFooter {
//                 continue
//             }

//             if newTableChildren {
//                 pageBreak = blockLevelPageBreak(
//                     newTableChildren[-1], group)
//                 if pageBreak := range ("page", "recto", "verso", "left", "right") {
//                     nextPage["break"] = pageBreak
//                     resumeAt = (indexGroup, None)
//                     break
//                 }
//             }

//             newGroup, resumeAt, nextPage = groupLayout(
//                 group, positionY, maxPositionY, pageIsEmpty, skipStack)
//             skipStack = None

//             if newGroup  == nil  {
//                 if newTableChildren {
//                     previousGroup = newTableChildren[-1]
//                     pageBreak = blockLevelPageBreak(previousGroup, group)
//                     if pageBreak == "avoid" {
//                         earlierPageBreak = findEarlierPageBreak(
//                             newTableChildren, absoluteBoxes, fixedBoxes)
//                         if earlierPageBreak  != nil  {
//                             newTableChildren, resumeAt = earlierPageBreak
//                             break
//                         }
//                     } resumeAt = (indexGroup, None)
//                 } else {
//                     return None, None, nextPage, positionY
//                 } break
//             }

//             newTableChildren.append(newGroup)
//             positionY += newGroup.height + borderSpacingY
//             pageIsEmpty = false

//             if resumeAt {
//                 resumeAt = (indexGroup, resumeAt)
//                 break
//             }

//         return newTableChildren, resumeAt, nextPage, positionY

//     // Layout for row groups, rows && cells
//     positionY = table.contentBoxY() + borderSpacingY
//     initialPositionY = positionY

//     def allGroupsLayout() {
//         if table.children && table.children[0].isHeader {
//             header = table.children[0]
//             header, resumeAt, nextPage = groupLayout(
//                 header, positionY, maxPositionY,
//                 skipStack=None, pageIsEmpty=false)
//             if header && ! resumeAt {
//                 headerHeight = header.height + borderSpacingY
//             } else:  // Header too big for the page
//                 header = None
//         } else {
//             header = None
//         }
//     }

//         if table.children && table.children[-1].isFooter {
//             footer = table.children[-1]
//             footer, resumeAt, nextPage = groupLayout(
//                 footer, positionY, maxPositionY,
//                 skipStack=None, pageIsEmpty=false)
//             if footer && ! resumeAt {
//                 footerHeight = footer.height + borderSpacingY
//             } else:  // Footer too big for the page
//                 footer = None
//         } else {
//             footer = None
//         }

//         // Don"t remove headers && footers if breaks are avoided := range line groups
//         skip = skipStack[0] if skipStack else 0
//         avoidBreaks = false
//         for group := range table.children[skip:] {
//             if ! group.isHeader && ! group.isFooter {
//                 avoidBreaks = (
//                     group.style["breakInside"] := range ("avoid", "avoid-page"))
//                 break
//             }
//         }

//         if header && footer {
//             // Try with both the header && footer
//             newTableChildren, resumeAt, nextPage, endPositionY = (
//                 bodyGroupsLayout(
//                     skipStack,
//                     positionY=positionY + headerHeight,
//                     maxPositionY=maxPositionY - footerHeight,
//                     pageIsEmpty=avoidBreaks))
//             if newTableChildren || ! pageIsEmpty {
//                 footer.translate(dy=endPositionY - footer.positionY)
//                 endPositionY += footerHeight
//                 return (header, newTableChildren, footer,
//                         endPositionY, resumeAt, nextPage)
//             } else {
//                 // We could ! fit any content, drop the footer
//                 footer = None
//             }
//         }

//         if header && ! footer {
//             // Try with just the header
//             newTableChildren, resumeAt, nextPage, endPositionY = (
//                 bodyGroupsLayout(
//                     skipStack,
//                     positionY=positionY + headerHeight,
//                     maxPositionY=maxPositionY,
//                     pageIsEmpty=avoidBreaks))
//             if newTableChildren || ! pageIsEmpty {
//                 return (header, newTableChildren, footer,
//                         endPositionY, resumeAt, nextPage)
//             } else {
//                 // We could ! fit any content, drop the header
//                 header = None
//             }
//         }

//         if footer && ! header {
//             // Try with just the footer
//             newTableChildren, resumeAt, nextPage, endPositionY = (
//                 bodyGroupsLayout(
//                     skipStack,
//                     positionY=positionY,
//                     maxPositionY=maxPositionY - footerHeight,
//                     pageIsEmpty=avoidBreaks))
//             if newTableChildren || ! pageIsEmpty {
//                 footer.translate(dy=endPositionY - footer.positionY)
//                 endPositionY += footerHeight
//                 return (header, newTableChildren, footer,
//                         endPositionY, resumeAt, nextPage)
//             } else {
//                 // We could ! fit any content, drop the footer
//                 footer = None
//             }
//         }

//         assert ! (header || footer)
//         newTableChildren, resumeAt, nextPage, endPositionY = (
//             bodyGroupsLayout(
//                 skipStack, positionY, maxPositionY, pageIsEmpty))
//         return (
//             header, newTableChildren, footer, endPositionY, resumeAt,
//             nextPage)

//     def getColumnCells(table, column) {
//         """Closure getting the column cells."""
//         return lambda: [
//             cell
//             for rowGroup := range table.children
//             for row := range rowGroup.children
//             for cell := range row.children
//             if cell.gridX == column.gridX]
//     }

//     header, newTableChildren, footer, positionY, resumeAt, nextPage = \
//         allGroupsLayout()

//     if newTableChildren  == nil  {
//         assert resumeAt  == nil
//         table = None
//         adjoiningMargins = []
//         collapsingThrough = false
//         return (
//             table, resumeAt, nextPage, adjoiningMargins, collapsingThrough)
//     }

//     table = table.copyWithChildren(
//         ([header] if header  != nil  else []) +
//         newTableChildren +
//         ([footer] if footer  != nil  else []),
//         isStart=skipStack  == nil , isEnd=resumeAt  == nil )
//     if table.Style.GetBorderCollapse() == "collapse" {
//         table.skippedRows = skippedRows
//     }

//     // If the height property has a bigger value, just add blank space
//     // below the last row group.
//     table.height = max(
//         table.height if table.height != "auto" else 0,
//         positionY - table.contentBoxY())

//     // Layout for column groups && columns
//     columnsHeight = positionY - initialPositionY
//     if table.children {
//         // The last border spacing is below the columns.
//         columnsHeight -= borderSpacingY
//     } for group := range table.columnGroups {
//         for column := range group.children {
//             resolvePercentages(column, containingBlock=table)
//             if column.gridX < len(columnPositions) {
//                 column.positionX = columnPositions[column.gridX]
//                 column.positionY = initialPositionY
//                 column.width = columnWidths[column.gridX]
//                 column.height = columnsHeight
//             } else {
//                 // Ignore extra empty columns
//                 column.positionX = 0
//                 column.positionY = 0
//                 column.width = 0
//                 column.height = 0
//             } resolvePercentages(group, containingBlock=table)
//             column.getCells = getColumnCells(table, column)
//         } first = group.children[0]
//         last = group.children[-1]
//         group.positionX = first.positionX
//         group.positionY = initialPositionY
//         group.width = last.positionX + last.width - first.positionX
//         group.height = columnsHeight
//     }

//     if resumeAt && ! pageIsEmpty && (
//             table.Style["breakInside"] := range ("avoid", "avoid-page")) {
//             }
//         table = None
//         resumeAt = None
//     adjoiningMargins = []
//     collapsingThrough = false
//     return table, resumeAt, nextPage, adjoiningMargins, collapsingThrough

// Increase the top padding of a box. This also translates the children.
func addTopPadding(box *bo.BoxFields, extraPadding pr.Float) {
	box.PaddingTop = box.PaddingTop.V() + extraPadding
	for _, child := range box.Children {
		child.Translate(child, 0, extraPadding, false)
	}
}

// Run the fixed table layout && return a list of column widths
//     http://www.w3.org/TR/CSS21/tables.html#fixed-table-layout
//
func fixedTableLayout(box) {
    table = box.getWrappedTable()
    assert table.Width != "auto"
}
    allColumns = [column for columnGroup := range table.columnGroups
                   for column := range columnGroup.children]
    if table.children && table.children[0].children {
        firstRowgroup = table.children[0]
        firstRowCells = firstRowgroup.children[0].children
    } else {
        firstRowCells = []
    } numColumns = max(
        len(allColumns),
        sum(cell.colspan for cell := range firstRowCells)
    )
    // ``None`` means ! know yet.
    columnWidths = [None] * numColumns

    // `width` on column boxes
    for i, column := range enumerate(allColumns) {
        resolveOnePercentage(column, "width", table.Width)
        if column.width != "auto" {
            columnWidths[i] = column.width
        }
    }

    if table.Style.GetBorderCollapse() == "separate" {
        borderSpacingX, _ = table.Style["borderSpacing"]
    } else {
        borderSpacingX = 0
    }

    // `width` on cells of the first row.
    i = 0
    for cell := range firstRowCells {
        resolvePercentages(cell, table)
        if cell.width != "auto" {
            width = cell.borderWidth()
            width -= borderSpacingX * (cell.colspan - 1)
            // In the general case, this width affects several columns (through
            // colspan) some of which already have a width. Subtract these
            // known widths && divide among remaining columns.
            columnsWithoutWidth = []  // && occupied by this cell
            for j := range range(i, i + cell.colspan) {
                if columnWidths[j]  == nil  {
                    columnsWithoutWidth.append(j)
                } else {
                    width -= columnWidths[j]
                }
            } if columnsWithoutWidth {
                widthPerColumn = width / len(columnsWithoutWidth)
                for j := range columnsWithoutWidth {
                    columnWidths[j] = widthPerColumn
                }
            } del width
        } i += cell.colspan
    } del i

    // Distribute the remaining space equally on columns that do ! have
    // a width yet.
    allBorderSpacing = borderSpacingX * (numColumns + 1)
    minTableWidth = (sum(w for w := range columnWidths if w  != nil ) +
                       allBorderSpacing)
    columnsWithoutWidth = [i for i, w := range enumerate(columnWidths)
                             if w  == nil ]
    if columnsWithoutWidth && table.Width >= minTableWidth {
        remainingWidth = table.Width - minTableWidth
        widthPerColumn = remainingWidth / len(columnsWithoutWidth)
        for i := range columnsWithoutWidth {
            columnWidths[i] = widthPerColumn
        }
    } else {
        // XXX this is bad, but we were given a broken table to work with...
        for i := range columnsWithoutWidth {
            columnWidths[i] = 0
        }
    }

    // If the sum is less than the table width,
    // distribute the remaining space equally
    extraWidth = table.Width - sum(columnWidths) - allBorderSpacing
    if extraWidth <= 0 {
        // substract a negative: widen the table
        table.Width -= extraWidth
    } else if numColumns {
        extraPerColumn = extraWidth / numColumns
        columnWidths = [w + extraPerColumn for w := range columnWidths]
    }

    // Now we have table.Width == sum(columnWidths) + allBorderSpacing
    // with possible floating point rounding errors.
    // (unless there is zero column)
    table.columnWidths = columnWidths

func sum(l []pr.Float) pr.Float {
	var out pr.Float
	for _, v := range l {
		out += v
	}
	return out
}

// Run the auto table layout and return a list of column widths.
// http://www.w3.org/TR/CSS21/tables.html#auto-table-layout
func autoTableLayout(context LayoutContext, box bo.BoxFields, containingBlock bo.Point) {
	table_ := box.GetWrappedTable()
	table := table_.Box()
	tmp := tableAndColumnsPreferredWidths(context, box, false)
	var margins pr.Float
	if !box.MarginLeft.Auto() {
		margins += box.MarginLeft.V()
	}
	if !box.MarginRight.Auto() {
		margins += box.MarginRight.V()
	}
	paddings := table.PaddingLeft.V() + table.PaddingRight.V()

	cbWidth := containingBlock[0]
	availableWidth := cbWidth - margins - paddings

	if table.Style.GetBorderCollapse() == "collapse" {
		availableWidth -= table.BorderLeftWidth.V() + table.BorderRightWidth.V()
	}

	if table.Width.Auto() {
		if availableWidth <= tmp.tableMinContentWidth {
			table.Width = tmp.tableMinContentWidth
		} else if availableWidth < tmp.tableMaxContentWidth {
			table.Width = availableWidth
		} else {
			table.Width = tmp.tableMaxContentWidth
		}
	} else {
		if table.Width.V() < tmp.tableMinContentWidth {
			table.Width = tmp.tableMinContentWidth
		}
	}

	if len(tmp.grid) == 0 {
		*table_.ColumnWidths() = nil
		return
	}

	assignableWidth := table.Width.V() - tmp.totalHorizontalBorderSpacing
	minContentGuess := append([]pr.Float{}, tmp.columnMinContentWidths...)
	minContentPercentageGuess := append([]pr.Float{}, tmp.columnMinContentWidths...)
	minContentSpecifiedGuess := append([]pr.Float{}, tmp.columnMinContentWidths...)
	maxContentGuess := append([]pr.Float{}, tmp.columnMaxContentWidths...)
	L := 4
	guesses := [4]*[]pr.Float{&minContentGuess, &minContentPercentageGuess,
		&minContentSpecifiedGuess, &maxContentGuess}
	for i := range tmp.grid {
		if tmp.columnIntrinsicPercentages[i] != 0 {
			minContentPercentageGuess[i] = pr.Max(
				tmp.columnIntrinsicPercentages[i]/100*assignableWidth,
				tmp.columnMinContentWidths[i])
			minContentSpecifiedGuess[i] = minContentPercentageGuess[i]
			maxContentGuess[i] = minContentPercentageGuess[i]
		} else if tmp.constrainedness[i] {
			minContentSpecifiedGuess[i] = tmp.columnMinContentWidths[i]
		}
	}

	if assignableWidth <= sum(maxContentGuess) {
		// Default values shouldn't be used, but we never know.
		// See https://github.com/Kozea/WeasyPrint/issues/770
		lowerGuess := guesses[0]
		upperGuess := guesses[L-1]

		// We have to work around floating point rounding errors here.
		// The 1e-9 value comes from PEP 485.
		for _, guess := range guesses {
			if sum(*guess) <= assignableWidth*(1+1e-9) {
				lowerGuess = guess
			} else {
				break
			}
		}
		for i := range guesses {
			guess := guesses[L-1-i]
			if sum(*guess) >= assignableWidth*(1-1e-9) {
				upperGuess = guess
			} else {
				break
			}
		}
		if upperGuess == lowerGuess {
			// TODO: Uncomment the assert when bugs #770 && #628 are closed
			// Equivalent to "assert assignableWidth == sum(upperGuess)"
			// assert abs(assignableWidth - sum(upperGuess)) <= (assignableWidth * 1e-9)
			*table_.ColumnWidths() = *upperGuess
		} else {
			addedWidths := make([]pr.Float, len(tmp.grid))
			var sl, saw pr.Float
			for i := range tmp.grid {
				addedWidths[i] = (*upperGuess)[i] - (*lowerGuess)[i]
				sl += (*lowerGuess)[i]
				saw += addedWidths[i]
			}
			availableRatio := (assignableWidth - sl) / saw
			cw := make([]pr.Float, len(tmp.grid))
			for i := range tmp.grid {
				cw[i] = (*lowerGuess)[i] + addedWidths[i]*availableRatio
			}
			*table_.ColumnWidths() = cw
		}
	} else {
		*table_.ColumnWidths() = maxContentGuess
		excessWidth := assignableWidth - sum(maxContentGuess)
		excessWidth = distributeExcessWidth(context, tmp.grid, excessWidth, *table_.ColumnWidths(), tmp.constrainedness,
			tmp.columnIntrinsicPercentages, tmp.columnMaxContentWidths, [2]int{0, len(tmp.grid)})
		if excessWidth != 0 {
			if tmp.tableMinContentWidth < table.Width.V()-excessWidth {
				// Reduce the width of the size from the excess width that has
				// not been distributed.
				table.Width = table.Width.V() - excessWidth
			} else {
				// Break rules
				var columns []int
				for i, column := range tmp.grid {
					anyColumn := false
					for _, b := range column {
						if b != nil {
							anyColumn = true
							break
						}
					}
					if anyColumn {
						columns = append(columns, i)
					}
				}
				cws := table_.ColumnWidths()
				for _, i := range columns {
					(*cws)[i] += excessWidth / pr.Float(len(columns))
				}
			}
		}
	}
}

// Find the width of each column and derive the wrapper width.
func tableWrapperWidth(context LayoutContext, wrapper *bo.BoxFields, containingBlock bo.MaybePoint) {
	table := wrapper.GetWrappedTable()
	resolvePercentages(table, containingBlock, "")

	if table.Box().Style.GetTableLayout() == "fixed" && !table.Box().Width.Auto() {
		fixedTableLayout(wrapper)
	} else {
		autoTableLayout(context, *wrapper, containingBlock.V())
	}

	wrapper.Width = table.Box().BorderWidth()
}

// Return the y position of a cell’s baseline from the top of its border box.
// See http://www.w3.org/TR/CSS21/tables.html#height-layout
func cellBaseline(cell Box) pr.Float {
	result := findInFlowBaseline(cell, false, bo.TypeLineBox, bo.TypeTableRowBox)
	if result != nil {
		return result.V() - cell.Box().PositionY
	} else {
		// Default to the bottom of the content area.
		return cell.Box().BorderTopWidth.V() + cell.Box().PaddingTop.V() + cell.Box().Height.V()
	}
}

// Return the absolute Y position for the first (or last) in-flow baseline
// if any or nil. Can't return "auto".
// last=false, baselineTypes=(boxes.LineBox,)
func findInFlowBaseline(box Box, last bool, baselineTypes ...bo.BoxType) pr.MaybeFloat {
	if len(baselineTypes) == 0 {
		baselineTypes = []bo.BoxType{bo.TypeLineBox}
	}
	// TODO: synthetize baseline when needed
	// See https://www.w3.org/TR/css-align-3/#synthesize-baseline
	for _, type_ := range baselineTypes { // if isinstance(box, baselineTypes)
		if type_.IsInstance(box) {
			return pr.Float(box.Box().PositionY + box.Box().Baseline)
		}
	}
	if bo.IsParentBox(box) && !bo.TypeTableCaptionBox.IsInstance(box) {
		children := box.Box().Children
		if last {
			children = reverseBoxes(children)
		}
		for _, child := range children {
			if child.Box().IsInNormalFlow() {
				result := findInFlowBaseline(child, last, baselineTypes...)
				if result != nil {
					return result
				}
			}
		}
	}
	return nil
}

type indexCol struct {
	i      int
	column []Box
}

// Distribute available width to columns.
//
// Return excess width left (>0) when it's impossible without breaking rules, or 0
//
// See http://dbaron.org/css/intrinsic/#distributetocols
func distributeExcessWidth(context LayoutContext, grid [][]bo.Box, excessWidth pr.Float, columnWidths []pr.Float,
	constrainedness []bool, columnIntrinsicPercentages, columnMaxContentWidths []pr.Float, columnSlice [2]int) pr.Float {
	// First group
	var (
		columns       []indexCol
		currentWidths []pr.Float
	)
	for i, column := range grid[columnSlice[0]:columnSlice[1]] {
		if !constrainedness[i+columnSlice[0]] && columnIntrinsicPercentages[i+columnSlice[0]] == 0 &&
			columnMaxContentWidths[i+columnSlice[0]] > 0 {
			v := indexCol{i: i + columnSlice[0], column: column}
			columns = append(columns, v)
			currentWidths = append(currentWidths, columnWidths[v.i])
		}
	}
	if len(columns) != 0 {
		var (
			sumDifferences pr.Float
			differences    []pr.Float
		)
		for i := range columnMaxContentWidths {
			v := pr.Max(0, columnMaxContentWidths[i]-currentWidths[i])
			sumDifferences += v
			differences = append(differences, v)
		}
		if sumDifferences > excessWidth {
			for i, difference := range differences {
				differences[i] = difference / sumDifferences * excessWidth
			}
		}
		excessWidth -= sumDifferences
		for i, difference := range differences {
			columnWidths[columns[i].i] += difference
		}
	}
	if excessWidth <= 0 {
		return 0
	}

	// Second group
	var columns_ []int
	for i := range grid[columnSlice[0]:columnSlice[1]] {
		if !constrainedness[i+columnSlice[0]] && columnIntrinsicPercentages[i+columnSlice[0]] == 0 {
			columns_ = append(columns_, i+columnSlice[0])
		}
	}

	if l := pr.Float(len(columns_)); l != 0 {
		for _, i := range columns_ {
			columnWidths[i] += excessWidth / l
		}
		return 0
	}

	// Third group
	columns, currentWidths = nil, nil
	for i, column := range grid[columnSlice[0]:columnSlice[1]] {
		if constrainedness[i+columnSlice[0]] && columnIntrinsicPercentages[i+columnSlice[0]] == 0 &&
			columnMaxContentWidths[i+columnSlice[0]] > 0 {
			v := indexCol{i + columnSlice[0], column}
			columns = append(columns, v)
			currentWidths = append(currentWidths, columnWidths[v.i])
		}
	}
	if len(columns) != 0 {
		var (
			sumDifferences pr.Float
			differences    []pr.Float
		)
		for i := range columnMaxContentWidths {
			v := pr.Max(0, columnMaxContentWidths[i]-currentWidths[i])
			sumDifferences += v
			differences = append(differences, v)
		}
		if sumDifferences > excessWidth {
			for i, difference := range differences {
				differences[i] = difference / sumDifferences * excessWidth
			}
		}
		excessWidth -= sumDifferences
		for i, difference := range differences {
			columnWidths[columns[i].i] += difference
		}
	}
	if excessWidth <= 0 {
		return 0
	}

	// Fourth group
	columns = nil
	mapIndex := map[int]bool{}
	for i, column := range grid[columnSlice[0]:columnSlice[1]] {
		if columnIntrinsicPercentages[i+columnSlice[0]] > 0 {
			v := indexCol{i: i + columnSlice[0], column: column}
			columns = append(columns, v)
			mapIndex[v.i] = true
		}
	}
	if L := len(columns); L != 0 {
		var fixedWidth pr.Float
		for j := range grid {
			if !mapIndex[j] {
				fixedWidth += columnWidths[j]
			}
		}
		var percentageWidth pr.Float
		for _, tmp := range columns {
			percentageWidth += columnIntrinsicPercentages[tmp.i]
		}
		var ratio pr.Float
		if fixedWidth != 0 && percentageWidth >= 100 {
			// Sum of the percentages are greater than 100%
			ratio = excessWidth
		} else if fixedWidth == 0 {
			// No fixed width, let's take the whole excess width
			ratio = excessWidth
		} else {
			ratio = fixedWidth / (100 - percentageWidth)
		}

		widths, currentWidths, differences := make([]pr.Float, L), make([]pr.Float, L), make([]pr.Float, L)
		var sumDifferences pr.Float
		for index, tmp := range columns {
			widths[index] = columnIntrinsicPercentages[tmp.i] * ratio
			currentWidths[index] = columnWidths[tmp.i]
			// Allow to reduce the size of the columns to respect the percentage
			differences[index] = widths[index] - currentWidths[index]
			sumDifferences += differences[index]
		}

		if sumDifferences > excessWidth {
			for i, difference := range differences {
				differences[i] = difference / sumDifferences * excessWidth
			}
		}
		excessWidth -= sumDifferences
		for i, difference := range differences {
			columnWidths[columns[i].i] += difference
		}
	}
	if excessWidth <= 0 {
		return 0
	}

	// Bonus: we've tried our best to distribute the extra size, but we
	// failed. Instead of blindly distributing the size among all the colums
	// and breaking all the rules (as said in the draft), let's try to
	// change the columns with no constraint at all, then resize the table,
	// and at least break the rules to make the columns fill the table.

	// Fifth group, part 1
	columns_ = nil
	for i, column := range grid[columnSlice[0]:columnSlice[1]] {
		anyColumn, anyMaxContent := false, false
		for _, cell := range column {
			if cell != nil {
				anyColumn = true
				if maxContentWidth(context, cell, true) != 0 {
					anyMaxContent = true
				}
			}
		}
		if anyColumn && columnIntrinsicPercentages[i+columnSlice[0]] == 0 && anyMaxContent {
			columns_ = append(columns_, i+columnSlice[0])
		}
	}
	if L := pr.Float(len(columns_)); L != 0 {
		for _, i := range columns_ {
			columnWidths[i] += excessWidth / L
		}
		return 0
	}

	// Fifth group, part 2, aka abort
	return excessWidth
}
