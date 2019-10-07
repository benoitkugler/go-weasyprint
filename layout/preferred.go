package layout

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Preferred and minimum preferred width, aka. max-content and min-content
// width, aka. the shrink-to-fit algorithm.

// Terms used (max-content width, min-content width) are defined in David
// Baron's unofficial draft (http://dbaron.org/css/intrinsic/).

// Return the shrink-to-fit width of ``box``.
// *Warning:* both availableOuterWidth and the return value are
// for width of the *content area*, not margin area.
// http://www.w3.org/TR/CSS21/visudet.html#float-width
func shrinkToFit(context LayoutContext, box bo.Box, availableWidth float32) float32 {
	return utils.Min(utils.Max(minContentWidth(context, box, false), availableWidth), maxContentWidth(context, box, false))
}

// Return the min-content width for ``box``.
// This is the width by breaking at every line-break opportunity.
// outer=true
func minContentWidth(context LayoutContext, box Box, outer bool) float32 {
    if box.Box().IsTableWrapper {
        return tableAndColumnsPreferredWidths(context, box, outer)[0]
    } else if bo.TypeTableCellBox.IsInstance(box) {
        return tableCellMinContentWidth(context, box, outer)
    } else if bo.IsBlockContainer(box) || bo.IsTableColumnBox(box) || bo.IsFlexBox(box) {
        return blockMinContentWidth(context, box, outer)
    } else if bo.TypeTableColumnGroupBox.IsInstance(box) {
        return columnGroupContentWidth(context, box)
    } else if bo.TypeInlineBox.IsInstance(box) || bo.TypeLineBox.IsInstance(box) {
        return inlineMinContentWidth(context, box, outer, true)
    } else if bo.IsReplacedBox(box) {
        return replacedMinContentWidth(box, outer)
    } else if bo.IsFlexContainerBox(box) {
        return flexMinContentWidth(context, box, outer)
    } else {
		log.Fatalf("min-content width for %T not handled yet", box)
    }
} 

// Return the max-content width for ``box``.
// This is the width by only breaking at forced line breaks.
// outer=true
func maxContentWidth(context LayoutContext, box bo.Box, outer bool) float32 {
	if box.Box().IsTableWrapper {
        return tableAndColumnsPreferredWidths(context, box, outer)[0]
    } else if bo.TypeTableCellBox.IsInstance(box) {
        return tableCellMaxContentWidth(context, box, outer)
    } else if bo.IsBlockContainer(box) || bo.IsTableColumnBox(box) || bo.IsFlexBox(box) {
        return blockMaxContentWidth(context, box, outer)
    } else if bo.TypeTableColumnGroupBox.IsInstance(box) {
        return columnGroupContentWidth(context, box)
    } else if bo.TypeInlineBox.IsInstance(box) || bo.TypeLineBox.IsInstance(box) {
        return inlineMaxContentWidth(context, box, outer, true)
    } else if bo.IsReplacedBox(box) {
        return replacedMaxContentWidth(box, outer)
    } else if bo.IsFlexContainerBox(box) {
        return flexMaxContentWidth(context, box, outer)
    } else {
		log.Fatalf("max-content width for %T not handled yet", box)
	}
} 

type fnBlock = func(LayoutContext, Box, bool) float32

// Helper to create ``block*ContentWidth.``
func blockContentWidth(context LayoutContext, box bo.BoxFields, function fnBlock, outer bool) {
	width := box.Style.GetWidth()
	var widthValue float32
    if width.String == "auto" || width.Unit == pr.Percentage {
        // "percentages on the following properties are treated instead as
        // though they were the following: width: auto"
		// http://dbaron.org/css/intrinsic/#outer-intrinsic
		var max float32 = 0
		for _, child := range box.Children {
			if !child.Box().IsAbsolutelyPositioned() {
				v := function(context, child, true)
				if v > max {
					max = v
				}
			}
		}
        widthValue = max
    } else {
		if width.Unit != pr.Px {
			log.Fatalf("expected Px got %d", width.Unit)
		}
        widthValue = width.Value
    }
    return adjust(box, outer, widthValue)
}

// Get box width from given width and box min- and max-widths.
func minMax(box bo.BoxFields, width float32) float32 {
    minWidth := box.Style.GetMinWidth()
	maxWidth := box.Style.GetMaxWidth()
	var resMin, resMax float32
    if minWidth.String == "auto" || minWidth.Unit == pr.Percentage {
        resMin = 0
    } else {
        resMin = minWidth.Value
	} 
	if maxWidth.String == "auto" || maxWidth.Unit == pr.Percentage {
        resMax = pr.Inf
    } else {
        resMax = maxWidth.Value
	} 
	return utils.Max(resMin, utils.Min(width, resMax))
} 

// Add box paddings, borders and margins to ``width``.
// left=true, right=true
func marginWidth(box bo.BoxFields, width float32, left, right bool) {
	var percentages float32 
	var cases []string
	if left {
		cases =append(cases, "margin_left", "padding_left")
}
if right {
	cases = append(cases, "margin_right", "padding_right")
}
    for _, value := range cases {
        styleValue := box.Style[value].(pr.Value)
        if styleValue.String != "auto" {
            switch styleValue.Unit  {
			case  pr.Px:
				width += styleValue.Value
			case pr.Percentage:
                percentages += styleValue.Value
			default:
				log.Fatalf("expected Px or Percentage, got %d", styleValue.Unit)
            }
        }
    }

    if left {
        width += box.Style.GetBorderLeftWidth().Value
	} 
	if right {
        width += box.Style.GetBorderRightWidth().Value
    }

    if percentages < 100 {
        return width / (1 - percentages / 100.)
    } else {
        // Pathological case, ignore
        return 0
	}	
}

// Respect min/max && adjust width depending on ``outer``.
//     If ``outer`` is set to ``true``, return margin width, else return content
//     width.
// left=true, right=true
func adjust(box bo.BoxFields, outer bool, width float32, left, right bool) float32 {
    fixed := minMax(box, width)

    if outer {
        return marginWidth(box, fixed, left, right)
    } else {
        return fixed
    }
}

// Return the min-content width for a ``BlockBox``.
// outer=true
func blockMinContentWidth(context LayoutContext, box bo.Box, outer bool) {
    return blockContentWidth(context, *box.Box(), minContentWidth, outer)
} 

// Return the max-content width for a ``BlockBox``.
// outer=true
func blockMaxContentWidth(context LayoutContext, box bo.Box, outer bool) {
    return blockContentWidth(context, *box.Box(), maxContentWidth, outer)
} 

// Return the min-content width for an ``InlineBox``.
// 
// The width is calculated from the lines from ``skipStack``. If
// ``firstLine`` is ``true``, only the first line minimum width is
// calculated.
// outer=true, skipStack=None, firstLine=false, isLineStart=false
func inlineMinContentWidth(context LayoutContext, box bo.BoxFields, outer bool, skipStack *bo.SkipStack,
                             firstLine, isLineStart bool) float32 {
    widths := inlineLineWidths(context, box, outer, isLineStart, true, skipStack, firstLine)

    if firstLine {
        widths = widths[0:1]
    } else {
        widths[len(widths)-1] -= trailingWhitespaceSize(context, box)
	} 
	return adjust(box, outer, utils.Maxs(widths))
	}

// Return the max-content width for an ``InlineBox``.
// outer=true, isLineStart=false
func inlineMaxContentWidth(context LayoutContext, box bo.BoxFields, outer, isLineStart bool) float32 {
    widths := inlineLineWidths(context, box, outer, isLineStart, false)
	widths[len(widths)-1] -= trailingWhitespaceSize(context, box)
    return adjust(box, outer, utils.Maxs(widths))
} 




// Return the *-content width for a ``TableColumnGroupBox``.
func columnGroupContentWidth(context LayoutContext, box bo.BoxFields) float32 {
    width := box.Style.GetWidth()
	var width_ float32
	if width.String == "auto" || width.Unit == pr.Percentage {
		width_ = 0
	} else if width.Unit == pr.Px {
		width_ = width.Value
	} else {
        log.Fatalf("expected Px got %d", width.Unit) 
    }

    return adjust(box, false, width_, true, true)
}

// Return the min-content width for a ``TableCellBox``.
func tableCellMinContentWidth(context LayoutContext, box_ bo.Box, outer bool) {
	box := box.Box()
	var maxChildrenWidths float32
	for _, child := range box.children {
		if ! child.Box().IsAbsolutelyPositioned() {
			v := minContentWidth(context, child, true)
			if v > maxChildrenWidths {
				maxChildrenWidths = v
			}
		}
	}
	childrenMinWidth := marginWidth(*box, maxChildrenWidths, true, true)

	width := box.Style.GetWidth()
	var cellMinWidth float32
    if width.String != "auto" && width.Unit == pr.Px{
        cellMinWidth = adjust(box, outer, width.Value, true, true)
    }

    return utils.Max(childrenMinWidth, cellMinWidth)
}

// Return the max-content width for a ``TableCellBox``.
func tableCellMaxContentWidth(context LayoutContext, box bo.Box, outer bool) float32 {
    return utils.Max(tableCellMinContentWidth(context, box, outer), blockMaxContentWidth(context, box, outer))
} 

// firstLine=false
func inlineLineWidths(context LayoutContext, box bo.BoxFields, outer, isLineStart, 
	minimum bool, skipStack *bo.SkipStack, firstLine bool) []float32 {
   var (
	   textIndent, currentLine float32
	   skip int
	   out []float32
   )
		if box.Style.GetTextIndent().Unit == pr.Percentage {
        // TODO: this is wrong, text-indent percentages should be resolved
        // before calling this function.
        textIndent = 0
    } else {
        textIndent = box.Style.GetTextIndent().Value
	} 
	currentLine = 0
    if skipStack != nil {
        skip, skipStack = skipStack.Skip, skipStack.Stack
	} 
	for _, child := range box.Children[skip:] {
        if child.Box().IsAbsolutelyPositioned() {
            continue  // Skip
        }
		textBox, isTextBox := child.(*bo.TextBox)
		var lines []float32
        if bo.TypeInlineBox.Isinstance(child) {
			lines = inlineLineWidths(context, child, outer, isLineStart, minimum, 
				skipStack, firstLine)
            if firstLine {
                lines = lines[0:1]
            } 
			if len(lines) == 1 {
                lines[0] = adjust(child, outer, lines[0])
            } else {
                lines[0] = adjust(child, outer, lines[0], true, false)
                lines[len(lines)-1] = adjust(child, outer, lines[len(lines)-1], false, true)
            }
        } else if isTextBox {
			wp := textBox.Style.GetWhiteSpace()
            spaceCollapse := wp == "normal" || wp == "nowrap" || wp == "pre-line"
            if skipStack == nil {
                skip = 0
            } else {
				skip, skipStack = skipStack.Skip, skipStack.Stack
                if skipStack != nil {
					log.Fatalf("expected empty SkipStack, got %v", skipStack)
				}
			} 
			childText := string([]rune(child.Text)[skip:])
            if isLineStart && spaceCollapse {
                childText = strings.TrimLeft(childText, " ")
			}
			if minimum && childText == " " {
                lines = []float32{0, 0}
            } else {
				var maxWidth *int
				if minimum != nil {
					maxWidth = new(int)
				}
				resumeAt := 0
				newResumeAt := new(int)
                for newResumeAt != nil {
                    resumeAt += *newResumeAt
                   tmp := pdf.SplitFirstLine(
                            childText[resumeAt:], child.style, context,
                            maxWidth, child.justificationSpacing,
							true)
					newResumeAt, width = tmp.ResumeAt, tmp.Width
                    lines = append(lines, width)
                    if firstLine {
                        break
                    }
				} 
				if firstLine && newResumeAt != nil && *newResumeAt != 0{
                    currentLine += lines[0]
                    break
                }
            }
        } else {
            // http://www.w3.org/TR/css3-text/#line-break-details
            // "The line breaking behavior of a replaced element
            //  or other atomic inline is equivalent to that
            //  of the Object Replacement Character (U+FFFC)."
            // http://www.unicode.org/reports/tr14/#DescriptionOfProperties
            // "By default, there is a break opportunity
            //  both before and after any inline object."
            if minimum != nil {
                lines = []float32{0, maxContentWidth(context, child), 0}
            } else {
                lines = []float32{maxContentWidth(context, child)}
            }
		} 
		// The first text line goes on the current line
        currentLine += lines[0]
        if len(lines) > 1 {
            // Forced line break
            out = append(out, currentLine + textIndent)
            textIndent = 0
            if len(lines) > 2 {
                for line := range lines[1:-1] {
                    out = append(out, line)
                }
			} 
			currentLine = lines[-1]
		} 
		isLineStart = lines[len(lines)-1] == 0
		skipStack = nil
			}
	out =append(out, currentLine + textIndent)
	return out
			}

// Return the percentage contribution of a cell, column or column group.
// http://dbaron.org/css/intrinsic/#pct-contrib
func percentageContribution(box bo.BoxFields) float32 {
	var minWidth, width float32 
	maxWidth := pr.Inf
	miw, maw, w := box.Style.GetMinWidth(), box.Style.GetMaxWidth(),box.Style.GetWidth()
         if miw.String != "auto" && miw.Unit == pr.Percentage {
 minWidth = miw.Value
		}
         if maw.String != "auto" && maw.Unit == pr.Percentage {
 maxWidth = maw.Value
		}
         if w.String != "auto" && w.Unit == pr.Percentage {
 width= w.Value
		}
    return utils.Max(minWidth, utils.Min(width, maxWidth))
} 

type tableContentWidths struct {
    tableMinContentWidth float32
 tableMaxContentWidth float32
       columnMinContentWidths []float32
 columnMaxContentWidths []float32
       columnIntrinsicPercentages []float32
 constrainedness []float32
       totalHorizontalBorderSpacing float32
 grid []float32
}

// Return content widths for the auto layout table and its columns.
//     http://dbaron.org/css/intrinsic/
// outer = true
func tableAndColumnsPreferredWidths(context LayoutContext, box Box, outer bool) tableContentWidths {
	table_ := box.Box().GetWrappedTable()
	table := table.Box()
    result := context.tables[table]
    if result != nil {
        return result[outer]
    }

    // Create the grid
    var gridWidth, gridHeight int
    rowNumber := 0
    for _, rowGroup := range table.Children {
        for _,  row := range rowGroup.Box().Children {
            for _, cell := range row.Box().Children {
                gridWidth = utils.MaxInt(cell.Box().GridX + cell.Box().Colspan, gridWidth)
                gridHeight = utils.MaxInt(rowNumber + cell.Box().Rowspan, gridHeight)
			} 
			rowNumber += 1
        }
	} 
	// zippedGrid = list(zip(*grid)), which is the transpose of grid
	grid, zippedGrid := make([][]bo.Box, gridHeight), make([][]bo.Box, gridWidth)
	for i := range grid {
		grid[i] = make([]bo.Box, gridWidth)
	}
	for j := range zippedGrid {
		grid[j] = make([]bo.Box, gridHeight)
	}
    rowNumber = 0
    for _, rowGroup := range table.Children {
        for _, row := range rowGroup.Children {
            for _, cell := range row.Children {
                grid[rowNumber][cell.Box().GridX] = cell
                zippedGrid[cell.Box().GridX][rowNumber] = cell
			} 
			rowNumber += 1
        }
    }

	// Define the total horizontal border spacing
	var totalHorizontalBorderSpacing float32
    if table.Style.GetBorderCollapse() == "separate" && gridWidth > 0 {
		tot := 1
		for _, column := range zippedGrid {
			any := false
			for _, b := range column {
				if b != nil {
					any = true
					break
				}
			}
			if any {
				tot += 1
			}
		}
        totalHorizontalBorderSpacing = table.Style.GetBorderSpacing()[0] * tot
    } 

    if gridWidth == 0 || gridHeight == 0 {
		table.children = nil
        minWidth := blockMinContentWidth(context, table_, false)
        maxWidth := blockMaxContentWidth(context, table_, false)
        outerMinWidth := adjust(box, true, blockMinContentWidth(context, table_, true))
        outerMaxWidth := adjust(box, true, blockMaxContentWidth(context, table_, true))
		context.tables[table] = map[bool]tableContentWidths{
			false: tableContentWidths{
				tableMinContentWidth: minWidth, 
				tableMaxContentWidth: maxWidth,
				 totalHorizontalBorderSpacing: totalHorizontalBorderSpacing,
				},
			true: tableContentWidths{
				tableMinContentWidth: outerMinWidth, 
				tableMaxContentWidth: outerMaxWidth,
				 totalHorizontalBorderSpacing: totalHorizontalBorderSpacing,
				},
        }
        return context.tables[table][outer]
    }

    columnGroups := make([]bo.Box, gridWidth)
    columns := make([]bo.Box, gridWidth)
	columnNumber := 0
	outerLoop:
    for _, columnGroup := range table.ColumnGroups {
        for _, column := range columnGroup.Box().Children {
            columnGroups[columnNumber] = columnGroup
            columns[columnNumber] = column
            columnNumber += 1
            if columnNumber == gridWidth {
                break outerLoop
            }
    }

    var colspanCells []bo.Box

    // Define the intermediate content widths
    minContentWidths := make([]float32, gridWidth)
    maxContentWidths := make([]float32, gridWidth)
    intrinsicPercentages := make([]float32, gridWidth)

    // Intermediate content widths for span 1
    for i := range range(gridWidth) {
        for groups := range (columnGroups, columns) {
            if groups[i] {
                minContentWidths[i] = max(
                    minContentWidths[i],
                    minContentWidth(context, groups[i]))
                maxContentWidths[i] = max(
                    maxContentWidths[i],
                    maxContentWidth(context, groups[i]))
                intrinsicPercentages[i] = max(
                    intrinsicPercentages[i],
                    percentageContribution(groups[i]))
            }
        } for cell := range zippedGrid[i] {
            if cell {
                if cell.colspan == 1 {
                    minContentWidths[i] = max(
                        minContentWidths[i],
                        minContentWidth(context, cell))
                    maxContentWidths[i] = max(
                        maxContentWidths[i],
                        maxContentWidth(context, cell))
                    intrinsicPercentages[i] = max(
                        intrinsicPercentages[i],
                        percentageContribution(cell))
                } else {
                    colspanCells.append(cell)
                }
            }
        }
    }

    // Intermediate content widths for span > 1 is wrong := range the 4.1 section, as
    // explained := range its third issue. Min- && max-content widths are handled by
    // the excess width distribution method, && percentages do not distribute
    // widths to columns that have originating cells.

    // Intermediate intrinsic percentage widths for span > 1
    for span := range range(1, gridWidth) {
        percentageContributions = []
        for i := range range(gridWidth) {
            percentageContribution = intrinsicPercentages[i]
            for j, cell := range enumerate(zippedGrid[i]) {
                indexes = [k for k := range range(i + 1) if grid[j][k]]
                if not indexes {
                    continue
                } origin = max(indexes)
                originCell = grid[j][origin]
                if originCell.colspan - 1 != span {
                    continue
                } cellSlice = slice(origin, origin + originCell.colspan)
                baselinePercentage = sum(intrinsicPercentages[cellSlice])
            }
        }
    }

                // Cell contribution to intrinsic percentage width
                if intrinsicPercentages[i] == 0 {
                    diff = max(
                        0,
                        percentageContribution(originCell) -
                        baselinePercentage)
                    otherColumnsContributions = [
                        maxContentWidths[j]
                        for j := range range(
                            origin, origin + originCell.colspan)
                        if intrinsicPercentages[j] == 0]
                    otherColumnsContributionsSum = sum(
                        otherColumnsContributions)
                    if otherColumnsContributionsSum == 0 {
                        if otherColumnsContributions {
                            ratio = 1 / len(otherColumnsContributions)
                        } else {
                            ratio = 1
                        }
                    } else {
                        ratio = (
                            maxContentWidths[i] /
                            otherColumnsContributionsSum)
                    } percentageContribution = max(
                        percentageContribution,
                        diff * ratio)
                }

            percentageContributions.append(percentageContribution)

        intrinsicPercentages = percentageContributions

    // Define constrainedness
    constrainedness = [false for i := range range(gridWidth)]
    for i := range range(gridWidth) {
        if (columnGroups[i] && columnGroups[i].style["width"] != "auto" and
                columnGroups[i].style["width"].unit != "%") {
                }
            constrainedness[i] = true
            continue
        if (columns[i] && columns[i].style["width"] != "auto" and
                columns[i].style["width"].unit != "%") {
                }
            constrainedness[i] = true
            continue
        for cell := range zippedGrid[i] {
            if (cell && cell.colspan == 1 and
                    cell.style["width"] != "auto" and
                    cell.style["width"].unit != "%") {
                    }
                constrainedness[i] = true
                break
        }
    }

    intrinsicPercentages = [
        min(percentage, 100 - sum(intrinsicPercentages[:i]))
        for i, percentage := range enumerate(intrinsicPercentages)]

    // Max- && min-content widths for span > 1
    for cell := range colspanCells {
        minContent = minContentWidth(context, cell)
        maxContent = maxContentWidth(context, cell)
        columnSlice = slice(cell.gridX, cell.gridX + cell.colspan)
        columnsMinContent = sum(minContentWidths[columnSlice])
        columnsMaxContent = sum(maxContentWidths[columnSlice])
        if table.style["borderCollapse"] == "separate" {
            spacing = (cell.colspan - 1) * table.style["borderSpacing"][0]
        } else {
            spacing = 0
        }
    }

        if minContent > columnsMinContent + spacing {
            excessWidth = minContent - (columnsMinContent + spacing)
            distributeExcessWidth(
                context, zippedGrid, excessWidth, minContentWidths,
                constrainedness, intrinsicPercentages, maxContentWidths,
                columnSlice)
        }

        if maxContent > columnsMaxContent + spacing {
            excessWidth = maxContent - (columnsMaxContent + spacing)
            distributeExcessWidth(
                context, zippedGrid, excessWidth, maxContentWidths,
                constrainedness, intrinsicPercentages, maxContentWidths,
                columnSlice)
        }

    // Calculate the max- && min-content widths of table && columns
    smallpercentageContributions = [
        maxContentWidths[i] / (intrinsicPercentages[i] / 100.)
        for i := range range(gridWidth)
        if intrinsicPercentages[i]]
    largepercentageContributionNumerator = sum(
        maxContentWidths[i] for i := range range(gridWidth)
        if intrinsicPercentages[i] == 0)
    largepercentageContributionDenominator = (
        (100 - sum(intrinsicPercentages)) / 100.)
    if largepercentageContributionDenominator == 0 {
        if largepercentageContributionNumerator == 0 {
            largepercentageContribution = 0
        } else {
            // "the large percentage contribution of the table [is] an
            // infinitely large number if the numerator is nonzero [and] the
            // denominator of that ratio is 0."
            #
            // http://dbaron.org/css/intrinsic/#autotableintrinsic
            #
            // Please note that "an infinitely large number" is not "infinite",
            // && that"s probably not a coincindence: putting "inf" here breaks
            // some cases (see #305).
            largepercentageContribution = sys.maxsize
        }
    } else {
        largepercentageContribution = (
            largepercentageContributionNumerator /
            largepercentageContributionDenominator)
    }

    tableMinContentWidth = (
        totalHorizontalBorderSpacing + sum(minContentWidths))
    tableMaxContentWidth = (
        totalHorizontalBorderSpacing + max(
            [sum(maxContentWidths), largepercentageContribution] +
            smallpercentageContributions))

    if table.style["width"] != "auto" && table.style["width"].unit == "px" {
        // "percentages on the following properties are treated instead as
        // though they were the following: width: auto"
        // http://dbaron.org/css/intrinsic/#outer-intrinsic
        tableMinWidth = tableMaxWidth = table.style["width"].value
    } else {
        tableMinWidth = tableMinContentWidth
        tableMaxWidth = tableMaxContentWidth
    }

    tableMinContentWidth = max(
        tableMinContentWidth, adjust(
            table, outer=false, width=tableMinWidth))
    tableMaxContentWidth = max(
        tableMaxContentWidth, adjust(
            table, outer=false, width=tableMaxWidth))
    tableOuterMinContentWidth = marginWidth(
        table, marginWidth(box, tableMinContentWidth))
    tableOuterMaxContentWidth = marginWidth(
        table, marginWidth(box, tableMaxContentWidth))

    result = (
        minContentWidths, maxContentWidths, intrinsicPercentages,
        constrainedness, totalHorizontalBorderSpacing, zippedGrid)
    context.tables[table] = result = {
        false: (tableMinContentWidth, tableMaxContentWidth) + result,
        true: (
            (tableOuterMinContentWidth, tableOuterMaxContentWidth) +
            result),
    }
    return result[outer]


// Return the min-content width for an ``InlineReplacedBox``.
func replacedMinContentWidth(box, outer=true) {
    width = box.style["width"]
    if width == "auto" {
        height = box.style["height"]
        if height == "auto" || height.unit == "%" {
            height = "auto"
        } else {
            assert height.unit == "px"
            height = height.value
        } if (box.style["maxWidth"] != "auto" and
                box.style["maxWidth"].unit == "%") {
                }
            // See https://drafts.csswg.org/css-sizing/#intrinsic-contribution
            width = 0
        else {
            image = box.replacement
            iwidth, iheight = image.getIntrinsicSize(
                box.style["imageResolution"], box.style["fontSize"])
            width, _ = defaultImageSizing(
                iwidth, iheight, image.intrinsicRatio, "auto", height,
                defaultWidth=300, defaultHeight=150)
        }
    } else if box.style["width"].unit == "%" {
        // See https://drafts.csswg.org/css-sizing/#intrinsic-contribution
        width = 0
    } else {
        assert width.unit == "px"
        width = width.value
    } return adjust(box, outer, width)
} 

// Return the max-content width for an ``InlineReplacedBox``.
func replacedMaxContentWidth(box, outer=true) {
    width = box.style["width"]
    if width == "auto" {
        height = box.style["height"]
        if height == "auto" || height.unit == "%" {
            height = "auto"
        } else {
            assert height.unit == "px"
            height = height.value
        } image = box.replacement
        iwidth, iheight = image.getIntrinsicSize(
            box.style["imageResolution"], box.style["fontSize"])
        width, _ = defaultImageSizing(
            iwidth, iheight, image.intrinsicRatio, "auto", height,
            defaultWidth=300, defaultHeight=150)
    } else if box.style["width"].unit == "%" {
        // See https://drafts.csswg.org/css-sizing/#intrinsic-contribution
        width = 0
    } else {
        assert width.unit == "px"
        width = width.value
    } return adjust(box, outer, width)
} 

// Return the min-content width for an ``FlexContainerBox``.
func flexMinContentWidth(context, box, outer=true) {
    // TODO: use real values, see
    // https://www.w3.org/TR/css-flexbox-1/#intrinsic-sizes
    minContents = [
        minContentWidth(context, child, outer=true)
        for child := range box.children if child.isFlexItem]
    if not minContents {
        return adjust(box, outer, 0)
    } if (box.style["flexDirection"].startswith("row") and
            box.style["flexWrap"] == "nowrap") {
            }
        return adjust(box, outer, sum(minContents))
    else {
        return adjust(box, outer, max(minContents))
    }
} 

// Return the max-content width for an ``FlexContainerBox``.
func flexMaxContentWidth(context, box, outer=true) {
    // TODO: use real values, see
    // https://www.w3.org/TR/css-flexbox-1/#intrinsic-sizes
    maxContents = [
        maxContentWidth(context, child, outer=true)
        for child := range box.children if child.isFlexItem]
    if not maxContents {
        return adjust(box, outer, 0)
    } if box.style["flexDirection"].startswith("row") {
        return adjust(box, outer, sum(maxContents))
    } else {
        return adjust(box, outer, max(maxContents))
    }
} 

// Return the size of the trailing whitespace of ``box``.
func trailingWhitespaceSize(context, box) {
    from .inlines import splitTextBox, splitFirstLine
} 
    while isinstance(box, (boxes.InlineBox, boxes.LineBox)) {
        if not box.children {
            return 0
        } box = box.children[-1]
    } if not (isinstance(box, boxes.TextBox) && box.text and
            box.style["whiteSpace"] := range ("normal", "nowrap", "pre-line")) {
            }
        return 0
    strippedText = box.text.rstrip(" ")
    if box.style["fontSize"] == 0 || len(strippedText) == len(box.text) {
        return 0
    } if strippedText {
        resume = 0
        while resume is not None {
            oldResume = resume
            oldBox, resume, _ = splitTextBox(context, box, None, resume)
        } assert oldBox
        strippedBox = box.copyWithText(strippedText)
        strippedBox, resume, _ = splitTextBox(
            context, strippedBox, None, oldResume)
        assert strippedBox is not None
        assert resume == nil
        return oldBox.width - strippedBox.width
    } else {
        _, _, _, width, _, _ = splitFirstLine(
            box.text, box.style, context, None, box.justificationSpacing)
        return width

