package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
)

// Leader management.

// get the index of the first leader box in ``box``.
func leaderIndex(box Box) (*tree.IntList, Box) {
	for i, child := range box.Box().Children {
		if child.Box().IsLeader {
			return &tree.IntList{Value: i}, child
		}
		if bo.ParentBoxT.IsInstance(child) {
			childLeaderIndex, childLeader := leaderIndex(child)
			if childLeaderIndex != nil {
				return &tree.IntList{Value: i, Next: childLeaderIndex}, childLeader
			}
		}
	}
	return nil, nil
}

// Find a leader box in ``line`` and handle its text and its position.
func handleLeader(context *layoutContext, line *bo.LineBox, containingBlock containingBlock) {
	index, leaderBox_ := leaderIndex(line)
	var extraWidth pr.Float
	if index != nil && len(leaderBox_.Box().Children) != 0 {
		leaderBox := leaderBox_.Box()
		textBox_ := leaderBox.Children[0]
		textBox := textBox_.Box()

		// Abort if the leader text has no width
		if textBox.Width.V() <= 0 {
			return
		}

		// Extra width is the additional width taken by the leader box
		var sum pr.Float
		for _, child := range line.Children {
			if child.Box().IsInNormalFlow() {
				sum += child.Box().Width.V()
			}
		}
		cbWidth, _ := containingBlock.ContainingBlock()
		extraWidth = cbWidth.V() - sum

		// Take care of excluded shapes
		for _, shape := range *context.excludedShapes {
			if shape.PositionY+shape.Height.V() > line.PositionY {
				extraWidth -= shape.Width.V()
			}
		}

		// Available width is the width available for the leader box
		availableWidth := extraWidth + textBox.Width.V()
		line.Width = cbWidth

		// Add text boxes into the leader box
		numberOfLeaders := int(line.Width.V()) / int(textBox.Width.V())
		positionX := line.PositionX + line.Width.V()
		var children []Box
		for i := 0; i < numberOfLeaders; i++ {
			positionX -= textBox.Width.V()
			if positionX < leaderBox.PositionX {
				// Don’t add leaders behind the text on the left
				continue
			} else if positionX+textBox.Width.V() >
				leaderBox.PositionX+availableWidth {
				// Don’t add leaders behind the text on the right
				continue
			}
			textBox_ = textBox_.Copy()
			textBox = textBox_.Box()
			textBox.PositionX = positionX
			children = append(children, textBox_)
		}
		leaderBox.Children = children

		if line.Style.GetDirection() == "rtl" {
			leaderBox_.Translate(leaderBox_, extraWidth, 0, false)
		}
	}

	// Widen leader parent boxes and translate following boxes
	var box Box = line
	for index != nil {
		for _, child := range box.Box().Children[index.Value+1:] {
			if child.Box().IsInNormalFlow() {
				if line.Style.GetDirection() == "ltr" {
					child.Translate(child, extraWidth, 0, false)
				} else {
					child.Translate(child, -extraWidth, 0, false)
				}
			}
		}
		box = box.Box().Children[index.Value]
		box.Box().Width = box.Box().Width.V() + extraWidth
		index = index.Next
	}
}
