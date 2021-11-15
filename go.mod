module github.com/benoitkugler/go-weasyprint

go 1.16

require (
	github.com/benoitkugler/cascadia v1.2.0
	github.com/benoitkugler/oksvg v0.0.0-20211103171435-b8520e63fd29
	github.com/benoitkugler/pdf v0.0.0-20211113115023-b48773b758da
	github.com/benoitkugler/textlayout v0.0.3
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	golang.org/x/text v0.3.7
)

replace github.com/benoitkugler/oksvg => ../oksvg

replace github.com/benoitkugler/pdf => ../pdf
