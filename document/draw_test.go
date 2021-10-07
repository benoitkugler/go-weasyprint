package document

import (
	"io/ioutil"
	"testing"
	"text/template"
)

func TestSVG(t *testing.T) {
	tmp := headerSVG + crop + cross
	tp := template.Must(template.New("svg").Parse(tmp))
	if err := tp.Execute(ioutil.Discard, svgArgs{}); err != nil {
		t.Fatal(err)
	}
}
