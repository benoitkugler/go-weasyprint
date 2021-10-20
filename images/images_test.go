package images

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/go-weasyprint/utils"
)

func TestLoadLocalImages(t *testing.T) {
	paths := []string{
		"../resources_test/blue.jpg",
		"../resources_test/icon.png",
		"../resources_test/pattern.gif",
		"../resources_test/pattern.svg",
	}
	for _, path := range paths {
		url, err := utils.Path2url(path)
		if err != nil {
			t.Fatal(err)
		}
		out, err := GetImageFromUri(NewCache(), utils.DefaultUrlFetcher, false, url, "")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%T\n", out)
	}
}
