module github.com/benoitkugler/go-weasyprint

go 1.16

require (
	github.com/benoitkugler/cascadia v1.2.0
	github.com/benoitkugler/oksvg v0.0.0-20201219152846-654d2ff256cf
	github.com/benoitkugler/pdf v0.0.0-20211103165629-04ad9d83f560
	github.com/benoitkugler/textlayout v0.0.3
	golang.org/x/net v0.0.0-20211101193420-4a448f8816b3
	golang.org/x/text v0.3.7
)

replace github.com/benoitkugler/oksvg => ../oksvg

replace github.com/benoitkugler/pdf => ../pdf
