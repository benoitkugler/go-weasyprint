// Transform a "before layout" box tree into an "after layout" tree.
// (Surprising, hu?)

// Break boxes across lines and pages; determine the size and dimension
// of each box fragement.

// Boxes in the new tree have *used values* in their ``position_x``,
// ``position_y``, ``width`` and ``height`` attributes, amongst others.

// See http://www.w3.org/TR/CSS21/cascade.html#used-value

// :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.
package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/boxes/counters"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/layout/text/hyphen"
	"github.com/benoitkugler/go-weasyprint/logger"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/textlayout/pango"
)

type Box = bo.Box

// Initialize ``context.pageMaker``.
// Collect the pagination's states required for page based counters.
func initializePageMaker(context *LayoutContext, rootBox bo.BoxFields) {
	context.pageMaker = nil

	// Special case the root box
	pageBreak := rootBox.Style.GetBreakBefore()

	// TODO: take care of text direction and writing mode
	// https://www.w3.org/TR/css3-page/#progression
	var rightPage bool
	switch pageBreak {
	case "right":
		rightPage = true
	case "left":
		rightPage = false
	case "recto":
		rightPage = rootBox.Style.GetDirection() == "ltr"
	case "verso":
		rightPage = rootBox.Style.GetDirection() == "rtl"
	default:
		rightPage = rootBox.Style.GetDirection() == "ltr"
	}
	pv, _ := rootBox.PageValues()
	nextPage := tree.PageBreak{Break: "any", Page: pv}

	// pageState is prerequisite for filling in missing page based counters
	// although neither a variable quoteDepth nor counterScopes are needed
	// in page-boxes -- reusing
	// `formattingStructure.build.updateCounters()` to avoid redundant
	// code requires a full `state`.
	// The value of **pages**, of course, is unknown until we return and
	// might change when "contentChanged" triggers re-pagination...
	// So we start with an empty state
	pageState := tree.PageState{
		// Shared mutable objects:
		QuoteDepth:    []int{0}, // quoteDepth: single integer
		CounterValues: tree.CounterValues{"pages": []int{0}},
		CounterScopes: []utils.Set{utils.NewSet("pages")}, // counterScopes
	}

	// Initial values
	remakeState := tree.RemakeState{}
	context.pageMaker = append(context.pageMaker, tree.PageMaker{
		InitialResumeAt: nil, InitialNextPage: nextPage, RightPage: rightPage,
		InitialPageState: pageState, RemakeState: remakeState,
	})
}

// Lay out and yield the fixed boxes of ``pages``.
func layoutFixedBoxes(context *LayoutContext, pages []*bo.PageBox, containingPage *bo.PageBox) []Box {
	var out []Box
	for _, page := range pages {
		for _, box := range page.FixedBoxes {
			// As replaced boxes are never copied during layout, ensure that we
			// have different boxes (with a possibly different layout) for
			// each pages.
			if bo.ReplacedBoxT.IsInstance(box) {
				box = box.Copy()
			}
			// Absolute boxes in fixed boxes are rendered as fixed boxes'
			// children, even when they are fixed themselves.
			var absoluteBoxes []*AbsolutePlaceholder
			out = append(out, absoluteBoxLayout(context, box, containingPage, &absoluteBoxes))
			for len(absoluteBoxes) != 0 {
				var newAbsoluteBoxes []*AbsolutePlaceholder
				for _, box := range absoluteBoxes {
					absoluteLayout(context, box, containingPage, &newAbsoluteBoxes)
				}
				absoluteBoxes = newAbsoluteBoxes
			}
		}
	}
	return out
}

// Lay out the whole document (maxLoops=-1).
// This includes line breaks, page breaks, absolute size and position for all
// boxes. Page based counters might require multiple passes.
// :param rootBox: root of the box tree (formatting structure of the html)
//                  the pages' boxes are created from that tree, i.e. this
//                  structure is not lost during pagination
// :returns: a list of laid out Page objects.
func LayoutDocument(html *tree.HTML, rootBox bo.BlockLevelBoxITF, context *LayoutContext, maxLoops int) []*bo.PageBox {
	initializePageMaker(context, *rootBox.Box())
	if maxLoops == -1 {
		maxLoops = 8 // default value
	}
	var pages []*bo.PageBox
	actualTotalPages := 0

	for loop := 0; loop < maxLoops; loop += 1 {
		if loop > 0 {
			logger.ProgressLogger.Printf("Step 5 - Creating layout - Repagination #%d \n", loop)
		}

		initialTotalPages := actualTotalPages
		pages = makeAllPages(context, rootBox, html, pages)
		actualTotalPages = len(pages)

		// Check whether another round is required
		reloopContent := false
		reloopPages := false
		for _, pageData := range context.pageMaker {
			// Update pages
			pageState, remakeState := pageData.InitialPageState, pageData.RemakeState
			pageCounterValues := pageState.CounterValues
			pageCounterValues["pages"] = []int{actualTotalPages}
			if remakeState.ContentChanged {
				reloopContent = true
			}
			if remakeState.PagesWanted {
				reloopPages = initialTotalPages != actualTotalPages
			}
		}

		// No need for another loop, stop here
		if !reloopContent && !reloopPages {
			break
		}
	}

	// Calculate string-sets and bookmark-label containing page based counters
	// when pagination is finished. No need to do that (maybe multiple times) in
	// makePage because they dont create boxes, only appear in MarginBoxes and
	// in the final PDF.
	for i, page := range pages {
		// We need the updated pageCounterValues
		pageCounterValues := context.pageMaker[i+1].InitialPageState.CounterValues

		for _, child := range bo.Descendants(page) {
			// TODO: remove attribute or set a default value in Box class
			if mLink := child.MissingLink(); mLink != nil {
				for key, item := range context.TargetCollector.CounterLookupItems {
					box, cssToken := key.SourceBox, key.CssToken
					if mLink == box && cssToken != "content" {
						item.ParseAgain(pageCounterValues)
					}
				}
			}
			// Collect the stringSets in the LayoutContext
			stringSets := child.Box().StringSet
			for _, stringSet := range stringSets {
				stringName, text := stringSet.Type, string(stringSet.Content.(pr.String))
				dict := context.stringSet[stringName]
				if dict == nil {
					dict = make(map[int][]string)
				}
				dict[i+1] = append(dict[i+1], text)
				context.stringSet[stringName] = dict
			}
		}
	}
	out := make([]*bo.PageBox, len(pages))
	// Add margin boxes
	for i, page := range pages {
		var rootChildren []Box
		root := page.Box().Children[0]
		rootChildren = append(rootChildren, layoutFixedBoxes(context, pages[:i], page)...)
		rootChildren = append(rootChildren, root.Box().Children...)
		rootChildren = append(rootChildren, layoutFixedBoxes(context, pages[i+1:], page)...)
		root.Box().Children = rootChildren
		context.currentPage = i + 1 // pageNumber starts at 1

		// pageMaker's pageState is ready for the MarginBoxes
		state := context.pageMaker[context.currentPage].InitialPageState
		page.Children = append([]Box{root}, makeMarginBoxes(context, page, state)...)
		layoutBackgrounds(page, context.GetImageFromUri)
		out[i] = page
	}
	return out
}

var _ text.TextLayoutContext = (*LayoutContext)(nil)

// LayoutContext stores the global context needed during layout,
// such as various caches.
type LayoutContext struct {
	// caches
	stringSet    map[string]map[int][]string
	strutLayouts map[text.StrutLayoutKey][2]pr.Float
	tables       map[*bo.TableBox]map[bool]tableContentWidths

	runningElements     map[string]map[int]Box
	GetImageFromUri     bo.Gifu
	fontConfig          *text.FontConfiguration
	TargetCollector     *tree.TargetCollector
	counterStyle        counters.CounterStyle
	dictionaries        map[text.HyphenDictKey]hyphen.Hyphener
	StyleFor            *tree.StyleFor
	pageMaker           []tree.PageMaker
	excludedShapes      []bo.BoxFields
	excludedShapesLists [][]bo.BoxFields
	currentPage         int
	marginClearance     bool
	forcedBreak         bool
}

func NewLayoutContext(getImageFromUri bo.Gifu,
	fontConfig *text.FontConfiguration, counterStyle counters.CounterStyle, targetCollector *tree.TargetCollector) *LayoutContext {
	self := LayoutContext{}
	self.GetImageFromUri = getImageFromUri
	self.fontConfig = fontConfig
	self.TargetCollector = targetCollector
	self.counterStyle = counterStyle
	self.runningElements = map[string]map[int]Box{}

	// Cache
	self.stringSet = make(map[string]map[int][]string)
	self.dictionaries = make(map[text.HyphenDictKey]hyphen.Hyphener)
	self.strutLayouts = make(map[text.StrutLayoutKey][2]pr.Float)
	self.tables = map[*bo.TableBox]map[bool]tableContentWidths{}
	return &self
}

func (self LayoutContext) RunningElements() map[string]map[int]Box { return self.runningElements }

func (self LayoutContext) CurrentPage() int { return self.currentPage }

func (self *LayoutContext) Fontmap() pango.FontMap { return self.fontConfig.Fontmap }

func (self *LayoutContext) HyphenCache() map[text.HyphenDictKey]hyphen.Hyphener {
	return self.dictionaries
}

func (self *LayoutContext) StrutLayoutsCache() map[text.StrutLayoutKey][2]pr.Float {
	return self.strutLayouts
}

func (self *LayoutContext) createBlockFormattingContext() {
	self.excludedShapes = nil
	self.excludedShapesLists = append(self.excludedShapesLists, self.excludedShapes)
}

func (self *LayoutContext) finishBlockFormattingContext(rootBox_ Box) {
	// See http://www.w3.org/TR/CSS2/visudet.html#root-height
	rootBox := rootBox_.Box()
	if rootBox.Style.GetHeight().String == "auto" && len(self.excludedShapes) != 0 {
		boxBottom := rootBox.ContentBoxY() + rootBox.Height.V()
		maxShapeBottom := boxBottom
		for _, shape := range self.excludedShapes {
			v := shape.PositionY + shape.MarginHeight()
			if v > maxShapeBottom {
				maxShapeBottom = v
			}
		}
		rootBox.Height = rootBox.Height.V() + maxShapeBottom - boxBottom
	}
	self.excludedShapesLists = self.excludedShapesLists[:len(self.excludedShapesLists)-1]
	if L := len(self.excludedShapesLists); L != 0 {
		self.excludedShapes = self.excludedShapesLists[L-1]
	} else {
		self.excludedShapes = nil
	}
}

// Resolve value of string function (as set by string set).
// We"ll have something like this that represents all assignments on a
// given page:
//
// {1: [u"First Header"], 3: [u"Second Header"],
//  4: [u"Third Header", u"3.5th Header"]}
//
// Value depends on current page.
// http://dev.w3.org/csswg/css-gcpm/#funcdef-string
//
// `keyword` indicates which value of the named string to use.
// 	Default is the first assignment on the current page
//  else the most recent assignment (entry value)
// keyword="first"
func (self LayoutContext) GetStringSetFor(page Box, name, keyword string) string {
	if currentS, in := self.stringSet[name][self.currentPage]; in {
		// A value was assigned on this page
		firstString := currentS[0]
		lastString := currentS[len(currentS)-1]
		switch keyword {
		case "first":
			return firstString
		case "start":
			element := page
			for element != nil {
				if element.Box().Style.GetStringSet().String != "none" {
					for _, v := range element.Box().Style.GetStringSet().Contents {
						if v.String == name {
							return firstString
						}
					}
				}
				if bo.ParentBoxT.IsInstance(element) {
					if len(element.Box().Children) > 0 {
						element = element.Box().Children[0]
						continue
					}
				}
				break
			}
		case "last":
			return lastString
		}
	}
	// Search backwards through previous pages
	for previousPage := self.currentPage - 1; previousPage > 0; previousPage -= 1 {
		if currentS, in := self.stringSet[name][previousPage]; in {
			return currentS[len(currentS)-1]
		}
	}
	return ""
}
