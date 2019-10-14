package boxes

import (
	"fmt"
	"testing"
)

func TestInheritance(t *testing.T) {
	// u := NewInlineBox("", nil, nil)
	// u.removeDecoration(nil, true, true)

}

func TestReplaced(t *testing.T) {
	var i InstanceReplacedBox
	i = new(ReplacedBox)
	fmt.Println(i)
	i = new(BlockReplacedBox)
	fmt.Println(i)
	i = new(InlineReplacedBox)
	fmt.Println(i)
}
