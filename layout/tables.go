package layout

import (
	"log"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
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

// Layout for a table box.
func tableLayout(context *LayoutContext, table_ bo.InstanceTableBox, maxPositionY pr.Float, skipStack *tree.SkipStack,
	containingBlock, pageIsEmpty bool, absoluteBoxes, fixedBoxes []*AbsolutePlaceholder) blockLayout {
		table := table_.Table()
		columnWidths := table.ColumnWidths

		var borderSpacingX, borderSpacingY pr.Float
	    if table.Style.GetBorderCollapse() == "separate" {
			tmp := table.Style.GetBorderSpacing()
			borderSpacingX, borderSpacingY = tmp[0], tmp[1]
	    } 

	    // TODO: reverse this for direction: rtl
		var columnPositions []pr.Float
		table.ColumnPositions = nil
	    positionX := table.ContentBoxX()
	    rowsX := positionX + borderSpacingX
	    for _, width := range columnWidths {
	        positionX += borderSpacingX
	        columnPositions = append(columnPositions, positionX)
	        positionX += width
		} 
		rowsWidth := positionX - rowsX

	    if table.Style.GetBorderCollapse() == "collapse" {
			skippedRows := 0
	        if skipStack != nil {
	            skippedGroups, groupSkipStack := skipStack.Skip, skipStack.Stack
				skippedRows = 0
	            if groupSkipStack != nil {
	                skippedRows = groupSkipStack.Skip
	            } 
								for _, group := range table.Children[:skippedGroups] {
	                skippedRows += len(group.Box().Children)
	            }
	        } 
			horizontalBorders := table.CollapsedBorderGrid.Horizontal
	        if len(horizontalBorders) != 0 {
				var max float32
				for _, tmp := range horizontalBorders[skippedRows] {
					if tmp.Width > max { 
						max = tmp.Width
					}
				}
				table.BorderTopWidth = pr.Float(max /  2)
	        }
	    }

	    // Make this a sub-function so that many local variables like rowsX
	    // don't need to be passed as parameters.
	    groupLayout := func(group_ Box, positionY, maxPositionY pr.Float,pageIsEmpty bool, skipStack *tree.SkipStack) {
	        var resumeAt *tree.SkipStack
	        nextPage := tree.PageBreak{Break: "any",Page: ""}
	        originalPageIsEmpty := pageIsEmpty
			resolvePercentages2(group_, table_,"")
			group := group_.Box()
	        group.PositionX = rowsX
	        group.PositionY = positionY
	        group.Width = rowsWidth
	        var newGroupChildren []Box
	        // For each rows, cells for which this is the last row (with rowspan)
	        endingCellsByRow := make([][]Box, len(group.Children))

	        isGroupStart := skipStack  == nil
			skip := 0
	        if !isGroupStart  {
	            skip, skipStack = skipStack.Skip, skipStack.Stack
	            if skipStack != nil {// No breaks inside rows for now
					log.Fatalf("expected empty skipStack here, got %v", skipStack)
				}  
			} 
			for i, row_ := range group.Children[skip:] {
				row := row_.Box()
				indexRow := i + skip
	            row_.index = indexRow
	        
	            if newGroupChildren {
	                pageBreak := blockLevelPageBreak( newGroupChildren[len(newGroupChildren)-1], row)
	                if pageBreak == "page"|| pageBreak ==  "recto"|| pageBreak ==  "verso"|| pageBreak ==  "left"|| pageBreak ==  "right" {
	                    nextPage.Break = pageBreak
	                    resumeAt = &tree.SkipStack{Skip:indexRow}
	                    break
	                }
	            }

	            resolvePercentages2(row, table_, "")
	            row.PositionX = rowsX
	            row.PositionY = positionY
	            row.Width = rowsWidth
	            // Place cells at the top of the row and layout their content
	            var newRowChildren []Box
	            for cellIndex, cell_ := range row.Children {
					cell := cell_.Box()
	                spannedWidths := columnWidths[cell.GridX:][:cell.Colspan]
	                // In the fixed layout the grid width is set by cells in
	                // the first row and column elements.
	                // This may be less than the previous value of cell.colspan
	                // if that would bring the cell beyond the grid width.
	                cell.Colspan = len(spannedWidths)
	                if cell.Colspan == 0 {
	                    // The cell is entierly beyond the grid width, remove it
	                    // entierly. Subsequent cells in the same row have greater
	                    // gridX, so they are beyond too.
	                    ignoredCells := row.Children[cellIndex:]
	                    log.Printf("This table row has more columns than the table, ignored %d cells: %v",
	                                   len(ignoredCells), ignoredCells)
	                    break
					} 
					resolvePercentages2(cell_, table, "")
	                cell.PositionX = columnPositions[cell.gridX]
	                cell.PositionY = row.PositionY
	                cell.MarginTop = pr.Float(0)
	                cell.MarginLeft = pr.Float(0)
	                cell.Width = pr.Float(0)
	                bordersPlusPadding := cell.BorderWidth()  // with width==0
	                // TODO: we should remove the number of columns with no
					// originating cells to cell.colspan, see testLayoutTableAuto49
					cell.Width = borderSpacingX * (cell.colspan - 1) -  bordersPlusPadding
					for _, sw := range spannedWidths {
						cell.Width += sw
					}
	                // The computed height is a minimum
	                cell.ComputedHeight = cell.Height
	                cell.Height = pr.Auto	
	                cell = blockContainerLayout(context, cell, pr.Inf, nil, true, absoluteBoxes, fixedBoxes, nil).newBox
					any := false
					for _, child := range cell.Children {
						if child.isFloated() || child.isInNormalFlow() {
							any = true 
							break
						}
					}
	                cell.Empty = !any
	                    
	                cell.ContentHeight = cell.Height
	                if cell.ComputedHeight != pr.Auto {
	                    cell.Height = pr.Max(cell.Height.V(), cell.ComputedHeight.V())
					} 
					newRowChildren = append(newRowChildren, cell)
	            }

	            row = bo.CopyWithChildren(row_, newRowChildren, true, true)

	            // Table height algorithm
	            // http://www.w3.org/TR/CSS21/tables.html#height-layout

	            // cells with vertical-align: baseline
	            var baselineCells []Box
	            for _, cell_ := range row.Children {
					cell := cell_.Box()
	                verticalAlign := cell.Style.GetVerticalAlign()
	                if verticalAlign.String == "top" || verticalAlign.String ==  "middle" || verticalAlign.String ==  "bottom" {
	                    cell.VerticalAlign = verticalAlign.String
	                } else {
	                    // Assume "baseline" for any other value
	                    cell.VerticalAlign = "baseline"
	                    cell.Baseline = cellBaseline(cell)
	                    baselineCells = append(baselineCells, cell)
	                }
				} 
				if len(baselineCells) != 0 {
					for _, cell := range baselineCells {
						if bs:= cell.Box().Baseline; bs > row.Baseline {
							row.Baseline = bs
						}
					}
	                for _, cell := range baselineCells {
	                    extra := row.Baseline - cell.Box().Baseline
	                    if cell.Box().Baseline != row.Baseline && extra != 0 {
	                        addTopPadding(cell, extra)
	                    }
	                }
	            }

	            // row height
	            for _, cell := range row.Children {
	                endingCellsByRow[cell.Box().Rowspan - 1] = append(endingCellsByRow[cell.Box().Rowspan - 1], cell)
				} 
				endingCells := endingCellsByRow[0]
				endingCellsByRow = endingCellsByRow[1:]
				var rowBottomY pr.Float
				if len(endingCells) != 0 {  // in this row
	                if row.Height == pr.Auto {
						for _, cell := range endingCells {
													if v := cell.Box().PositionY + cell.Box().BorderHeight(); v > rowBottomY {
								rowBottomY = v
							}
						}
	                    row.Height = pr.Max(rowBottomY - row.PositionY, 0)
	                } else {
						var m pr.Float
						for _, rowCell := range endingCells {
							if v := rowCell.Box().Height; v > m {
								m = v
							}
						}
	                    row.Height = pr.Max(row.Height, m)
	                    rowBottomY = row.PositionY + row.Height
	                }
	            } else {
	                rowBottomY = row.positionY
	                row.Height = 0
	            }

	            if len(baselineCells) != 0 {
	                row.Baseline = rowBottomY
	            }

	            // Add extra padding to make the cells the same height as the row
	            // and honor vertical-align
	            for _, cell_ := range endingCells {
					cell := cell_.Box()
	                cellBottomY := cell.PositionY + cell.BorderHeight()
	                extra := rowBottomY - cellBottomY
	                if extra != 0 {
	                    if cell.VerticalAlign == "bottom" {
	                        addTopPadding(cell, extra)
	                    } else if cell.VerticalAlign == "middle" {
	                        extra /= 2.
	                        addTopPadding(cell, extra)
	                        cell.PaddingBottom += extra
	                    } else {
	                        cell.PaddingBottom += extra
	                    }
					} 
					if cell.ComputedHeight != "auto" {
	                    var verticalAlignShift pr.Float
	                    if cell.VerticalAlign == "middle" {
	                        verticalAlignShift = ( cell.ComputedHeight - cell.ContentHeight) / 2
	                    } else if cell.VerticalAlign == "bottom" {
	                        verticalAlignShift = cell.ComputedHeight - cell.ContentHeight
						} 
						if verticalAlignShift > 0 {
	                        for _, child := range cell.Children {
	                            child.Translate(child,0 ,verticalAlignShift, false)
	                        }
	                    }
	                }
	            }

	            nextPositionY := row.PositionY + row.Height.V() + borderSpacingY
	            // Break if this row overflows the page, unless there is no
	            // other content on the page.
	            if nextPositionY > maxPositionY && ! pageIsEmpty {
	                if len(newGroupChildren) != 0 {
	                    previousRow := newGroupChildren[len(newGroupChildren)-1]
	                    pageBreak := blockLevelPageBreak(previousRow, row)
	                    if pageBreak == "avoid" {
	                        newGroupChildren, resumeAt = findEarlierPageBreak( newGroupChildren, absoluteBoxes, fixedBoxes)
	                        if newGroupChildren != nil || resumeAt != nil {
	                            break
	                        }
	                    } else {
	                        resumeAt = &tree.SkipStack{Skip:indexRow}
	                        break
	                    }
					} 
					if originalPageIsEmpty {
	                    resumeAt = &tree.SkipStack{Skip:indexRow}
	                } else {
	                    return nil, nil, nextPage
					} 
					break
	            }

	            positionY = nextPositionY
	            newGroupChildren = append(newGroupChildren, row)
	            pageIsEmpty = false
			}

	        // Do not keep the row group if we made a page break
	        // before any of its rows or with "avoid"
	        if bi := group.Style.GetBreakInside(); resumeAt != nil && ! originalPageIsEmpty && (
	                bi == "avoid" || bi ==  "avoid-page" || len(newGroupChildren) == 0 ) {
				return nil, nil, nextPage
					}

	        group_ = bo.CopyWithChildren(group_,  newGroupChildren, isGroupStart, resumeAt  == nil )
			group = group_.Box()
	        // Set missing baselines in a second loop because of rowspan
	        for _, row_ := range group.Children {
				row := row_.Box()
	            if row.Baseline  == nil  {
	                if len(row.Children) != 0 {
	                    // lowest bottom content edge
	                    row.Baseline = max(
	                        cell.contentBoxY() + cell.height
	                        for cell := range row.children) - row.positionY
	                } else {
	                    row.Baseline = 0
	                }
	            }
			} 
			group.height = positionY - group.positionY
	        if group.children {
	            // The last border spacing is outside of the group.
	            group.height -= borderSpacingY
	        }

	        return group, resumeAt, nextPage
		}

	    bodyGroupsLayout := func(skipStack, positionY, maxPositionY, pageIsEmpty) {        
	        if skipStack  == nil  {
	            skip = 0
	        } else {
	            skip, skipStack = skipStack
			} 
			newTableChildren = []
	        resumeAt = None
	        nextPage = {"break": "any", "page": None}

	        for i, group := range enumerate(table.children[skip:]) {
	            indexGroup = i + skip
	            group.index = indexGroup


	            if group.isHeader || group.isFooter {
	                continue
	            }

	            if newTableChildren {
	                pageBreak = blockLevelPageBreak(
	                    newTableChildren[-1], group)
	                if pageBreak := range ("page", "recto", "verso", "left", "right") {
	                    nextPage["break"] = pageBreak
	                    resumeAt = (indexGroup, None)
	                    break
	                }
	            }

	            newGroup, resumeAt, nextPage = groupLayout(
	                group, positionY, maxPositionY, pageIsEmpty, skipStack)
	            skipStack = None

	            if newGroup  == nil  {
	                if newTableChildren {
	                    previousGroup = newTableChildren[-1]
	                    pageBreak = blockLevelPageBreak(previousGroup, group)
	                    if pageBreak == "avoid" {
	                        earlierPageBreak = findEarlierPageBreak(
	                            newTableChildren, absoluteBoxes, fixedBoxes)
	                        if earlierPageBreak  != nil  {
	                            newTableChildren, resumeAt = earlierPageBreak
	                            break
	                        }
	                    } resumeAt = (indexGroup, None)
	                } else {
	                    return None, None, nextPage, positionY
	                } break
	            }

	            newTableChildren.append(newGroup)
	            positionY += newGroup.height + borderSpacingY
	            pageIsEmpty = false

	            if resumeAt {
	                resumeAt = (indexGroup, resumeAt)
	                break
	            }

	        return newTableChildren, resumeAt, nextPage, positionY
			}

	    // Layout for row groups, rows && cells
	    positionY = table.contentBoxY() + borderSpacingY
	    initialPositionY = positionY

	    allGroupsLayout := func() {
	        if table.children && table.children[0].isHeader {
	            header = table.children[0]
	            header, resumeAt, nextPage = groupLayout(
	                header, positionY, maxPositionY,
	                skipStack=None, pageIsEmpty=false)
	            if header && ! resumeAt {
	                headerHeight = header.height + borderSpacingY
	            } else:  // Header too big for the page
	                header = None
	        } else {
	            header = None
	        }
	    
	        if table.children && table.children[-1].isFooter {
	            footer = table.children[-1]
	            footer, resumeAt, nextPage = groupLayout(
	                footer, positionY, maxPositionY,
	                skipStack=None, pageIsEmpty=false)
	            if footer && ! resumeAt {
	                footerHeight = footer.height + borderSpacingY
	            } else:  // Footer too big for the page
	                footer = None
	        } else {
	            footer = None
	        }

	        // Don"t remove headers && footers if breaks are avoided := range line groups
	        skip = skipStack[0] if skipStack else 0
	        avoidBreaks = false
	        for group := range table.children[skip:] {
	            if ! group.isHeader && ! group.isFooter {
	                avoidBreaks = (
	                    group.style["breakInside"] := range ("avoid", "avoid-page"))
	                break
	            }
	        }

	        if header && footer {
	            // Try with both the header && footer
	            newTableChildren, resumeAt, nextPage, endPositionY = (
	                bodyGroupsLayout(
	                    skipStack,
	                    positionY=positionY + headerHeight,
	                    maxPositionY=maxPositionY - footerHeight,
	                    pageIsEmpty=avoidBreaks))
	            if newTableChildren || ! pageIsEmpty {
	                footer.translate(dy=endPositionY - footer.positionY)
	                endPositionY += footerHeight
	                return (header, newTableChildren, footer,
	                        endPositionY, resumeAt, nextPage)
	            } else {
	                // We could ! fit any content, drop the footer
	                footer = None
	            }
	        }

	        if header && ! footer {
	            // Try with just the header
	            newTableChildren, resumeAt, nextPage, endPositionY = (
	                bodyGroupsLayout(
	                    skipStack,
	                    positionY=positionY + headerHeight,
	                    maxPositionY=maxPositionY,
	                    pageIsEmpty=avoidBreaks))
	            if newTableChildren || ! pageIsEmpty {
	                return (header, newTableChildren, footer,
	                        endPositionY, resumeAt, nextPage)
	            } else {
	                // We could ! fit any content, drop the header
	                header = None
	            }
	        }

	        if footer && ! header {
	            // Try with just the footer
	            newTableChildren, resumeAt, nextPage, endPositionY = (
	                bodyGroupsLayout(
	                    skipStack,
	                    positionY=positionY,
	                    maxPositionY=maxPositionY - footerHeight,
	                    pageIsEmpty=avoidBreaks))
	            if newTableChildren || ! pageIsEmpty {
	                footer.translate(dy=endPositionY - footer.positionY)
	                endPositionY += footerHeight
	                return (header, newTableChildren, footer,
	                        endPositionY, resumeAt, nextPage)
	            } else {
	                // We could ! fit any content, drop the footer
	                footer = None
	            }
	        }

	        assert ! (header || footer)
	        newTableChildren, resumeAt, nextPage, endPositionY = (
	            bodyGroupsLayout(
	                skipStack, positionY, maxPositionY, pageIsEmpty))
	        return (
	            header, newTableChildren, footer, endPositionY, resumeAt,
	            nextPage)
			}

	    // Closure getting the column cells.
	    getColumnCells := func(table, column) {
	        return lambda: [
	            cell
	            for rowGroup := range table.children
	            for row := range rowGroup.children
	            for cell := range row.children
	            if cell.gridX == column.gridX]
	    }

	    header, newTableChildren, footer, positionY, resumeAt, nextPage = allGroupsLayout()

	    if newTableChildren  == nil  {
	        assert resumeAt  == nil
	        table = None
	        adjoiningMargins = []
	        collapsingThrough = false
	        return (table, resumeAt, nextPage, adjoiningMargins, collapsingThrough)
	    }

	    table = table.copyWithChildren(
	        ([header] if header  != nil  else []) +
	        newTableChildren +
	        ([footer] if footer  != nil  else []),
	        isStart=skipStack  == nil , isEnd=resumeAt  == nil )
	    if table.Style.GetBorderCollapse() == "collapse" {
	        table.skippedRows = skippedRows
	    }

	    // If the height property has a bigger value, just add blank space
	    // below the last row group.
	    table.height = max(
	        table.height if table.height != "auto" else 0,
	        positionY - table.contentBoxY())

	    // Layout for column groups && columns
	    columnsHeight = positionY - initialPositionY
	    if table.children {
	        // The last border spacing is below the columns.
	        columnsHeight -= borderSpacingY
		} 
		for group := range table.columnGroups {
	        for column := range group.children {
	            resolvePercentages(column, containingBlock=table)
	            if column.gridX < len(columnPositions) {
	                column.positionX = columnPositions[column.gridX]
	                column.positionY = initialPositionY
	                column.width = columnWidths[column.gridX]
	                column.height = columnsHeight
	            } else {
	                // Ignore extra empty columns
	                column.positionX = 0
	                column.positionY = 0
	                column.width = 0
	                column.height = 0
	            } resolvePercentages(group, containingBlock=table)
	            column.getCells = getColumnCells(table, column)
			} 
			first = group.children[0]
	        last = group.children[-1]
	        group.positionX = first.positionX
	        group.positionY = initialPositionY
	        group.width = last.positionX + last.width - first.positionX
	        group.height = columnsHeight
	    }

	    if resumeAt && ! pageIsEmpty && (
	            table.Style["breakInside"] := range ("avoid", "avoid-page")) {
	            }
	        table = None
	        resumeAt = None
	    adjoiningMargins = []
	    collapsingThrough = false
	    return table, resumeAt, nextPage, adjoiningMargins, collapsingThrough
}

// Increase the top padding of a box. This also translates the children.
func addTopPadding(box *bo.BoxFields, extraPadding pr.Float) {
	box.PaddingTop = box.PaddingTop.V() + extraPadding
	for _, child := range box.Children {
		child.Translate(child, 0, extraPadding, false)
	}
}

// Run the fixed table layout and return a list of column widths
// http://www.w3.org/TR/CSS21/tables.html#fixed-table-layout
func fixedTableLayout(box *bo.BoxFields) {
	table := box.GetWrappedTable().Table()
	if table.Width== pr.Auto {
		log.Fatalf("table width can't be auto here")
	}
	var allColumns []Box
	for _, columnGroup := range table.ColumnGroups {
		for _, column := range columnGroup.Box().Children {
			allColumns = append(allColumns, column)
		}
	}

	var firstRowCells []Box
	if len(table.Children) != 0 && len(table.Children[0].Box().Children) != 0 {
		firstRowgroup := table.Children[0].Box()
		firstRowCells = firstRowgroup.Children[0].Box().Children
	}
	var sum int
	for _, cell := range firstRowCells {
		sum += cell.Box().Colspan
	}
	numColumns := utils.MaxInt(len(allColumns), sum)
	// ``None`` means not know yet.
	columnWidths := make([]pr.MaybeFloat, numColumns)

	// `width` on column boxes
	for i, column_ := range allColumns {
		column := column_.Box()
		column.Width = resolveOnePercentage(pr.MaybeFloatToValue(column.Width), "width", table.Width.V(), "")
		if !column.Width== pr.Auto {
			columnWidths[i] = column.Width
		}
	}

	var borderSpacingX pr.Float
	if table.Style.GetBorderCollapse() == "separate" {
		borderSpacingX = table.Style.GetBorderSpacing()[0].Value
	}

	// `width` on cells of the first row.
	i := 0
	for _, cell_ := range firstRowCells {
		cell := cell_.Box()
		resolvePercentages2(cell_, table, "")
		if !cell.Width== pr.Auto {
			width := cell.BorderWidth()
			width -= borderSpacingX * pr.Float(cell.Colspan-1)
			// In the general case, this width affects several columns (through
			// colspan) some of which already have a width. Subtract these
			// known widths and divide among remaining columns.
			var columnsWithoutWidth []int // and occupied by this cell
			for j := i; j < i+cell.Colspan; j++ {
				if columnWidths[j] == nil {
					columnsWithoutWidth = append(columnsWithoutWidth, j)
				} else {
					width -= columnWidths[j].V()
				}
			}
			if len(columnsWithoutWidth) != 0 {
				widthPerColumn := width / pr.Float(len(columnsWithoutWidth))
				for _, j := range columnsWithoutWidth {
					columnWidths[j] = widthPerColumn
				}
			}
		}
		i += cell.Colspan
	}

	// Distribute the remaining space equally on columns that do not have
	// a width yet.
	allBorderSpacing := borderSpacingX * pr.Float(numColumns+1)
	var columnsWithoutWidth []int
	minTableWidth := allBorderSpacing
	for i, w := range columnWidths {
		if w != nil {
			minTableWidth += w.V()
		} else {
			columnsWithoutWidth = append(columnsWithoutWidth, i)
		}
	}
	if len(columnsWithoutWidth) != 0 && table.Width.V() >= minTableWidth {
		remainingWidth := table.Width.V() - minTableWidth
		widthPerColumn := remainingWidth / pr.Float(len(columnsWithoutWidth))
		for _, i := range columnsWithoutWidth {
			columnWidths[i] = widthPerColumn
		}
	} else {
		// XXX this is bad, but we were given a broken table to work with...
		for _, i := range columnsWithoutWidth {
			columnWidths[i] = pr.Float(0)
		}
	}
	outCW := make([]pr.Float, len(columnWidths))
	var sumColumnWidths pr.Float
	for i, v := range columnWidths {
		sumColumnWidths += v.V()
		outCW[i] = v.V()
	}
	// If the sum is less than the table width,
	// distribute the remaining space equally
	extraWidth := table.Width.V() - sumColumnWidths - allBorderSpacing
	if extraWidth <= 0 {
		// substract a negative: widen the table
		table.Width = table.Width.V() - extraWidth
	} else if numColumns != 0 {
		extraPerColumn := extraWidth / pr.Float(numColumns)
		for i, w := range outCW {
			outCW[i] = w + extraPerColumn
		}
	}

	// Now we have table.Width == sum(columnWidths) + allBorderSpacing
	// with possible floating point rounding errors.
	// (unless there is zero column)
	table.ColumnWidths = outCW
}

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
	table := table_.Table()
	tmp := tableAndColumnsPreferredWidths(context, box, false)
	var margins pr.Float
	if !box.MarginLeft== pr.Auto {
		margins += box.MarginLeft.V()
	}
	if !box.MarginRight== pr.Auto {
		margins += box.MarginRight.V()
	}
	paddings := table.PaddingLeft.V() + table.PaddingRight.V()

	cbWidth := containingBlock[0]
	availableWidth := cbWidth - margins - paddings

	if table.Style.GetBorderCollapse() == "collapse" {
		availableWidth -= table.BorderLeftWidth.V() + table.BorderRightWidth.V()
	}

	if table.Width== pr.Auto {
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
		table.ColumnWidths = nil
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
			table.ColumnWidths = *upperGuess
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
			table.ColumnWidths = cw
		}
	} else {
		table.ColumnWidths = maxContentGuess
		excessWidth := assignableWidth - sum(maxContentGuess)
		excessWidth = distributeExcessWidth(context, tmp.grid, excessWidth, table.ColumnWidths, tmp.constrainedness,
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
				for _, i := range columns {
					table.ColumnWidths[i] += excessWidth / pr.Float(len(columns))
				}
			}
		}
	}
}

// Find the width of each column and derive the wrapper width.
func tableWrapperWidth(context LayoutContext, wrapper *bo.BoxFields, containingBlock bo.MaybePoint) {
	table := wrapper.GetWrappedTable()
	resolvePercentages(table, containingBlock, "")

	if table.Box().Style.GetTableLayout() == "fixed" && !table.Box().Width== pr.Auto {
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
