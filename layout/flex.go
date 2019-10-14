package layout

import (
	"github.com/benoitkugler/go-weasyprint/utils"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
)


// Layout for flex containers && flex-items.

type oneFlex struct {
	index int 
	box Box 
}

type FlexLine struct {
	line []oneFlex
	crossSize float32 
}

func (f FlexLine) reverse() {
	for left, right := 0, len(f.line)-1; left < right; left, right = left+1, right-1 {
		f.line[left], f.line[right] = f.line[right], f.line[left]
	}
}

func (f FlexLine) sum() float32 {
	var sum float32
	for _, child := range f.line {
		sum += child.box.Box().HypotheticalMainSize 
	}
	return sum
}

func (f FlexLine) allFrozen() bool {
	for _, child := range f.line {
		if !child.box.Box().Frozen {
			return false
		}
	}
	return true
}

func (f FlexLine) adjustements() float32 {
	var sum float32
	for _, child := range f.line {
		sum += child.box.Box().Adjustment 
	}
	return sum
}

func reverse(f []FlexLine) {
	for left, right := 0, len(f)-1; left < right; left, right = left+1, right-1 {
		f[left], f[right] = f[right], f[left]
	}
}

func getAttr(box_ Box, axis, min string) pr.MaybeFloat {
	box := box_.Box()
	var boxAxis pr.MaybeFloat
	if axis == "width" {
		boxAxis = box.Width
		if min == "min" {
			boxAxis = box.MinWidth
		} else if min == "max" {
			boxAxis = box.MaxWidth
		}
	} else {
		boxAxis = box.Height
		if min == "min" {
			boxAxis = box.MinHeight
		} else if min == "max" {
			boxAxis = box.MaxHeight
		}
	}
	return boxAxis
}


func flexLayout(context *LayoutContext, box_ Box, maxPositionY flexLayout, skipStack *bo.SkipStack, containingBlock Box,
                pageIsEmpty bool, absoluteBoxes []*AbsolutePlaceholder, fixedBoxes []Box) blockLayout  {

    context.createBlockFormattingContext()
    var resumeAt *bo.SkipStack
	box := box_.Box()
    // Step 1 is done in formattingStructure.Boxes
    // Step 2
	axis, cross := "height", "width"
    if strings.HasPrefix(string(box.Style.GetFlexDirection(), "row")) {
        axis, cross = "width", "height"
    } 

				 var marginLeft float32 
	if !box.MarginLeft.Auto() {
			 marginLeft :=box.MarginLeft.V() 
	 }
	 			 var marginRight float32
	 if !box.MarginRight.Auto() {
			 marginRight :=box.MarginRight.V() 
	 }
	 			 var marginTop float32
	 if !box.MarginTop.Auto() {
			 marginTop :=box.MarginTop.V() 
	 }
	 			 var marginBottom float32
	 if !box.MarginBottom.Auto() {
			 marginBottom :=box.MarginBottom.V() 
	 }
var availableMainSpace float32
	 boxAxis := getAttr(box, axis)
    if !boxAxis.Auto(){
        availableMainSpace = boxAxis.V()
    } else {
        if axis == "width" {
            availableMainSpace = containingBlock.Width - marginLeft - marginRight -
                box.PaddingLeft - box.PaddingRight - box.BorderLeftWidth.V() - box.BorderRightWidth.V()
        } else {
            mainSpace := maxPositionY - box.PositionY
            if he := containingBlock.Box().Height; ! he.Auto() {
                // if hasattr(containingBlock.Height, "unit") {
                //     assert containingBlock.Height.unit == "px"
                //     mainSpace = min(mainSpace, containingBlock.Height.value)
                 mainSpace = utils.Min(mainSpace, he.V())
			} 
			availableMainSpace = mainSpace - marginTop - marginBottom -
                box.PaddingTop - box.PaddingBottom - box.BorderTopWidth.V() - box.BorderBottomWidth.V()
        }
	}
	var availableCrossSpace float32
	boxCross := getAttr(box, cross)
    if !boxCross.Auto() {
        availableCrossSpace = boxCross.V()
    } else {
        if cross == "height" {
            mainSpace := maxPositionY - box.ContentBoxY()
            if he:= containingBlock.Box().Height; !he.Auto(){
                    mainSpace = utils.Min(mainSpace, he.V())
			} 
			availableCrossSpace = mainSpace -marginTop - marginBottom -
                box.PaddingTop - box.PaddingBottom -  box.BorderTopWidth.V() - box.BorderBottomWidth.V()
        } else {
            availableCrossSpace = containingBlock.Box().Width - marginLeft - marginRight -
                box.PaddingLeft - box.PaddingRight - box.BorderLeftWidth.V() - box.BorderRightWidth.V()
        }
    }

    // Step 3
    children := box.Children
	parentBox_ := bo.CopyWithChildren(box, children)
	parentBox := parentBox_.Box()
    resolvePercentages2(parentBox_, containingBlock, "")
    // TODO: removing auto margins is OK for this step, but margins should be
    // calculated later.
    if parentBox.MarginTop.Auto() {
		box.MarginTop =  pr.Float(0)
		parentBox.MarginTop = pr.Float(0)
	}
	 if parentBox.MarginBottom.Auto() {
		box.MarginBottom =  pr.Float(0)
		parentBox.MarginBottom = pr.Float(0)
	}
	 if parentBox.MarginLeft.Auto() {
		box.MarginLeft =  pr.Float(0)
		parentBox.MarginLeft = pr.Float(0)
	}
	 if parentBox.MarginRight.Auto() {
		box.MarginRight =  pr.Float(0)
		parentBox.MarginRight = pr.Float(0)
	}
	 if bo.TypeFlexBox.IsInstance(parentBox_){
        blockLevelWidth(parentBox, containingBlock)
    } else {
        parentBox.Width = pr.Float(flexMaxContentWidth(context, parentBox, true))
	}
	originalSkipStack := skipStack
    if skipStack != nil {
        if strings.HasSuffix(string(box.Style.GetFlexDirection()), "-reverse") {
            children = children[:skipStack.Skip + 1]
        } else {
            children = children[skipStack.Skip:]
		} 
		skipStack = skipStack.Stack
    } else {
        skipStack = nil
	} 
	childSkipStack := skipStack
    for  _, child_ := range children {
		child := child_.Box()
        if !child.IsFlexItem {
            continue
        }
    
		// See https://www.W3.org/TR/css-flexbox-1/#min-size-auto
		
		mainFlexDirection := ""
        if child.Style.GetOverflow() == "visible" {
            mainFlexDirection = axis
		}
		
		resolvePercentages2(child, containingBlock, mainFlexDirection)
        child.PositionX = parentBox.ContentBoxX()
        child.PositionY = parentBox.ContentBoxY()
        if child.MinWidth.Auto() {
            specifiedSize := pr.Inf
                 if !child.Width.Auto() {
					specifiedSize = child.Width.V()
				}
                newChild := bo.Copy(child_)
            if bo.IsParentBox(child_) {
                newChild = bo.CopyWithChildren(child_, child.Children, true, true)
            } 
			newChild.Box().Style = child.Style.Copy()
            newChild.Box().Style.SetWidth("auto")
            newChild.Box().Style.SetMinWidth(Dimension(0, "px"))
            newChild.Box().Style.SetMaxWidth(Dimension(pr.Inf, "px"))
            contentSize:= minContentWidth(context, newChild, false)
            child.MinWidth = pr.Float(utils.Min(float32(specifiedSize), contentSize))
        } else if child.MinHeight.Auto() {
            // TODO: find a way to get min-content-height
            specifiedSize := pr.Inf
                 if !child.Height.Auto() {
					specifiedSize = child.Height.V()
				}
				newChild := bo.Copy(child_)
				if bo.IsParentBox(child_) {
					newChild = bo.CopyWithChildren(child_, child.Children, true, true)
				} 
			newChild.Style = child.Style.Copy()
			newChild.Box().Style.SetHeight("auto")
            newChild.Box().Style.SetMinHeight(Dimension(0, "px"))
            newChild.Box().Style.SetMaxHeight(Dimension(pr.Inf, "px"))
            newChild = blockLevelLayout(context, newChild, pr.Inf, childSkipStack, parentBox, pageIsEmpty, nil,nil,nil)[0]
            contentSize = newChild.Box().Height
            child.MinHeight = pr.Float(utils.Min(float32(specifiedSize), contentSize))
        }

        child.Style = child.Style.Copy()
		var flexBasis pr.Value
        if child.Style.GetFlexBasis().String == "content" {
			flexBasis = pr.SToV("content")
			child.FlexBasis = flexBasis
        } else {
            child.FlexBasis = pr.MaybeFloatToValue(resolveOnePercentage(child.Style.GetFlexBasis(), "flex_basis", availableMainSpace, ""))
            flexBasis = child.FlexBasis
        }

        // "If a value would resolve to auto for width, it instead resolves
        // to content for flex-basis." Let's do this for height too.
		// See https://www.W3.org/TR/css-flexbox-1/#propdef-flex-basis
		target := &child.Height
		if axis == "width" {
			target = &child.Width
		}
        *target = resolveOnePercentage(pr.MaybeFloatToValue(*target), axis, availableMainSpace, "")
        if flexBasis.String == "auto" {
            if child.Style[axis] == "auto" {
                flexBasis = pr.SToV("content")
            } else {
                if axis == "width" {
                    flexBasis_ := child.BorderWidth()
                    if !child.MarginLeft.Auto() {
                        flexBasis_ += child.MarginLeft.V()
					} 
					if !child.MarginRight.Auto() {
                        flexBasis_ += child.MarginRight.V()
					}
					flexBasis = pr.FToV(flexBasis_)
                } else {
                    flexBasis_ := child.BorderHeight()
                    if !child.MarginTop.Auto() {
                        flexBasis_ += child.MarginTop.V()
					} 
					if !child.MarginBottom.Auto() {
                        flexBasis_ += child.MarginBottom.V()
					}
					flexBasis = pr.FToV(flexBasis_)
                }
            }
        }

        // Step 3.A
        if flexBasis.String != "content" {
            child.FlexBaseSize = flexBasis.Value

        // TODO: Step 3.B
        // TODO: Step 3.C

        // Step 3.D is useless, as we never have infinite sizes on paged media

        // Step 3.E
        } else {
            child.Style[axis] = pr.SToV("max-content")
			styleAxis := child.Style[axis].(pr.Value)
            // TODO: don"t set style value, support *-content values instead
            if styleAxis.String == "max-content" {
                child.Style[axis] = pr.SToV("auto")
                if axis == "width" {
                    child.FlexBaseSize = maxContentWidth(context, child, true)
                } else {
					newChild := bo.Copy(child_)
				if bo.IsParentBox(child_) {
					newChild = bo.CopyWithChildren(child_, child.Children, true, true)
				} 
					newChild.Box().Width = pr.Inf
                    newChild = blockLevelLayout(context, newChild, pr.Inf, childSkipStack,
                        parentBox, pageIsEmpty, absoluteBoxes, fixedBoxes, nil)[0]
                    child.FlexBaseSize = newChild.MarginHeight()
                }
            } else if styleAxis.String == "min-content" {
                child.Style[axis] = pr.SToV("auto")
                if axis == "width" {
                    child.FlexBaseSize = minContentWidth(context, child)
                } else {
					newChild := bo.Copy(child_)
					if bo.IsParentBox(child_) {
						newChild = bo.CopyWithChildren(child_, child.Children, true, true)
					} 
					newChild.Box().Width = pr.Float(0)
                    newChild = blockLevelLayout(context, newChild, pr.Inf, childSkipStack,
                        parentBox, pageIsEmpty, absoluteBoxes, fixedBoxes, nil)[0]
                    child.FlexBaseSize = newChild.MarginHeight()
                }
            } else if styleAxis.Unit == pr.Px {
                // TODO: should we add padding, borders and margins?
				child.FlexBaseSize = styleAxis.Value
			} else {
				log.Fatalf("unexpected Style[axis] : %v", styleAxis)
            }
		}
		if axis == "width" {
			child.HypotheticalMainSize = utils.Max(child.MinWidth, utils.Min(child.FlexBaseSize, child.MaxWidth))
			} else {
				child.HypotheticalMainSize = utils.Max(child.MinHeight, utils.Min(child.FlexBaseSize, child.MaxHeight))
		}

        // Skip stack is only for the first child
        childSkipStack = nil
	}

    // Step 4
    // TODO: the whole step has to be fixed
    if axis == "width" {
        blockLevelWidth(box, containingBlock)
    } else {
        if he := box.Style.GetHeight(); he.String != "auto" {
            box.Height = he.Value
        } else {
            box.Height = pr.Float(0)
            for i, child_ := range children {
				child := child_.Box()
				if ! child.isFlexItem {
                    continue
				} 
				childHeight := child.HypotheticalMainSize +child.BorderTopWidth + child.BorderBottomWidth +
                    child.PaddingTop + child.PaddingBottom
                if getAttr(box, axis).Auto() && childHeight + box.Height.V() > availableMainSpace {
                    resumeAt = &bo.SkipStack{Skip: i}
                    children = children[:i + 1]
					break
				}
                box.Height =pr.Float(box.Height.V() + childHeight)
            }
        }
    }

    // Step 5
    var flexLines []FlexLine

    var line FlexLine
    var lineSize float32
	axisSize := getAttr(box, axis)
	children = append([]Box{}, children...)
	sort.Slice(children, func(i,j int) bool {
		return children[i].Box().Style.GetOrder() < children[j].Box().Style.GetOrder()
	})
    for i, child_ := range children {
		child := child_.Box()
        if ! child.isFlexItem {
            continue
		} 
		lineSize += child.HypotheticalMainSize
        if box.Style.GetFlexWrap() != "nowrap" && lineSize > axisSize {
            if len(line) != 0 {
                flexLines = append(flexLines, line)
                line = FlexLine{{index: i,box: child}}
                lineSize = child.HypotheticalMainSize
            } else {
                line = append(line, FlexLine{index: i,box: child})
                flexLines = append(flexLines, line)
                line = nil
                lineSize = 0
            }
        } else {
            line = append(line, FlexLine{index: i,box: child})
        }
    if len(line) !=  0{
        flexLines = append(flexLines, line)
    }

    // TODO: handle *-reverse using the terminology from the specification
    if box.Style.GetFlexWrap() == "wrap-reverse" {
        reverse(flexLines)
	} 
	if strings.HasPrefix(box.Style.GetFlexDirection(), "-reverse") {
        for _, line := range flexLines {
            line.reverse()
        }
    }

    // Step 6
    // See https://www.W3.org/TR/css-flexbox-1/#resolve-flexible-lengths
    for _, line := range flexLines {
        // Step 6 - 9.7.1
        hypotheticalMainSize := line.Sum()
		flexFactorType := "shrink"
        if hypotheticalMainSize < availableMainSpace {
            flexFactorType = "grow"
        } 

        // Step 6 - 9.7.2
        for _, v := range line {
			i, child := v.index, v.box.Box()
            if flexFactorType == "grow" {
                child.FlexFactor = child.Style.GetFlexGrow()
            } else {
                child.FlexFactor = child.Style.GetFlexShrink()
			} 
			if child.FlexFactor == 0 ||
                (flexFactorType == "grow" &&  child.FlexBaseSize > child.HypotheticalMainSize) ||
                (flexFactorType == "shrink" && child.FlexBaseSize < child.HypotheticalMainSize) {
                child.TargetMainSize = child.HypotheticalMainSize
                child.Frozen = true
            } else {
                child.Frozen = false
            }
        }

        // Step 6 - 9.7.3
        initialFreeSpace := availableMainSpace
        for _, v := range line {
			i, child:= v.index, v.box.Box()
            if child.Frozen {
                initialFreeSpace -= child.TargetMainSize
            } else {
                initialFreeSpace -= child.FlexBaseSize
            }
        }

        // Step 6 - 9.7.4
        for !line.AllFrozen() {
            unfrozenFactorSum := 0
            remainingFreeSpace := availableMainSpace

            // Step 6 - 9.7.4.B
            for _, v := range line {
				i, child:= v.index, v.box.Box()
                if child.Frozen {
                    remainingFreeSpace -= child.TargetMainSize
                } else {
                    remainingFreeSpace -= child.FlexBaseSize
                    unfrozenFactorSum += child.FlexFactor
                }
            }

            if unfrozenFactorSum < 1 {
                initialFreeSpace *= unfrozenFactorSum
            }

            if initialFreeSpace == pr.Inf {
                initialFreeSpace = math.MaxInt32
			} 
			if remainingFreeSpace == pr.Inf {
                remainingFreeSpace = math.MaxInt32
            }

            initialMagnitude := -pr.Inf 
                 if initialFreeSpace > 0 {
					 initialMagnitude = float32(math.Round(math.Log10(float64(initialFreeSpace))))
				 }
            remainingMagnitude := -pr.Inf 
                 if remainingFreeSpace > 0 {
					 remainingMagnitude = float32(math.Round(math.Log10(float64(remainingFreeSpace))))
				 }
            if initialMagnitude < remainingMagnitude {
                remainingFreeSpace = initialFreeSpace
            }

            // Step 6 - 9.7.4.c
            if remainingFreeSpace == 0 {
                // "Do nothing", but we at least set the flexBaseSize as
                // targetMainSize for next step.
                for _, v := range line {
				i, child:= v.index, v.box.Box()
                    if ! child.Frozen {
                        child.TargetMainSize = child.FlexBaseSize
                    }
                }
            } else {
                var scaledFlexShrinkFactorsSum , flexGrowFactorsSum pr.Float
                for _, v := range line {
				i, child:= v.index, v.box.Box()
                    if ! child.Frozen {
                        child.ScaledFlexShrinkFactor = child.FlexBaseSize * child.Style.GetFlexShrink()
                        scaledFlexShrinkFactorsSum += child.ScaledFlexShrinkFactor
                        flexGrowFactorsSum += child.Style.GetFlexGrow()
                    }
				} 
				for _, v := range line {
				i, child:= v.index, v.box.Box()
                    if ! child.Frozen {
                        if flexFactorType == "grow" {
                            ratio := child.Style.GetFlexGrow() / flexGrowFactorsSum
                            child.TargetMainSize = child.FlexBaseSize + remainingFreeSpace * ratio
                        } else if flexFactorType == "shrink" {
                            if scaledFlexShrinkFactorsSum == 0 {
                                child.TargetMainSize = child.FlexBaseSize
                            } else {
                                ratio = child.ScaledFlexShrinkFactor /scaledFlexShrinkFactorsSum
                                child.TargetMainSize = child.FlexBaseSize + remainingFreeSpace * ratio
                            }
                        }
                    }
                }
            }

            // Step 6 - 9.7.4.d
            // TODO: First part of this step is useless until 3.E is correct
            for _, v := range line {
				i, child:= v.index, v.box.Box()
                child.Adjustment = 0
                if ! child.Frozen && child.TargetMainSize < 0 {
                    child.Adjustment = -child.TargetMainSize
                    child.TargetMainSize = 0
                }
            }

            // Step 6 - 9.7.4.e
            adjustments := line.adjustements()
            for _, v := range line {
				i, child:= v.index, v.box.Box()
                if adjustments == 0 {
                    child.Frozen = true
                } else if adjustments > 0 && child.Adjustment > 0 {
                    child.Frozen = true
                } else if adjustments < 0 && child.Adjustment < 0 {
                    child.Frozen = true
                }
            }
		}
        // Step 6 - 9.7.5
        for _, v := range line {
				i, child:= v.index, v.box.Box()
            if axis == "width" {
                child.Width =child.TargetMainSize - child.PaddingLeft - child.PaddingRight -
                    child.BorderLeftWidth.V() - child.BorderRightWidth.V()
                if !child.MarginLeft.Auto() {
                    child.Width = child.Width.V() - child.MarginLeft.V()
				} 
				if !child.MarginRight.Auto() {
                    child.Width = child.Width.V() - child.MarginRight.V()
                }
            } else {
                child.Height = child.TargetMainSize - child.PaddingTop - child.PaddingBottom -
                    child.BorderTopWidth.V() - child.BorderTopWidth.V()
                if !child.MarginLeft.Auto() {
                    child.Height = child.Height.V() - child.MarginLeft.V()
				} 
				if !child.MarginRight.Auto() {
                    child.Height = child.Height.V() - child.MarginRight.V()
                }
            }
		}
	}

    // Step 7
    // TODO: Fix TODO in build.FlexChildren
    // TODO: Handle breaks
    var newFlexLines []FlexLine
    childSkipStack = skipStack
    for _, line := range flexLines {
        var newFlexLine FlexLine
        for _, v := range line {
			child_ := v.box
				child:= child_.Box()
            // TODO: Find another way than calling blockLevelLayoutSwitch to
            // get baseline and child.Height
            if child.MarginTop == "auto" {
                child.MarginTop = 0
			}
			 if child.MarginBottom == "auto" {
                child.MarginBottom = 0
			} 
			childCopy := bo.Copy(child_)
					if bo.IsParentBox(child_) {
						childCopy = bo.CopyWithChildren(child_, child.Children, true, true)
					} 

			blockLevelWidth(childCopy, parentBox)
            newChild, _, _, adjoiningMargins, _ = blockLevelLayoutSwitch(
                    context, childCopy, pr.Inf, childSkipStack,
                    parentBox, pageIsEmpty, absoluteBoxes, fixedBoxes,nil)
		
			child.Baseline = 0
			if bl := findInFlowBaseline(newChild, false, nil); bl != nil {
				child.Baseline = bl.V()
			}
            if cross == "height" {
                child.Height = newChild.Height
                // As flex items margins never collapse (with other flex items
                // or with the flex container), we can add the adjoining margins
                // to the child bottom margin.
                child.MarginBottom = pr.Float(child.MarginBottom.V() + collapseMargin(adjoiningMargins))
            } else {
                child.Width = pr.Float(minContentWidth(context, child, false))
            }

            newFlexLine = append(newFlexLine, oneFlex{index: v.index, box: child})

            // Skip stack is only for the first child
            childSkipStack = nil
		}
        if len(newFlexLine) != 0 {
            newFlexLines = append(newFlexLines, newFlexLine)
		}
	}
    flexLines = newFlexLines

    // Step 8
    crossSize := getAttr(box, cross)
    if len(flexLines) == 1 && !crossSize.Auto() {
        flexLines[0].crossSize = crossSize.V()
    } else {
        for index , line := range flexLines {
            var collectedItems, notCollectedItems []*bo.BoxFields
            for _, v := range line.line {
				child := v.box.Box()
                alignSelf := child.Style.GetAlignSelf()
                if strings.HasPrefix(box.Style.GetFlexDirection(), "row") && alignSelf == "baseline" &&
                        !child.MarginTop.Auto() && !child.MarginBottom.Auto() {
                    collectedItems = append(collectedItems, child)
                } else {
                    notCollectedItems = append(notCollectedItems, child)
                }
			} 
			var crossStartDistance, crossEndDistance float32
            for _, child := range collectedItems {
                baseline := child.Baseline - child.PositionY
                crossStartDistance = utils.Max(crossStartDistance, baseline)
                crossEndDistance = utils.Max(crossEndDistance, child.MarginHeight() - baseline)
			}
			collectedCrossSize := crossStartDistance + crossEndDistance
            var nonCollectedCrossSize float32
            if len(notCollectedItems) != 0 {
                nonCollectedCrossSize = - pr.Inf
                for _, child := range notCollectedItems {
                    if cross == "height" {
                        childCrossSize := child.BorderHeight()
                        if !child.MarginTop.Auto() {
                            childCrossSize += child.MarginTop.V()
						} 
						if !child.MarginBottom.Auto() {
                            childCrossSize += child.MarginBottom.V()
                        }
                    } else {
                        childCrossSize = child.BorderWidth()
                        if !child.MarginLeft.Auto() {
                            childCrossSize += child.MarginLeft.V()
						}
						 if !child.MarginRight.Auto() {
                            childCrossSize += child.MarginRight.V()
                        }
					} 
					nonCollectedCrossSize = utils.Max( childCrossSize, nonCollectedCrossSize)
                }
			} 
			line.crossSize = utils.Max(collectedCrossSize, nonCollectedCrossSize)
			flexLines[index] = line
        }
    }

    if len(flexLines) == 1 {
        line := flexLines[0]
        minCrossSize := getAttr(box, cross, "min")
        if minCrossSize.Auto() {
            minCrossSize = - pr.Inf
		} 
		maxCrossSize := getattr(box, cross, "max")
        if maxCrossSize.Auto() {
            maxCrossSize = pr.Inf
		} 
		line.crossSize = utils.Max(minCrossSize.V(), min(line.crossSize, maxCrossSize.V()))
    }

    // Step 9
    if box.Style["alignContent"] == "stretch" {
        definiteCrossSize = None
        if cross == "height" && box.Style["height"] != "auto" {
            definiteCrossSize = box.Style["height"].value
        } else if cross == "width" {
            if isinstance(box, boxes.FlexBox) {
                if box.Style["width"] == "auto" {
                    definiteCrossSize = availableCrossSpace
                } else {
                    definiteCrossSize = box.Style["width"].value
                }
            }
        } if definiteCrossSize is ! None {
            extraCrossSize = definiteCrossSize - sum(
                line.crossSize for line := range flexLines)
            if extraCrossSize {
                for line := range flexLines {
                    line.crossSize += extraCrossSize / len(flexLines)
                }
            }
        }
    }

    // TODO: Step 10

    // Step 11
    for line := range flexLines {
        for _, v := range line {
				i, child:= v.index, v.box.Box()
            alignSelf = child.Style["alignSelf"]
            if alignSelf == "auto" {
                alignSelf = box.Style["alignItems"]
            } if alignSelf == "stretch" && child.Style[cross] == "auto" {
                crossMargins = (
                    (child.MarginTop, child.MarginBottom)
                    if cross == "height"
                    else (child.MarginLeft, child.MarginRight))
                if child.Style[cross] == "auto" {
                    if "auto" ! := range crossMargins {
                        crossSize = line.crossSize
                        if cross == "height" {
                            crossSize -= (
                                child.MarginTop + child.MarginBottom +
                                child.PaddingTop + child.PaddingBottom +
                                child.BorderTopWidth +
                                child.BorderBottomWidth)
                        } else {
                            crossSize -= (
                                child.MarginLeft + child.MarginRight +
                                child.PaddingLeft + child.PaddingRight +
                                child.BorderLeftWidth +
                                child.BorderRightWidth)
                        } setattr(child, cross, crossSize)
                        // TODO: redo layout?
                    }
                }
            } // else: Cross size has been set by step 7
        }
    }

    // Step 12
    // TODO: handle rtl
    originalPositionAxis = (
        box.contentBoxX() if axis == "width"
        else box.contentBoxY())
    justifyContent = box.Style["justifyContent"]
    if box.Style["flexDirection"].endswith("-reverse") {
        if justifyContent == "flex-start" {
            justifyContent = "flex-end"
        } else if justifyContent == "flex-end" {
            justifyContent = "flex-start"
        }
    }

    for line := range flexLines {
        positionAxis = originalPositionAxis
        if axis == "width" {
            freeSpace = box.Width
            for _, v := range line {
				i, child:= v.index, v.box.Box()
                freeSpace -= child.BorderWidth()
                if child.MarginLeft != "auto" {
                    freeSpace -= child.MarginLeft
                } if child.MarginRight != "auto" {
                    freeSpace -= child.MarginRight
                }
            }
        } else {
            freeSpace = box.Height
            for _, v := range line {
				i, child:= v.index, v.box.Box()
                freeSpace -= child.BorderHeight()
                if child.MarginTop != "auto" {
                    freeSpace -= child.MarginTop
                } if child.MarginBottom != "auto" {
                    freeSpace -= child.MarginBottom
                }
            }
        }
    }

        margins = 0
        for _, v := range line {
				i, child:= v.index, v.box.Box()
            if axis == "width" {
                if child.MarginLeft == "auto" {
                    margins += 1
                } if child.MarginRight == "auto" {
                    margins += 1
                }
            } else {
                if child.MarginTop == "auto" {
                    margins += 1
                } if child.MarginBottom == "auto" {
                    margins += 1
                }
            }
        } if margins {
            freeSpace /= margins
            for _, v := range line {
				i, child:= v.index, v.box.Box()
                if axis == "width" {
                    if child.MarginLeft == "auto" {
                        child.MarginLeft = freeSpace
                    } if child.MarginRight == "auto" {
                        child.MarginRight = freeSpace
                    }
                } else {
                    if child.MarginTop == "auto" {
                        child.MarginTop = freeSpace
                    } if child.MarginBottom == "auto" {
                        child.MarginBottom = freeSpace
                    }
                }
            } freeSpace = 0
        }

        if justifyContent == "flex-end" {
            positionAxis += freeSpace
        } else if justifyContent == "center" {
            positionAxis += freeSpace / 2
        } else if justifyContent == "space-around" {
            positionAxis += freeSpace / len(line) / 2
        } else if justifyContent == "space-evenly" {
            positionAxis += freeSpace / (len(line) + 1)
        }

        for _, v := range line {
				i, child:= v.index, v.box.Box()
            if axis == "width" {
                child.PositionX = positionAxis
                if justifyContent == "stretch" {
                    child.Width += freeSpace / len(line)
                }
            } else {
                child.PositionY = positionAxis
            } positionAxis += (
                child.MarginWidth() if axis == "width"
                else child.MarginHeight())
            if justifyContent == "space-around" {
                positionAxis += freeSpace / len(line)
            } else if justifyContent == "space-between" {
                if len(line) > 1 {
                    positionAxis += freeSpace / (len(line) - 1)
                }
            } else if justifyContent == "space-evenly" {
                positionAxis += freeSpace / (len(line) + 1)
            }
        }

    // Step 13
    positionCross = (
        box.contentBoxY() if cross == "height"
        else box.contentBoxX())
    for line := range flexLines {
        line.lowerBaseline = 0
        // TODO: don"t duplicate this loop
        for _, v := range line {
				i, child:= v.index, v.box.Box()
            alignSelf = child.Style["alignSelf"]
            if alignSelf == "auto" {
                alignSelf = box.Style["alignItems"]
            } if alignSelf == "baseline" && axis == "width" {
                // TODO: handle vertical text
                child.Baseline = child.Baseline - positionCross
                line.lowerBaseline = utils.Max(line.lowerBaseline, child.Baseline)
            }
        } for _, v := range line {
				i, child:= v.index, v.box.Box()
            crossMargins = (
                (child.MarginTop, child.MarginBottom) if cross == "height"
                else (child.MarginLeft, child.MarginRight))
            autoMargins = sum([margin == "auto" for margin := range crossMargins])
            if autoMargins {
                extraCross = line.crossSize
                if cross == "height" {
                    extraCross -= child.BorderHeight()
                    if child.MarginTop != "auto" {
                        extraCross -= child.MarginTop
                    } if child.MarginBottom != "auto" {
                        extraCross -= child.MarginBottom
                    }
                } else {
                    extraCross -= child.BorderWidth()
                    if child.MarginLeft != "auto" {
                        extraCross -= child.MarginLeft
                    } if child.MarginRight != "auto" {
                        extraCross -= child.MarginRight
                    }
                } if extraCross > 0 {
                    extraCross /= autoMargins
                    if cross == "height" {
                        if child.MarginTop == "auto" {
                            child.MarginTop = extraCross
                        } if child.MarginBottom == "auto" {
                            child.MarginBottom = extraCross
                        }
                    } else {
                        if child.MarginLeft == "auto" {
                            child.MarginLeft = extraCross
                        } if child.MarginRight == "auto" {
                            child.MarginRight = extraCross
                        }
                    }
                } else {
                    if cross == "height" {
                        if child.MarginTop == "auto" {
                            child.MarginTop = 0
                        } child.MarginBottom = extraCross
                    } else {
                        if child.MarginLeft == "auto" {
                            child.MarginLeft = 0
                        } child.MarginRight = extraCross
                    }
                }
            } else {
                // Step 14
                alignSelf = child.Style["alignSelf"]
                if alignSelf == "auto" {
                    alignSelf = box.Style["alignItems"]
                } position = "positionY" if cross == "height" else "positionX"
                setattr(child, position, positionCross)
                if alignSelf == "flex-end" {
                    if cross == "height" {
                        child.PositionY += (
                            line.crossSize - child.MarginHeight())
                    } else {
                        child.PositionX += (
                            line.crossSize - child.MarginWidth())
                    }
                } else if alignSelf == "center" {
                    if cross == "height" {
                        child.PositionY += (
                            line.crossSize - child.MarginHeight()) / 2
                    } else {
                        child.PositionX += (
                            line.crossSize - child.MarginWidth()) / 2
                    }
                } else if alignSelf == "baseline" {
                    if cross == "height" {
                        child.PositionY += (
                            line.lowerBaseline - child.Baseline)
                    } else {
                        // Handle vertical text
                        pass
                    }
                } else if alignSelf == "stretch" {
                    if child.Style[cross] == "auto" {
                        if cross == "height" {
                            margins = child.MarginTop + child.MarginBottom
                        } else {
                            margins = child.MarginLeft + child.MarginRight
                        } if child.Style["boxSizing"] == "content-box" {
                            if cross == "height" {
                                margins += (
                                    child.BorderTopWidth +
                                    child.BorderBottomWidth +
                                    child.PaddingTop + child.PaddingBottom)
                            } else {
                                margins += (
                                    child.BorderLeftWidth +
                                    child.BorderRightWidth +
                                    child.PaddingLeft + child.PaddingRight)
                            }
                        } // TODO: don"t set style width, find a way to avoid
                        // width re-calculation after Step 16
                        child.Style[cross] = Dimension(
                            line.crossSize - margins, "px")
                    }
                }
            }
        } positionCross += line.crossSize
    }

    // Step 15
    if box.Style[cross] == "auto" {
        // TODO: handle min-max
        setattr(box, cross, sum(line.crossSize for line := range flexLines))
    }

    // Step 16
    else if len(flexLines) > 1 {
        extraCrossSize = getattr(box, cross) - sum(
            line.crossSize for line := range flexLines)
        direction = "positionY" if cross == "height" else "positionX"
        if extraCrossSize > 0 {
            crossTranslate = 0
            for line := range flexLines {
                for _, v := range line {
				i, child:= v.index, v.box.Box()
                    if child.isFlexItem {
                        currentValue = getattr(child, direction)
                        currentValue += crossTranslate
                        setattr(child, direction, currentValue)
                        if box.Style["alignContent"] == "flex-end" {
                            setattr(
                                child, direction,
                                currentValue + extraCrossSize)
                        } else if box.Style["alignContent"] == "center" {
                            setattr(
                                child, direction,
                                currentValue + extraCrossSize / 2)
                        } else if box.Style["alignContent"] == "space-around" {
                            setattr(
                                child, direction,
                                currentValue + extraCrossSize /
                                len(flexLines) / 2)
                        } else if box.Style["alignContent"] == "space-evenly" {
                            setattr(
                                child, direction,
                                currentValue + extraCrossSize /
                                (len(flexLines) + 1))
                        }
                    }
                } if box.Style["alignContent"] == "space-between" {
                    crossTranslate += extraCrossSize / (len(flexLines) - 1)
                } else if box.Style["alignContent"] == "space-around" {
                    crossTranslate += extraCrossSize / len(flexLines)
                } else if box.Style["alignContent"] == "space-evenly" {
                    crossTranslate += extraCrossSize / (len(flexLines) + 1)
                }
            }
        }
    }

    // TODO: don"t use blockBoxLayout, see TODOs := range Step 14 and
    // build.FlexChildren.
    box = box.copy()
    box.children = []
    childSkipStack = skipStack
    for line := range flexLines {
        for _, v := range line {
				i, child:= v.index, v.box.Box()
            if child.isFlexItem {
                newChild, childResumeAt = blocks.BlockLevelLayoutSwitch(
                    context, child, maxPositionY, childSkipStack, box,
                    pageIsEmpty, absoluteBoxes, fixedBoxes,
                    adjoiningMargins=[])[:2]
                if newChild is None {
                    if resumeAt && resumeAt[0] {
                        resumeAt = (resumeAt[0] + i - 1, None)
                    }
                } else {
                    box.children = append(children, newChild)
                    if childResumeAt is ! None {
                        if originalSkipStack {
                            firstLevelSkip = originalSkipStack[0]
                        } else {
                            firstLevelSkip = 0
                        } if resumeAt {
                            firstLevelSkip += resumeAt[0]
                        } resumeAt = (firstLevelSkip + i, childResumeAt)
                    }
                } if resumeAt {
                    break
                }
            }
        }
    }

            // Skip stack is only for the first child
            childSkipStack = None
        if resumeAt {
            break
        }

    // Set box height
    // TODO: this is probably useless because of step #15
    if axis == "width" && box.Height == "auto" {
        if flexLines {
            box.Height = sum(line.crossSize for line := range flexLines)
        } else {
            box.Height = 0
        }
    }

    // Set baseline
    // See https://www.W3.org/TR/css-flexbox-1/#flex-baselines
    // TODO: use the real algorithm
    if isinstance(box, boxes.InlineFlexBox) {
        if axis == "width":  // && main text direction is horizontal
            box.Baseline = flexLines[0].lowerBaseline if flexLines else 0
        else {
            box.Baseline = ((
                findInFlowBaseline(box.children[0])
                if box.children else 0) || 0)
        }
    }

    context.FinishBlockFormattingContext(box)

    // TODO: check these returned values
    return box, resumeAt, {"break": "any", "page": None}, [], false
				}