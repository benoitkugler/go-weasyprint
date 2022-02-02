module github.com/benoitkugler/go-weasyprint

go 1.16

require (
	github.com/benoitkugler/pdf v0.0.0-20211224133419-7f3afc58affd
	github.com/benoitkugler/textlayout v0.0.9
	github.com/benoitkugler/webrender v0.0.1
)

replace github.com/benoitkugler/pdf => ../pdf

replace github.com/benoitkugler/textlayout => ../textlayout

replace github.com/benoitkugler/webrender => ../webrender
