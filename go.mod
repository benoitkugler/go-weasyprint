module github.com/benoitkugler/go-weasyprint

go 1.17

require (
	github.com/benoitkugler/pdf v0.0.1
	github.com/benoitkugler/textlayout v0.3.0
	github.com/benoitkugler/textprocessing v0.0.3
	github.com/benoitkugler/webrender v0.0.5
)

require (
	github.com/benoitkugler/pstokenizer v1.0.1 // indirect
	github.com/hhrutter/lzw v1.0.0 // indirect
	golang.org/x/image v0.7.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/text v0.9.0 // indirect
)

replace github.com/benoitkugler/webrender => ../webrender
