package goweasyprint

import (
	"os"
	"testing"
	"text/template"
)

func TestSVG(t *testing.T) {
	tmp := headerSVG + crop + cross
	tp := template.Must(template.New("svg").Parse(tmp))
	tp.Execute(os.Stdout, svgArgs{})
}
