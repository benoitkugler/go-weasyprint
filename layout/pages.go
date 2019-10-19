package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
)

// Layout for pages and CSS3 margin boxes.

type contentSizer interface {
	minContentSize() pr.Float
	maxContentSize() pr.Float
}

type orientedBox interface {
	baseBox() *OrientedBox
	restoreBoxAttributes() 
}

type OrientedBox struct {
	context LayoutContext
	box *bo.MarginBox 
	paddingPlusBorder pr.Float
	marginA, marginB, inner pr.MaybeFloat

	// abstract, must be implemented by subclasses
	contentSizer
}

func (o *OrientedBox) baseBox() *OrientedBox {
	return o
}
    
    func (o OrientedBox) sugar() pr.Float {
        return o.paddingPlusBorder + o.marginA + o.marginB
    }

    
    func (o OrientedBox) outer() pr.Float {
        return o.sugar + o.inner.V()
    }

    
    func (o OrientedBox) outerMinContentSize() pr.Float {
		 if o.inner.Auto() {
			return o.sugar  + o.minContentSize()
		}
		return o.sugar  + o.inner.V()
    }

    
    func (o OrientedBox) outerMaxContentSize() pr.Float {
		 if o.inner.Auto() {
			return o.sugar  + o.maxContentSize()
		}
		return o.sugar  + o.inner.V()
    }

    func (o *OrientedBox) shrinkToFit(available pr.Float) {
        o.inner = pr.Min( pr.Max(o.minContentSize(), available), o.maxContentSize())
    }


type VerticalBox struct {
	OrientedBox
}
    func NewVerticalBox(context LayoutContext, box *bo.MarginBox) *VerticalBox {
		self := new(VerticalBox)
        self.context = context
		self.box = box
        // Inner dimension: that of the content area, as opposed to the
        // outer dimension: that of the margin area.
        self.inner = box.Height
        self.marginA = box.MarginTop.V()
        self.marginB = box.MarginBottom.V()
        self.paddingPlusBorder = box.PaddingTop.V() + box.PaddingBottom.V() +
			box.BorderTopWidth.V() + box.BorderBottomWidth.V()
			self.OrientedBox.contentSizer = self
			return self
    }

    func (self *VerticalBox	) restoreBoxAttributes() {
        box := self.box.Box()
        box.Height = self.inner
        box.MarginTop = self.marginA
        box.MarginBottom = self.marginB
    }

    // TODO: Define what are the min-content && max-content heights
    func (self *VerticalBox	) minContentSize() pr.Float{
        return 0
    }

    func (self *VerticalBox	) maxContentSize() pr.Float {
        return 1e6
    }


type HorizontalBox struct {
	OrientedBox
	_minContentSize , _maxContentSize pr.MaybeFloat
}

    func NewHorizontalBox(context LayoutContext, box *bo.MarginBox) *HorizontalBox {
		self := new(HorizontalBox)
		self.context = context
		self.box = box 
        self.inner = box.Width
        self.marginA = box.MarginLeft.V()
        self.marginB = box.MarginRight.V()
        self.paddingPlusBorder =  box.PaddingLeft.V() + box.PaddingRight.V() +
			box.BorderLeftWidth.V() + box.BorderRightWidth.V()
			self.OrientedBox.contentSizer = self
    }

    func (self *HorizontalBox	) restoreBoxAttributes() {
        box := self.box.Box()
        box.Width = self.inner
        box.MarginLeft = self.marginA
        box.MarginRight = self.marginB
    }

    func (self *HorizontalBox	) minContentSize() pr.Float {
        if self._minContentSize  == nil  {
            self._minContentSize =pr.Float(minContentWidth( self.context, self.box, false))
		} 
		return self._minContentSize.V()
    }

    func (self *HorizontalBox	) maxContentSize() pr.Float {
        if self._maxContentSize  == nil  {
            self._maxContentSize =pr.Float(maxContentWidth(self.context, self.box, false))
		} 
		return self._maxContentSize.V()
    }

func countAuto(values ...pr.MaybeFloat) int {
	out := 0
	for _, v := range values {
		if v != nil && v.Auto() {
			out += 1
		}
	}
	return out
}

//     Compute and set a margin box fixed dimension on ``box``, as described in:
//     http://dev.w3.org/csswg/css3-page/#margin-constraints
//     :param box:
//         The margin box to work on
//     :param outer:
//         The target outer dimension (value of a page margin)
//     :param vertical:
//         true to set height, margin-top and margin-bottom; false for width,
//         margin-left and margin-right
//     :param topOrLeft:
//         true if the margin box in if the top half (for vertical==true) or
//         left half (for vertical==false) of the page.
//         This determines which margin should be "auto" if the values are
//         over-constrained. (Rule 3 of the algorithm.)
func computeFixedDimension(context LayoutContext, box_ * bo.MarginBox, outer pr.Float, vertical, topOrLeft bool) {
	var boxOriented orientedBox
	if vertical {
		boxOriented = NewVerticalBox(context, box_)
	} else {
		boxOriented = NewHorizontalBox(context, box_)
	}
	box := boxOriented.baseBox()

    // Rule 2
    total := box.paddingPlusBorder 
    for _, value := range [3]pr.MaybeFloat{box.marginA, box.marginB, box.inner} {
        if !value.Auto() {
			total += value.V()
		}
	}
    if total > outer {
        if box.marginA.Auto() {
            box.marginA = pr.Float(0)
		} 
		if box.marginB.Auto() {
            box.marginB = pr.Float(0)
		}
		 if box.inner.Auto() {
            // XXX this is not in the spec, but without it box.inner
            // would end up with a negative value.
            // Instead, this will trigger rule 3 below.
            // http://lists.w3.org/Archives/Public/www-style/2012Jul/0006.html
            box.inner = pr.Float(0)
        }
	} 
	// Rule 3
    if  countAuto(box.marginA, box.marginB, box.inner) == 0 {
        // Over-constrained
        if topOrLeft {
            box.marginA = pr.Auto
        } else {
            box.marginB = pr.Auto
        }
	} 
	// Rule 4
    if countAuto(box.marginA, box.marginB, box.inner) == 1 {
        if box.inner.Auto() {
            box.inner = outer - box.paddingPlusBorder - box.marginA.V() - box.marginB.V()
        } else if box.marginA.Auto() {
            box.marginA = outer - box.paddingPlusBorder -  box.marginB.V() - box.inner.V()
        } else if box.marginB.Auto() {
            box.marginB = outer - box.paddingPlusBorder -  box.marginA.V() - box.inner.V()
        }
	} 
	// Rule 5
    if box.inner.Auto() {
        if box.marginA.Auto() {
            box.marginA = pr.Float(0)
		} 
		if box.marginB.Auto() {
            box.marginB = pr.Float(0)
		} 
		box.inner = outer - box.paddingPlusBorder - box.marginA.V() - box.marginB.V()
	} 
	// Rule 6
    if box.marginA.Auto() && box.marginB.Auto() {
		v := (  outer - box.PaddingPlusBorder - box.inner.V()) / 2
		box.marginA = v
		 box.marginB = v
    }
	if countAuto(box.marginA, box.marginB, box.inner) > 0 {
		log.Fatalf("unexpected auto value in %v", box)
	}

    boxOriented.restoreBoxAttributes()
}



// Compute and set a margin box fixed dimension on ``box``, as described in:
// http://dev.w3.org/csswg/css3-page/#margin-dimension
// :param sideBoxes: Three boxes on a same side (as opposed to a corner.)
//     A list of:
//     - A @*-left or @*-top margin box
//     - A @*-center or @*-middle margin box
//     - A @*-right or @*-bottom margin box
// :param vertical:
//     true to set height, margin-top and margin-bottom; false for width,
//     margin-left and margin-right
// :param outerSum:
//     The target total outer dimension (max box width or height)
func computeVariableDimension(context LayoutContext, sideBoxes_ [3]*bo.MarginBox, vertical bool, outerSum pr.Float) {
	var sideBoxes [3]orientedBox
	for i, box_ := range sideBoxes_ {
		if vertical {
			sideBoxes[i] = NewVerticalBox(context, box_)
		} else {
			sideBoxes[i] = NewHorizontalBox(context, box_)
		}
	}
    boxA, boxB, boxC := sideBoxes[0].baseBox(), sideBoxes[1].baseBox(), sideBoxes[2].baseBox()

	for _, box_ := range sideBoxes {
		box := box_.baseBox()
        if box.marginA.Auto() {
            box.marginA = pr.Float(0)
		} 
		if box.marginB.Auto() {
            box.marginB = pr.Float(0)
        }
    }

    if boxB.box.Box().IsGenerated {
        if boxB.inner.Auto() {
            acMaxContentSize := 2 * pr.Max(boxA.outerMaxContentSize(), boxC.outerMaxContentSize())
            if outerSum >= (boxB.outerMaxContentSize() + acMaxContentSize) {
                boxB.inner = boxB.maxContentSize()
            } else {
                acMinContentSize := 2 * pr.Max(  boxA.outerMinContentSize(), boxC.outerMinContentSize())
                boxB.inner = boxB.minContentSize()
                available := outerSum - boxB.outer() - acMinContentSize
                if available > 0 {
                    weightAc := acMaxContentSize - acMinContentSize
                    weightB := boxB.maxContentSize() - boxB.minContentSize()
                    weightSum := weightAc + weightB
                    // By definition of maxContentSize and minContentSize,
                    // weights can not be negative. weightSum == 0 implies that
                    // maxContentSize == minContentSize for each box, in
                    // which case the sum can not be both <= and > outerSum
                    // Therefore, one of the last two "if" statements would not
                    // have lead us here.
                    if weightSum <= 0 {
						log.Fatalf("weightSum must be > 0, got %f", weightSum)
					}
                    boxB.inner = boxB.inner.V() + available * weightB / weightSum
                }
            }
		} 
		if boxA.inner.Auto() {
            boxA.shrinkToFit((outerSum - boxB.outer()) / 2 - boxA.sugar())
		} 
		if boxC.inner.Auto() {
            boxC.shrinkToFit((outerSum - boxB.outer()) / 2 - boxC.sugar())
        }
    } else {
        // Non-generated boxes get zero for every box-model property
        if boxB.inner.V() != 0 {
			log.Fatalf("expected boxB.inner == 0, got %v", boxB.inner)
		}
        if boxA.inner.Auto() && boxC.inner.Auto() {
            if outerSum >= (boxA.outerMaxContentSize() + boxC.outerMaxContentSize()) {  
                boxA.inner = boxA.maxContentSize()
                boxC.inner = boxC.maxContentSize()
            }else {
                boxA.inner = boxA.minContentSize()
                boxC.inner = boxC.minContentSize()
                available := outerSum - boxA.outer() - boxC.outer()
                if available > 0 {
                    weightA := boxA.maxContentSize() - boxA.minContentSize()
                    weightC := boxC.maxContentSize() - boxC.minContentSize()
                    weightSum := weightA + weightC
                    // By definition of maxContentSize and minContentSize,
                    // weights can ! be negative. weightSum == 0 implies that
                    // maxContentSize == minContentSize for each box, in
                    // which case the sum can ! be both <= and > outerSum
                    // Therefore, one of the last two "if" statements would not
                    // have lead us here.
					if weightSum <= 0 {
						log.Fatalf("weightSum must be > 0, got %f", weightSum)
					}                    
					boxA.inner = boxA.inner.V() + available * weightA / weightSum
                    boxC.inner = boxC.inner.V() + available * weightC / weightSum
                }
            }
        } else if boxA.inner.Auto() {
            boxA.shrinkToFit(outerSum - boxC.outer.V()	 - boxA.sugar())
        } else if boxC.inner.Auto() {
            boxC.shrinkToFit(outerSum - boxA.outer.V()	 - boxC.sugar())
        }
    }

	// And, we’re done!
	if countAuto(boxA.inner, boxB.inner, boxC.inner) > 0 {
		log.Fatalln("unexpected auto value")
	}

	// Set the actual attributes back.
    for _, box := range sideBoxes {
        box.restoreBoxAttributes()
    }
}

// Drop "pages" counter from style in @page and @margin context.
// Ensure `counter-increment: page` for @page context if not otherwise
// manipulated by the style.
func standardizePageBasedCounters(style pr.Properties, pseudoType string) {
	pageCounterTouched := false
    // XXX "counter-set` not yet supported
    for _, propname := range [2]string{"counter_reset", "counter_increment"} {
		prop := style[propname].(pr.SIntStrings)
		if prop.String == "auto" {
            style[propname] = pr.SIntStrings{Values: pr.IntStrings{}}
            continue
		} 
		var justifiedValues pr.IntStrings
        for _, v := range prop.Values {
            if v.String == "page" {
                pageCounterTouched = true
			} 
			if v.String != "pages" {
                justifiedValues = append(justifiedValues, v)
            }
		}
		 style[propname] = pr.SIntStrings{Values: justifiedValues}
    }

    if pseudoType  == ""  && ! pageCounterTouched {
		current := style.GetCounterIncrement()
		newInc := append(pr.IntStrings{{String: "page", Int: 1}},  current.Values...)
		style.SetCounterIncrement(pr.SIntStrings{Values: newInc})
	}
}


// Yield laid-out margin boxes for this page.
// ``state`` is the actual, up-to-date page-state from
// ``context.pageMaker[context.currentPage]``.
func makeMarginBoxes(context LayoutContext, page *bo.PageBox, state bo.StateShared) []Box {
    // This is a closure only to make calls shorter
	
	// Return a margin box with resolved percentages.
	// The margin box may still have "auto" values.
	// Return ``None`` if this margin box should not be generated.
	// :param atKeyword: which margin box to return, eg. "@top-left"
	// :param containingBlock: as expected by :func:`resolvePercentages`.
    makeBox := func(atKeyword string, containingBlock pr.MaybePoint) *bo.MarginBox {
        style := context.styleFor.Get(page.pageType, atKeyword)
        if style  == nil  {
            // doesn't affect counters
            style = tree.ComputedFromCascaded(nil, nil, page.style)
		} 
		standardizePageBasedCounters(style, atKeyword)
        box := &bo.NewMarginBox(atKeyword, style)
        // Empty boxes should not be generated, but they may be needed for
        // the layout of their neighbors.
		// TODO: should be the computed value.
		ct := style.GetContent().String
        box.IsGenerated = !(ct ==  "normal" || ct == "inhibit" || ct == "none")
        // TODO: get actual counter values at the time of the last page break
        if box.IsGenerated {
            // @margins mustn't manipulate page-context counters
            marginState := state.Copy()
            // quoteDepth, counterValues, counterScopes = marginState
            // TODO: check this, probably useless
            marginState.CounterScopes = append(marginState.CounterScopes, pr.NewSet())
            bo.UpdateCounters(&marginState, box.Style)
            box.Children = bo.ContentToBoxes(
                box.Style, box, marginState.QuoteDepth, marginState.CounterValues,
                context.getImageFromUri, context.targetCollector, context,
                page)
            bo.ProcessWhitespace(box, false)
            box = bo.AnonymousTableBoxes(box)
            box = bo.FlexBoxes(box)
            box = bo.IlineInBlock(box)
			box = bo.BlockInInline(box)
			return box.(*bo.MarginBox)
		} 
		resolvePercentages(box, containingBlock, "")
		boxF := box.Box()
        if !BoxF.IsGenerated {
			BoxF.Width = pr.Float(0)
			BoxF.Height = pr.Float(0)
            for side := range [4]string{"top", "right", "bottom", "left"} {
                BoxF.ResetSpacing(side)
            }
		} 
		return box
    }

    marginTop := page.MarginTop.V()
    marginBottom := page.MarginBottom
    marginLeft := page.MarginLeft.V()
    marginRight := page.MarginRight
    maxBoxWidth := page.BorderWidth()
    maxBoxHeight := page.BorderHeight()

    // bottom right corner of the border box
    pageEndX := marginLeft + maxBoxWidth
    pageEndY := marginTop + maxBoxHeight

    // Margin box dimensions, described in
    // http://dev.w3.org/csswg/css3-page/#margin-box-dimensions
    var generatedBoxes []*bo.MarginBox
	prefixs := [4]string{"top","bottom","left","right"}
	verticals := [4]bool{false,		false,		true,		true}
	containingBlocks := [4]pr.MaybePoint{
		{maxBoxWidth, marginTop},
			{maxBoxWidth, marginBottom},
			{marginLeft, maxBoxHeight},
			{marginRight, maxBoxHeight},
	}
	positionXs := [4]pr.Float{  marginLeft,		marginLeft,		 0,		pageEndX}
	positionYs := [4]pr.Float{0,		pageEndY,		marginTop,		marginTop}
	
    for i := range prefixes {
		prefix, vertical, containingBlock, positionX, positionY := prefixs[i], verticals[i], containingBlocks[i], positionXs[i], positionYs[i]
		
		suffixes := [3]string{"left", "center", "right"}
		variableOuter, fixedOuter := containingBlock[0], containingBlock[1]
		if vertical {
            suffixes = [3]string{"top", "middle", "bottom"}
            fixedOuter, variableOuter = containingBlock[0], containingBlock[1]
        } 
		var sideBoxes [3]*bo.MarginBox
		anyIsGenerated := false
		for i, suffix := range suffixes {
			sideBoxes[i] = makeBox(fmt.Sprintf("@%s-%s",prefix, suffix), containingBlock)
			if sideBoxes[i].IsGenerated {
				anyIsGenerated =true
			}
		}
	
        if ! anyIsGenerated {
            continue
		} 
		// We need the three boxes together for the variable dimension:
		computeVariableDimension(context, sideBoxes, vertical, variableOuter.V())
		offsets := []pr.Float{0, 0.5, 1}
        for i := range sideBoxes {
			box, offset := sideBoxes[i], offsets[i]
            if ! box.IsGenerated {
                continue
			} 
			box.PositionX = positionX
            box.PositionY = positionY
            if vertical {
                box.PositionY += offset * (variableOuter - box.MarginHeight())
            } else {
                box.PositionX += offset * (  - box.MarginWidth())
			} 
			computeFixedDimension( context, box, fixedOuter, ! vertical, prefix == "top" || prefix == "left")
            generatedBoxes = append(generatedBoxes, box)
        }
    }

	atKeywords := [4]string{"@top-left-corner","@top-right-corner","@bottom-left-corner","@bottom-right-corner"}
	cbWidths := [4]pr.MaybeFloat{ marginLeft,		marginRight,		marginLeft,		marginRight	}
	cbHeights := [4]pr.MaybeFloat{ marginTop,marginTop,marginBottom,marginBottom	}
	positionXs = [4]pr.Float{0,pageEndX,		0,		pageEndX}
	positionYs = [4]pr.Float{0,0,		pageEndY,		pageEndY}
    // Corner boxes
    for i := range atKeywords {
		atKeyword, cbWidth, cbHeight, positionX, positionY := atKeywords[i], cbWidths[i], cbHeights[i], positionXs[i], positionYs[i]
        box := makeBox(atKeyword, bo.MaybePoint{cbWidth, cbHeight})
        if ! box.IsGenerated {
            continue
		} 
		box.PositionX = positionX
        box.PositionY = positionY
        computeFixedDimension( context, box, cbHeight, true,  strings.Contains(atKeyword,"top") )
        computeFixedDimension( context, box, cbWidth, false,  strings.Contains(atKeyword,"left") )
        generatedBoxes = append(generatedBoxes, box)
    }

	out := make([]Box, len(generatedBoxes))
    for i, box := range generatedBoxes {
        out[i] = marginBoxContentLayout(context, page, box)
	}
	return out
}


// Layout a margin box’s content once the box has dimensions.
func marginBoxContentLayout(context LayoutContext, page *bo.PageBox, Mbox *bo.MarginBox) Box {
	tmp := blockContainerLayout(context, Mbox,  pr.Inf, nil, true, 
		new([]*AbsolutePlaceholder),new([]*AbsolutePlaceholder),nil)
		
    if tmp.resumeAt  != nil {
		log.Fatalf("resumeAt should be nil, got %v", tmp.resumeAt)
	}	
	box := tmp.newBox.Box()
    verticalAlign := box.Style.GetVerticalAlign()
    // Every other value is read as "top", ie. no change.
    if L := len(box.Children); (verticalAlign.String == "middle" || verticalAlign.String == "bottom") && L != 0 {
        firstChild := box.Children[0]
        lastChild := box.Children[L-1].Box()
        top := firstChild.Box().PositionY
        // Not always exact because floating point errors
        // assert top == box.contentBoxY()
        bottom := lastChild.PositionY + lastChild.MarginHeight()
        contentHeight := bottom - top
        offset := box.Height.V() - contentHeight
        if verticalAlign.String == "middle" {
            offset /= 2
		} 
		for _, child := range box.Children {
            child.Translate(child, 0, offset, false)
        }
	}
	return tmp.newBox
}


// Take a :class:`OrientedBox` object and set either width, margin-left
// and margin-right; or height, margin-top and margin-bottom.
// "The width and horizontal margins of the page box are then calculated
//  exactly as for a non-replaced block element := range normal flow. The height
//  and vertical margins of the page box are calculated analogously (instead
//  of using the block height formulas). In both cases if the values are
//  over-constrained, instead of ignoring any margins, the containing block
//  is resized to coincide with the margin edges of the page box."
// http://dev.w3.org/csswg/css3-page/#page-box-page-rule
// http://www.w3.org/TR/CSS21/visudet.html#blockwidth
func pageWidthOrHeight(box_ orientedBox, containingBlockSize pr.Float) {
	box := box_.baseBox()
	remaining := containingBlockSize - box.paddingPlusBorder
    if box.inner.Auto() {
        if box.marginA.Auto() {
            box.marginA = pr.Float(0)
		} 
		if box.marginB.Auto() {
            box.marginB = pr.Float(0)
		}
		 box.inner = remaining - box.marginA.V()- box.marginB.V()
    } else if box.marginA.Auto() && box.marginB.Auto() {
		box.marginA = (remaining - box.inner.Auto()) / 2
		box.marginB = box.marginA
    } else if box.marginA.Auto() {
        box.marginA = remaining - box.inner.V() - box.marginB.V()
    } else if box.marginB.Auto() {
        box.marginB = remaining - box.inner.V() - box.marginA.V()
	}
	box_.restoreBoxAttributes()
} 

var (
	pageWidth = handleMinMaxWidth(pageWidth_)
	pageHeight = handleMinMaxHeight(pageHeight_)
)

// @handleMinMaxWidth
func pageWidth_(box Box, context LayoutContext, containingBlockWidth block) (bool, float32) {
	pageWidthOrHeight(NewHorizontalBox(context, box.(*bo.MarginBox)), containingBlockWidth)
	return false, 0
	} 
	
	// @handleMinMaxHeight
	func pageHeight_(box Box, context LayoutContext, containingBlockHeight block) (bool, float32) {
		pageWidthOrHeight(NewVerticalBox(context, box.(*bo.MarginBox)), containingBlockHeight)
		return false, 0
} 

// Take just enough content from the beginning to fill one page.
//
// Return ``(page, finished)``. ``page`` is a laid out PageBox object
// and ``resumeAt`` indicates where in the document to start the next page,
// or is ``None`` if this was the last page.
//
// :param pageNumber: integer, start at 1 for the first page
// :param resumeAt: as returned by ``makePage()`` for the previous page,
// 	or ``None`` for the first page.
func makePage(context *LayoutContext, rootBox Box, pageType utils.PageElement , resumeAt *tree.SkipStack,
	pageNumber int, pageState *bo.StateShared) (*bo.PageBox ,*tree.SkipStack, page) {
    style := context.styleFor.Get(pageType, "")

    // Propagated from the root or <body>.
    style.SetOverflow(pr.String(rootBox.Box().viewportOverflow)) 
    page := &bo.NewPageBox(pageType, style)

    deviceSize_ := page.Style.GetSize()
    cbWidth, cbHeight = deviceSize_[0], deviceSize_[1] 
    resolvePercentages(page, bo.MaybePoint{cbWidth, cbHeight}, "")

    page.PositionX = 0
    page.PositionY = 0
    pageWidth(page, context, cbWidth)
    pageHeight(page, context, cbHeight)

    rootBox.Box().PositionX = page.ContentBoxX()
    rootBox.Box().PositionY = page.ContentBoxY()
    pageContentBottom := rootBox.Box().PositionY + page.Height.V()
    initialContainingBlock := page

	var previousResumeAt *tree.SkipStack
    if pageType.blank {
        previousResumeAt = resumeAt
        rootBox = bo.CopyWithChildren(rootBox, nil, true, true)
    }

    // TODO: handle cases where the root element is something else.
    // See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
    if !(bo.TypeBlockBox.IsInstance(rootBox) || bo.IsFlexContainerBox(rootBox)) {
		log.Fatalf("expected Block or FlexContainer, got %s", rootBox)
	}
    context.createBlockFormattingContext()
    context.currentPage = pageNumber
    pageIsEmpty := true
    var adjoiningMargins []pr.Float
    var positionedBoxes []*AbsolutePlaceholder  // Mixed absolute && fixed
    tmp := blockLevelLayout(context, rootBox, pageContentBottom, resumeAt,
		initialContainingBlock, pageIsEmpty, &positionedBoxes, &positionedBoxes, adjoiningMargins)
		rootBox, resumeAt = tmp.newBox, tmp.resumeAt
    if rootBox == nil {
		log.Fatalln("expected newBox got nil")
	}

    for _, placeholder := range positionedBoxes {
        if placeholder.Box.Box().Style.GetPosition() == "fixed" {
			page.FixedBoxes  = append(page.FixedBoxes , placeholder.Box)
		}
	}
    for _, absoluteBox := range positionedBoxes {
        absoluteLayout(context, absoluteBox, page, positionedBoxes)
	} 
	context.finishBlockFormattingContext(rootBox)

    page.Children = []Box{rootBox}
    descendants := page.Descendants()

    // Update page counter values
    standardizePageBasedCounters(style, "")
    bo.UpdateCounters(pageState, style)
    pageCounterValues := pageState.CounterValues
    // pageCounterValues will be cached in the pageMaker

    targetCollector := context.targetCollector
    pageMaker := context.pageMaker

    // remakeState tells the makeAllPages-loop in layoutDocument()
	// whether and what to re-make.
    remakeState := pageMaker[pageNumber - 1].remakeState

    // Evaluate and cache page values only once (for the first LineBox)
    // otherwise we suffer endless loops when the target/pseudo-element
    // spans across multiple pages
	cachedAnchors := pr.NewSet()
    cachedLookups := map[*tree.CounterLookupItem]bool{}
	
    for _ , v := range pageMaker[:pageNumber - 1] {
		cachedAnchors.Extend(v.RemakeState.Anchors)
		for _, u := range v.RemakeState.ContentLookups {
			cachedLookups[u] = true
		}
    }

    for _,  child := range descendants {
        // Cache target's page counters
        anchor := child.Box().Style.GetAnchor()
        if anchor != "" && !cachedAnchors.Has(anchor) {
            remakeState.Anchors = append(remakeState.Anchors, anchor)
            cachedAnchors.Add(anchor)
            // Re-make of affected targeting boxes is inclusive
            targetCollector.CacheTargetPageCounters(anchor, pageCounterValues, pageNumber - 1, pageMaker)
        }
    
        // string-set and bookmark-labels don't create boxes, only `content`
        // requires another call to makePage. There is maximum one "content"
        // item per box.
        // TODO: remove attribute or set a default value in Box class
		var counterLookup *tree.CounterLookupItem
		if missingLink := child.MissingLink(); missingLink != nil {
			// A CounterLookupItem exists for the css-token "content"
			key := tree.NewFunctionKey(missingLink, "content")
            counterLookup = targetCollector.CounterLookupItems[key]
        } 

        // Resolve missing (page based) counters
        if counterLookup  != nil  {
            callParseAgain := false
    
            // Prevent endless loops
            refreshMissingCounters := ! cachedLookups[counterLookup]
            if refreshMissingCounters {
                remakeState.ContentLookups = append(remakeState.ContentLookups, counterLookup)
                cachedLookups = append(cachedLookups, counterLookup)
                counterLookup.pageMakerIndex = pageNumber - 1
            }

            // Step 1: page based back-references
            // Marked as pending by targetCollector.cacheTargetPageCounters
            if counterLookup.Pending {
                if ! pageCounterValues.Equal(counterLookup.CachedPageCounterValues) {
					counterLookup.CachedPageCounterValues =  pageCounterValues.Copy()
				}
                counterLookup.Pending = false
                callParseAgain = true
            }

            // Step 2: local counters
            // If the box mixed-in page counters changed, update the content
            // && cache the new values.
            missingCounters := counterLookup.MissingCounters
            if len(missingCounters) != 0 {
                if missingCounters.Has("pages") {
                    remakeState.PagesWanted = true
				} 
				if refreshMissingCounters && !pageCounterValues.Equal(counterLookup.CachedPageCounterValues) {
                    counterLookup.CachedPageCounterValues = pageCounterValues.Copy()
                    for counterName := range missingCounters {
                        counterValue := pageCounterValues[counterName]
                        if counterValue  != nil  {
                            callParseAgain = true
                            // no need to loop them all
                            break
                        }
					}
				}
            }

            // Step 3: targeted counters
            targetMissing := counterLookup.MissingTargetCounters
            for anchorName, missedCounters := range targetMissing {
                if !missedCounters.Has("pages") {
                    continue
				} 
				// Adjust "pagesWanted"
                item := targetCollector.TargetLookupItems[anchorName]
                pageMakerIndex := item.PageMakerIndex
                if pageMakerIndex >= 0 && cachedAnchors.Has(anchorName) {
                    pageMaker[pageMakerIndex].RemakeState.PagesWanted = true
				} 
				// "contentChanged" is triggered in
                // targets.cacheTargetPageCounters()
            }

            if callParseAgain {
                remakeState.ContentChanged = true
                counterLookup.ParseAgain(pageCounterValues)
			}
		}
	}

    if pageType.blank {
        resumeAt = previousResumeAt
    }

    return page, resumeAt, tmp.nextPage
	}

// Set style for page types and pseudo-types matching ``pageType``.
func setPageTypeComputedStyles(pageType utils.PageElement, html tree.Html, styleFor tree.StyleFor) {
    styleFor.AddPageDeclarations(pageType)

    // Apply style for page
    styleFor.SetComputedStyles( pageType,
        // @page inherits from the root element :
        // http://lists.w3.org/Archives/Public/www-style/2012Jan/1164.html
        html.Root, html.Root, "", html.BaseUrl, nil)

    // Apply style for page pseudo-elements (margin boxes)
    for key := range styleFor.CascadedStyles {
        if key.PseudoType != "" && key.PageType == pageType {
            styleFor.SetComputedStyles(
                key.PageType, key.PageType,
                // The pseudo-element inherits from the element.
                html.Root, pseudoType, html.BaseUrl, nil)
        }
    }
}

// Return one laid out page without margin boxes.
// Start with the initial values from ``context.pageMaker[index]``.
// The resulting values / initial values for the next page are stored in
// the ``pageMaker``.
// As the function"s name suggests: the plan is ! to make all pages
// repeatedly when a missing counter was resolved, but rather re-make the
// single page where the ``contentChanged`` happened.
func remakePage(index int, context LayoutContext, rootBox Box, html tree.Html) {
    pageMaker := context.pageMaker
    tmp := pageMaker[index]
	// initialResumeAt, initialNextPage, rightPage, initialPageState,remakeState := tmp

    // PageType for current page, values for pageMaker[index + 1].
    // Don"t modify actual pageMaker[index] values!
    // TODO: should we store (and reuse) pageType := range the pageMaker?
    pageState = copy.deepcopy(initialPageState)
    nextPageName = initialNextPage["page"]
    first = index == 0
    if initialNextPage["break"] := range ("left", "right") {
        nextPageSide = initialNextPage["break"]
    } else if initialNextPage["break"] := range ("recto", "verso") {
        directionLtr = rootBox.style["direction"] == "ltr"
        breakVerso = initialNextPage["break"] == "verso"
        nextPageSide = "right" if directionLtr ^ breakVerso else "left"
    } else {
        nextPageSide = None
	}
	 blank = ((nextPageSide == "left" && rightPage) or
             (nextPageSide == "right" && ! rightPage))
    if blank {
        nextPageName = ""
	} 
	side = "right" if rightPage else "left"
    pageType = PageType(side, blank, first, index, name=nextPageName)
    setPageTypeComputedStyles(pageType, html, context.styleFor)

    context.forcedBreak = (
        initialNextPage["break"] != "any" || initialNextPage["page"])
    context.marginClearance = false

    // makePage wants a pageNumber of index + 1
    pageNumber = index + 1
    page, resumeAt, nextPage = makePage(
        context, rootBox, pageType, initialResumeAt,
        pageNumber, pageState)
    assert nextPage
    if blank {
        nextPage["page"] = initialNextPage["page"]
	} 
	rightPage = ! rightPage

    // Check whether we need to append || update the next pageMaker item
    if index + 1 >= len(pageMaker) {
        // New page
        pageMakerNextChanged = true
    } else {
        // Check whether something changed
        // TODO: Find what we need to compare. Is resumeAt enough?
        (nextResumeAt, nextNextPage, nextRightPage,
         nextPageState, ) = pageMaker[index + 1]
        pageMakerNextChanged = (
            nextResumeAt != resumeAt or
            nextNextPage != nextPage or
            nextRightPage != rightPage or
            nextPageState != pageState)
    }

    if pageMakerNextChanged {
        // Reset remakeState
        remakeState = {
            "contentChanged": false,
            "pagesWanted": false,
            "anchors": [],
            "contentLookups": [],
        }
        // Setting contentChanged to true ensures remake.
        // If resumeAt  == nil  (last page) it must be false to prevent endless
        // loops && list index out of range (see #794).
        remakeState["contentChanged"] = resumeAt  != nil 
        // pageState is already a deepcopy
        item = resumeAt, nextPage, rightPage, pageState, remakeState
        if index + 1 >= len(pageMaker) {
            pageMaker.append(item)
        } else {
            pageMaker[index + 1] = item
        }
    }

	return page, resumeAt
}


// Return a list of laid out pages without margin boxes.
// Re-make pages only if necessary.
func makeAllPages(context LayoutContext, rootBox Box, html tree.Html, pages []*bo.PageBox) []*bo.PageBox {
	var out []*bo.PageBox
	i := 0
    for {
		remakeState := context.pageMaker[i].RemakeState
		var resumeAt *tree.SkipStack
        if (len(pages) == 0 || remakeState.ContentChanged || remakeState.PagesWanted) {
            logger.ProgressLogger.Printf("Step 5 - Creating layout - Page %d", i + 1)
            // Reset remakeState
            context.pageMaker[i].RemakeState = tree.RemakeState{}
            page, resumeAt = remakePage(i, context, rootBox, html)
            out = append(out, page)
        } else {
            logger.ProgressLogger.Printf("Step 5 - Creating layout - Page %d (up-to-date)", i + 1)
            resumeAt = context.pageMaker[i + 1].NextResumeAt
            out = append(out, pages[i])
        }

        i += 1
        if resumeAt  == nil  {
            // Throw away obsolete pages
            context.pageMaker = context.pageMaker[:i + 1]
			return out
		}
	}
}
