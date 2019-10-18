package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
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
//     ``state`` is the actual, up-to-date page-state from
//     ``context.pageMaker[context.currentPage]``.
func makeMarginBoxes(context LayoutContext, page *bo.PageBox, state bo.StateShared) {
    // This is a closure only to make calls shorter
	
	// Return a margin box with resolved percentages.
	// The margin box may still have "auto" values.
	// Return ``None`` if this margin box should not be generated.
	// :param atKeyword: which margin box to return, eg. "@top-left"
	// :param containingBlock: as expected by :func:`resolvePercentages`.
    makeBox := func(atKeyword string, containingBlock block) Box {
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
            marginState.counterScopes.append(set())
            build.updateCounters(marginState, box.style)
            box.children = build.contentToBoxes(
                box.style, box, quoteDepth, counterValues,
                context.getImageFromUri, context.targetCollector, context,
                page)
            build.processWhitespace(box)
            box = build.anonymousTableBoxes(box)
            box = build.flexBoxes(box)
            box = build.inlineInBlock(box)
            box = build.blockInInline(box)
		} 
		resolvePercentages(box, containingBlock)
        if ! box.isGenerated {
            box.Width = box.Height = 0
            for side := range ("top", "right", "bottom", "left") {
                box.ResetSpacing(side)
            }
        } return box
    }
} 
    marginTop = page.marginTop
    marginBottom = page.marginBottom
    marginLeft = page.marginLeft
    marginRight = page.marginRight
    maxBoxWidth = page.borderWidth()
    maxBoxHeight = page.borderHeight()

    // bottom right corner of the border box
    pageEndX = marginLeft + maxBoxWidth
    pageEndY = marginTop + maxBoxHeight

    // Margin box dimensions, described in
    // http://dev.w3.org/csswg/css3-page/#margin-box-dimensions
    generatedBoxes = []

    for prefix, vertical, containingBlock, positionX, positionY := range [
        ("top", false, (maxBoxWidth, marginTop),
            marginLeft, 0),
        ("bottom", false, (maxBoxWidth, marginBottom),
            marginLeft, pageEndY),
        ("left", true, (marginLeft, maxBoxHeight),
            0, marginTop),
        ("right", true, (marginRight, maxBoxHeight),
            pageEndX, marginTop),
    ] {
        if vertical {
            suffixes = ["top", "middle", "bottom"]
            fixedOuter, variableOuter = containingBlock
        } else {
            suffixes = ["left", "center", "right"]
            variableOuter, fixedOuter = containingBlock
        } sideBoxes = [makeBox("@%s-%s" % (prefix, suffix), containingBlock)
                      for suffix := range suffixes]
        if ! any(box.isGenerated for box := range sideBoxes) {
            continue
        } // We need the three boxes together for the variable dimension {
        } computeVariableDimension(
            context, sideBoxes, vertical, variableOuter)
        for box, offset := range zip(sideBoxes, [0, 0.5, 1]) {
            if ! box.isGenerated {
                continue
            } box.PositionX = positionX
            box.PositionY = positionY
            if vertical {
                box.PositionY += offset * (
                    variableOuter - box.marginHeight())
            } else {
                box.PositionX += offset * (
                    variableOuter - box.marginWidth())
            } computeFixedDimension(
                context, box, fixedOuter, ! vertical,
                prefix := range ["top", "left"])
            generatedBoxes.append(box)
        }
    }

    // Corner boxes

    for atKeyword, cbWidth, cbHeight, positionX, positionY := range [
        ("@top-left-corner", marginLeft, marginTop, 0, 0),
        ("@top-right-corner", marginRight, marginTop, pageEndX, 0),
        ("@bottom-left-corner", marginLeft, marginBottom, 0, pageEndY),
        ("@bottom-right-corner", marginRight, marginBottom,
            pageEndX, pageEndY),
    ] {
        box = makeBox(atKeyword, (cbWidth, cbHeight))
        if ! box.isGenerated {
            continue
        } box.PositionX = positionX
        box.PositionY = positionY
        computeFixedDimension(
            context, box, cbHeight, true, "top" := range atKeyword)
        computeFixedDimension(
            context, box, cbWidth, false, "left" := range atKeyword)
        generatedBoxes.append(box)
    }

    for box := range generatedBoxes {
        yield marginBoxContentLayout(context, page, box)
    }


// Layout a margin box’s content once the box has dimensions.
func marginBoxContentLayout(context, page, box) {
    box, resumeAt, nextPage, _, _ = blockContainerLayout(
        context, box,
        maxPositionY=float("inf"), skipStack=None,
        pageIsEmpty=true, absoluteBoxes=[], fixedBoxes=[])
    assert resumeAt  == nil 
} 
    verticalAlign = box.style["verticalAlign"]
    // Every other value is read as "top", ie. no change.
    if verticalAlign := range ("middle", "bottom") && box.children {
        firstChild = box.children[0]
        lastChild = box.children[-1]
        top = firstChild.positionY
        // Not always exact because floating point errors
        // assert top == box.contentBoxY()
        bottom = lastChild.positionY + lastChild.marginHeight()
        contentHeight = bottom - top
        offset = box.Height - contentHeight
        if verticalAlign == "middle" {
            offset /= 2
        } for child := range box.children {
            child.translate(0, offset)
        }
    } return box


// Take a :class:`OrientedBox` object && set either width, margin-left
//     && margin-right; || height, margin-top && margin-bottom.
//     "The width && horizontal margins of the page box are then calculated
//      exactly as for a non-replaced block element := range normal flow. The height
//      && vertical margins of the page box are calculated analogously (instead
//      of using the block height formulas). In both cases if the values are
//      over-constrained, instead of ignoring any margins, the containing block
//      is resized to coincide with the margin edges of the page box."
//     http://dev.w3.org/csswg/css3-page/#page-box-page-rule
//     http://www.w3.org/TR/CSS21/visudet.html#blockwidth
//     
func pageWidthOrHeight(box, containingBlockSize) {
    remaining = containingBlockSize - box.PaddingPlusBorder
    if box.inner == "auto" {
        if box.marginA == "auto" {
            box.marginA = 0
        } if box.marginB == "auto" {
            box.marginB = 0
        } box.inner = remaining - box.marginA - box.marginB
    } else if box.marginA == box.marginB == "auto" {
        box.marginA = box.marginB = (remaining - box.inner) / 2
    } else if box.marginA == "auto" {
        box.marginA = remaining - box.inner - box.marginB
    } else if box.marginB == "auto" {
        box.marginB = remaining - box.inner - box.marginA
    } box.restoreBoxAttributes()
} 

@handleMinMaxWidth
func pageWidth(box, context, containingBlockWidth) {
    pageWidthOrHeight(HorizontalBox(context, box), containingBlockWidth)
} 

@handleMinMaxHeight
func pageHeight(box, context, containingBlockHeight) {
    pageWidthOrHeight(VerticalBox(context, box), containingBlockHeight)
} 

func makePage(context, rootBox, pageType, resumeAt, pageNumber,
              pageState) {
    """Take just enough content from the beginning to fill one page.

    Return ``(page, finished)``. ``page`` is a laid out PageBox object
    && ``resumeAt`` indicates where := range the document to start the next page,
    || is ``None`` if this was the last page.

    :param pageNumber: integer, start at 1 for the first page
    :param resumeAt: as returned by ``makePage()`` for the previous page,
                      || ``None`` for the first page.

              }
    """
    style = context.styleFor(pageType)

    // Propagated from the root || <body>.
    style["overflow"] = rootBox.viewportOverflow
    page = boxes.PageBox(pageType, style)

    deviceSize = page.style["size"]

    resolvePercentages(page, deviceSize)

    page.positionX = 0
    page.positionY = 0
    cbWidth, cbHeight = deviceSize
    pageWidth(page, context, cbWidth)
    pageHeight(page, context, cbHeight)

    rootBox.PositionX = page.contentBoxX()
    rootBox.PositionY = page.contentBoxY()
    pageContentBottom = rootBox.PositionY + page.height
    initialContainingBlock = page

    if pageType.blank {
        previousResumeAt = resumeAt
        rootBox = rootBox.copyWithChildren([])
    }

    // TODO: handle cases where the root element is something else.
    // See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
    assert isinstance(rootBox, (boxes.BlockBox, boxes.FlexContainerBox))
    context.createBlockFormattingContext()
    context.currentPage = pageNumber
    pageIsEmpty = true
    adjoiningMargins = []
    positionedBoxes = []  // Mixed absolute && fixed
    rootBox, resumeAt, nextPage, _, _ = blockLevelLayout(
        context, rootBox, pageContentBottom, resumeAt,
        initialContainingBlock, pageIsEmpty, positionedBoxes,
        positionedBoxes, adjoiningMargins)
    assert rootBox

    page.fixedBoxes = [
        placeholder.Box for placeholder := range positionedBoxes
        if placeholder.Box.style["position"] == "fixed"]
    for absoluteBox := range positionedBoxes {
        absoluteLayout(context, absoluteBox, page, positionedBoxes)
    } context.finishBlockFormattingContext(rootBox)

    page.children = [rootBox]
    descendants = page.descendants()

    // Update page counter values
    StandardizePageBasedCounters(style, None)
    build.updateCounters(pageState, style)
    pageCounterValues = pageState[1]
    // pageCounterValues will be cached := range the pageMaker

    targetCollector = context.targetCollector
    pageMaker = context.pageMaker

    // remakeState tells the makeAllPages-loop := range layoutDocument()
    // whether && what to re-make.
    remakeState = pageMaker[pageNumber - 1][-1]

    // Evaluate && cache page values only once (for the first LineBox)
    // otherwise we suffer endless loops when the target/pseudo-element
    // spans across multiple pages
    cachedAnchors = []
    cachedLookups = []
    for (_, _, _, _, xRemakeState) := range pageMaker[:pageNumber - 1] {
        cachedAnchors.extend(xRemakeState.get("anchors", []))
        cachedLookups.extend(xRemakeState.get("contentLookups", []))
    }

    for child := range descendants {
        // Cache target"s page counters
        anchor = child.style["anchor"]
        if anchor && anchor ! := range cachedAnchors {
            remakeState["anchors"].append(anchor)
            cachedAnchors.append(anchor)
            // Re-make of affected targeting boxes is inclusive
            targetCollector.cacheTargetPageCounters(
                anchor, pageCounterValues, pageNumber - 1, pageMaker)
        }
    }

        // string-set && bookmark-labels don"t create boxes, only `content`
        // requires another call to makePage. There is maximum one "content"
        // item per box.
        // TODO: remove attribute || set a default value := range Box class
        if hasattr(child, "missingLink") {
            // A CounterLookupItem exists for the css-token "content"
            counterLookup = targetCollector.counterLookupItems.get(
                (child.missingLink, "content"))
        } else {
            counterLookup = None
        }

        // Resolve missing (page based) counters
        if counterLookup  != nil  {
            callParseAgain = false
        }

            // Prevent endless loops
            counterLookupId = id(counterLookup)
            refreshMissingCounters = counterLookupId ! := range cachedLookups
            if refreshMissingCounters {
                remakeState["contentLookups"].append(counterLookupId)
                cachedLookups.append(counterLookupId)
                counterLookup.pageMakerIndex = pageNumber - 1
            }

            // Step 1: page based back-references
            // Marked as pending by targetCollector.cacheTargetPageCounters
            if counterLookup.pending {
                if (pageCounterValues !=
                        counterLookup.cachedPageCounterValues) {
                        }
                    counterLookup.cachedPageCounterValues = copy.deepcopy(
                        pageCounterValues)
                counterLookup.pending = false
                callParseAgain = true
            }

            // Step 2: local counters
            // If the box mixed-in page counters changed, update the content
            // && cache the new values.
            missingCounters = counterLookup.missingCounters
            if missingCounters {
                if "pages" := range missingCounters {
                    remakeState["pagesWanted"] = true
                } if refreshMissingCounters && pageCounterValues != \
                        counterLookup.cachedPageCounterValues {
                        }
                    counterLookup.cachedPageCounterValues = \
                        copy.deepcopy(pageCounterValues)
                    for counterName := range missingCounters {
                        counterValue = pageCounterValues.get(
                            counterName, None)
                        if counterValue  != nil  {
                            callParseAgain = true
                            // no need to loop them all
                            break
                        }
                    }
            }

            // Step 3: targeted counters
            targetMissing = counterLookup.missingTargetCounters
            for anchorName, missedCounters := range targetMissing.items() {
                if "pages" ! := range missedCounters {
                    continue
                } // Adjust "pagesWanted"
                item = targetCollector.targetLookupItems.get(
                    anchorName, None)
                pageMakerIndex = item.pageMakerIndex
                if pageMakerIndex >= 0 && anchorName := range cachedAnchors {
                    pageMaker[pageMakerIndex][-1]["pagesWanted"] = true
                } // "contentChanged" is triggered in
                // targets.cacheTargetPageCounters()
            }

            if callParseAgain {
                remakeState["contentChanged"] = true
                counterLookup.parseAgain(pageCounterValues)
            }

    if pageType.blank {
        resumeAt = previousResumeAt
    }

    return page, resumeAt, nextPage


// Set style for page types && pseudo-types matching ``pageType``.
func setPageTypeComputedStyles(pageType, html, styleFor) {
    styleFor.addPageDeclarations(pageType)
} 
    // Apply style for page
    styleFor.setComputedStyles(
        pageType,
        // @page inherits from the root element {
        } // http://lists.w3.org/Archives/Public/www-style/2012Jan/1164.html
        root=html.etreeElement, parent=html.etreeElement,
        baseUrl=html.baseUrl)

    // Apply style for page pseudo-elements (margin boxes)
    for element, pseudoType := range styleFor.getCascadedStyles() {
        if pseudoType && element == pageType {
            styleFor.setComputedStyles(
                element, pseudoType=pseudoType,
                // The pseudo-element inherits from the element.
                root=html.etreeElement, parent=element,
                baseUrl=html.baseUrl)
        }
    }


// Return one laid out page without margin boxes.
//     Start with the initial values from ``context.pageMaker[index]``.
//     The resulting values / initial values for the next page are stored in
//     the ``pageMaker``.
//     As the function"s name suggests: the plan is ! to make all pages
//     repeatedly when a missing counter was resolved, but rather re-make the
//     single page where the ``contentChanged`` happened.
//     
func remakePage(index, context, rootBox, html) {
    pageMaker = context.pageMaker
    (initialResumeAt, initialNextPage, rightPage, initialPageState,
     remakeState) = pageMaker[index]
} 
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
    } blank = ((nextPageSide == "left" && rightPage) or
             (nextPageSide == "right" && ! rightPage))
    if blank {
        nextPageName = ""
    } side = "right" if rightPage else "left"
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
    } rightPage = ! rightPage

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


// Return a list of laid out pages without margin boxes.
//     Re-make pages only if necessary.
//     
func makeAllPages(context, rootBox, html, pages) {
    i = 0
    while true {
        remakeState = context.pageMaker[i][-1]
        if (len(pages) == 0 or
                remakeState["contentChanged"] or
                remakeState["pagesWanted"]) {
                }
            PROGRESSLOGGER.info("Step 5 - Creating layout - Page %i", i + 1)
            // Reset remakeState
            remakeState["contentChanged"] = false
            remakeState["pagesWanted"] = false
            remakeState["anchors"] = []
            remakeState["contentLookups"] = []
            page, resumeAt = remakePage(i, context, rootBox, html)
            yield page
        else {
            PROGRESSLOGGER.info(
                "Step 5 - Creating layout - Page %i (up-to-date)", i + 1)
            resumeAt = context.pageMaker[i + 1][0]
            yield pages[i]
        }
    }
} 
        i += 1
        if resumeAt  == nil  {
            // Throw away obsolete pages
            context.pageMaker = context.pageMaker[:i + 1]
            return
