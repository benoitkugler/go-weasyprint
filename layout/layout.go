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
	"github.com/benoitkugler/go-weasyprint/structure"
)

// Lay out and yield the fixed boxes of ``pages``.
func layoutFixedBoxes(context LayoutContext, pages []Page) {
	var out []structure.Box
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
	fixedBoxes []structure.Box
}

type LayoutContext struct{}
