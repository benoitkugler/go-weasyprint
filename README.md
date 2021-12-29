# go-weasyprint

Golang port of [Weasyprint](https://github.com/Kozea/WeasyPrint) python Html to Pdf library.

This is an **ongoing work**, not production ready just yet.

## Outline

This package converts an HTML document (with its associated CSS files) to a PDF file.
The heavy lifting is actually delegated to [webrender](github.com/benoitkugler/webrender), but this package implements a backend for PDF files, relying on [benoitkugler/pdf](github.com/benoitkugler/pdf).
