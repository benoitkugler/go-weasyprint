// Package svg implements parsing of SVG images.
// It transforms SVG text files into an in-memory structure
// that is easy to draw.
// CSS is supported via the style and cascadia packages.
package svg

import (
	"fmt"
	"io"
)

// convert from an svg tree to the final form

type SVGImage = *svgNode

// Parse parsed the given SVG data. Warnings are
// logged for unsupported elements.
// An error is returned for invalid documents.
// `baseURL` is used as base path for url resources.
func Parse(svg io.Reader, baseURL string) (SVGImage, error) {
	out, err := buildSVGTree(svg, baseURL)
	if err != nil {
		return nil, err
	}

	return out.postProcess()
}

type svgNode struct {
	content  drawable
	children []*svgNode
	attributes
}

type drawable interface{}

// attributes stores the SVG attributes
// shared by all node types.
type attributes struct {
	transform, clipPath, mask, filter         string
	marker, markerStart, markerMid, markerEnd string
	fontSize                                  value

	x, y, width, height value

	opacity Fl // default to 1

	display, visible bool
}

// Build the drawable items by parsing attributes
func (tree *svgTree) postProcess() (SVGImage, error) {
	return processNode(tree.root)
}

func processNode(node *cascadedNode) (*svgNode, error) {
	var out svgNode
	err := node.attrs.parseCommonAttributes(&out.attributes)
	if err != nil {
		return nil, err
	}

	builder := elementBuilders[node.tag]
	if builder == nil {
		fmt.Println(node.tag)
		// return nil, fmt.Errorf("unsupported element %s", node.tag)
	} else {
		out.content, err = builder(node)
		if err != nil {
			return nil, fmt.Errorf("invalid element %s: %s", node.tag, err)
		}
	}

	out.children = make([]*svgNode, len(node.children))
	for i, c := range node.children {
		out.children[i], err = processNode(c)
		if err != nil {
			return nil, err
		}
	}

	return &out, nil
}

func (na nodeAttributes) parseCommonAttributes(out *attributes) error {
	var err error
	out.fontSize, err = na.fontSize()
	if err != nil {
		return err
	}
	out.opacity, err = na.opacity()
	if err != nil {
		return err
	}
	out.transform = na["transform"]
	out.filter = parseURLFragment(na["filter"])
	out.clipPath = parseURLFragment(na["clip-path"])
	out.mask = parseURLFragment(na["mask"])

	out.marker = parseURLFragment(na["marker"])
	out.markerStart = parseURLFragment(na["marker-start"])
	out.markerMid = parseURLFragment(na["marker-mid"])
	out.markerEnd = parseURLFragment(na["marker-end"])

	out.x, err = parseValue(na["x"])
	if err != nil {
		return err
	}
	out.y, err = parseValue(na["y"])
	if err != nil {
		return err
	}
	out.width, err = parseValue(na["width"])
	if err != nil {
		return err
	}
	out.height, err = parseValue(na["height"])
	if err != nil {
		return err
	}

	out.display = na.display()
	out.visible = na.visible()
	return nil
}
