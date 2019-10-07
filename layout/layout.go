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
	"github.com/benoitkugler/go-weasyprint/boxes"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/fonts"
	"github.com/benoitkugler/go-weasyprint/style/tree"
)

// Lay out and yield the fixed boxes of ``pages``.
func layoutFixedBoxes(context LayoutContext, pages []Page) {
	var out []boxes.Box
	for _, page := range pages {
		for _, box := range page.fixedBoxes {
			// Use an empty list as last argument because the fixed boxes in the
			// fixed box has already been added to page.fixedBoxes, we don"t
			// want to get them again
			out = append(out, absoluteBoxLayout(context, box, page, nil))
		}
	}
}

type Page struct {
	fixedBoxes []boxes.Box
}

type LayoutContext struct {
	enableHinting       bool
	styleFor            tree.StyleFor
	getImageFromUri     boxes.Gifu
	fontConfig          *fonts.FontConfiguration
	targetCollector     *tree.TargetCollector
	excludedShapes      []shape
	excludedShapesLists [][]shape
	stringSet           map[string]map[string][]string
	runningElements     map[string]int
	currentPage         int
	forcedBreak         bool
	strutLayouts        map[string]int
	fontFeatures        map[string]int
	tables              map[*bo.BoxFields]map[bool]tableContentWidths
	dictionaries        map[string]int
}

type shape struct {
	positionY float32
}

func (s shape) marginHeight() float32 {}

func NewLayoutContext(enableHinting bool, styleFor tree.StyleFor, getImageFromUri boxes.Gifu,
	fontConfig *fonts.FontConfiguration, targetCollector *tree.TargetCollector) *LayoutContext {
	self := LayoutContext{}
	self.enableHinting = enableHinting
	self.styleFor = styleFor
	self.getImageFromUri = getImageFromUri
	self.fontConfig = fontConfig
	self.targetCollector = targetCollector
	self.runningElements = map[string]int{}
	// Cache
	self.strutLayouts = map[string]int{}
	self.fontFeatures = map[string]int{}
	self.tables = map[string]int{}
	self.dictionaries = map[string]int{}
	return &self
}

func (self *LayoutContext) createBlockFormattingContext() {
	self.excludedShapes = nil
	self.excludedShapesLists = append(self.excludedShapesLists, self.excludedShapes)
}

func (self *LayoutContext) finishBlockFormattingContext(rootBox_ Box) {
	// See http://www.w3.org/TR/CSS2/visudet.html#root-height
	rootBox := rootBox_.Box()
	if rootBox.style.GetHeight() == "auto" && self.excludedShapes {
		boxBottom = rootBox.contentBoxY() + rootBox.height
		maxShapeBottom := boxBottom
		for _, shape := range self.excludedShapes {
			v := shape.positionY + shape.marginHeight()
			if v > maxShapeBottom {
				maxShapeBottom = v
			}
		}
		rootBox.height += maxShapeBottom - boxBottom
	}
	self.ExcludedShapesLists.pop()
	if self.ExcludedShapesLists {
		self.excludedShapes = self.ExcludedShapesLists[-1]
	} else {
		self.excludedShapes = None
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
func (self LayoutContext) getStringSetFor(page boxes.Box, name, keyword string) string {
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
				if boxes.IsParentBox(element) {
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
