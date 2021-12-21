package svg

import (
	"bytes"
	"strings"

	"github.com/benoitkugler/go-weasyprint/utils"
)

// these attributes are not cascaded
var notInheritedAttributes = utils.NewSet(
	"clip",
	"clip-path",
	"filter",
	"height",
	"id",
	"mask",
	"opacity",
	"overflow",
	"rotate",
	"stop-color",
	"stop-opacity",
	"style",
	"transform",
	"transform-origin",
	"viewBox",
	"width",
	"x",
	"y",
	"dx",
	"dy",
	"href",
)

var (
	replacerPreserve   = strings.NewReplacer("\n", " ", "\r", " ", "\t", " ")
	replacerNoPreserve = strings.NewReplacer("\n", "", "\r", "", "\t", " ")
)

// replace newlines by spaces, and merge spaces if not preserved.
func processWhitespace(text []byte, preserveSpace bool) []byte {
	if preserveSpace {
		return []byte(replacerPreserve.Replace(string(text)))
	}
	return []byte(replacerNoPreserve.Replace(string(text)))
}

// handle text node by fixing whitespaces and flattening tails,
// updating node 'children' and 'text'
func handleText(node *SVGNode, trailingSpace, textRoot bool, defs map[string]SVGNode) bool {
	preserve := node.spacePreserve()
	node.text = processWhitespace(node.text, preserve)
	if trailingSpace && !preserve {
		node.text = bytes.TrimLeft(node.text, " ")
	}

	if len(node.text) != 0 {
		trailingSpace = bytes.HasSuffix(node.text, []byte{' '})
	}

	var newChildren []SVGNode
	for _, child := range node.children {
		if child.tag == "tref" {
			// child_node = Node(child_element, self._style)
			// child_node._etree_node.tag = 'tspan'
			// Retrieve the referenced node and get its flattened text
			// and remove the node children.
			id := parseURLFragment(child.nodeAttributes["xlink:href"])
			node.text = append(node.text, defs[id].text...)
			continue
		}

		trailingSpace = handleText(&child, trailingSpace, false, defs)

		newChildren = append(newChildren, child)
	}

	if textRoot && len(newChildren) == 0 && !preserve {
		node.text = bytes.TrimRight(node.text, " ")
	}

	node.children = newChildren

	return trailingSpace
}

// finalize the parsing by applying the steps
// which require to have seen the whole document
func (svg *SVGImage) postProcessNode(node *SVGNode) {
	if node.tag == "text" || node.tag == "textPath" || node.tag == "a" {
		handleText(node, true, true, svg.defs)
		return
	}

	// recurse
	for i := range node.children {
		svg.postProcessNode(&node.children[i])
	}
}
