package boxes

import (
	"fmt"
	"log"
	"testing"

	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
)

func TestInheritance(t *testing.T) {
	// u := NewInlineBox("", nil, nil)
	// u.RemoveDecoration(nil, true, true)
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

func TestBlockLevel(t *testing.T) {
	var i InstanceBlockLevelBox
	i = new(BlockBox)
	fmt.Println(i)
	i = new(BlockReplacedBox)
	fmt.Println(i)
	i = new(TableBox)
	fmt.Println(i)
	i = new(FlexBox)
	fmt.Println(i)
}

func TestTable(t *testing.T) {
	var i InstanceTableBox
	i = new(TableBox)
	fmt.Println(i)
	i = new(InlineTableBox)
	fmt.Println(i)
}

//  Test that the "before layout" box tree is correctly constructed.

func fakeHTML(html *tree.HTML) *tree.HTML {
	html.UAStyleSheet = tree.TestUAStylesheet
	return html
}

func parseBase(content utils.ContentInput, baseUrl string) {
	html, err := tree.NewHTML(content, baseUrl, utils.DefaultUrlFetcher, "")
	if err != nil {
		log.Fatal(err)
	}
	tree.GetAllComputedStyles()
}

// func _parse_base(html_content, base_url=BASE_URL):
//     document = FakeHTML(string=html_content, base_url=base_url)
//     counter_style = CounterStyle()
//     style_for = get_all_computed_styles(document, counter_style=counter_style)
//     get_image_from_uri = functools.partial(
//         images.get_image_from_uri, cache={}, url_fetcher=document.url_fetcher,
//         optimize_size=())
//     target_collector = TargetCollector()
//     return (
//         document.etree_element, style_for, get_image_from_uri, base_url,
//         target_collector, counter_style)

// func getGrid(html string)
