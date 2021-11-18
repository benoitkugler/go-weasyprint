module github.com/benoitkugler/go-weasyprint

go 1.16

require (
	github.com/benoitkugler/cascadia v1.2.1
	github.com/benoitkugler/oksvg v0.0.0-20211103171435-b8520e63fd29
	github.com/benoitkugler/pdf v0.0.0-20211117135622-641dd4463e8a
	github.com/benoitkugler/textlayout v0.0.3
	golang.org/x/net v0.0.0-20211118161319-6a13c67c3ce4
	golang.org/x/text v0.3.7
)

replace github.com/benoitkugler/oksvg => ../oksvg

replace github.com/benoitkugler/pdf => ../pdf

replace github.com/benoitkugler/textlayout => ../textlayout
