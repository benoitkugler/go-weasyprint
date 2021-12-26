module github.com/benoitkugler/go-weasyprint

go 1.16

require (
	github.com/benoitkugler/cascadia v1.2.1
	github.com/benoitkugler/oksvg v0.0.0-20211213162813-d17e098b5309
	github.com/benoitkugler/pdf v0.0.0-20211224133419-7f3afc58affd
	github.com/benoitkugler/textlayout v0.0.5
	golang.org/x/net v0.0.0-20211216030914-fe4d6282115f
	golang.org/x/text v0.3.7
)

replace github.com/benoitkugler/oksvg => ../oksvg

replace github.com/benoitkugler/pdf => ../pdf

replace github.com/benoitkugler/textlayout => ../textlayout
